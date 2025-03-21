package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/daxiong/tool-kit/rabbitmq"
	"github.com/daxiong/tool-kit/rabbitmq/producer"
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

	// 创建生产者配置
	producerConfig := producer.Config{
		Exchange:     "test_exchange",
		ExchangeType: "direct",
		RoutingKey:   "test_key",
		Durable:      true,
		DeliveryMode: 2, // 持久化消息
		ContentType:  "application/json",
	}

	// 创建生产者
	p, err := producer.New(connManager, producerConfig)
	if err != nil {
		log.Fatalf("创建生产者失败: %v", err)
	}
	defer p.Close()

	// 发送普通消息
	ctx := context.Background()
	msg := map[string]interface{}{
		"message": "这是一条测试消息",
		"time":    time.Now().Format(time.RFC3339),
	}

	// 使用JSON格式发送消息
	err = p.PublishJSON(ctx, msg)
	if err != nil {
		log.Fatalf("发送消息失败: %v", err)
	}
	fmt.Println("消息已发送")

	// 发送带延迟的消息（需要RabbitMQ安装delay插件）
	delayMsg := map[string]interface{}{
		"message": "这是一条延迟消息",
		"time":    time.Now().Format(time.RFC3339),
	}

	delayMsgBytes, _ := json.Marshal(delayMsg)
	err = p.PublishWithDelay(ctx, delayMsgBytes, map[string]interface{}{
		"content-type": "application/json",
	}, 5*time.Second)

	if err != nil {
		log.Printf("发送延迟消息失败: %v", err)
	} else {
		fmt.Println("延迟消息已发送，将在5秒后被消费")
	}

	// 等待一段时间确保消息被发送
	time.Sleep(1 * time.Second)
}
