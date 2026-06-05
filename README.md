# AI-Spoken-English-Practice-Partner

面向英语爱好者的 AI 英语口语陪练助手

基于 Go + Vue 3 的 AI 英语口语陪练平台，支持实时语音对话和文字对话双模式。

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端框架 | Go + Gin |
| 实时通信 | WebSocket (Gorilla) |
| 大模型 | 火山引擎豆包 Pro |
| 前端框架 | Vue 3 + TypeScript + Vite |
| UI 组件库 | Element Plus |
| 状态管理 | Pinia |
| 路由 | Vue Router 4 |

## 功能特性

- 🎯 **三场景切换**：日常对话 / 商务英语 / 考试模拟（雅思托福）
- 🎤 **语音对话**：按住说话，AI 实时语音回复
- 💬 **文字兜底**：文字输入始终可用
- 📝 **语法纠错**：温柔纠正语法和用词错误
- 🔊 **发音评分**：0-100 分发音评估
- 📜 **历史记录**：按场景筛选查看过往对话
- ⚡ **流式回复**：LLM 实时流式输出，秒级响应

## 项目结构

```
├── cmd/main.go                  # Go 后端入口
├── internal/model/              # 数据模型 + DB 操作
│   ├── db.go                    # 数据库初始化
│   └── conversation.go          # 会话 & 消息 CRUD
├── internal/pkg/                # 第三方服务封装
│   ├── asr/aliyun_asr.go        # 阿里云语音识别
│   └── tts/aliyun_tts.go        # 阿里云语音合成
├── frontend/                    # Vue 3 前端
│   └── src/
│       ├── pages/               # 页面组件
│       ├── components/          # 通用 & 聊天组件
│       ├── stores/              # Pinia 状态管理
│       ├── api/                 # API 层
│       ├── router/              # 路由配置
│       └── types/               # 类型定义
└── docs/superpowers/specs/      # 设计文档
```

## 快速启动

### 1. 后端

```bash
# 设置环境变量
export MYSQL_DSN="user:password@tcp(localhost:3306)/english_tutor?charset=utf8mb4&parseTime=True"
export ALIYUN_ACCESS_KEY_ID="your_key"
export ALIYUN_ACCESS_KEY_SECRET="your_secret"

# 运行
go run cmd/main.go
# 服务器启动在 :8080
```

### 2. 前端

```bash
cd frontend
npm install
npm run dev
# 开发服务器 → http://localhost:3000
```

### 3. 数据库

```sql
CREATE TABLE user_conversations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    scene VARCHAR(20) NOT NULL COMMENT 'daily/business/exam',
    title VARCHAR(100) DEFAULT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_scene (user_id, scene)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE user_messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    conversation_id BIGINT NOT NULL,
    role VARCHAR(10) NOT NULL COMMENT 'user/assistant',
    content TEXT NOT NULL,
    audio_url VARCHAR(255) DEFAULT NULL,
    correction TEXT DEFAULT NULL,
    pronunciation_score TINYINT DEFAULT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_conversation (conversation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/conversations` | 创建新会话 |
| GET | `/api/conversations` | 获取会话列表 |
| GET | `/api/conversations/:id` | 获取会话历史 |
| POST | `/api/message/text` | 文字消息兜底 |
| WS | `/ws/chat/:id` | WebSocket 实时对话 |
