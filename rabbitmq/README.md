# RabbitMQ 工具包

这是一个用于简化 RabbitMQ 操作的 Go 语言工具包，提供了连接管理、生产者和消费者的封装实现。

## 功能特性

- 连接管理：自动重连、连接池、TLS 支持
- 生产者：支持确认模式、延迟消息、JSON 消息
- 消费者：支持自动/手动确认、预取设置
- 错误处理：统一的错误类型和处理机制

## 安装

```bash
go get github.com/daxiong/tool-kit/rabbitmq
```

## 快速开始

### 创建连接

```go
config := rabbitmq.Config{
    Host:                 "localhost",
    Port:                 5672,
    Username:             "guest",
    Password:             "guest",
    VHost:                "/",
    ReconnectDelay:       5 * time.Second,
    MaxReconnectAttempts: 10,
    Heartbeat:            10 * time.Second,
}

connManager, err := rabbitmq.NewConnectionManager(config)
if err != nil {
    log.Fatalf("创建连接管理器失败: %v", err)
}
defer connManager.Close()
```

### 生产者示例

```go
producerConfig := producer.Config{
    Exchange:     "test_exchange",
    ExchangeType: "direct",
    RoutingKey:   "test_key",
    Durable:      true,
    DeliveryMode: 2, // 持久化消息
    ContentType:  "application/json",
}

p, err := producer.New(connManager, producerConfig)
if err != nil {
    log.Fatalf("创建生产者失败: %v", err)
}
defer p.Close()

// 发送JSON消息
msg := map[string]interface{}{
    "message": "这是一条测试消息",
    "time":    time.Now().Format(time.RFC3339),
}

ctx := context.Background()
err = p.PublishJSON(ctx, msg)
if err != nil {
    log.Fatalf("发送消息失败: %v", err)
}
```

### 消费者示例

```go
consumerConfig := consumer.Config{
    Queue:          "test_queue",
    Exchange:       "test_exchange",
    ExchangeType:   "direct",
    RoutingKey:     "test_key",
    QueueDurable:   true,
    ExchangeDurable: true,
    AutoAck:        false,
    PrefetchCount:  1,
}

c, err := consumer.New(connManager, consumerConfig)
if err != nil {
    log.Fatalf("创建消费者失败: %v", err)
}
defer c.Close()

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

deliveries, err := c.Consume(ctx)
if err != nil {
    log.Fatalf("开始消费失败: %v", err)
}

for delivery := range deliveries {
    // 处理消息
    fmt.Printf("收到消息: %s\n", string(delivery.Body))
    
    // 手动确认消息
    err := delivery.Ack(false)
    if err != nil {
        log.Printf("确认消息失败: %v", err)
    }
}
```

## 运行示例

1. 启动 RabbitMQ 服务

```bash
cd examples
docker-compose up -d
```

2. 运行消费者

```bash
go run examples/consumer_example.go
```

3. 在另一个终端运行生产者

```bash
go run examples/producer_example.go
```

## 高级功能

### 延迟消息

```go
// 发送5秒后处理的延迟消息
err = p.PublishWithDelay(ctx, msgBytes, headers, 5*time.Second)
```

### TLS 连接

```go
config := rabbitmq.Config{
    Host:      "localhost",
    Port:      5671, // TLS端口通常是5671
    Username:  "guest",
    Password:  "guest",
    EnableTLS: true,
    TLSConfig: &rabbitmq.TLSConfig{
        CertFile:   "/path/to/client.crt",
        KeyFile:    "/path/to/client.key",
        CACertFile: "/path/to/ca.crt",
        Verify:     true,
    },
}
```

## 注意事项

- 确保在应用程序退出前正确关闭连接和通道
- 对于生产环境，建议设置适当的重连策略
- 处理消息时注意异常情况的处理
