# Day2 缺口补齐：对话历史 + OSS 上传

**日期**: 2026-06-07  
**状态**: approved  
**范围**: `cmd/main.go` + 新建 `internel/momel/package/pkg/oss/aliyun_oss.go`

---

## 1. 对话历史上下文

### 问题
当前 `streamLLM` 每次只发送 `{system prompt, 当前消息}`，LLM 无对话记忆。

### 方案
LLM 调用时加载最近 10 条历史消息拼入 Messages 数组。

**streamLLM 签名变更：**
```go
// 旧
func streamLLM(systemPrompt, userText string, conn *websocket.Conn) (string, error)
// 新
func streamLLM(systemPrompt string, history []model.Message, userText string, conn *websocket.Conn) (string, error)
```

**Handler 调用处：**
```go
history, _ := model.GetMessagesByConversationID(cid)
if len(history) > 10 {
    history = history[len(history)-10:] // 最近10条
}
streamLLM(systemPrompt, history, userText, conn)
```

**TextMessageHandler 同样处理。**

---

## 2. 阿里云 OSS 上传

### 新建文件
`internel/momel/package/pkg/oss/aliyun_oss.go`

### 配置项
| 环境变量 | 必填 | 说明 |
|----------|------|------|
| `ALIYUN_OSS_ENDPOINT` | 是 | 如 `oss-cn-beijing.aliyuncs.com` |
| `ALIYUN_OSS_BUCKET` | 是 | Bucket 名称 |
| `ALIYUN_ACCESS_KEY_ID` | 是 | 复用已有 |
| `ALIYUN_ACCESS_KEY_SECRET` | 是 | 复用已有 |

### 接口
```go
type AliyunOSS struct { ... }
func NewAliyunOSS() *AliyunOSS    // 未配置返回 nil
func (o *AliyunOSS) UploadMP3(data []byte, filename string) (string, error)
```

### 调用处 (main.go)
TTS 生成 MP3 后：
```go
if ossClient != nil {
    if url, err := ossClient.UploadMP3(audioData, audioFilename); err == nil {
        audioURL = url
    }
}
if audioURL == "" {
    // 降级：WriteFile 到 ./audio/（现有逻辑）
}
```

---

## 3. 改动文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `internel/momel/package/pkg/oss/aliyun_oss.go` | 新建 | OSS 上传模块 |
| `cmd/main.go` | 修改 | streamLLM 加 history、Handler 加载历史、OSS 调用 |
| `.env.example` | 修改 | 补 OSS 配置项 |

## 4. 不改动
- conversation.go、db.go（已有 GetMessagesByConversationID）
- ASR、TTS
- WebSocket 路由、REST API
- 前端
