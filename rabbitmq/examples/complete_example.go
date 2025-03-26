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
	queueName := "fake_live_task_priority_queue"
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
		ExchangeType: "fanout",
		RoutingKey:   routingKey,
		Durable:      true,
		DeliveryMode: 2, // 持久化消息
		ContentType:  "application/json",
		// 添加优先级配置
		MaxPriority:     5, // 设置最大优先级为5
		DefaultPriority: 1, // 设置默认优先级为1
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

		// 根据消息ID决定优先级
		priority := uint8(counter%5 + 1) // 优先级从1到5循环
		msg["priority"] = priority

		// 使用不同的优先级发送消息
		jsonData, err := json.Marshal(msg)
		if err != nil {
			log.Printf("消息序列化失败: %v", err)
			continue
		}

		// 使用优先级发送消息
		err = p.PublishWithPriority(ctx, jsonData, nil, priority)
		if err != nil {
			log.Printf("发送消息失败: %v", err)
		} else {
			fmt.Printf("消息已发送 (优先级: %d): %+v\n", priority, msg)
		}

		// 每5条消息发送一条高优先级消息
		if counter%5 == 0 {
			highPriorityMsg := map[string]interface{}{
				"id":        counter * 100,
				"message":   fmt.Sprintf("这是一条高优先级消息(%d)", counter),
				"time":      time.Now().Format(time.RFC3339),
				"task_type": "fake_live_task",
				"priority":  5,
			}

			// 序列化高优先级消息
			highPriorityData, err := json.Marshal(highPriorityMsg)
			if err != nil {
				log.Printf("高优先级消息序列化失败: %v", err)
				continue
			}

			// 使用JSON格式发送高优先级消息
			err = p.PublishWithPriority(ctx, highPriorityData, nil, 5) // 使用最高优先级
			if err != nil {
				log.Printf("发送高优先级消息失败: %v", err)
			} else {
				fmt.Printf("高优先级消息已发送: %+v\n", highPriorityMsg)
			}
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
		ExchangeType:    "fanout",
		RoutingKey:      routingKey,
		QueueDurable:    true,
		ExchangeDurable: true,
		AutoAck:         false,
		PrefetchCount:   1,
		// 添加优先级配置
		MaxPriority: 5, // 设置最大优先级为5
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
