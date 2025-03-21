package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/daxiong/tool-kit/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Config 包含生产者的配置参数
type Config struct {
	// Exchange 交换机名称
	Exchange string

	// ExchangeType 交换机类型 (direct, fanout, topic, headers)
	ExchangeType string

	// RoutingKey 路由键
	RoutingKey string

	// Durable 是否持久化
	Durable bool

	// Mandatory 是否要求消息必须被路由到队列
	Mandatory bool

	// Immediate 是否要求消息必须被立即消费
	Immediate bool

	// ConfirmMode 是否启用发布确认模式
	ConfirmMode bool

	// ConfirmTimeout 发布确认超时时间
	ConfirmTimeout time.Duration

	// DeliveryMode 投递模式 (1=非持久化, 2=持久化)
	DeliveryMode uint8

	// ContentType 内容类型
	ContentType string

	// ContentEncoding 内容编码
	ContentEncoding string

	// ExchangeArgs 交换机参数
	ExchangeArgs map[string]interface{}
}

// Producer RabbitMQ消息生产者
type Producer struct {
	config        Config
	connManager   *rabbitmq.ConnectionManager
	channel       *amqp.Channel
	notifyClose   chan *amqp.Error
	notifyReturn  chan amqp.Return
	notifyConfirm chan amqp.Confirmation
	mutex         sync.Mutex
	closed        bool
}

// New 创建一个新的生产者
func New(connManager *rabbitmq.ConnectionManager, config Config) (*Producer, error) {
	// 设置默认值
	if config.ExchangeType == "" {
		config.ExchangeType = "direct"
	}

	if config.DeliveryMode == 0 {
		config.DeliveryMode = 1 // 默认非持久化
	}

	if config.ContentType == "" {
		config.ContentType = "application/octet-stream"
	}

	if config.ConfirmTimeout == 0 {
		config.ConfirmTimeout = 5 * time.Second
	}

	p := &Producer{
		config:      config,
		connManager: connManager,
		closed:      false,
	}

	// 初始化通道
	if err := p.initChannel(); err != nil {
		return nil, err
	}

	// 设置连接状态回调
	connManager.SetCallbacks(
		func() { p.handleReconnect() },
		func(err error) { /* 断开连接时的处理 */ },
	)

	return p, nil
}

// initChannel 初始化通道
func (p *Producer) initChannel() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.closed {
		return rabbitmq.ErrConnectionClosed
	}

	// 创建通道
	ch, err := p.connManager.CreateChannel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	// 声明交换机
	if p.config.Exchange != "" {
		err = ch.ExchangeDeclare(
			p.config.Exchange,     // name
			p.config.ExchangeType, // type
			p.config.Durable,      // durable
			false,                 // auto-deleted
			false,                 // internal
			false,                 // no-wait
			p.config.ExchangeArgs, // arguments
		)
		if err != nil {
			ch.Close()
			return rabbitmq.WrapError(rabbitmq.ErrExchangeDeclare, err)
		}
	}

	// 设置发布确认模式
	if p.config.ConfirmMode {
		err = ch.Confirm(false) // noWait = false
		if err != nil {
			ch.Close()
			return rabbitmq.WrapError(rabbitmq.ErrConfirmMode, err)
		}

		p.notifyConfirm = ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	}

	// 设置通道关闭通知
	p.notifyClose = make(chan *amqp.Error, 1)
	ch.NotifyClose(p.notifyClose)

	// 设置退回消息通知
	if p.config.Mandatory || p.config.Immediate {
		p.notifyReturn = make(chan amqp.Return, 1)
		ch.NotifyReturn(p.notifyReturn)
	}

	p.channel = ch

	// 启动监控
	go p.monitor()

	return nil
}

// monitor 监控通道状态
func (p *Producer) monitor() {
	for {
		select {
		case err, ok := <-p.notifyClose:
			if !ok {
				return
			}
			// 通道关闭，等待连接管理器重连
			fmt.Printf("producer channel closed: %v\n", err)
			return

		case ret, ok := <-p.notifyReturn:
			if !ok {
				return
			}
			// 处理退回的消息
			fmt.Printf("message returned: %s, reason: %s\n", ret.RoutingKey, ret.ReplyText)
		}
	}
}

// handleReconnect 处理重连
func (p *Producer) handleReconnect() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.closed {
		return
	}

	// 重新初始化通道
	err := p.initChannel()
	if err != nil {
		fmt.Printf("failed to initialize channel after reconnect: %v\n", err)
	}
}

// Publish 发布消息
func (p *Producer) Publish(ctx context.Context, body []byte, headers map[string]interface{}) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.closed {
		return rabbitmq.ErrConnectionClosed
	}

	if p.channel == nil {
		return rabbitmq.ErrChannelClosed
	}

	// 准备消息
	msg := amqp.Publishing{
		Headers:         headers,
		ContentType:     p.config.ContentType,
		ContentEncoding: p.config.ContentEncoding,
		Body:            body,
		DeliveryMode:    p.config.DeliveryMode,
		Timestamp:       time.Now(),
	}

	// 发布消息
	err := p.channel.PublishWithContext(
		ctx,
		p.config.Exchange,   // exchange
		p.config.RoutingKey, // routing key
		p.config.Mandatory,  // mandatory
		p.config.Immediate,  // immediate
		msg,
	)

	if err != nil {
		return rabbitmq.WrapError(rabbitmq.ErrPublishFailed, err)
	}

	// 如果启用了确认模式，等待确认
	if p.config.ConfirmMode {
		select {
		case confirm := <-p.notifyConfirm:
			if !confirm.Ack {
				return fmt.Errorf("message not confirmed")
			}
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(p.config.ConfirmTimeout):
			return fmt.Errorf("timeout waiting for confirmation")
		}
	}

	return nil
}

// PublishWithDelay 发布延迟消息
func (p *Producer) PublishWithDelay(ctx context.Context, body []byte, headers map[string]interface{}, delay time.Duration) error {
	if headers == nil {
		headers = make(map[string]interface{})
	}

	// 设置延迟参数
	headers["x-delay"] = int(delay.Milliseconds())

	return p.Publish(ctx, body, headers)
}

// PublishJSON 发布JSON格式的消息
func (p *Producer) PublishJSON(ctx context.Context, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return p.Publish(ctx, body, map[string]interface{}{"content-type": "application/json"})
}

// Close 关闭生产者
func (p *Producer) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	if p.channel != nil {
		err := p.channel.Close()
		if err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
		p.channel = nil
	}

	return nil
}
