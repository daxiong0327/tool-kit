package consumer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/daxiong/tool-kit/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Config
type Config struct {
	// Queue 队列
	Queue string

	// Exchange 交换机
	Exchange string

	// ExchangeType 交换机类型 (direct, fanout, topic, headers)
	ExchangeType string

	// RoutingKey 路由键
	RoutingKey string

	// QueueDurable 队列是否持久化
	QueueDurable bool

	// QueueAutoDelete 队列是否自动删除
	QueueAutoDelete bool

	// QueueExclusive 队列是否独占
	QueueExclusive bool

	// ExchangeDurable 交换机是否持久化
	ExchangeDurable bool

	// AutoAck 是否自动确认
	AutoAck bool

	// PrefetchCount 预取数量
	PrefetchCount int

	// PrefetchSize 预取大小
	PrefetchSize int

	// ConsumerTag 消费者标签
	ConsumerTag string

	// QueueArgs 队列参数
	QueueArgs map[string]interface{}

	// ExchangeArgs 交换机参数
	ExchangeArgs map[string]interface{}

	// BindingArgs 绑定参数
	BindingArgs map[string]interface{}
}

// Consumer RabbitMQ 消费者
type Consumer struct {
	config      Config
	connManager *rabbitmq.ConnectionManager
	channel     *amqp.Channel
	notifyClose chan *amqp.Error
	mutex       sync.Mutex
	closed      bool
	consuming   bool
}

// New 创建一个新的消费者
func New(connManager *rabbitmq.ConnectionManager, config Config) (*Consumer, error) {
	// 设置默认值
	if config.ExchangeType == "" {
		config.ExchangeType = "direct"
	}

	if config.PrefetchCount <= 0 {
		config.PrefetchCount = 1
	}

	c := &Consumer{
		config:      config,
		connManager: connManager,
		closed:      false,
		consuming:   false,
	}

	// 初始化通道
	if err := c.initChannel(); err != nil {
		return nil, err
	}

	// 设置连接状态
	connManager.SetCallbacks(
		func() { c.handleReconnect() },
		func(err error) { /* do nothing */ },
	)

	return c, nil
}

// initChannel 初始化通道
func (c *Consumer) initChannel() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return rabbitmq.ErrConnectionClosed
	}

	// 创建通道
	ch, err := c.connManager.CreateChannel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	// 设置预取数量
	err = ch.Qos(
		c.config.PrefetchCount, // prefetch count
		c.config.PrefetchSize,  // prefetch size
		false,                  // global
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to set prefetch count: %w", err)
	}

	// 设置交换机
	if c.config.Exchange != "" {
		err = ch.ExchangeDeclare(
			c.config.Exchange,        // name
			c.config.ExchangeType,    // type
			c.config.ExchangeDurable, // durable
			false,                    // auto-deleted
			false,                    // internal
			false,                    // no-wait
			c.config.ExchangeArgs,    // arguments
		)
		if err != nil {
			ch.Close()
			return rabbitmq.WrapError(rabbitmq.ErrExchangeDeclare, err)
		}
	}

	// 声明队列
	q, err := ch.QueueDeclare(
		c.config.Queue,           // name
		c.config.QueueDurable,    // durable
		c.config.QueueAutoDelete, // auto-delete
		c.config.QueueExclusive,  // exclusive
		false,                    // no-wait
		c.config.QueueArgs,       // arguments
	)
	if err != nil {
		ch.Close()
		return rabbitmq.WrapError(rabbitmq.ErrQueueDeclare, err)
	}

	// 如果没有指定队列名称，则使用生成的队列名称
	if c.config.Queue == "" {
		c.config.Queue = q.Name
	}

	// 绑定队列到交换机
	if c.config.Exchange != "" {
		err = ch.QueueBind(
			c.config.Queue,       // queue name
			c.config.RoutingKey,  // routing key
			c.config.Exchange,    // exchange
			false,                // no-wait
			c.config.BindingArgs, // arguments
		)
		if err != nil {
			ch.Close()
			return rabbitmq.WrapError(rabbitmq.ErrQueueBind, err)
		}
	}

	// 设置通道关闭通知
	c.notifyClose = make(chan *amqp.Error, 1)
	ch.NotifyClose(c.notifyClose)

	c.channel = ch

	// 启动监控
	go c.monitor()

	return nil
}

// monitor 监控连接状态
func (c *Consumer) monitor() {
	for {
		select {
		case err, ok := <-c.notifyClose:
			if !ok {
				return
			}
			// 连接关闭
			fmt.Printf("connection closed: %v\n", err)
			return
		}
	}
}

// handleReconnect 	处理连接断开
func (c *Consumer) handleReconnect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return
	}

	// 重新初始化通道
	err := c.initChannel()
	if err != nil {
		fmt.Printf("failed to reinitialize channel: %v\n", err)
		return
	}

	// 如果正在消费，重新启动
	if c.consuming {
		ctx := context.Background()
		_, err := c.Consume(ctx)
		if err != nil {
			fmt.Printf("failed to restart consumption: %v\n", err)
		}
	}
}

// Consume 启动消费
func (c *Consumer) Consume(ctx context.Context) (<-chan amqp.Delivery, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return nil, rabbitmq.ErrConnectionClosed
	}

	if c.channel == nil {
		return nil, rabbitmq.ErrChannelClosed
	}

	// 如果没有指定消费者标签，自动生成
	consumerTag := c.config.ConsumerTag
	if consumerTag == "" {
		consumerTag = fmt.Sprintf("consumer-%d", time.Now().UnixNano())
	}

	// 启动消费
	deliveries, err := c.channel.Consume(
		c.config.Queue,   // queue
		consumerTag,      // consumer tag
		c.config.AutoAck, // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
	if err != nil {
		return nil, rabbitmq.WrapError(rabbitmq.ErrConsumeFailed, err)
	}

	// 标记正在消费
	c.consuming = true

	// 创建消费通道
	ch := make(chan amqp.Delivery)

	// 启动消费
	go func() {
		defer close(ch)

		for {
			select {
			case d, ok := <-deliveries:
				if !ok {
					// 消费者关闭
					return
				}
				// 将消息发送到消费通道
				select {
				case ch <- d:
					// 消息已发送
				case <-ctx.Done():
					// 上下文已取消
					return
				}
			case <-ctx.Done():
				// 上下文已取消
				return
			}
		}
	}()

	return ch, nil
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	c.consuming = false

	if c.channel != nil {
		err := c.channel.Close()
		if err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
		c.channel = nil
	}

	return nil
}
