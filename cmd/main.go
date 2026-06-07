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

	BusinessPrompt = `你是专业的商务英语口语教练，只用英文对话。
模拟职场场景：会议、邮件、谈判、客户沟通。
纠正商务用语错误，推荐更专业的表达。
保持正式但友好的语气，不要说中文。
禁止解答中小学、高考全学科任何题目、试卷、作文，只做英语口语交流和纠错。`

	ExamPrompt = `你是雅思口语考官，严格按照雅思标准进行模拟考试。
分三个部分：自我介绍、话题陈述、深度讨论。
严格纠正语法、发音、流利度问题，给出评分和改进建议。
只用英文对话，不要说中文。
禁止解答中小学、高考全学科任何题目、试卷、作文，只做英语口语交流和纠错。`
)

func GetSystemPrompt(scene string) string {
	switch scene {
	case "business":
		return BusinessPrompt
	case "exam":
		return ExamPrompt
	default:
		return DailyPrompt
	}
}

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("⚠ 未找到 .env 文件，使用系统环境变量")
	}

	// 初始化LLM客户端
	llmAPIKey := os.Getenv("DOUBAO_API_KEY")
	if llmAPIKey == "" {
		log.Fatal("❌ 缺少 DOUBAO_API_KEY 环境变量，请在 .env 中配置")
	}
	config := openai.DefaultConfig(llmAPIKey)
	config.BaseURL = os.Getenv("DOUBAO_BASE_URL")
	if config.BaseURL == "" {
		config.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
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

	// 初始化Redis缓存（非致命）
	if err := model.InitRedis(); err != nil {
		log.Println("⚠ Redis缓存连接失败:", err)
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

	// REST API接口
	api := r.Group("/api")
	{
		api.POST("/conversations", CreateConversationHandler)
		api.GET("/conversations", ListConversationsHandler)
		api.GET("/conversations/:id", GetConversationHandler)
		api.POST("/message/text", TextMessageHandler)
		api.DELETE("/conversations/:id", DeleteConversationHandler)
	}

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

		// 确定场景提示词
		systemPrompt := DailyPrompt
		if cid > 0 {
			if scene, err := model.GetConversationScene(cid); err == nil {
				systemPrompt = GetSystemPrompt(scene)
			}
		}

		// === LLM 流式生成回复 ===
		fullText, err := streamLLM(systemPrompt, userText, conn)
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
func streamLLM(systemPrompt, userText string, conn *websocket.Conn) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: "doubao-seed-2-0-pro-260215",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
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

// ==================== REST API Handlers ====================

// CreateConversationHandler POST /api/conversations
func CreateConversationHandler(c *gin.Context) {
	var req struct {
		UserID int64  `json:"user_id"`
		Scene  string `json:"scene"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": -1, "message": "invalid request"})
		return
	}

	id, err := model.CreateConversation(req.UserID, req.Scene)
	if err != nil {
		log.Println("创建会话失败:", err)
		c.JSON(500, gin.H{"code": -1, "message": "创建失败"})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"conversation_id": id,
			"created_at":      time.Now().Format("2006-01-02T15:04:05Z"),
		},
	})
}

// ListConversationsHandler GET /api/conversations?user_id=&scene=
func ListConversationsHandler(c *gin.Context) {
	userIDStr := c.Query("user_id")
	scene := c.DefaultQuery("scene", "")

	var userID int64
	fmt.Sscanf(userIDStr, "%d", &userID)

	convs, err := model.GetConversationsByUserID(userID, scene)
	if err != nil {
		log.Println("获取会话列表失败:", err)
		c.JSON(500, gin.H{"code": -1, "message": "获取失败"})
		return
	}
	if convs == nil {
		convs = []model.Conversation{}
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    convs,
	})
}

// GetConversationHandler GET /api/conversations/:id
func GetConversationHandler(c *gin.Context) {
	idStr := c.Param("id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	messages, err := model.GetMessagesByConversationID(id)
	if err != nil {
		log.Println("获取会话历史失败:", err)
		c.JSON(500, gin.H{"code": -1, "message": "获取失败"})
		return
	}
	if messages == nil {
		messages = []model.Message{}
	}

	// 获取会话基本信息
	scene, _ := model.GetConversationScene(id)
	if scene == "" {
		scene = "daily"
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"conversation_id": id,
			"scene":           scene,
			"title":           "",
			"messages":        messages,
		},
	})
}

// TextMessageHandler POST /api/message/text (文字兜底)
func TextMessageHandler(c *gin.Context) {
	var req struct {
		ConversationID int64  `json:"conversation_id"`
		Content        string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": -1, "message": "invalid request"})
		return
	}

	// 保存用户消息
	if req.ConversationID > 0 {
		model.SaveMessage(req.ConversationID, "user", req.Content, "")
	}

	// 确定场景提示词
	systemPrompt := DailyPrompt
	if req.ConversationID > 0 {
		if scene, err := model.GetConversationScene(req.ConversationID); err == nil {
			systemPrompt = GetSystemPrompt(scene)
		}
	}

	// 调用 LLM（非流式，通过管道收集）
	pr, pw := io.Pipe()
	var fullText strings.Builder

	go func() {
		defer pw.Close()
		llmReq := openai.ChatCompletionRequest{
			Model: "doubao-seed-2-0-pro-260215",
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
				{Role: openai.ChatMessageRoleUser, Content: req.Content},
			},
			Stream:      true,
			Temperature: 0.7,
			MaxTokens:   500,
		}
		stream, err := llmClient.CreateChatCompletionStream(context.Background(), llmReq)
		if err != nil {
			return
		}
		defer stream.Close()
		for {
			resp, err := stream.Recv()
			if err != nil {
				break
			}
			if len(resp.Choices) > 0 && resp.Choices[0].Delta.Content != "" {
				pw.Write([]byte(resp.Choices[0].Delta.Content))
			}
		}
	}()

	data := make([]byte, 4096)
	for {
		n, err := pr.Read(data)
		if err != nil {
			break
		}
		fullText.Write(data[:n])
	}

	// TTS 合成
	audioURL := ""
	audioData, err := ttsClient.TextToMP3(fullText.String())
	if err == nil {
		filename := fmt.Sprintf("tts_%d.mp3", time.Now().UnixNano())
		path := fmt.Sprintf("./audio/%s", filename)
		if os.WriteFile(path, audioData, 0644) == nil {
			audioURL = fmt.Sprintf("/audio/%s", filename)
		}
	}

	// 保存 AI 消息
	if req.ConversationID > 0 {
		model.SaveMessage(req.ConversationID, "assistant", fullText.String(), audioURL)
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"content":   fullText.String(),
			"audio_url": audioURL,
		},
	})
}

// DeleteConversationHandler DELETE /api/conversations/:id
func DeleteConversationHandler(c *gin.Context) {
	idStr := c.Param("id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	if err := model.DeleteConversation(id); err != nil {
		log.Println("删除会话失败:", err)
		c.JSON(500, gin.H{"code": -1, "message": "删除失败"})
		return
	}

	c.JSON(200, gin.H{"code": 0, "message": "success"})
}
