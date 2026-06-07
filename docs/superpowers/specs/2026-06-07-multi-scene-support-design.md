# 多场景支持（商务 + 考试）设计

**日期**: 2026-06-07  
**状态**: approved  
**范围**: `cmd/main.go` 单文件改动

---

## 1. 目标

在现有 `daily` 场景基础上增加 `business` 和 `exam` 两个场景，让用户创建会话时可以通过 `scene` 参数选择不同对话模式。

---

## 2. 新增 Prompt

| 场景 | 常量名 | 角色定位 |
|------|--------|----------|
| daily | `DailyPrompt`（已有） | 日常闲聊，温和纠错 |
| business | `BusinessPrompt`（新增） | 职场场景，专业商务用语 |
| exam | `ExamPrompt`（新增） | 雅思口语考官，严格评分 |

Prompt 内容直接使用文档 4.1 节原文。

---

## 3. 选择函数

```go
func GetSystemPrompt(scene string) string {
    switch scene {
    case "business": return BusinessPrompt
    case "exam":     return ExamPrompt
    default:         return DailyPrompt
    }
}
```

---

## 4. 改动点

### 4.1 streamLLM 签名变更

```go
// 旧: func streamLLM(userText string, conn *websocket.Conn) (string, error)
// 新:
func streamLLM(systemPrompt, userText string, conn *websocket.Conn) (string, error)
```

将 `DailyPrompt` 硬编码改为 `systemPrompt` 参数。

### 4.2 WebSocketHandler

- 有 `convID` → 查会话获取 `scene` → 选 prompt
- 无 `convID` → 默认 `daily`

### 4.3 TextMessageHandler

- 查会话获取 `scene` → 选 prompt → 传入 LLM 调用

### 4.4 conversation.go

新增一个小查询函数：

```go
func GetConversationScene(conversationID int64) (string, error)
```

用于快速查询会话的 scene 字段。

---

## 5. 不改动的

- REST API（POST /api/conversations 已支持 scene）
- WebSocket 路由
- 缓存 / ASR / TTS 逻辑
- 前端
