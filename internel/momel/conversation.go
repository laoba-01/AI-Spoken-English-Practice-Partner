// internal/model/conversation.go
package model

import "database/sql"

// 创建新会话
func CreateConversation(userID int64, scene string) (int64, error) {
	if DB != nil {
		res, err := DB.Exec(
			"INSERT INTO user_conversations (user_id, scene) VALUES (?, ?)",
			userID, scene,
		)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}

	// 内存 fallback
	memMu.Lock()
	defer memMu.Unlock()
	id := memConvIDSeq
	memConvIDSeq++
	memConversations = append(memConversations, Conversation{
		ID: id, UserID: userID, Scene: scene, Title: "", CreatedAt: memNow(),
	})
	return id, nil
}

// 获取会话历史消息
func GetMessagesByConversationID(conversationID int64) ([]Message, error) {
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
	return result, nil
}

// 获取用户会话列表
func GetConversationsByUserID(userID int64, scene string) ([]Conversation, error) {
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
	return result, nil
}

// 删除会话及其所有消息
func DeleteConversation(conversationID int64) error {
	if DB != nil {
		_, err := DB.Exec("DELETE FROM user_messages WHERE conversation_id = ?", conversationID)
		if err != nil {
			return err
		}
		_, err = DB.Exec("DELETE FROM user_conversations WHERE id = ?", conversationID)
		return err
	}

	// 内存 fallback
	memMu.Lock()
	defer memMu.Unlock()

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
	return nil
}

// 保存消息
func SaveMessage(conversationID int64, role, content, audioURL string) error {
	if DB != nil {
		_, err := DB.Exec(
			"INSERT INTO user_messages (conversation_id, role, content, audio_url) VALUES (?, ?, ?, ?)",
			conversationID, role, content, audioURL,
		)
		return err
	}

	// 内存 fallback
	memMu.Lock()
	defer memMu.Unlock()
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
	return nil
}
