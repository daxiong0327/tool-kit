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
	"github.com/daxiong0327/tool-kit/rabbitmq/producer"
)

func main() {
	// 创建RabbitMQ连接配置
	config := rabbitmq.Config{
		Host:                 "127.0.0.1",
		Port:                 5672,
		Username:             "admin",
		Password:             "admin",
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

	// 设置Exchange和Queue名称
	exchangeName := "fake_live_task_test"
	queueName := "fake_live_task_queue"
	routingKey := "fake_live_task"

	// 启动生产者
	go runProducer(connManager, exchangeName, routingKey)

	// 启动消费者
	go runConsumer(connManager, exchangeName, queueName, routingKey)

	// 设置信号处理，以便优雅关闭
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("程序已启动，按Ctrl+C退出")

	// 等待退出信号
	<-sigCh
	fmt.Println("接收到退出信号，正在关闭程序...")
	time.Sleep(2 * time.Second) // 给一些时间完成清理工作
}

// 运行生产者，定期发送消息
func runProducer(connManager *rabbitmq.ConnectionManager, exchange, routingKey string) {
	// 创建生产者配置
	producerConfig := producer.Config{
		Exchange:     exchange,
		ExchangeType: "direct",
		RoutingKey:   routingKey,
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

	// 定期发送消息
	ctx := context.Background()
	counter := 1

	for {
		// 创建消息内容
		msg := map[string]interface{}{
			"id":        counter,
			"message":   fmt.Sprintf("这是第%d条测试消息", counter),
			"time":      time.Now().Format(time.RFC3339),
			"task_type": "fake_live_task",
		}

		// 使用JSON格式发送消息
		err := p.PublishJSON(ctx, msg)
		if err != nil {
			log.Printf("发送消息失败: %v", err)
		} else {
			fmt.Printf("消息已发送: %+v\n", msg)
		}

		counter++
		time.Sleep(5 * time.Second) // 每5秒发送一条消息
	}
}

// 运行消费者，处理接收到的消息
func runConsumer(connManager *rabbitmq.ConnectionManager, exchange, queue, routingKey string) {
	// 创建消费者配置
	consumerConfig := consumer.Config{
		Queue:           queue,
		Exchange:        exchange,
		ExchangeType:    "direct",
		RoutingKey:      routingKey,
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

	fmt.Println("消费者已启动，等待接收消息...")

	// 处理消息
	for delivery := range deliveries {
		// 处理消息
		processMessage(delivery.Body)

		// 手动确认消息
		err := delivery.Ack(false) // multiple=false，只确认当前消息
		if err != nil {
			log.Printf("确认消息失败: %v", err)
		}
	}
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
