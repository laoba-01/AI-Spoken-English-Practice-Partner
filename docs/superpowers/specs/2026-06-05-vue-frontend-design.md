# Vue 3 前端设计方案 — AI 英语口语陪练

## 概述

基于 Vue 3 + Vite + TypeScript 构建 AI 英语口语陪练前端。通过 WebSocket（实时对话）和 REST API（会话管理）与现有 Go 后端通信，后端已运行在 `:8080`。

## 技术栈

- Vue 3（Composition API + `<script setup>`）
- Vite + TypeScript
- Element Plus（UI 组件库）
- Pinia（状态管理）
- Vue Router 4

## 路由

| 路径 | 页面 | 说明 |
|------|------|------|
| `/` | 首页 | 三大场景选择（日常/商务/考试），创建或继续会话 |
| `/chat/:id` | 对话页 | WebSocket 实时对话，语音 + 文字双模式 |
| `/history` | 历史页 | 按场景筛选查看过往会话列表 |

## 组件树

```
App.vue
├── layout/
│   ├── AppLayout.vue        — 左侧边栏 + 右侧主内容区
│   └── Sidebar.vue           — 会话列表、场景筛选、新建对话按钮
├── pages/
│   ├── HomePage.vue          — 场景选择卡片、快速开始
│   ├── ChatPage.vue          — WebSocket 聊天、语音录制、音频播放
│   └── HistoryPage.vue       — 分页会话列表
├── chat/
│   ├── ChatWindow.vue        — 可滚动消息列表 + 输入区域
│   ├── MessageBubble.vue     — 单条消息（文字、纠错、评分、语音播放）
│   ├── TextInput.vue         — 文字输入框 + 发送按钮
│   └── VoiceRecorder.vue     — 按住说话按钮
├── common/
│   ├── SceneSelector.vue     — 三张场景卡片
│   ├── AudioPlayer.vue       — TTS 语音播放按钮
│   └── EmptyState.vue        — 空状态占位
```

## 数据流与 Store 设计

### conversationStore（Pinia）
- `currentScene: 'daily' | 'business' | 'exam'`
- `conversationId: number | null`
- `messages: Message[]`
- `conversations: Conversation[]`
- Actions: `createConversation()`、`fetchHistory()`、`fetchConversations()`、`addMessage()`

### websocketStore（Pinia）
- `connected: boolean`
- `streamingText: string` — 当前正在接收的 LLM 流式文本
- Actions: `connect(id)`、`sendText(text)`、`sendVoice(mp3Blob)`、`disconnect()`
- 处理消息类型：`asr_result`、`llm_chunk`、`audio_result`、`error`

### API 接口层（纯 TS 函数）
- `POST /api/conversations` — 创建新会话
- `GET /api/conversations?user_id=&scene=` — 获取会话列表
- `GET /api/conversations/:id` — 获取会话消息历史
- `POST /api/message/text` — 文字消息兜底
- `WS /ws/chat/:id` — WebSocket 实时对话

## WebSocket 消息协议

```ts
// 服务端 → 客户端
{ type: 'asr_result',  text: string }       // 语音识别结果
{ type: 'llm_chunk',   text: string }       // AI 流式回复片段
{ type: 'audio_result', audio_url: string } // TTS 语音文件地址
{ type: 'error',       message: string }    // 错误消息

// 客户端 → 服务端
// 文字模式：直接发送文本字符串
// 语音模式：发送二进制 MP3 Blob
```

## 关键交互行为

- **语音录制**：使用 MediaRecorder API，MP3 格式，按住说话松开发送
- **流式文字**：`llm_chunk` 消息实时追加到同一个 AI 气泡中
- **语音播放**：收到 `audio_result` 后通过 `<audio>` 标签播放 TTS 音频
- **错误处理**：错误消息以 Toast 提示，WebSocket 断线自动重连（最多 3 次，间隔 1s）
- **兜底模式**：文字输入始终可用，不受语音模式影响
- **历史回顾**：进入历史会话只读展示，不建立 WebSocket 连接
