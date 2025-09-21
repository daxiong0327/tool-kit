# Log 模块

基于 [zap](https://github.com/uber-go/zap) 的高性能日志模块，提供简单易用的日志接口。

## 特性

- 🚀 高性能：基于 zap 实现，性能优异
- 🎨 多种格式：支持 JSON 和 Console 格式
- 📁 多种输出：支持标准输出和文件输出
- 🔧 灵活配置：支持自定义日志级别、文件轮转等
- 🧪 易于测试：提供测试友好的接口
- 📊 结构化日志：支持字段和错误记录

## 安装

```bash
go get github.com/daxiong0327/tool-kit/log
```

## 快速开始

### 基本使用

```go
package main

import (
    "github.com/daxiong0327/tool-kit/log"
)

func main() {
    // 使用默认配置
    logger, err := log.New(nil)
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    logger.Info("Hello, World!")
    logger.Infof("User %s logged in", "alice")
}
```

### 自定义配置

```go
package main

import (
    "github.com/daxiong0327/tool-kit/log"
)

func main() {
    config := &log.Config{
        Level:  "debug",
        Format: "console",
        Output: "stdout",
        File:   "logs/app.log",
    }

    logger, err := log.New(config)
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    logger.Debug("Debug message")
    logger.Info("Info message")
    logger.Warn("Warning message")
    logger.Error("Error message")
}
```

### 结构化日志

```go
package main

import (
    "errors"
    "github.com/daxiong0327/tool-kit/log"
)

func main() {
    logger, _ := log.New(nil)
    defer logger.Sync()

    // 添加字段
    logger.WithField("user_id", "123").Info("User action")
    
    // 添加多个字段
    logger.WithFields(map[string]interface{}{
        "user_id": "123",
        "action":  "login",
        "ip":      "192.168.1.1",
    }).Info("User logged in")

    // 添加错误
    err := errors.New("something went wrong")
    logger.WithError(err).Error("Operation failed")
}
```

### 预设配置

```go
package main

import (
    "github.com/daxiong0327/tool-kit/log"
)

func main() {
    // 开发环境
    devLogger, _ := log.NewDevelopment()
    devLogger.Debug("Debug message")

    // 生产环境
    prodLogger, _ := log.NewProduction()
    prodLogger.Info("Production message")

    // 测试环境
    testLogger, _ := log.NewTest()
    testLogger.Error("Test error")
}
```

### 多输出支持

```go
package main

import (
    "github.com/daxiong0327/tool-kit/log"
)

func main() {
    // 方式1：使用NewMultiOutput同时输出到多个地方
    outputs := []log.OutputConfig{
        {Type: "stdout"},
        {Type: "file", File: "logs/app.log"},
        {Type: "file", File: "logs/error.log"},
    }
    
    logger, err := log.NewMultiOutput("info", "json", outputs...)
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    logger.Info("这条日志会同时输出到控制台和两个文件")
    logger.Error("错误日志也会同时输出到所有地方")
}
```

### 便捷的多输出方法

```go
package main

import (
    "github.com/daxiong0327/tool-kit/log"
)

func main() {
    // 方式2：使用NewFileAndConsole同时输出到文件和控制台
    logger, err := log.NewFileAndConsole("info", "console", "logs/app.log")
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    logger.Info("这条日志会同时显示在控制台和保存到文件")
}
```

### 日志轮转功能

```go
package main

import (
    "github.com/daxiong0327/tool-kit/log"
)

func main() {
    // 方式1：创建带轮转功能的文件日志
    // 参数：级别, 格式, 文件路径, 最大大小(MB), 最大备份数, 最大保存天数
    logger, err := log.NewRotatingFile("info", "json", "logs/app.log", 10, 5, 30)
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    logger.Info("这条日志会保存到带时间戳的文件中")
    // 文件命名格式: app_2025.09.21_20:28:54.685.log
}
```

### 轮转文件 + 控制台输出

```go
package main

import (
    "github.com/daxiong0327/tool-kit/log"
)

func main() {
    // 同时输出到轮转文件和控制台
    logger, err := log.NewRotatingFileAndConsole("debug", "console", "logs/app.log", 5, 3, 7)
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    logger.Debug("调试信息")
    logger.Info("应用信息")
}
```

### 自定义轮转配置

```go
package main

import (
    "github.com/daxiong0327/tool-kit/log"
)

func main() {
    outputs := []log.OutputConfig{
        {Type: "stdout"},
        {
            Type:        "file",
            File:        "logs/app.log",
            MaxSize:     10,  // 10MB
            MaxBackups:  5,   // 保留5个备份
            MaxAge:      30,  // 保留30天
            Compress:    true,
            UseRotation: true,
            TimeFormat:  "2006.01.02_15:04:05.000", // 自定义时间格式
        },
    }
    
    logger, err := log.NewMultiOutput("info", "json", outputs...)
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    logger.Info("自定义轮转配置的日志")
}
```

### 自定义输出

```go
package main

import (
    "bytes"
    "github.com/daxiong0327/tool-kit/log"
)

func main() {
    var buf bytes.Buffer
    
    logger, err := log.NewWithWriter("info", "json", &buf)
    if err != nil {
        panic(err)
    }

    logger.Info("This will be written to buffer")
    fmt.Println(buf.String())
}
```

## 配置选项

### 基本配置

| 选项 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| Level | string | "info" | 日志级别: debug, info, warn, error, fatal |
| Format | string | "json" | 日志格式: json, console |
| Outputs | []OutputConfig | stdout | 输出配置列表（推荐使用） |

### 输出配置 (OutputConfig)

| 选项 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| Type | string | "stdout" | 输出类型: stdout, file |
| File | string | "" | 文件路径（当Type为file时） |
| MaxSize | int | 100 | 日志文件最大大小(MB) |
| MaxBackups | int | 3 | 最大备份文件数 |
| MaxAge | int | 7 | 最大保存天数 |
| Compress | bool | true | 是否压缩备份文件 |
| TimeFormat | string | "2006.01.02_15:04:05.000" | 时间格式，用于文件名 |
| UseRotation | bool | false | 是否启用日志轮转 |

### 向后兼容配置（已废弃）

| 选项 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| Output | string | "stdout" | 输出方式: stdout, file (已废弃，使用Outputs) |
| File | string | "logs/app.log" | 日志文件路径 (已废弃，使用Outputs) |
| MaxSize | int | 100 | 日志文件最大大小(MB) (已废弃，使用Outputs) |
| MaxBackups | int | 3 | 最大备份文件数 (已废弃，使用Outputs) |
| MaxAge | int | 7 | 最大保存天数 (已废弃，使用Outputs) |
| Compress | bool | true | 是否压缩备份文件 (已废弃，使用Outputs) |

## API 参考

### Logger 接口

```go
type Logger interface {
    Debug(args ...interface{})
    Debugf(format string, args ...interface{})
    Info(args ...interface{})
    Infof(format string, args ...interface{})
    Warn(args ...interface{})
    Warnf(format string, args ...interface{})
    Error(args ...interface{})
    Errorf(format string, args ...interface{})
    Fatal(args ...interface{})
    Fatalf(format string, args ...interface{})
    WithField(key string, value interface{}) Logger
    WithFields(fields map[string]interface{}) Logger
    WithError(err error) Logger
    Sync() error
}
```

### 构造函数

- `New(config *Config) (Logger, error)` - 使用配置创建日志器
- `NewWithWriter(level, format string, writer io.Writer) (Logger, error)` - 使用自定义输出创建日志器
- `NewMultiOutput(level, format string, outputs ...OutputConfig) (Logger, error)` - 创建多输出日志器
- `NewFileAndConsole(level, format, filePath string) (Logger, error)` - 创建同时输出到文件和控制台的日志器
- `NewRotatingFile(level, format, filePath string, maxSize, maxBackups, maxAge int) (Logger, error)` - 创建带轮转功能的文件日志器
- `NewRotatingFileAndConsole(level, format, filePath string, maxSize, maxBackups, maxAge int) (Logger, error)` - 创建同时输出到轮转文件和控制台的日志器
- `NewDevelopment() (Logger, error)` - 创建开发环境日志器
- `NewProduction() (Logger, error)` - 创建生产环境日志器
- `NewTest() (Logger, error)` - 创建测试环境日志器

### 多输出相关

- `NewMultiWriter(writers ...io.Writer) *MultiWriter` - 创建多输出写入器
- `MultiWriter.AddWriter(w io.Writer)` - 动态添加写入器

### 轮转相关

- `GenerateTimeBasedFileName(basePath, timeFormat string) string` - 生成基于时间的文件名

## 完整使用示例

```go
package main

import (
    "github.com/daxiong0327/tool-kit/log"
)

func main() {
    // 创建多输出日志器：同时输出到控制台和文件
    logger, err := log.NewFileAndConsole("info", "json", "logs/app.log")
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    // 基本日志记录
    logger.Info("应用启动")
    logger.Infof("用户 %s 登录", "alice")

    // 结构化日志
    logger.WithFields(map[string]interface{}{
        "user_id": "123",
        "action":  "login",
        "ip":      "192.168.1.1",
    }).Info("用户登录成功")

    // 错误日志
    err = someOperation()
    if err != nil {
        logger.WithError(err).Error("操作失败")
    }

    // 高级多输出配置
    advancedLogger, err := log.NewMultiOutput("debug", "console", 
        log.OutputConfig{Type: "stdout"},
        log.OutputConfig{Type: "file", File: "logs/debug.log"},
        log.OutputConfig{Type: "file", File: "logs/error.log"},
    )
    if err != nil {
        panic(err)
    }
    defer advancedLogger.Sync()

    advancedLogger.Debug("调试信息")
    advancedLogger.Error("错误信息")
}
```

## 性能

基于 zap 的高性能实现，在大多数场景下性能优异：

- 零分配 JSON 编码
- 结构化日志记录
- 异步日志写入
- 内存池优化
- 多输出支持（性能影响极小）

## 许可证

MIT License
