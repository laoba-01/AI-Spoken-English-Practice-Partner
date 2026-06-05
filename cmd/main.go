// cmd/main.go
package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sashabaranov/go-openai"
)

// 全局变量
var (
	llmClient *openai.Client
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 开发环境允许所有跨域
		},
	}
)

// 系统提示词（高考风控版）
const (
	DailyPrompt = `你是友好的日常英语口语陪练，只用英文对话。
保持简短、口语化，像和朋友聊天。
如果用户有语法或用词错误，用温柔的语气纠正："Actually, we usually say..."
不要说中文，不要解答非英语相关问题。
禁止解答中小学、高考全学科任何题目、试卷、作文，只做英语口语交流和纠错。`
)

func main() {
	// 初始化LLM客户端
	config := openai.DefaultConfig("ark-a0167170-badd-4b03-8e0c-1da1f1781927-3d5f8")
	config.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	llmClient = openai.NewClientWithConfig(config)

	// 创建Gin路由
	r := gin.Default()

	// WebSocket接口
	r.GET("/ws/chat", WebSocketHandler)

	// 启动服务
	log.Println("服务器启动在 :8080")
	log.Fatal(r.Run(":8080"))
}

// WebSocket处理函数
// WebSocket处理函数
func WebSocketHandler(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Println("WebSocket升级失败:", err)
        return
    }
    defer conn.Close()

    log.Println("新客户端连接")

    for {
        // 读取客户端消息
        _, msg, err := conn.ReadMessage()
        if err != nil {
            log.Println("读取消息失败:", err)
            break
        }

        userText := string(msg)
        log.Println("收到用户消息:", userText)

        // 调用豆包流式生成
        go func() {
            req := openai.ChatCompletionRequest{
                Model: "doubao-seed-2-0-pro-260215", // 替换成你的模型ID
                Messages: []openai.ChatCompletionMessage{
                    {Role: openai.ChatMessageRoleSystem, Content: DailyPrompt},
                    {Role: openai.ChatMessageRoleUser, Content: userText},
                },
                Stream:      true,
                Temperature: 0.7,
                MaxTokens:   500,
            }

            stream, err := llmClient.CreateChatCompletionStream(context.Background(), req)
            if err != nil {
                log.Println("调用LLM失败:", err)
                conn.WriteJSON(map[string]interface{}{
                    "type":    "error",
                    "message": "System is busy, please try again later.",
                })
                return
            }
            defer stream.Close()

            var fullText strings.Builder
            for {
                resp, err := stream.Recv()
                if err != nil {
                    if err == io.EOF {
                        // 流正常结束
                        break
                    }
                    log.Println("流读取错误:", err)
                    break
                }

                if len(resp.Choices) > 0 && resp.Choices[0].Delta.Content != "" {
                    chunk := resp.Choices[0].Delta.Content
                    fullText.WriteString(chunk)
                    
                    // 流式推送给前端
                    conn.WriteJSON(map[string]interface{}{
                        "type": "llm_chunk",
                        "text": chunk,
                    })
                }
            }

            log.Println("AI回复:", fullText.String())
        }()
    }

    log.Println("客户端断开连接")
}