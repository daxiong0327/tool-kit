# HTTP 模块

基于 [req](https://github.com/imroc/req) 的高性能 HTTP 客户端模块，提供简单易用的 HTTP 请求接口。

## 特性

- 🚀 高性能：基于 req 实现，性能优异
- 🎯 简单易用：提供简洁的 API 接口
- 🔧 灵活配置：支持超时、重试、代理等配置
- 📦 多种请求：支持 GET、POST、PUT、DELETE、PATCH
- 🎨 JSON 支持：内置 JSON 序列化和反序列化
- 🛡️ 错误处理：完善的错误处理机制
- 🧪 易于测试：提供测试友好的接口

## 安装

```bash
go get github.com/daxiong0327/tool-kit/http
```

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "github.com/daxiong0327/tool-kit/http"
)

func main() {
    // 使用默认配置
    client := http.New(nil)
    
    ctx := context.Background()
    
    // GET 请求
    resp, err := client.Get(ctx, "https://api.example.com/users")
    if err != nil {
        panic(err)
    }
    
    println(resp.Text)
}
```

### 自定义配置

```go
package main

import (
    "context"
    "time"
    "github.com/daxiong0327/tool-kit/http"
)

func main() {
    config := &http.Config{
        BaseURL:    "https://api.example.com",
        Timeout:    30 * time.Second,
        RetryCount: 3,
        RetryDelay: 1 * time.Second,
        Headers: map[string]string{
            "User-Agent": "MyApp/1.0",
        },
    }

    client := http.New(config)
    
    ctx := context.Background()
    resp, err := client.Get(ctx, "/users")
    if err != nil {
        panic(err)
    }
    
    println(resp.Text)
}
```

### JSON 请求和响应

```go
package main

import (
    "context"
    "github.com/daxiong0327/tool-kit/http"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    client := http.New(&http.Config{
        BaseURL: "https://api.example.com",
    })
    
    ctx := context.Background()
    
    // POST JSON 请求
    user := User{Name: "Alice"}
    var result User
    err := client.PostJSON(ctx, "/users", user, &result)
    if err != nil {
        panic(err)
    }
    
    println(result.Name)
}
```

### 请求选项

```go
package main

import (
    "context"
    "github.com/daxiong0327/tool-kit/http"
)

func main() {
    client := http.New(nil)
    ctx := context.Background()
    
    // 使用选项
    resp, err := client.Get(ctx, "https://api.example.com/users",
        http.WithHeader("Authorization", "Bearer token"),
        http.WithQuery("page", "1"),
        http.WithQuery("limit", "10"),
        http.WithTimeout(10*time.Second),
    )
    if err != nil {
        panic(err)
    }
    
    println(resp.Text)
}
```

### 认证

```go
package main

import (
    "context"
    "github.com/daxiong0327/tool-kit/http"
)

func main() {
    client := http.New(nil)
    ctx := context.Background()
    
    // Basic 认证
    resp, err := client.Get(ctx, "https://api.example.com/protected",
        http.WithBasicAuth("username", "password"),
    )
    
    // Bearer Token 认证
    resp, err = client.Get(ctx, "https://api.example.com/protected",
        http.WithBearerToken("your-token"),
    )
}
```

### 错误处理

```go
package main

import (
    "context"
    "fmt"
    "github.com/daxiong0327/tool-kit/http"
)

func main() {
    client := http.New(nil)
    ctx := context.Background()
    
    resp, err := client.Get(ctx, "https://api.example.com/users")
    if err != nil {
        fmt.Printf("Request failed: %v\n", err)
        return
    }
    
    if resp.StatusCode >= 400 {
        fmt.Printf("HTTP error: %d %s\n", resp.StatusCode, resp.Text)
        return
    }
    
    fmt.Println("Success:", resp.Text)
}
```

## 配置选项

### 基本配置

| 选项 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| BaseURL | string | "" | 基础 URL |
| Timeout | time.Duration | 30s | 请求超时时间 |
| Headers | map[string]string | {} | 默认请求头 |
| UserAgent | string | "tool-kit-http-client/1.0.0" | User-Agent |
| Proxy | string | "" | 代理地址 |
| Insecure | bool | false | 是否跳过 SSL 验证 |
| Debug | bool | false | 是否开启调试模式 |

### 重试配置 (RetryConfig)

| 选项 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| MaxRetries | int | 3 | 最大重试次数 |
| BaseDelay | time.Duration | 1s | 基础延迟时间 |
| MaxDelay | time.Duration | 30s | 最大延迟时间 |
| Strategy | RetryStrategy | RetryStrategyExponential | 重试策略 |
| RetryableCodes | []int | [500,502,503,504,408,429] | 可重试的状态码 |

### 连接池配置 (PoolConfig)

| 选项 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| MaxIdleConns | int | 100 | 最大空闲连接数 |
| MaxIdleConnsPerHost | int | 10 | 每个主机的最大空闲连接数 |
| MaxConnsPerHost | int | 0 | 每个主机的最大连接数（0=无限制） |
| IdleConnTimeout | time.Duration | 90s | 空闲连接超时时间 |
| DisableKeepAlives | bool | false | 是否禁用Keep-Alive |

### 重试策略 (RetryStrategy)

| 策略 | 描述 | 延迟计算 |
|------|------|----------|
| RetryStrategyFixed | 固定延迟 | BaseDelay |
| RetryStrategyLinear | 线性增长 | BaseDelay × (attempt + 1) |
| RetryStrategyExponential | 指数退避 | BaseDelay × 2^attempt |

## API 参考

### 请求方法

```go
// 基本请求
Get(ctx, url, options...)
Post(ctx, url, body, options...)
Put(ctx, url, body, options...)
Delete(ctx, url, options...)
Patch(ctx, url, body, options...)

// JSON 请求
GetJSON(ctx, url, result, options...)
PostJSON(ctx, url, body, result, options...)
PutJSON(ctx, url, body, result, options...)
DeleteJSON(ctx, url, result, options...)
PatchJSON(ctx, url, body, result, options...)
```

### 请求选项

```go
WithHeader(key, value string) Option
WithHeaders(headers map[string]string) Option
WithQuery(key, value string) Option
WithQueries(queries map[string]string) Option
WithAuth(authType, token string) Option
WithBasicAuth(username, password string) Option
WithBearerToken(token string) Option
WithTimeout(timeout time.Duration) Option
WithRetry(count int, delay time.Duration) Option
```

### 响应结构

```go
type Response struct {
    StatusCode int               `json:"status_code"`
    Headers    map[string]string `json:"headers"`
    Body       []byte            `json:"body"`
    Text       string            `json:"text"`
}
```

## 高级用法

### 重试机制

```go
package main

import (
    "context"
    "time"
    "github.com/daxiong0327/tool-kit/http"
)

func main() {
    // 创建带重试配置的客户端
    retryConfig := &http.RetryConfig{
        MaxRetries:     3,                    // 最大重试次数
        BaseDelay:      1 * time.Second,      // 基础延迟时间
        MaxDelay:       30 * time.Second,     // 最大延迟时间
        Strategy:       http.RetryStrategyExponential, // 重试策略
        RetryableCodes: []int{500, 502, 503, 504, 408, 429}, // 可重试的状态码
    }
    
    client := http.NewWithRetry("https://api.example.com", retryConfig)
    ctx := context.Background()
    
    // 请求会自动重试
    resp, err := client.Get(ctx, "/users")
    if err != nil {
        panic(err)
    }
    
    println(resp.Text)
}
```

### 连接池配置

```go
package main

import (
    "time"
    "github.com/daxiong0327/tool-kit/http"
)

func main() {
    // 创建带连接池配置的客户端
    poolConfig := &http.PoolConfig{
        MaxIdleConns:        100,              // 最大空闲连接数
        MaxIdleConnsPerHost: 10,               // 每个主机的最大空闲连接数
        MaxConnsPerHost:     20,               // 每个主机的最大连接数
        IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
        DisableKeepAlives:   false,            // 是否禁用Keep-Alive
    }
    
    client := http.NewWithPool("https://api.example.com", poolConfig)
    
    // 使用连接池进行多次请求
    for i := 0; i < 100; i++ {
        resp, err := client.Get(ctx, "/users")
        // 处理响应...
    }
}
```

### 重试策略

```go
package main

import (
    "github.com/daxiong0327/tool-kit/http"
)

func main() {
    // 固定延迟重试
    fixedRetry := &http.RetryConfig{
        Strategy: http.RetryStrategyFixed,
        BaseDelay: 1 * time.Second,
    }
    
    // 线性增长重试
    linearRetry := &http.RetryConfig{
        Strategy: http.RetryStrategyLinear,
        BaseDelay: 1 * time.Second,
    }
    
    // 指数退避重试（推荐）
    exponentialRetry := &http.RetryConfig{
        Strategy: http.RetryStrategyExponential,
        BaseDelay: 1 * time.Second,
        MaxDelay: 30 * time.Second,
    }
    
    client1 := http.NewWithRetry("", fixedRetry)
    client2 := http.NewWithRetry("", linearRetry)
    client3 := http.NewWithRetry("", exponentialRetry)
}
```

### 组合配置

```go
package main

import (
    "github.com/daxiong0327/tool-kit/http"
)

func main() {
    // 重试 + 连接池组合配置
    retryConfig := &http.RetryConfig{
        MaxRetries:     3,
        BaseDelay:      1 * time.Second,
        Strategy:       http.RetryStrategyExponential,
        RetryableCodes: []int{500, 502, 503, 504},
    }
    
    poolConfig := &http.PoolConfig{
        MaxIdleConns:        50,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     60 * time.Second,
    }
    
    client := http.NewWithRetryAndPool("https://api.example.com", retryConfig, poolConfig)
}
```

### 自定义客户端配置

```go
package main

import (
    "github.com/daxiong0327/tool-kit/http"
)

func main() {
    client := http.New(nil)
    
    // 动态配置
    client.SetBaseURL("https://api.example.com")
    client.SetTimeout(60 * time.Second)
    client.SetHeader("X-API-Key", "your-api-key")
    
    // 重试配置
    retryConfig := &http.RetryConfig{
        MaxRetries: 5,
        BaseDelay:  2 * time.Second,
        Strategy:   http.RetryStrategyExponential,
    }
    client.SetRetryConfig(retryConfig)
    
    // 连接池配置
    poolConfig := &http.PoolConfig{
        MaxIdleConns: 100,
        MaxIdleConnsPerHost: 20,
    }
    client.SetPoolConfig(poolConfig)
}
```

### 获取底层客户端

```go
package main

import (
    "github.com/daxiong0327/tool-kit/http"
    "github.com/imroc/req/v3"
)

func main() {
    client := http.New(nil)
    
    // 获取底层 req 客户端进行高级操作
    reqClient := client.GetClient()
    reqClient.EnableDumpAll()
    reqClient.SetCommonRetryCount(3)
}
```

## 性能

基于 req 的高性能实现：

- 连接池复用
- 自动重试机制
- 请求/响应压缩
- HTTP/2 支持
- 零拷贝优化

## 许可证

MIT License
