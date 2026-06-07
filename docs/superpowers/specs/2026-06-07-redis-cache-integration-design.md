# Redis 缓存接入设计

**日期**: 2026-06-07  
**状态**: approved  
**方案**: Cache-Aside（旁路缓存），不改动现有存储逻辑，在读写前后插入缓存层

---

## 1. 目标

为 AI 英语口语陪练项目接入 Redis 缓存，加速以下数据访问：
- 会话列表查询 `GetConversationsByUserID`
- 历史消息查询 `GetMessagesByConversationID`

Redis 不可用时自动降级，系统照常运行。

---

## 2. 新增配置项

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `REDIS_ADDR` | - | Redis 地址，如 `localhost:6379`。不配置则不启用缓存 |
| `REDIS_PASSWORD` | 空 | Redis 密码 |
| `REDIS_DB` | `0` | Redis 数据库编号 |

---

## 3. 缓存 Key 设计

| 数据 | Key 格式 | 值类型 | TTL | 失效策略 |
|------|----------|--------|-----|----------|
| 会话列表 | `conv:list:{userID}:{scene}` | String(JSON数组) | 5分钟 | 创建/删除会话时删除 |
| 会话消息 | `conv:msgs:{convID}` | String(JSON数组) | 2分钟 | 新消息写入时删除 |

---

## 4. 改动涉及文件

### 4.1 `internal/model/db.go` — 新增 Redis 初始化

```go
var RedisClient *redis.Client

func InitRedis() error
```

- 读取 `REDIS_ADDR` 环境变量，未配置直接返回 nil（不启用缓存）
- 连接失败也返回 nil，打印 warning 日志
- 成功则设置 `RedisClient` 为非 nil，后续缓存逻辑自动生效

main.go 在 `InitDB()` 之后调用 `model.InitRedis()`。

### 4.2 `internal/model/conversation.go` — 5 个函数加入缓存逻辑

每个函数遵循相同模式：

```
读函数（Get/List）：
  1. 构造 key
  2. Redis GET → 命中则 Unmarshal 直接返回
  3. 未命中走原查询逻辑（MySQL / 内存）
  4. 查得结果后 Marshal → Redis SET + TTL → 返回

写函数（Create/Save/Delete）：
  1. 执行原写入逻辑（MySQL / 内存）
  2. 成功后 Redis DEL 相关 key
```

**具体改动：**

| 函数 | 类型 | 缓存 Key | 操作 |
|------|------|----------|------|
| `GetConversationsByUserID` | 读 | `conv:list:{userID}:{scene}` | GET → 查库 → SET |
| `GetMessagesByConversationID` | 读 | `conv:msgs:{convID}` | GET → 查库 → SET |
| `CreateConversation` | 写 | `conv:list:{userID}:{scene}` | 写库 → DEL |
| `SaveMessage` | 写 | `conv:msgs:{convID}` | 写库 → DEL |
| `DeleteConversation` | 写 | `conv:msgs:{convID}`（直接删），`conv:list:{userID}:{scene}`（需先查出会话信息再删） | 查会话→删库→DEL 两个 key |

### 4.3 `.env.example` — 补充 Redis 配置项

```
# Redis 缓存（可选，不配置则使用直查模式）
# REDIS_ADDR=localhost:6379
# REDIS_PASSWORD=
```

---

### 4.4 DeleteConversation 特殊处理

当前 `DeleteConversation(convID)` 不持有 userID 和 scene。缓存失效需要先查出这两个字段：

```go
func DeleteConversation(conversationID int64) error {
    // ① 先查出会话的 userID 和 scene（用于后续删缓存）
    var userID int64
    var scene string
    if RedisClient != nil {
        // 从内存或 MySQL 查出会话信息
    }

    // ② 执行原有删除逻辑 ...

    // ③ 删除缓存
    if RedisClient != nil {
        RedisClient.Del(ctx, fmt.Sprintf("conv:msgs:%d", conversationID))
        RedisClient.Del(ctx, fmt.Sprintf("conv:list:%d:%s", userID, scene))
    }
}
```

内存 fallback 中已有会话列表 `memConversations`，遍历即可查到；MySQL 中加一条 SELECT。

---

## 5. 降级行为

| 场景 | 行为 |
|------|------|
| 未配置 `REDIS_ADDR` | `RedisClient = nil`，所有缓存代码跳过 |
| Redis 连接失败 | 同上，打 warning 日志 |
| Redis 运行时断开（GET/SET/DEL 出错） | 忽略错误，走原有逻辑，不影响请求 |

任何时候 Redis 挂了，系统的读写能力不受任何影响。

---

## 6. 依赖

- `github.com/redis/go-redis/v9` — 已在 go.mod 中声明，无需新增
- `encoding/json` — Go 标准库，已有
- `context`、`fmt` — Go 标准库，已有

---

## 7. 不改动的部分

- MySQL 连接逻辑不变
- 内存 fallback 逻辑不变
- main.go 路由和 Handler 不变
- ASR、TTS、LLM 调用逻辑不变
- 前端代码不变
