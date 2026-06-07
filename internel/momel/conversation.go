// internal/model/conversation.go
package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// 创建新会话
func CreateConversation(userID int64, scene string) (int64, error) {
	var id int64
	var dbErr error

	if DB != nil {
		res, err := DB.Exec(
			"INSERT INTO user_conversations (user_id, scene) VALUES (?, ?)",
			userID, scene,
		)
		if err != nil {
			return 0, err
		}
		id, dbErr = res.LastInsertId()
		if dbErr != nil {
			return 0, dbErr
		}
	} else {
		// 内存 fallback
		memMu.Lock()
		id = memConvIDSeq
		memConvIDSeq++
		memConversations = append(memConversations, Conversation{
			ID: id, UserID: userID, Scene: scene, Title: "", CreatedAt: memNow(),
		})
		memMu.Unlock()
	}

	// 删除缓存（使下次列表查询重建）
	if RedisClient != nil {
		ctx := context.Background()
		key := fmt.Sprintf("conv:list:%d:%s", userID, scene)
		RedisClient.Del(ctx, key)
		// 也删除"全部场景"的缓存 key
		RedisClient.Del(ctx, fmt.Sprintf("conv:list:%d:", userID))
	}
	return id, nil
}

// 获取会话历史消息
func GetMessagesByConversationID(conversationID int64) ([]Message, error) {
	ctx := context.Background()

	// ① Redis 缓存查询
	if RedisClient != nil {
		key := fmt.Sprintf("conv:msgs:%d", conversationID)
		cached, err := RedisClient.Get(ctx, key).Result()
		if err == nil {
			var msgs []Message
			if json.Unmarshal([]byte(cached), &msgs) == nil {
				return msgs, nil
			}
		}
	}

	// ② 查 MySQL / 内存（原有逻辑）
	if DB != nil {
		rows, err := DB.Query(
			"SELECT id, conversation_id, role, content, audio_url, created_at FROM user_messages WHERE conversation_id = ? ORDER BY id ASC",
			conversationID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var messages []Message
		for rows.Next() {
			var msg Message
			err := rows.Scan(
				&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.AudioURL, &msg.CreatedAt,
			)
			if err != nil {
				return nil, err
			}
			messages = append(messages, msg)
		}

		// ③ 回填 Redis
		if RedisClient != nil && messages != nil {
			key := fmt.Sprintf("conv:msgs:%d", conversationID)
			if data, err := json.Marshal(messages); err == nil {
				RedisClient.Set(ctx, key, data, 2*time.Minute)
			}
		}
		return messages, nil
	}

	// 内存 fallback
	memMu.Lock()
	defer memMu.Unlock()
	var result []Message
	for _, m := range memMessages {
		if m.ConversationID == conversationID {
			result = append(result, m)
		}
	}

	// ③ 回填 Redis
	if RedisClient != nil && result != nil {
		key := fmt.Sprintf("conv:msgs:%d", conversationID)
		if data, err := json.Marshal(result); err == nil {
			RedisClient.Set(ctx, key, data, 2*time.Minute)
		}
	}
	return result, nil
}

// 获取用户会话列表
func GetConversationsByUserID(userID int64, scene string) ([]Conversation, error) {
	ctx := context.Background()

	// ① Redis 缓存查询
	if RedisClient != nil {
		key := fmt.Sprintf("conv:list:%d:%s", userID, scene)
		cached, err := RedisClient.Get(ctx, key).Result()
		if err == nil {
			var convs []Conversation
			if json.Unmarshal([]byte(cached), &convs) == nil {
				return convs, nil
			}
		}
	}

	// ② 查 MySQL / 内存（原有逻辑）
	if DB != nil {
		var rows *sql.Rows
		var err error
		if scene != "" {
			rows, err = DB.Query(
				"SELECT id, user_id, scene, COALESCE(title,''), created_at FROM user_conversations WHERE user_id = ? AND scene = ? ORDER BY id DESC",
				userID, scene,
			)
		} else {
			rows, err = DB.Query(
				"SELECT id, user_id, scene, COALESCE(title,''), created_at FROM user_conversations WHERE user_id = ? ORDER BY id DESC",
				userID,
			)
		}
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var convs []Conversation
		for rows.Next() {
			var c Conversation
			if err := rows.Scan(&c.ID, &c.UserID, &c.Scene, &c.Title, &c.CreatedAt); err != nil {
				return nil, err
			}
			convs = append(convs, c)
		}

		// ③ 回填 Redis
		if RedisClient != nil && convs != nil {
			key := fmt.Sprintf("conv:list:%d:%s", userID, scene)
			if data, err := json.Marshal(convs); err == nil {
				RedisClient.Set(ctx, key, data, 5*time.Minute)
			}
		}
		return convs, nil
	}

	// 内存 fallback
	memMu.Lock()
	defer memMu.Unlock()
	var result []Conversation
	for _, c := range memConversations {
		if c.UserID == userID && (scene == "" || c.Scene == scene) {
			result = append(result, c)
		}
	}

	// ③ 回填 Redis（内存结果也缓存）
	if RedisClient != nil && result != nil {
		key := fmt.Sprintf("conv:list:%d:%s", userID, scene)
		if data, err := json.Marshal(result); err == nil {
			RedisClient.Set(ctx, key, data, 5*time.Minute)
		}
	}
	return result, nil
}

// 删除会话及其所有消息
func DeleteConversation(conversationID int64) error {
	// ① 删除前查出会话信息（用于缓存失效）
	var targetUserID int64
	var targetScene string

	if DB != nil {
		row := DB.QueryRow(
			"SELECT user_id, scene FROM user_conversations WHERE id = ?",
			conversationID,
		)
		row.Scan(&targetUserID, &targetScene)

		_, err := DB.Exec("DELETE FROM user_messages WHERE conversation_id = ?", conversationID)
		if err != nil {
			return err
		}
		_, err = DB.Exec("DELETE FROM user_conversations WHERE id = ?", conversationID)
		if err != nil {
			return err
		}
	} else {
		// 内存 fallback
		memMu.Lock()
		// 先查会话信息
		for _, c := range memConversations {
			if c.ID == conversationID {
				targetUserID = c.UserID
				targetScene = c.Scene
				break
			}
		}
		// 删除消息
		var filteredMsgs []Message
		for _, m := range memMessages {
			if m.ConversationID != conversationID {
				filteredMsgs = append(filteredMsgs, m)
			}
		}
		memMessages = filteredMsgs
		// 删除会话
		var filteredConvs []Conversation
		for _, c := range memConversations {
			if c.ID != conversationID {
				filteredConvs = append(filteredConvs, c)
			}
		}
		memConversations = filteredConvs
		memMu.Unlock()
	}

	// ② 删除缓存
	if RedisClient != nil {
		ctx := context.Background()
		// 删除消息缓存
		RedisClient.Del(ctx, fmt.Sprintf("conv:msgs:%d", conversationID))
		// 删除会话列表缓存（指定场景 + 全部场景）
		if targetUserID > 0 {
			RedisClient.Del(ctx, fmt.Sprintf("conv:list:%d:%s", targetUserID, targetScene))
			RedisClient.Del(ctx, fmt.Sprintf("conv:list:%d:", targetUserID))
		}
	}
	return nil
}

// 保存消息
func SaveMessage(conversationID int64, role, content, audioURL string) error {
	if DB != nil {
		_, err := DB.Exec(
			"INSERT INTO user_messages (conversation_id, role, content, audio_url) VALUES (?, ?, ?, ?)",
			conversationID, role, content, audioURL,
		)
		if err != nil {
			return err
		}
	} else {
		// 内存 fallback
		memMu.Lock()
		id := memMsgIDSeq
		memMsgIDSeq++
		memMessages = append(memMessages, Message{
			ID:             id,
			ConversationID: conversationID,
			Role:           role,
			Content:        content,
			AudioURL:       audioURL,
			CreatedAt:      memNow(),
		})
		memMu.Unlock()
	}

	// 删除消息缓存（下次查询时重建）
	if RedisClient != nil {
		ctx := context.Background()
		key := fmt.Sprintf("conv:msgs:%d", conversationID)
		RedisClient.Del(ctx, key)
	}
	return nil
}

// 获取会话场景（用于选择 System Prompt）
func GetConversationScene(conversationID int64) (string, error) {
	if DB != nil {
		var scene string
		err := DB.QueryRow(
			"SELECT scene FROM user_conversations WHERE id = ?",
			conversationID,
		).Scan(&scene)
		return scene, err
	}

	memMu.Lock()
	defer memMu.Unlock()
	for _, c := range memConversations {
		if c.ID == conversationID {
			return c.Scene, nil
		}
	}
	return "daily", nil
}

// UpdateLastUserMessage 更新会话中最后一条用户消息的纠错和评分
func UpdateLastUserMessage(conversationID int64, correction string, score int) error {
	if DB != nil {
		_, err := DB.Exec(
			"UPDATE user_messages SET correction = ?, pronunciation_score = ? WHERE id = (SELECT id FROM (SELECT id FROM user_messages WHERE conversation_id = ? AND role = 'user' ORDER BY id DESC LIMIT 1) AS t)",
			correction, score, conversationID,
		)
		if err != nil {
			return err
		}
	} else {
		// 内存 fallback
		memMu.Lock()
		defer memMu.Unlock()
		for i := len(memMessages) - 1; i >= 0; i-- {
			if memMessages[i].ConversationID == conversationID && memMessages[i].Role == "user" {
				memMessages[i].Correction = correction
				memMessages[i].PronunciationScore = int8(score)
				break
			}
		}
	}

	// 删除缓存
	if RedisClient != nil {
		ctx := context.Background()
		key := fmt.Sprintf("conv:msgs:%d", conversationID)
		RedisClient.Del(ctx, key)
	}
	return nil
}
