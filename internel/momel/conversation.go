// internal/model/conversation.go
package model

// 创建新会话
func CreateConversation(userID int64, scene string) (int64, error) {
	res, err := DB.Exec(
		"INSERT INTO user_conversations (user_id, scene) VALUES (?, ?)",
		userID, scene,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// 获取会话历史消息
func GetMessagesByConversationID(conversationID int64) ([]Message, error) {
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

// 保存消息
func SaveMessage(conversationID int64, role, content, audioURL string) error {
	_, err := DB.Exec(
		"INSERT INTO user_messages (conversation_id, role, content, audio_url) VALUES (?, ?, ?, ?)",
		conversationID, role, content, audioURL,
	)
	return err
}