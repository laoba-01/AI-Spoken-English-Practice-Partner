# AI 英语口语陪练 (AI English Speaking Tutor)

基于 Go + Vue 3 的 AI 英语口语陪练平台，支持实时语音对话和文字对话双模式。

## 功能特性

### 核心功能
- 🎯 **三场景切换**：日常对话 / 商务英语 / 雅思托福考试模拟
- 🎤 **语音对话**：按住说话，ASR 流式识别（边说边转，<500ms）+ AI 语音回复
- 💬 **文字兜底**：文字输入始终可用，语音不可用时自动切换
- ⚡ **流式回复**：LLM 实时流式输出 + TTS 预生成（LLM 生成 3 句后并行合成语音）
- 📝 **语法纠错**：LLM 异步分析语法/用词/发音错误，分类高亮展示
- 🔊 **发音评分**：0-100 分综合评分（绿≥80 / 黄≥60 / 红<60）
- 📜 **历史记录**：按场景筛选查看过往对话，支持删除
- 🛡️ **限流保护**：IP 滑动窗口限流，每 IP 每分钟 10 次请求
- 📖 **Swagger 文档**：`/swagger/index.html` 交互式 API 文档

### 性能优化
| 指标 | 目标 | 实现 |
|------|------|------|
| ASR 识别延迟 | < 500ms | 流式识别 + 中间结果实时推送 |
| LLM 首字延迟 | < 1s | 豆包 Pro 流式输出 |
| TTS 生成延迟 | < 2s | 预生成：LLM 生成 3 句后即并行合成 |
| 并发支持 | ≥ 100 路 | Go 协程 + WebSocket 长连接 |

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端框架 | Go 1.22+ + Gin |
| 实时通信 | WebSocket (Gorilla) |
| 大模型 LLM | 火山引擎 豆包 Pro 4.0 |
| 语音识别 ASR | 阿里云智能语音（流式） |
| 语音合成 TTS | 阿里云智能语音（MP3） |
| 文件存储 | 阿里云 OSS + CDN 加速 |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7.0 |
| 前端框架 | Vue 3 + TypeScript + Vite |
| UI 组件库 | Element Plus |
| 状态管理 | Pinia |
| API 文档 | Swagger (swaggo) |

## 项目结构

```
├── cmd/main.go                       # Go 后端入口（所有逻辑）
├── internel/momel/                   # 数据模型 + DB/Redis 操作
│   ├── db.go                         # 数据库 & Redis 初始化
│   └── conversation.go               # 会话 & 消息 CRUD（含缓存）
├── internel/momel/package/pkg/       # 第三方服务封装
│   ├── asr/aliyun_asr.go             # 阿里云流式语音识别
│   ├── tts/aliyun_tts.go             # 阿里云语音合成
│   └── oss/aliyun_oss.go             # 阿里云 OSS + CDN
├── docs/                             # Swagger 自动生成文档
├── frontend/                         # Vue 3 前端
│   └── src/
│       ├── pages/                    # ChatPage / HomePage / HistoryPage
│       ├── components/chat/          # ChatWindow / MessageBubble / VoiceRecorder / TextInput
│       ├── components/common/        # AudioPlayer / SceneSelector
│       ├── stores/                   # Pinia: conversation / websocket
│       ├── api/                      # REST API 封装
│       ├── router/                   # Vue Router 配置
│       └── types/                    # TypeScript 类型定义
├── Dockerfile                        # 多阶段构建
├── docker-compose.yml                # 一键部署 (app + mysql + redis)
└── .env.example                      # 环境变量模板
```

## 快速启动

### 1. 环境变量

```bash
cp .env.example .env
# 编辑 .env，填入你的 API 密钥
```

必填：
- `DOUBAO_API_KEY` — 火山引擎豆包 API Key
- `ALIYUN_ACCESS_KEY_ID` / `ALIYUN_ACCESS_KEY_SECRET` — 阿里云 AK
- `ALIYUN_ASR_APPKEY` / `ALIYUN_TTS_APPKEY` — 阿里云语音 AppKey

可选：
- `MYSQL_DSN` — MySQL 连接串（不配则用内存存储，重启丢失）
- `REDIS_ADDR` — Redis 地址（不配则不使用缓存）
- `ALIYUN_OSS_ENDPOINT` / `ALIYUN_OSS_BUCKET` — OSS 配置（不配则音频存本地）
- `ALIYUN_OSS_DOMAIN` — CDN 加速域名

### 2. 后端

```bash
go run cmd/main.go
# 服务器启动在 :8080
# Swagger: http://localhost:8080/swagger/index.html
```

### 3. 前端

```bash
cd frontend
npm install
npm run dev
# 开发服务器 → http://localhost:3000
```

### 4. 数据库（可选，用于持久化）

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
    correction TEXT DEFAULT NULL COMMENT '纠错 JSON',
    pronunciation_score TINYINT DEFAULT NULL COMMENT '0-100',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_conversation (conversation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## Docker 部署

```bash
# 一键启动（含 MySQL + Redis）
docker-compose up -d

# 查看日志
docker-compose logs -f app

# 停止
docker-compose down
```

服务端口：
- 后端 API：`http://localhost:8080`
- Swagger：`http://localhost:8080/swagger/index.html`

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/conversations` | 创建新会话 |
| GET | `/api/conversations?user_id=&scene=` | 获取会话列表 |
| GET | `/api/conversations/:id` | 获取会话历史 |
| POST | `/api/message/text` | 文字消息兜底 |
| DELETE | `/api/conversations/:id` | 删除会话 |
| WS | `/ws/chat/:id` | WebSocket 实时对话 |

### WebSocket 消息类型

| type | 方向 | 说明 |
|------|------|------|
| `asr_interim` | 后端→前端 | ASR 中间识别结果（实时展示） |
| `asr_result` | 后端→前端 | ASR 最终识别结果 |
| `llm_chunk` | 后端→前端 | LLM 流式回复片段 |
| `audio_result` | 后端→前端 | TTS 合成语音 URL |
| `correction_result` | 后端→前端 | 纠错分析 + 发音评分 |
| `error` | 后端→前端 | 错误消息 |

### 限流

所有 `/api/*` 接口：每 IP 每分钟最多 10 次请求，超限返回 `429 Too Many Requests`。

## 架构

```
浏览器 ──WebSocket──▶ 后端 ──流式──▶ 阿里云 ASR（语音→文字）
   │                    │
   │                    ├──流式──▶ 豆包 Pro（生成回复 + 纠错）
   │                    │
   │                    ├───────▶ 阿里云 TTS（文字→MP3）
   │                    │
   │                    ├───────▶ 阿里云 OSS / CDN（存储音频）
   │                    │
   │                    ├───────▶ MySQL（持久化）
   │                    │
   │                    └───────▶ Redis（缓存）
   │
   └── HTTP ──▶ Vite 代理 ──▶ 后端 REST API
```
