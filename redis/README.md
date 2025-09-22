# Redis 模块

基于 [go-redis v9](https://github.com/redis/go-redis) 的 Redis 客户端封装，提供简单易用的 Redis 操作接口。

## 特性

- 🚀 **高性能**: 基于 go-redis v9，支持连接池和管道操作
- 🔧 **易用性**: 提供简洁的 API 接口，支持链式调用
- 🛡️ **类型安全**: 完整的类型支持，减少运行时错误
- 📦 **模块化**: 按功能模块组织，支持按需使用
- 🔄 **事务支持**: 完整的事务和管道操作支持
- 📡 **发布订阅**: 支持 Redis 发布订阅功能
- ⚙️ **配置灵活**: 支持多种配置方式和连接选项
- 🧪 **测试完备**: 包含完整的单元测试和示例代码

## 安装

```bash
go get github.com/daxiong0327/tool-kit/redis
```

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "log"
    
    "github.com/daxiong0327/tool-kit/redis"
)

func main() {
    // 创建客户端
    config := redis.DefaultConfig()
    config.Addr = "localhost:6379"
    
    client, err := redis.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    ctx := context.Background()
    
    // 测试连接
    err = client.Ping(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // 字符串操作
    stringOps := client.NewString()
    err = stringOps.Set(ctx, "key", "value", 0)
    if err != nil {
        log.Fatal(err)
    }
    
    value, err := stringOps.Get(ctx, "key")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("Value:", value)
}
```

### 从 URL 创建客户端

```go
client, err := redis.NewFromURL("redis://localhost:6379/0")
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

## 配置选项

### 基本配置

```go
config := &redis.Config{
    Addr:     "localhost:6379",  // Redis 地址
    Password: "",                // 密码
    DB:       0,                 // 数据库编号
    Username: "",                // 用户名
    Protocol: 3,                 // 协议版本 (2 或 3)
}
```

### 连接池配置

```go
config := &redis.Config{
    Addr:            "localhost:6379",
    PoolSize:        10,                    // 连接池大小
    MinIdleConns:    5,                     // 最小空闲连接数
    MaxIdleConns:    10,                    // 最大空闲连接数
    ConnMaxIdleTime: 30 * time.Minute,     // 连接最大空闲时间
    ConnMaxLifetime: 0,                     // 连接最大生存时间
}
```

### 超时配置

```go
config := &redis.Config{
    Addr:         "localhost:6379",
    DialTimeout:  5 * time.Second,  // 连接超时
    ReadTimeout:  3 * time.Second,  // 读取超时
    WriteTimeout: 3 * time.Second,  // 写入超时
}
```

### 重试配置

```go
config := &redis.Config{
    Addr:            "localhost:6379",
    MaxRetries:      3,                        // 最大重试次数
    MinRetryBackoff: 8 * time.Millisecond,    // 最小重试间隔
    MaxRetryBackoff: 512 * time.Millisecond,  // 最大重试间隔
}
```

## 数据类型操作

### 字符串操作

```go
stringOps := client.NewString()

// 基本操作
err := stringOps.Set(ctx, "key", "value", 0)
value, err := stringOps.Get(ctx, "key")

// 数值操作
err := stringOps.Set(ctx, "counter", "10", 0)
count, err := stringOps.Incr(ctx, "counter")
count, err := stringOps.IncrBy(ctx, "counter", 5)

// 条件设置
success, err := stringOps.SetNX(ctx, "key", "value", 0)  // 仅当键不存在时
success, err := stringOps.SetXX(ctx, "key", "value", 0)  // 仅当键存在时

// 批量操作
err := stringOps.MSet(ctx, "key1", "value1", "key2", "value2")
values, err := stringOps.MGet(ctx, "key1", "key2")

// 位操作
err := stringOps.SetBit(ctx, "bitkey", 0, 1)
bit, err := stringOps.GetBit(ctx, "bitkey", 0)
```

### 哈希操作

```go
hashOps := client.NewHash()

// 基本操作
err := hashOps.HSet(ctx, "user:1", "name", "张三", "age", "25")
value, err := hashOps.HGet(ctx, "user:1", "name")
all, err := hashOps.HGetAll(ctx, "user:1")

// 批量操作
err := hashOps.HMSet(ctx, "user:1", "name", "李四", "age", "30")
values, err := hashOps.HMGet(ctx, "user:1", "name", "age")

// 数值操作
count, err := hashOps.HIncrBy(ctx, "user:1", "age", 1)
score, err := hashOps.HIncrByFloat(ctx, "user:1", "score", 0.5)

// 其他操作
exists, err := hashOps.HExists(ctx, "user:1", "name")
length, err := hashOps.HLen(ctx, "user:1")
keys, err := hashOps.HKeys(ctx, "user:1")
values, err := hashOps.HVals(ctx, "user:1")
```

### 列表操作

```go
listOps := client.NewList()

// 推入操作
err := listOps.LPush(ctx, "list", "item1", "item2")
err := listOps.RPush(ctx, "list", "item3", "item4")

// 弹出操作
item, err := listOps.LPop(ctx, "list")
item, err := listOps.RPop(ctx, "list")

// 阻塞操作
items, err := listOps.BLPop(ctx, 5*time.Second, "list")
items, err := listOps.BRPop(ctx, 5*time.Second, "list")

// 范围操作
items, err := listOps.LRange(ctx, "list", 0, -1)
item, err := listOps.LIndex(ctx, "list", 0)

// 其他操作
length, err := listOps.LLen(ctx, "list")
err := listOps.LSet(ctx, "list", 0, "newitem")
err := listOps.LRem(ctx, "list", 1, "item")
```

### 集合操作

```go
setOps := client.NewSet()

// 基本操作
err := setOps.SAdd(ctx, "set", "member1", "member2", "member3")
err := setOps.SRem(ctx, "set", "member1")

// 查询操作
members, err := setOps.SMembers(ctx, "set")
exists, err := setOps.SIsMember(ctx, "set", "member1")
count, err := setOps.SCard(ctx, "set")

// 集合运算
union, err := setOps.SUnion(ctx, "set1", "set2")
intersection, err := setOps.SInter(ctx, "set1", "set2")
difference, err := setOps.SDiff(ctx, "set1", "set2")

// 存储运算结果
count, err := setOps.SUnionStore(ctx, "result", "set1", "set2")
count, err := setOps.SInterStore(ctx, "result", "set1", "set2")
count, err := setOps.SDiffStore(ctx, "result", "set1", "set2")
```

### 有序集合操作

```go
zsetOps := client.NewZSet()

// 添加成员
err := zsetOps.ZAdd(ctx, "zset", redis.Z{Score: 100, Member: "member1"})
err := zsetOps.ZAdd(ctx, "zset", redis.Z{Score: 200, Member: "member2"})

// 范围查询
members, err := zsetOps.ZRange(ctx, "zset", 0, -1)
members, err := zsetOps.ZRevRange(ctx, "zset", 0, -1)
members, err := zsetOps.ZRangeWithScores(ctx, "zset", 0, -1)

// 按分数查询
members, err := zsetOps.ZRangeByScore(ctx, "zset", &redis.ZRangeBy{
    Min: "100",
    Max: "200",
})

// 排名查询
rank, err := zsetOps.ZRank(ctx, "zset", "member1")
rank, err := zsetOps.ZRevRank(ctx, "zset", "member1")
score, err := zsetOps.ZScore(ctx, "zset", "member1")

// 数值操作
newScore, err := zsetOps.ZIncrBy(ctx, "zset", 50, "member1")

// 删除操作
err := zsetOps.ZRem(ctx, "zset", "member1")
count, err := zsetOps.ZRemRangeByRank(ctx, "zset", 0, 1)
count, err := zsetOps.ZRemRangeByScore(ctx, "zset", "100", "200")
```

## 高级功能

### Lua脚本支持

Redis模块提供了完整的Lua脚本支持，包括脚本管理、执行、缓存和监控功能。

#### 基本使用

```go
script := client.NewScript()

// 注册脚本
scriptInfo := &redis.ScriptInfo{
    Name:        "hello_script",
    Source:      `return "Hello, " .. ARGV[1]`,
    Keys:        []string{},
    Args:        []string{"name"},
    Description: "问候脚本",
    Timeout:     5 * time.Second,
}

err := script.Register(ctx, scriptInfo)
if err != nil {
    log.Fatal(err)
}

// 执行脚本
opts := &redis.ScriptOptions{
    Keys:        []string{},
    Args:        []interface{}{"World"},
    Timeout:     5 * time.Second,
    RetryCount:  3,
    RetryDelay:  100 * time.Millisecond,
    UseCache:    true,
    ForceReload: false,
}

result, err := script.Execute(ctx, "hello_script", opts)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Result:", result.Value)
```

#### 脚本模板

```go
templates := script.NewScriptTemplates()

// 注册常用脚本模板
err := templates.RegisterCommonScripts(ctx)
if err != nil {
    log.Fatal(err)
}

// 使用分布式锁脚本
lockOpts := &redis.ScriptOptions{
    Keys:        []string{"my:lock"},
    Args:        []interface{}{"lock_value", 10},
    Timeout:     5 * time.Second,
    RetryCount:  3,
    RetryDelay:  100 * time.Millisecond,
    UseCache:    true,
    ForceReload: false,
}

result, err := script.Execute(ctx, "distributed_lock", lockOpts)
if err != nil {
    log.Fatal(err)
}

if result.Value == int64(1) {
    fmt.Println("成功获取锁")
}
```

#### 脚本管理器

```go
manager := script.NewScriptManager()

// 注册脚本
scriptInfo := &redis.ScriptInfo{
    Name:        "my_script",
    Source:      `return "Script result: " .. ARGV[1]`,
    Keys:        []string{},
    Args:        []string{"param"},
    Description: "我的脚本",
    Timeout:     5 * time.Second,
}

err := manager.RegisterScript(ctx, scriptInfo)
if err != nil {
    log.Fatal(err)
}

// 执行脚本
opts := &redis.ScriptOptions{
    Keys:        []string{},
    Args:        []interface{}{"test"},
    Timeout:     5 * time.Second,
    RetryCount:  3,
    RetryDelay:  100 * time.Millisecond,
    UseCache:    true,
    ForceReload: false,
}

result, err := manager.ExecuteScript(ctx, "my_script", opts)
if err != nil {
    log.Fatal(err)
}

// 获取统计信息
stats, exists := manager.GetScriptStats("my_script")
if exists {
    fmt.Printf("执行次数: %d\n", stats.Executions)
    fmt.Printf("平均执行时间: %v\n", stats.AverageTime)
    fmt.Printf("成功率: %.2f%%\n", stats.SuccessRate)
}
```

#### 脚本监控

```go
// 添加告警规则
alertRule := &redis.AlertRule{
    Name:        "high_error_rate",
    ScriptName:  "my_script",
    Condition:   "error_rate",
    Threshold:   50.0,
    Duration:    1 * time.Minute,
    Enabled:     true,
}

manager.AddAlertRule(alertRule)

// 启动监控
manager.StartMonitor()
defer manager.StopMonitor()
```

#### 常用脚本模板

Redis模块提供了丰富的脚本模板，包括：

- **分布式锁**: `distributed_lock`, `distributed_unlock`
- **限流**: `rate_limit`
- **计数器**: `counter`, `atomic_increment`, `atomic_decrement`
- **比较并交换**: `compare_and_swap`
- **批量操作**: `batch_set`, `batch_get`
- **列表操作**: `atomic_list_push`, `atomic_list_pop`
- **集合操作**: `atomic_set_add`, `atomic_set_remove`
- **哈希操作**: `atomic_hash_set`, `atomic_hash_get`
- **有序集合操作**: `atomic_zset_add`, `atomic_zset_remove`, `atomic_zset_increment`

### 管道操作

```go
pipeline := client.NewPipeline()

// 添加命令到管道
pipeline.Set(ctx, "key1", "value1", 0)
pipeline.Set(ctx, "key2", "value2", 0)
pipeline.Get(ctx, "key1")
pipeline.Get(ctx, "key2")

// 执行管道
cmds, err := pipeline.Exec(ctx)
if err != nil {
    log.Fatal(err)
}

// 处理结果
for i, cmd := range cmds {
    if cmd.Err() != nil {
        log.Printf("Command %d failed: %v", i, cmd.Err())
    }
}
```

### 发布订阅

```go
// 发布者
publisher := client.NewPublisher()
_, err := publisher.Publish(ctx, "channel", "Hello World!")

// 订阅者
subscriber := client.NewSubscriber()
err = subscriber.Subscribe(ctx, "channel")

// 监听消息
err = subscriber.Listen(ctx, func(msg *redis.Message) {
    log.Printf("Received: %s", msg.Payload)
})

// 关闭订阅
subscriber.Close()
```

### 事务操作

```go
// 使用事务
err := client.WithTransaction(ctx, func(tx *redis.Tx) error {
    tx.Set(ctx, "key1", "value1", 0)
    tx.Set(ctx, "key2", "value2", 0)
    tx.Set(ctx, "key3", "value3", 0)
    return nil
}, "key1", "key2", "key3")

// 带选项的事务
opts := &redis.TransactionOptions{
    MaxRetries: 3,
    Timeout:    5 * time.Second,
}

err := client.WithTransactionOptions(ctx, func(tx *redis.Tx) error {
    // 事务逻辑
    return nil
}, []string{"key1", "key2"}, opts)
```

## 实际应用示例

### 缓存实现

```go
func GetUserFromCache(ctx context.Context, client *redis.Client, userID string) (*User, error) {
    stringOps := client.NewString()
    
    cacheKey := fmt.Sprintf("user:%s", userID)
    
    // 尝试从缓存获取
    cached, err := stringOps.Get(ctx, cacheKey)
    if err == nil {
        var user User
        err = json.Unmarshal([]byte(cached), &user)
        if err == nil {
            return &user, nil
        }
    }
    
    // 缓存未命中，从数据库获取
    user, err := getUserFromDB(userID)
    if err != nil {
        return nil, err
    }
    
    // 存储到缓存
    userJSON, _ := json.Marshal(user)
    stringOps.Set(ctx, cacheKey, userJSON, 30*time.Minute)
    
    return user, nil
}
```

### 分布式锁

```go
func AcquireLock(ctx context.Context, client *redis.Client, resource string, timeout time.Duration) (bool, error) {
    stringOps := client.NewString()
    
    lockKey := fmt.Sprintf("lock:%s", resource)
    lockValue := generateLockValue()
    
    success, err := stringOps.SetNX(ctx, lockKey, lockValue, timeout)
    if err != nil {
        return false, err
    }
    
    return success, nil
}

func ReleaseLock(ctx context.Context, client *redis.Client, resource string, lockValue string) error {
    // 使用 Lua 脚本确保原子性
    script := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `
    
    _, err := client.GetClient().Eval(ctx, script, []string{fmt.Sprintf("lock:%s", resource)}, lockValue).Result()
    return err
}
```

### 限流器

```go
func RateLimit(ctx context.Context, client *redis.Client, key string, limit int, window time.Duration) (bool, error) {
    stringOps := client.NewString()
    
    windowKey := fmt.Sprintf("%s:%d", key, time.Now().Unix()/int64(window.Seconds()))
    
    // 检查当前计数
    current, err := stringOps.GetInt(ctx, windowKey)
    if err != nil && err != redis.Nil {
        return false, err
    }
    
    if current >= int64(limit) {
        return false, nil // 限流
    }
    
    // 增加计数
    newCount, err := stringOps.Incr(ctx, windowKey)
    if err != nil {
        return false, err
    }
    
    // 设置过期时间
    if newCount == 1 {
        client.Expire(ctx, windowKey, window)
    }
    
    return true, nil
}
```

## 测试

运行测试：

```bash
# 运行所有测试
go test ./redis -v

# 运行特定测试
go test ./redis -run TestString -v

# 运行测试并显示覆盖率
go test ./redis -cover -v
```

## 示例代码

查看 `examples/` 目录下的示例代码：

- `basic_example.go` - 基本使用示例
- `advanced_example.go` - 高级功能示例
- `lua_example.go` - Lua脚本使用示例

运行示例：

```bash
# 基本示例
go run examples/basic_example.go

# 高级示例
go run examples/advanced_example.go

# Lua脚本示例
go run examples/lua_example.go
```

## 性能优化

### 连接池配置

```go
config := &redis.Config{
    Addr:            "localhost:6379",
    PoolSize:        20,                    // 增加连接池大小
    MinIdleConns:    10,                    // 保持更多空闲连接
    ConnMaxIdleTime: 30 * time.Minute,     // 连接空闲时间
}
```

### 管道优化

```go
// 批量操作使用管道
pipeline := client.NewPipeline()
for i := 0; i < 1000; i++ {
    pipeline.Set(ctx, fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i), 0)
}
cmds, err := pipeline.Exec(ctx)
```

### 缓冲区配置

```go
config := &redis.Config{
    Addr:            "localhost:6379",
    ReadBufferSize:  1024 * 1024,  // 1MB 读缓冲区
    WriteBufferSize: 1024 * 1024,  // 1MB 写缓冲区
}
```

## 错误处理

```go
value, err := stringOps.Get(ctx, "key")
if err != nil {
    if err == redis.Nil {
        // 键不存在
        log.Println("Key not found")
    } else {
        // 其他错误
        log.Printf("Redis error: %v", err)
    }
    return
}
```

## 监控和调试

### 连接池统计

```go
stats := client.GetStats()
log.Printf("Pool stats: %+v", stats)
```

### 健康检查

```go
err := client.Ping(ctx)
if err != nil {
    log.Printf("Redis health check failed: %v", err)
}
```

## 最佳实践

1. **连接管理**: 使用连接池，避免频繁创建和关闭连接
2. **错误处理**: 始终检查错误，特别是 `redis.Nil` 错误
3. **超时设置**: 为所有操作设置合理的超时时间
4. **管道使用**: 批量操作使用管道提高性能
5. **键命名**: 使用有意义的键命名规范
6. **过期时间**: 为缓存数据设置合理的过期时间
7. **监控**: 监控连接池状态和 Redis 性能指标

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 更新日志

### v1.0.0
- 初始版本发布
- 支持所有 Redis 数据类型操作
- 支持管道、发布订阅、事务
- 完整的测试覆盖
- 详细的文档和示例
