package main

import (
	"fmt"
	"time"

	"github.com/daxiong0327/tool-kit/log"
)

func main() {
	fmt.Println("=== Log 轮转功能演示 ===")

	// 示例1：创建带轮转功能的文件日志
	fmt.Println("\n1. 创建轮转文件日志:")
	logger1, err := log.NewRotatingFile("info", "json", "logs/app.log", 1, 3, 7)
	if err != nil {
		panic(err)
	}
	defer logger1.Sync()

	logger1.Info("这条日志会保存到带时间戳的文件中")
	logger1.WithField("timestamp", time.Now()).Info("带时间戳的日志")

	// 示例2：同时输出到轮转文件和控制台
	fmt.Println("\n2. 轮转文件 + 控制台输出:")
	logger2, err := log.NewRotatingFileAndConsole("debug", "console", "logs/debug.log", 2, 5, 10)
	if err != nil {
		panic(err)
	}
	defer logger2.Sync()

	logger2.Debug("调试信息")
	logger2.Info("应用信息")
	logger2.Warn("警告信息")

	// 示例3：自定义轮转配置
	fmt.Println("\n3. 自定义轮转配置:")
	outputs := []log.OutputConfig{
		{Type: "stdout"},
		{
			Type:        "file",
			File:        "logs/custom.log",
			MaxSize:     1, // 1MB
			MaxBackups:  2, // 保留2个备份
			MaxAge:      3, // 保留3天
			Compress:    true,
			UseRotation: true,
			TimeFormat:  "2006.01.02_15:04:05.000",
		},
	}

	logger3, err := log.NewMultiOutput("info", "json", outputs...)
	if err != nil {
		panic(err)
	}
	defer logger3.Sync()

	logger3.Info("自定义轮转配置的日志")
	logger3.WithFields(map[string]interface{}{
		"service": "demo",
		"version": "1.0.0",
	}).Info("服务启动")

	// 示例4：演示文件名生成
	fmt.Println("\n4. 文件名生成演示:")
	basePath := "logs/demo.log"
	timeFormat := "2006.01.02_15:04:05.000"
	fileName := log.GenerateTimeBasedFileName(basePath, timeFormat)
	fmt.Printf("基础路径: %s\n", basePath)
	fmt.Printf("时间格式: %s\n", timeFormat)
	fmt.Printf("生成的文件名: %s\n", fileName)

	// 示例5：不同时间格式
	fmt.Println("\n5. 不同时间格式演示:")
	formats := []string{
		"2006.01.02_15:04:05.000",
		"2006-01-02_15-04-05",
		"20060102_150405",
		"2006.01.02",
	}

	for _, format := range formats {
		fileName := log.GenerateTimeBasedFileName("test.log", format)
		fmt.Printf("格式 %s -> %s\n", format, fileName)
	}

	fmt.Println("\n=== 演示完成 ===")
	fmt.Println("请检查 logs/ 目录下的日志文件")
	fmt.Println("文件命名格式: app_年.月.日_时:分:秒:毫秒.log")
}
