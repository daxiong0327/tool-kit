# Tool Kit

一个 Go 语言工具库集合，提供常用的工具模块，帮助开发者快速构建应用程序。

## 模块列表

### 📝 [Log 模块](./log/)
基于 [zap](https://github.com/uber-go/zap) 的高性能日志模块，提供简单易用的日志接口。

**特性：**
- 🚀 高性能：基于 zap 实现，性能优异
- 🎨 多种格式：支持 JSON 和 Console 格式
- 📁 多种输出：支持标准输出和文件输出
- 🔧 灵活配置：支持自定义日志级别、文件轮转等
- 🧪 易于测试：提供测试友好的接口

### 🌐 [HTTP 模块](./http/)
基于 [req](https://github.com/imroc/req) 的高性能 HTTP 客户端模块，提供简单易用的 HTTP 请求接口。

**特性：**
- 🚀 高性能：基于 req 实现，性能优异
- 🎯 简单易用：提供简洁的 API 接口
- 🔧 灵活配置：支持超时、重试、代理等配置
- 📦 多种请求：支持 GET、POST、PUT、DELETE、PATCH
- 🎨 JSON 支持：内置 JSON 序列化和反序列化

### 🐰 [RabbitMQ 模块](./rabbitmq/)
基于 [amqp091-go](https://github.com/rabbitmq/amqp091-go) 的 RabbitMQ 客户端模块，提供消息队列功能。

**特性：**
- 📨 消息发布和消费
- 🔄 自动重连机制
- 🛡️ 错误处理和重试
- 📊 连接池管理
- 🧪 完整的示例代码

### ⚡ [SingleFlight 模块](./singleflight/)
基于 [golang.org/x/sync](https://pkg.go.dev/golang.org/x/sync) 的单次执行模块，防止重复请求。

**特性：**
- 🚫 防止重复请求
- 💾 缓存结果
- ⚡ 高性能
- 🔧 易于集成

## 快速开始

### 安装

```bash
go get github.com/daxiong0327/tool-kit
```

### 使用示例

```go
package main

import (
    "context"
    "github.com/daxiong0327/tool-kit/log"
    "github.com/daxiong0327/tool-kit/http"
)

func main() {
    // 创建日志器
    logger, _ := log.New(nil)
    defer logger.Sync()
    
    // 创建 HTTP 客户端
    client := http.New(&http.Config{
        BaseURL: "https://api.example.com",
    })
    
    // 发送请求
    ctx := context.Background()
    resp, err := client.Get(ctx, "/users")
    if err != nil {
        logger.Error("Request failed", err)
        return
    }
    
    logger.Info("Response received", "status", resp.StatusCode)
}
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License