package main

import (
	"fmt"
	"time"

	"github.com/daxiong0327/tool-kit/log"
)

func main() {
	fmt.Println("=== Log 多输出功能演示 ===")

	// 示例1：同时输出到控制台和文件
	fmt.Println("\n1. 同时输出到控制台和文件:")
	logger1, err := log.NewFileAndConsole("info", "console", "logs/demo.log")
	if err != nil {
		panic(err)
	}
	defer logger1.Sync()

	logger1.Info("这条日志会同时显示在控制台和保存到文件")
	logger1.WithField("timestamp", time.Now()).Info("带时间戳的日志")

	// 示例2：多文件输出
	fmt.Println("\n2. 输出到多个文件:")
	outputs := []log.OutputConfig{
		{Type: "stdout"},
		{Type: "file", File: "logs/app.log"},
		{Type: "file", File: "logs/error.log"},
	}

	logger2, err := log.NewMultiOutput("debug", "json", outputs...)
	if err != nil {
		panic(err)
	}
	defer logger2.Sync()

	logger2.Debug("调试信息")
	logger2.Info("应用信息")
	logger2.Error("错误信息")

	// 示例3：使用MultiWriter
	fmt.Println("\n3. 使用MultiWriter:")
	var buf1, buf2 []byte
	writer1 := &bufferWriter{&buf1}
	writer2 := &bufferWriter{&buf2}

	multiWriter := log.NewMultiWriter(writer1, writer2)
	logger3, err := log.NewWithWriter("info", "json", multiWriter)
	if err != nil {
		panic(err)
	}

	logger3.Info("这条日志会写入到两个buffer")
	fmt.Printf("Buffer1: %s\n", string(buf1))
	fmt.Printf("Buffer2: %s\n", string(buf2))

	// 示例4：动态添加输出
	fmt.Println("\n4. 动态添加输出:")
	var buf3 []byte
	writer3 := &bufferWriter{&buf3}
	multiWriter.AddWriter(writer3)

	logger3.Info("现在这条日志会写入到三个buffer")
	fmt.Printf("Buffer1: %s\n", string(buf1))
	fmt.Printf("Buffer2: %s\n", string(buf2))
	fmt.Printf("Buffer3: %s\n", string(buf3))

	fmt.Println("\n=== 演示完成 ===")
	fmt.Println("请检查 logs/ 目录下的日志文件")
}

// bufferWriter 实现io.Writer接口，用于演示
type bufferWriter struct {
	buf *[]byte
}

func (bw *bufferWriter) Write(p []byte) (n int, err error) {
	*bw.buf = append(*bw.buf, p...)
	return len(p), nil
}
