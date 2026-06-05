// internal/model/db.go
package model

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() error {
	dsn := os.Getenv("MYSQL_DSN")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(20)
	DB = db
	log.Println("数据库连接成功")
	return nil
}

// Conversation 会话模型
type Conversation struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Scene     string `json:"scene"` // daily/business/exam
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

// Message 消息模型
type Message struct {
	ID              int64  `json:"id"`
	ConversationID  int64  `json:"conversation_id"`
	Role            string `json:"role"` // user/assistant
	Content         string `json:"content"`
	AudioURL        string `json:"audio_url"`
	Correction      string `json:"correction"`
	PronunciationScore int8 `json:"pronunciation_score"`
	CreatedAt       string `json:"created_at"`
}