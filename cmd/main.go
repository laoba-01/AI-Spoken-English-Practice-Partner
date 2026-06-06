// cmd/main.go
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"

	"english-tutor/internel/momel"
	"english-tutor/internel/momel/package/pkg/asr"
	"english-tutor/internel/momel/package/pkg/tts"
)

// 全局变量
var (
	llmClient *openai.Client
	asrClient *asr.AliyunASR
	ttsClient *tts.AliyunTTS
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
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("⚠ 未找到 .env 文件，使用系统环境变量")
	}

	// 初始化LLM客户端
	config := openai.DefaultConfig("ark-a0167170-badd-4b03-8e0c-1da1f1781927-3d5f8")
	config.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	llmClient = openai.NewClientWithConfig(config)

	// 初始化ASR客户端
	asrClient = asr.NewAliyunASR()
	logASRStatus()

	// 初始化TTS客户端
	ttsClient = tts.NewAliyunTTS()
	logTTSStatus()

	// 初始化数据库（非致命）
	if err := model.InitDB(); err != nil {
		log.Println("⚠ 数据库连接失败，对话记录不会保存:", err)
	}

	// 创建音频文件存储目录
	os.MkdirAll("./audio", 0755)

	// 创建Gin路由
	r := gin.Default()

	// 静态文件服务：提供TTS生成的音频文件
	r.Static("/audio", "./audio")

	// WebSocket接口
	r.GET("/ws/chat", WebSocketHandler)
	r.GET("/ws/chat/:id", WebSocketHandler)

	// 启动服务
	log.Println("服务器启动在 :8080")
	log.Fatal(r.Run(":8080"))
}

func logASRStatus() {
	if os.Getenv("ALIYUN_ACCESS_KEY_ID") == "" {
		log.Println("⚠ 阿里云ASR未配置（缺少 ALIYUN_ACCESS_KEY_ID），语音识别不可用")
	} else {
		log.Println("✓ 阿里云ASR已配置")
	}
}

func logTTSStatus() {
	if os.Getenv("ALIYUN_ACCESS_KEY_ID") == "" {
		log.Println("⚠ 阿里云TTS未配置（缺少 ALIYUN_ACCESS_KEY_ID），语音合成不可用")
	} else {
		log.Println("✓ 阿里云TTS已配置")
	}
}

// WebSocketHandler WebSocket处理函数
func WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket升级失败:", err)
		return
	}
	defer conn.Close()

	convID := c.Param("id")
	var cid int64
	if convID != "" {
		fmt.Sscanf(convID, "%d", &cid)
	}

	log.Printf("新客户端连接 (conversation=%s)", convID)

	for {
		// 读取客户端消息
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("读取消息失败:", err)
			break
		}

		var userText string

		if msgType == websocket.BinaryMessage {
			// === 语音消息：ASR → 发送识别结果 ===
			log.Printf("收到语音消息，大小: %d bytes", len(msg))

			recognizedText, err := asrClient.RecognizeWebM(msg)
			if err != nil {
				log.Println("ASR识别失败:", err)
				conn.WriteJSON(map[string]interface{}{
					"type":    "error",
					"message": "Sorry, I didn't catch that. Could you say it again?",
				})
				continue
			}

			userText = recognizedText
			log.Println("ASR识别结果:", userText)

			// 推送ASR识别结果给前端
			conn.WriteJSON(map[string]interface{}{
				"type": "asr_result",
				"text": userText,
			})
		} else {
			// === 文字消息 ===
			userText = strings.TrimSpace(string(msg))
			if userText == "" {
				continue
			}
			log.Println("收到用户消息:", userText)
		}

		// 保存用户消息
		if cid > 0 {
			model.SaveMessage(cid, "user", userText, "")
		}

		// === LLM 流式生成回复 ===
		fullText, err := streamLLM(userText, conn)
		if err != nil {
			log.Println("LLM调用失败:", err)
			conn.WriteJSON(map[string]interface{}{
				"type":    "error",
				"message": "System is busy, please try again later.",
			})
			continue
		}

		log.Println("AI回复:", fullText)

		// === TTS 语音合成 ===
		audioData, err := ttsClient.TextToMP3(fullText)
		if err != nil {
			log.Println("TTS合成失败（AI文字回复已发送，跳过语音）:", err)
			// 保存文字消息（无语音）
			if cid > 0 {
				model.SaveMessage(cid, "assistant", fullText, "")
			}
			continue
		}

		// 保存音频文件到本地
		audioFilename := fmt.Sprintf("tts_%d.mp3", time.Now().UnixNano())
		audioPath := fmt.Sprintf("./audio/%s", audioFilename)
		if err := os.WriteFile(audioPath, audioData, 0644); err != nil {
			log.Println("保存音频文件失败:", err)
			if cid > 0 {
				model.SaveMessage(cid, "assistant", fullText, "")
			}
			continue
		}

		// 推送TTS语音URL给前端
		audioURL := fmt.Sprintf("/audio/%s", audioFilename)
		conn.WriteJSON(map[string]interface{}{
			"type":      "audio_result",
			"audio_url": audioURL,
		})

		// 保存AI消息（含语音URL）
		if cid > 0 {
			model.SaveMessage(cid, "assistant", fullText, audioURL)
		}
	}

	log.Println("客户端断开连接")
}

// streamLLM 调用豆包LLM流式生成，实时推送每个chunk到前端
func streamLLM(userText string, conn *websocket.Conn) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: "doubao-seed-2-0-pro-260215",
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
		return "", err
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

	return fullText.String(), nil
}
