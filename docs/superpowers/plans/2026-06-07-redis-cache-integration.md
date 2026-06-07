# Redis 缓存接入实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 AI 英语口语陪练项目接入 Redis 缓存层（Cache-Aside 模式），加速会话列表和消息历史查询

**Architecture:** 在现有 `conversation.go` 每个读写函数前后插入 Redis GET/SET/DEL，Redis 不可用时自动降级走原有 MySQL/内存逻辑。所有缓存逻辑包裹在 `if RedisClient != nil` 保护中。

**Tech Stack:** Go 1.25, github.com/redis/go-redis/v9, encoding/json

---

### Task 1: 新增 Redis 初始化函数

**Files:**
- Modify: `internel/momel/db.go`

- [ ] **Step 1: 在 db.go 文件顶部新增 import 和全局变量**

在现有 import 块中新增：

```go
import (
    "context"
    "database/sql"
    "log"
    "os"
    "sync"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/redis/go-redis/v9"
)
```

在 `var DB *sql.DB` 下方新增：

```go
var RedisClient *redis.Client
```

- [ ] **Step 2: 在 db.go 末尾新增 InitRedis 函数**

在文件末尾新增：

```go
func InitRedis() error {
    addr := os.Getenv("REDIS_ADDR")
    if addr == "" {
        log.Println("⚠ 未配置 REDIS_ADDR，不使用缓存")
        return nil
    }
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: os.Getenv("REDIS_PASSWORD"),
        DB:       0,
    })
    ctx := context.Background()
    if err := rdb.Ping(ctx).Err(); err != nil {
        log.Println("⚠ Redis连接失败，不使用缓存:", err)
        return nil
    }
    RedisClient = rdb
    log.Println("✓ Redis缓存已启用 (addr=" + addr + ")")
    return nil
}
```

- [ ] **Step 3: 验证编译通过**

```bash
cd "D:\七牛云夏令营\AI英语口语陪练" && go build ./...
```

预期：编译成功，`RedisClient` 变量和 `InitRedis` 函数被声明但尚未被调用。

---

### Task 2: main.go 调用 InitRedis

**Files:**
- Modify: `cmd/main.go`

- [ ] **Step 1: 在 InitDB 之后添加 InitRedis 调用**

找到 `main()` 中这一段：

```go
// 初始化数据库（非致命）
if err := model.InitDB(); err != nil {
    log.Println("⚠ 数据库连接失败，对话记录不会保存:", err)
}
```

在其后新增：

```go
// 初始化Redis缓存（非致命）
if err := model.InitRedis(); err != nil {
    log.Println("⚠ Redis缓存连接失败:", err)
}
```

- [ ] **Step 2: 验证编译和启动日志**

```bash
cd "D:\七牛云夏令营\AI英语口语陪练" && go build -o cmd.exe ./cmd/
```

预期：编译成功。启动后会打印 `⚠ 未配置 REDIS_ADDR，不使用缓存`（因为当前 .env 没配 Redis）。

---

### Task 3: GetConversationsByUserID 加缓存

**Files:**
- Modify: `internel/momel/conversation.go`

- [ ] **Step 1: 在文件顶部新增 import**

```go
import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"
)
```

- [ ] **Step 2: 改造 GetConversationsByUserID 函数**

将原函数改为：

```go
func GetConversationsByUserID(userID int64, scene string) ([]Conversation, error) {
    ctx := context.Background()

    // ① Redis 缓存查询
    if RedisClient != nil {
        key := fmt.Sprintf("conv:list:%d:%s", userID, scene)
        cached, err := RedisClient.Get(ctx, key).Result()
        if err == nil {
            var convs []Conversation
            if json.Unmarshal([]byte(cached), &convs) == nil {
                return convs, nil
            }
        }
    }

    // ② 查 MySQL / 内存（原有逻辑）
    if DB != nil {
        var rows *sql.Rows
        var err error
        if scene != "" {
            rows, err = DB.Query(
                "SELECT id, user_id, scene, COALESCE(title,''), created_at FROM user_conversations WHERE user_id = ? AND scene = ? ORDER BY id DESC",
                userID, scene,
            )
        } else {
            rows, err = DB.Query(
                "SELECT id, user_id, scene, COALESCE(title,''), created_at FROM user_conversations WHERE user_id = ? ORDER BY id DESC",
                userID,
            )
        }
        if err != nil {
            return nil, err
        }
        defer rows.Close()

        var convs []Conversation
        for rows.Next() {
            var c Conversation
            if err := rows.Scan(&c.ID, &c.UserID, &c.Scene, &c.Title, &c.CreatedAt); err != nil {
                return nil, err
            }
            convs = append(convs, c)
        }

        // ③ 回填 Redis
        if RedisClient != nil && convs != nil {
            key := fmt.Sprintf("conv:list:%d:%s", userID, scene)
            if data, err := json.Marshal(convs); err == nil {
                RedisClient.Set(ctx, key, data, 5*time.Minute)
            }
        }
        return convs, nil
    }

    // 内存 fallback
    memMu.Lock()
    defer memMu.Unlock()
    var result []Conversation
    for _, c := range memConversations {
        if c.UserID == userID && (scene == "" || c.Scene == scene) {
            result = append(result, c)
        }
    }

    // ③ 回填 Redis（内存结果也缓存）
    if RedisClient != nil && result != nil {
        key := fmt.Sprintf("conv:list:%d:%s", userID, scene)
        if data, err := json.Marshal(result); err == nil {
            RedisClient.Set(ctx, key, data, 5*time.Minute)
        }
    }
    return result, nil
}
```

- [ ] **Step 3: 验证编译通过**

```bash
cd "D:\七牛云夏令营\AI英语口语陪练" && go build ./...
```

预期：编译成功。

---

### Task 4: GetMessagesByConversationID 加缓存

**Files:**
- Modify: `internel/momel/conversation.go`

- [ ] **Step 1: 改造 GetMessagesByConversationID 函数**

将原函数改为：

```go
func GetMessagesByConversationID(conversationID int64) ([]Message, error) {
    ctx := context.Background()

    // ① Redis 缓存查询
    if RedisClient != nil {
        key := fmt.Sprintf("conv:msgs:%d", conversationID)
        cached, err := RedisClient.Get(ctx, key).Result()
        if err == nil {
            var msgs []Message
            if json.Unmarshal([]byte(cached), &msgs) == nil {
                return msgs, nil
            }
        }
    }

    // ② 查 MySQL / 内存（原有逻辑）
    if DB != nil {
        rows, err := DB.Query(
            "SELECT id, conversation_id, role, content, audio_url, created_at FROM user_messages WHERE conversation_id = ? ORDER BY id ASC",
            conversationID,
        )
        if err != nil {
            return nil, err
        }
        defer rows.Close()

        var messages []Message
        for rows.Next() {
            var msg Message
            err := rows.Scan(
                &msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.AudioURL, &msg.CreatedAt,
            )
            if err != nil {
                return nil, err
            }
            messages = append(messages, msg)
        }

        // ③ 回填 Redis
        if RedisClient != nil && messages != nil {
            key := fmt.Sprintf("conv:msgs:%d", conversationID)
            if data, err := json.Marshal(messages); err == nil {
                RedisClient.Set(ctx, key, data, 2*time.Minute)
            }
        }
        return messages, nil
    }

    // 内存 fallback
    memMu.Lock()
    defer memMu.Unlock()
    var result []Message
    for _, m := range memMessages {
        if m.ConversationID == conversationID {
            result = append(result, m)
        }
    }

    // ③ 回填 Redis
    if RedisClient != nil && result != nil {
        key := fmt.Sprintf("conv:msgs:%d", conversationID)
        if data, err := json.Marshal(result); err == nil {
            RedisClient.Set(ctx, key, data, 2*time.Minute)
        }
    }
    return result, nil
}
```

- [ ] **Step 2: 验证编译通过**

```bash
cd "D:\七牛云夏令营\AI英语口语陪练" && go build ./...
```

预期：编译成功。

---

### Task 5: CreateConversation 加缓存失效

**Files:**
- Modify: `internel/momel/conversation.go`

- [ ] **Step 1: 改造 CreateConversation 函数**

将原函数改为：

```go
func CreateConversation(userID int64, scene string) (int64, error) {
    var id int64
    var dbErr error

    if DB != nil {
        res, err := DB.Exec(
            "INSERT INTO user_conversations (user_id, scene) VALUES (?, ?)",
            userID, scene,
        )
        if err != nil {
            return 0, err
        }
        id, dbErr = res.LastInsertId()
        if dbErr != nil {
            return 0, dbErr
        }
    } else {
        // 内存 fallback
        memMu.Lock()
        id = memConvIDSeq
        memConvIDSeq++
        memConversations = append(memConversations, Conversation{
            ID: id, UserID: userID, Scene: scene, Title: "", CreatedAt: memNow(),
        })
        memMu.Unlock()
    }

    // 删除缓存（使下次列表查询重建）
    if RedisClient != nil {
        ctx := context.Background()
        key := fmt.Sprintf("conv:list:%d:%s", userID, scene)
        RedisClient.Del(ctx, key)
    }
    return id, nil
}
```

- [ ] **Step 2: 验证编译通过**

```bash
cd "D:\七牛云夏令营\AI英语口语陪练" && go build ./...
```

---

### Task 6: SaveMessage 加缓存失效

**Files:**
- Modify: `internel/momel/conversation.go`

- [ ] **Step 1: 改造 SaveMessage 函数**

将原函数改为：

```go
func SaveMessage(conversationID int64, role, content, audioURL string) error {
    if DB != nil {
        _, err := DB.Exec(
            "INSERT INTO user_messages (conversation_id, role, content, audio_url) VALUES (?, ?, ?, ?)",
            conversationID, role, content, audioURL,
        )
        if err != nil {
            return err
        }
    } else {
        // 内存 fallback
        memMu.Lock()
        id := memMsgIDSeq
        memMsgIDSeq++
        memMessages = append(memMessages, Message{
            ID:             id,
            ConversationID: conversationID,
            Role:           role,
            Content:        content,
            AudioURL:       audioURL,
            CreatedAt:      memNow(),
        })
        memMu.Unlock()
    }

    // 删除消息缓存（下次查询时重建）
    if RedisClient != nil {
        ctx := context.Background()
        key := fmt.Sprintf("conv:msgs:%d", conversationID)
        RedisClient.Del(ctx, key)
    }
    return nil
}
```

- [ ] **Step 2: 验证编译通过**

```bash
cd "D:\七牛云夏令营\AI英语口语陪练" && go build ./...
```

---

### Task 7: DeleteConversation 加缓存失效

**Files:**
- Modify: `internel/momel/conversation.go`

- [ ] **Step 1: 改造 DeleteConversation 函数**

将原函数改为（删除前先查出 userID 和 scene 用于精确删缓存）：

```go
func DeleteConversation(conversationID int64) error {
    // ① 删除前查出会话信息（用于缓存失效）
    var targetUserID int64
    var targetScene string

    if DB != nil {
        row := DB.QueryRow(
            "SELECT user_id, scene FROM user_conversations WHERE id = ?",
            conversationID,
        )
        row.Scan(&targetUserID, &targetScene)

        _, err := DB.Exec("DELETE FROM user_messages WHERE conversation_id = ?", conversationID)
        if err != nil {
            return err
        }
        _, err = DB.Exec("DELETE FROM user_conversations WHERE id = ?", conversationID)
        if err != nil {
            return err
        }
    } else {
        // 内存 fallback
        memMu.Lock()
        // 先查会话信息
        for _, c := range memConversations {
            if c.ID == conversationID {
                targetUserID = c.UserID
                targetScene = c.Scene
                break
            }
        }
        // 删除消息
        var filteredMsgs []Message
        for _, m := range memMessages {
            if m.ConversationID != conversationID {
                filteredMsgs = append(filteredMsgs, m)
            }
        }
        memMessages = filteredMsgs
        // 删除会话
        var filteredConvs []Conversation
        for _, c := range memConversations {
            if c.ID != conversationID {
                filteredConvs = append(filteredConvs, c)
            }
        }
        memConversations = filteredConvs
        memMu.Unlock()
    }

    // ② 删除缓存
    if RedisClient != nil {
        ctx := context.Background()
        // 删除消息缓存
        RedisClient.Del(ctx, fmt.Sprintf("conv:msgs:%d", conversationID))
        // 删除会话列表缓存
        if targetUserID > 0 {
            RedisClient.Del(ctx, fmt.Sprintf("conv:list:%d:%s", targetUserID, targetScene))
        }
    }
    return nil
}
```

- [ ] **Step 2: 验证编译通过**

```bash
cd "D:\七牛云夏令营\AI英语口语陪练" && go build ./...
```

预期：编译成功，所有改动完成。

---

### Task 8: .env.example 补充 Redis 配置项

**Files:**
- Modify: `.env.example`

- [ ] **Step 1: 在 .env.example 末尾追加 Redis 配置**

在文件末尾添加：

```
# Redis 缓存（可选，不配置则使用直查模式）
# REDIS_ADDR=localhost:6379
# REDIS_PASSWORD=
```

- [ ] **Step 2: 最终验证编译**

```bash
cd "D:\七牛云夏令营\AI英语口语陪练" && go build -o cmd.exe ./cmd/ && echo "BUILD OK"
```

预期：`BUILD OK`

---

### 验证清单（全部 Tasks 完成后）

- [ ] 不配置 `REDIS_ADDR` 时系统启动正常，日志显示 "⚠ 未配置 REDIS_ADDR，不使用缓存"
- [ ] 配置 `REDIS_ADDR=localhost:6379` 且 Redis 可用时，日志显示 "✓ Redis缓存已启用"
- [ ] Redis 运行时，列表查询和消息查询均走缓存（二次查询毫秒级返回）
- [ ] 创建会话后列表缓存自动失效
- [ ] 发送消息后消息缓存自动失效
- [ ] 删除会话后对应列表缓存和消息缓存均失效
- [ ] 杀掉 Redis 后系统不受影响，所有请求正常返回
