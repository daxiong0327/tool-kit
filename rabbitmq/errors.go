package rabbitmq

import (
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	// ErrNotConnected 表示没有连接到RabbitMQ服务器
	ErrNotConnected = errors.New("unconnected to RabbitMQ server")

	// ErrChannelClosed 表示通道已关闭
	ErrChannelClosed = errors.New("channel closed")

	// ErrConnectionClosed 表示连接已关闭
	ErrConnectionClosed = errors.New("connection closed")

	// ErrPublishFailed 表示发布消息失败
	ErrPublishFailed = errors.New("publish failed")

	// ErrConsumeFailed 表示消费消息失败
	ErrConsumeFailed = errors.New("consume failed")

	// ErrQueueDeclare 表示声明队列失败
	ErrQueueDeclare = errors.New("declare queue failed")

	// ErrExchangeDeclare 表示声明交换机失败
	ErrExchangeDeclare = errors.New("declare exchange failed")

	// ErrQueueBind 表示绑定队列失败
	ErrQueueBind = errors.New("bind queue failed")

	// ErrConfirmMode 表示设置确认模式失败
	ErrConfirmMode = errors.New("confirm mode failed")
)

// IsConnectionError 检查错误是否与连接相关
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrNotConnected) || errors.Is(err, ErrConnectionClosed) {
		return true
	}

	// 检查是否为amqp库的连接错误
	var amqpErr *amqp.Error
	if errors.As(err, &amqpErr) {
		// 连接错误代码范围
		return amqpErr.Code >= 300 && amqpErr.Code < 400
	}

	return false
}

// IsChannelError 检查错误是否与通道相关
func IsChannelError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrChannelClosed) {
		return true
	}

	// 检查是否为amqp库的通道错误
	var amqpErr *amqp.Error
	if errors.As(err, &amqpErr) {
		// 通道错误代码范围
		return amqpErr.Code >= 200 && amqpErr.Code < 300
	}

	return false
}

// WrapError 包装错误并添加上下文
func WrapError(baseErr error, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", baseErr, err)
}
