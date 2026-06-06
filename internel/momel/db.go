// internal/model/db.go
package model

import (
	"database/sql"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// 内存存储（当数据库不可用时的 fallback）
var (
	memConversations []Conversation
	memMessages      []Message
	memMu            sync.Mutex
	memConvIDSeq     int64 = 1
	memMsgIDSeq      int64 = 1
)

func InitDB() error {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Println("⚠ 未配置 MYSQL_DSN，使用内存存储（重启后数据丢失）")
		return nil
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println("⚠ MySQL连接失败，使用内存存储:", err)
		return nil
	}
	if err := db.Ping(); err != nil {
		log.Println("⚠ MySQL Ping失败，使用内存存储:", err)
		return nil
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
	ID                int64  `json:"id"`
	ConversationID    int64  `json:"conversation_id"`
	Role              string `json:"role"` // user/assistant
	Content           string `json:"content"`
	AudioURL          string `json:"audio_url"`
	Correction        string `json:"correction"`
	PronunciationScore int8  `json:"pronunciation_score"`
	CreatedAt         string `json:"created_at"`
}

// ─── 内存存储辅助 ───

func memNow() string {
	return time.Now().Format("2006-01-02T15:04:05Z")
}
