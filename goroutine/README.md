# 协程管理模块

一个强大的Go协程管理库，提供协程安全执行、崩溃恢复、协程池、监控和统计等功能，确保单个协程的崩溃不会影响整个服务的运行。

## 特性

- 🛡️ **协程安全执行**: 自动捕获和恢复协程panic，防止服务崩溃
- 📊 **详细统计**: 提供协程数量、执行时间、崩溃次数等统计信息
- 🏊 **协程池管理**: 高效的协程池，支持任务队列和负载均衡
- 📈 **实时监控**: 监控协程数量、内存使用、GC等系统指标
- 🔄 **重试机制**: 支持多种退避策略的重试机制
- ⚡ **熔断器**: 防止级联故障的熔断器模式
- 🚦 **限流器**: 控制协程执行频率的限流器
- 📦 **批量执行**: 高效的批量任务执行
- 🎯 **优雅关闭**: 支持优雅关闭和资源清理

## 安装

```bash
go get github.com/daxiong0327/tool-kit/goroutine
```

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "time"
    "github.com/daxiong0327/tool-kit/goroutine"
)

func main() {
    // 创建协程执行器
    executor := goroutine.NewSafeExecutor()
    
    // 安全启动协程
    executor.Go(func() {
        fmt.Println("Hello from safe goroutine!")
    })
    
    // 带上下文的协程
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    executor.GoWithContext(ctx, func() {
        fmt.Println("Hello from context goroutine!")
    })
    
    time.Sleep(1 * time.Second)
}
```

### 全局使用

```go
package main

import (
    "github.com/daxiong0327/tool-kit/goroutine"
)

func main() {
    // 使用全局执行器
    goroutine.Go(func() {
        fmt.Println("Hello from global goroutine!")
    })
    
    // 设置全局崩溃恢复处理器
    goroutine.SetGlobalRecoverHandler(func(panicValue interface{}, stack []byte, goroutineID string) {
        log.Printf("Goroutine panic recovered: %v", panicValue)
        log.Printf("Goroutine ID: %s", goroutineID)
        log.Printf("Stack trace:\n%s", string(stack))
    })
}
```

## 协程安全执行器

### 基本功能

```go
executor := goroutine.NewSafeExecutor()

// 基本协程执行
executor.Go(func() {
    // 你的代码
})

// 带上下文的协程执行
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
executor.GoWithContext(ctx, func() {
    // 你的代码
})

// 带超时的协程执行
executor.GoWithTimeout(func() {
    // 你的代码
}, 3*time.Second)

// 延迟执行
executor.GoWithDelay(func() {
    // 你的代码
}, 1*time.Second)

// 定时执行
stop := executor.GoWithInterval(func() {
    // 你的代码
}, 2*time.Second)

// 停止定时执行
close(stop)
```

### 崩溃恢复

```go
executor := goroutine.NewSafeExecutor()

// 设置崩溃恢复处理器
executor.SetRecoverHandler(func(panicValue interface{}, stack []byte, goroutineID string) {
    log.Printf("Goroutine panic recovered: %v", panicValue)
    log.Printf("Goroutine ID: %s", goroutineID)
    log.Printf("Stack trace:\n%s", string(stack))
    
    // 可以在这里添加告警、日志记录等逻辑
})

// 启动可能崩溃的协程
executor.Go(func() {
    panic("Something went wrong!")
    // 这行代码不会执行，但服务不会崩溃
})

// 其他协程继续正常运行
executor.Go(func() {
    fmt.Println("This goroutine continues to run normally")
})
```

### 统计信息

```go
executor := goroutine.NewSafeExecutor()

// 启动一些协程
for i := 0; i < 10; i++ {
    executor.Go(func() {
        time.Sleep(100 * time.Millisecond)
    })
}

time.Sleep(200 * time.Millisecond)

// 获取统计信息
stats := executor.GetStats()
fmt.Printf("Total goroutines: %d\n", stats.TotalGoroutines)
fmt.Printf("Active goroutines: %d\n", stats.ActiveGoroutines)
fmt.Printf("Completed goroutines: %d\n", stats.CompletedCount)
fmt.Printf("Panic count: %d\n", stats.PanicCount)
```

## 协程池

### 基本使用

```go
// 创建协程池
config := &goroutine.PoolConfig{
    MaxWorkers: 10,        // 最大工作协程数
    QueueSize:  100,       // 任务队列大小
    JobTimeout: 5*time.Second, // 任务超时时间
}

pool := goroutine.NewPool(config)
defer pool.Stop()

// 提交任务
err := pool.SubmitFunc("task-1", func() error {
    fmt.Println("Executing task 1")
    return nil
})

if err != nil {
    log.Printf("Failed to submit task: %v", err)
}

// 提交带超时的任务
err = pool.SubmitWithTimeout("task-2", func() error {
    fmt.Println("Executing task 2")
    return nil
}, 2*time.Second)
```

### 自定义任务

```go
// 实现Job接口
type MyJob struct {
    ID   string
    Data interface{}
}

func (j *MyJob) Execute() error {
    fmt.Printf("Processing job %s with data %v\n", j.ID, j.Data)
    return nil
}

func (j *MyJob) GetID() string {
    return j.ID
}

func (j *MyJob) GetTimeout() time.Duration {
    return 5 * time.Second
}

// 提交自定义任务
job := &MyJob{
    ID:   "custom-job",
    Data: "some data",
}

err := pool.Submit(job)
```

### 协程池统计

```go
stats := pool.GetStats()
fmt.Printf("Total jobs: %d\n", stats.TotalJobs)
fmt.Printf("Completed jobs: %d\n", stats.CompletedJobs)
fmt.Printf("Failed jobs: %d\n", stats.FailedJobs)
fmt.Printf("Active workers: %d\n", stats.ActiveWorkers)
fmt.Printf("Queued jobs: %d\n", stats.QueuedJobs)
```

## 批量执行

```go
executor := goroutine.NewSafeExecutor()
batch := goroutine.NewBatch(executor)

// 添加任务
for i := 0; i < 5; i++ {
    i := i // 捕获循环变量
    batch.Add(func() (interface{}, error) {
        result := fmt.Sprintf("Result %d", i)
        return result, nil
    })
}

// 等待所有任务完成
results, errors := batch.Wait()

fmt.Printf("Successfully completed %d tasks\n", len(results))
fmt.Printf("Failed %d tasks\n", len(errors))

// 处理结果
for _, result := range results {
    fmt.Printf("Result: %+v\n", result)
}
```

## 重试机制

```go
executor := goroutine.NewSafeExecutor()

// 创建重试器
retry := goroutine.NewRetry(executor, 3, 100*time.Millisecond, &goroutine.ExponentialBackoff{
    BaseDelay: 100 * time.Millisecond,
    MaxDelay:  1 * time.Second,
})

// 执行会失败的任务
err := retry.Execute(func() error {
    // 模拟可能失败的操作
    if rand.Float64() < 0.7 {
        return fmt.Errorf("operation failed")
    }
    fmt.Println("Operation succeeded!")
    return nil
})

if err != nil {
    log.Printf("Retry failed: %v", err)
}
```

### 退避策略

```go
// 固定延迟
fixedBackoff := &goroutine.FixedBackoff{
    Delay: 1 * time.Second,
}

// 指数退避
exponentialBackoff := &goroutine.ExponentialBackoff{
    BaseDelay: 100 * time.Millisecond,
    MaxDelay:  10 * time.Second,
}

// 线性退避
linearBackoff := &goroutine.LinearBackoff{
    BaseDelay: 100 * time.Millisecond,
    MaxDelay:  5 * time.Second,
}
```

## 熔断器

```go
executor := goroutine.NewSafeExecutor()
cb := goroutine.NewCircuitBreaker(executor, 5, 30*time.Second)

// 使用熔断器执行可能失败的操作
err := cb.Execute(func() error {
    // 调用外部服务
    return callExternalService()
})

if err != nil {
    log.Printf("Operation failed: %v", err)
}

// 检查熔断器状态
state := cb.GetState()
switch state {
case goroutine.StateClosed:
    fmt.Println("Circuit breaker is closed (normal operation)")
case goroutine.StateOpen:
    fmt.Println("Circuit breaker is open (failing fast)")
case goroutine.StateHalfOpen:
    fmt.Println("Circuit breaker is half-open (testing)")
}
```

## 限流器

```go
executor := goroutine.NewSafeExecutor()
rl := goroutine.NewRateLimiter(executor, 10, 1*time.Second) // 每秒最多10个请求
defer rl.Stop()

// 尝试执行操作
for i := 0; i < 20; i++ {
    if rl.Allow() {
        fmt.Printf("Request %d allowed\n", i+1)
        // 执行操作
    } else {
        fmt.Printf("Request %d rate limited\n", i+1)
    }
    time.Sleep(100 * time.Millisecond)
}
```

## 协程监控

### 基本监控

```go
executor := goroutine.NewSafeExecutor()
monitor := goroutine.NewMonitor(executor, 1*time.Second)

// 启动监控
monitor.Start()
defer monitor.Stop()

// 添加监控处理器
monitor.AddHandler(func(stats *goroutine.MonitorStats) {
    fmt.Printf("Goroutines: %d, Heap: %d KB\n", 
        stats.NumGoroutines, stats.HeapAlloc/1024)
})

// 运行一些协程
for i := 0; i < 10; i++ {
    executor.Go(func() {
        time.Sleep(1 * time.Second)
    })
}

time.Sleep(5 * time.Second)
```

### 告警配置

```go
// 设置告警配置
monitor.SetAlertConfig(&goroutine.AlertConfig{
    MaxGoroutines: 1000,
    MaxHeapAlloc:  100 * 1024 * 1024, // 100MB
    AlertHandler: func(alert *goroutine.Alert) {
        log.Printf("Alert: %s - %s", alert.Type, alert.Message)
        // 发送告警通知
    },
})
```

### 获取详细统计

```go
stats := monitor.GetStats()
fmt.Printf("Goroutines: %d\n", stats.NumGoroutines)
fmt.Printf("Heap Alloc: %d KB\n", stats.HeapAlloc/1024)
fmt.Printf("GC Count: %d\n", stats.NumGC)
fmt.Printf("Last GC: %v\n", stats.LastGC)

// 获取协程profile
goroutineProfile := monitor.GetGoroutineProfile()
fmt.Printf("Goroutine profile: %d records\n", len(goroutineProfile))

// 获取内存profile
memProfile := monitor.GetMemProfile()
fmt.Printf("Memory profile: %d records\n", len(memProfile))
```

## 优雅关闭

```go
executor := goroutine.NewSafeExecutor()
pool := goroutine.NewPool(&goroutine.PoolConfig{
    MaxWorkers: 10,
    QueueSize:  100,
})

// 设置优雅关闭信号
shutdown := make(chan os.Signal, 1)
signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

// 启动优雅关闭处理器
executor.Go(func() {
    <-shutdown
    fmt.Println("Shutting down gracefully...")
    
    // 停止接收新任务
    pool.StopGracefully(30 * time.Second)
    
    fmt.Println("Shutdown complete")
})

// 正常业务逻辑
// ...

// 等待关闭信号
<-shutdown
```

## 配置选项

### SafeExecutor 配置

| 选项 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| RecoverHandler | RecoverHandler | nil | 崩溃恢复处理器 |
| Logger | Logger | DefaultLogger | 日志器 |

### PoolConfig 配置

| 选项 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| MaxWorkers | int | 10 | 最大工作协程数 |
| QueueSize | int | 1000 | 任务队列大小 |
| WorkerTimeout | time.Duration | 30m | 工作协程超时时间 |
| JobTimeout | time.Duration | 5m | 任务超时时间 |

### AlertConfig 配置

| 选项 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| MaxGoroutines | int | 1000 | 最大协程数告警阈值 |
| MaxHeapAlloc | uint64 | 100MB | 最大堆内存告警阈值 |
| MaxStackInuse | uint64 | 10MB | 最大栈内存告警阈值 |
| CheckInterval | time.Duration | 30s | 检查间隔 |
| AlertHandler | AlertHandler | nil | 告警处理器 |

## 最佳实践

### 1. 合理设置协程池大小

```go
// 根据CPU核心数设置
numCPU := runtime.NumCPU()
config := &goroutine.PoolConfig{
    MaxWorkers: numCPU * 2, // 通常设置为CPU核心数的2倍
    QueueSize:  1000,
}
```

### 2. 设置合适的超时时间

```go
// 根据业务需求设置超时
executor.GoWithTimeout(func() {
    // 数据库操作
}, 5*time.Second)

executor.GoWithTimeout(func() {
    // HTTP请求
}, 10*time.Second)
```

### 3. 使用熔断器保护外部服务

```go
// 为每个外部服务创建熔断器
userServiceCB := goroutine.NewCircuitBreaker(executor, 5, 30*time.Second)
orderServiceCB := goroutine.NewCircuitBreaker(executor, 3, 20*time.Second)

// 使用熔断器调用服务
err := userServiceCB.Execute(func() error {
    return callUserService()
})
```

### 4. 监控协程数量

```go
// 设置告警阈值
monitor.SetAlertConfig(&goroutine.AlertConfig{
    MaxGoroutines: 1000,
    AlertHandler: func(alert *goroutine.Alert) {
        // 发送告警通知
        sendAlert(alert)
    },
})
```

### 5. 优雅关闭

```go
// 监听系统信号
shutdown := make(chan os.Signal, 1)
signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

// 优雅关闭
go func() {
    <-shutdown
    pool.StopGracefully(30 * time.Second)
}()
```

## 示例

查看 `examples/` 目录中的完整示例：

- `basic_example.go` - 基本功能示例
- `advanced_example.go` - 高级功能示例

## 许可证

MIT License
