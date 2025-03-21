package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daxiong0327/tool-kit/rabbitmq"
	"github.com/daxiong0327/tool-kit/rabbitmq/consumer"
)

func main() {
	// 创建RabbitMQ连接配置
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

	// 创建连接管理器
	connManager, err := rabbitmq.NewConnectionManager(config)
	if err != nil {
		log.Fatalf("创建连接管理器失败: %v", err)
	}
	defer connManager.Close()

	// 创建消费者配置
	consumerConfig := consumer.Config{
		Queue:           "test_queue",
		Exchange:        "test_exchange",
		ExchangeType:    "direct",
		RoutingKey:      "test_key",
		QueueDurable:    true,
		ExchangeDurable: true,
		AutoAck:         false,
		PrefetchCount:   1,
	}

	// 创建消费者
	c, err := consumer.New(connManager, consumerConfig)
	if err != nil {
		log.Fatalf("创建消费者失败: %v", err)
	}
	defer c.Close()

	// 开始消费消息
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deliveries, err := c.Consume(ctx)
	if err != nil {
		log.Fatalf("开始消费失败: %v", err)
	}

	// 设置信号处理，以便优雅关闭
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 处理消息
	go func() {
		for delivery := range deliveries {
			// 处理消息
			processMessage(delivery.Body)

			// 手动确认消息
			err := delivery.Ack(false) // multiple=false，只确认当前消息
			if err != nil {
				log.Printf("确认消息失败: %v", err)
			}
		}
	}()

	fmt.Println("消费者已启动，按Ctrl+C退出")

	// 等待退出信号
	<-sigCh
	fmt.Println("接收到退出信号，正在关闭消费者...")
	cancel()                    // 取消上下文，停止消费
	time.Sleep(1 * time.Second) // 给一些时间完成清理工作
}

// 处理接收到的消息
func processMessage(body []byte) {
	// 尝试解析为JSON
	var msg map[string]interface{}
	if err := json.Unmarshal(body, &msg); err == nil {
		fmt.Printf("收到消息: %+v\n", msg)
	} else {
		// 如果不是JSON，则作为字符串处理
		fmt.Printf("收到消息: %s\n", string(body))
	}
}
