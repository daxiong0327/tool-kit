package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// PubSub 发布订阅操作
type PubSub struct {
	client *redis.Client
	pubsub *redis.PubSub
}

// NewPubSub 创建发布订阅实例
func (c *Client) NewPubSub() *PubSub {
	return &PubSub{client: c.client}
}

// Subscribe 订阅频道
func (ps *PubSub) Subscribe(ctx context.Context, channels ...string) error {
	ps.pubsub = ps.client.Subscribe(ctx, channels...)
	return ps.pubsub.Ping(ctx)
}

// PSubscribe 订阅模式
func (ps *PubSub) PSubscribe(ctx context.Context, channels ...string) error {
	ps.pubsub = ps.client.PSubscribe(ctx, channels...)
	return ps.pubsub.Ping(ctx)
}

// SSubscribe 订阅共享频道
func (ps *PubSub) SSubscribe(ctx context.Context, channels ...string) error {
	ps.pubsub = ps.client.SSubscribe(ctx, channels...)
	return ps.pubsub.Ping(ctx)
}

// Publish 发布消息
func (ps *PubSub) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	return ps.client.Publish(ctx, channel, message).Result()
}

// Receive 接收消息
func (ps *PubSub) Receive(ctx context.Context) (interface{}, error) {
	if ps.pubsub == nil {
		return nil, redis.Nil
	}
	return ps.pubsub.Receive(ctx)
}

// ReceiveMessage 接收消息（类型化）
func (ps *PubSub) ReceiveMessage(ctx context.Context) (*redis.Message, error) {
	if ps.pubsub == nil {
		return nil, redis.Nil
	}
	return ps.pubsub.ReceiveMessage(ctx)
}

// ReceiveTimeout 带超时的接收消息
func (ps *PubSub) ReceiveTimeout(ctx context.Context, timeout time.Duration) (interface{}, error) {
	if ps.pubsub == nil {
		return nil, redis.Nil
	}
	return ps.pubsub.ReceiveTimeout(ctx, timeout)
}

// Channel 返回消息通道
func (ps *PubSub) Channel() <-chan *redis.Message {
	if ps.pubsub == nil {
		return nil
	}
	return ps.pubsub.Channel()
}

// ChannelSize 返回带缓冲的消息通道
func (ps *PubSub) ChannelSize(size int) <-chan *redis.Message {
	if ps.pubsub == nil {
		return nil
	}
	return ps.pubsub.Channel(redis.WithChannelSize(size))
}

// Close 关闭发布订阅
func (ps *PubSub) Close() error {
	if ps.pubsub == nil {
		return nil
	}
	return ps.pubsub.Close()
}

// Ping 测试连接
func (ps *PubSub) Ping(ctx context.Context, payload ...string) error {
	if ps.pubsub == nil {
		return redis.Nil
	}
	return ps.pubsub.Ping(ctx, payload...)
}

// Unsubscribe 取消订阅频道
func (ps *PubSub) Unsubscribe(ctx context.Context, channels ...string) error {
	if ps.pubsub == nil {
		return redis.Nil
	}
	return ps.pubsub.Unsubscribe(ctx, channels...)
}

// PUnsubscribe 取消订阅模式
func (ps *PubSub) PUnsubscribe(ctx context.Context, channels ...string) error {
	if ps.pubsub == nil {
		return redis.Nil
	}
	return ps.pubsub.PUnsubscribe(ctx, channels...)
}

// SUnsubscribe 取消订阅共享频道
func (ps *PubSub) SUnsubscribe(ctx context.Context, channels ...string) error {
	if ps.pubsub == nil {
		return redis.Nil
	}
	return ps.pubsub.SUnsubscribe(ctx, channels...)
}

// PubSubChannels 获取活跃频道列表
func (ps *PubSub) PubSubChannels(ctx context.Context, pattern string) ([]string, error) {
	return ps.client.PubSubChannels(ctx, pattern).Result()
}

// PubSubNumSub 获取频道订阅数量
func (ps *PubSub) PubSubNumSub(ctx context.Context, channels ...string) (map[string]int64, error) {
	return ps.client.PubSubNumSub(ctx, channels...).Result()
}

// PubSubNumPat 获取模式订阅数量
func (ps *PubSub) PubSubNumPat(ctx context.Context) (int64, error) {
	return ps.client.PubSubNumPat(ctx).Result()
}

// PubSubShardChannels 获取共享频道列表
func (ps *PubSub) PubSubShardChannels(ctx context.Context, pattern string) ([]string, error) {
	return ps.client.PubSubShardChannels(ctx, pattern).Result()
}

// PubSubShardNumSub 获取共享频道订阅数量
func (ps *PubSub) PubSubShardNumSub(ctx context.Context, channels ...string) (map[string]int64, error) {
	return ps.client.PubSubShardNumSub(ctx, channels...).Result()
}

// Publisher 发布者
type Publisher struct {
	client *redis.Client
}

// NewPublisher 创建发布者
func (c *Client) NewPublisher() *Publisher {
	return &Publisher{client: c.client}
}

// Publish 发布消息
func (p *Publisher) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	return p.client.Publish(ctx, channel, message).Result()
}

// PublishMany 批量发布消息
func (p *Publisher) PublishMany(ctx context.Context, channel string, messages ...interface{}) ([]int64, error) {
	var results []int64
	for _, message := range messages {
		count, err := p.Publish(ctx, channel, message)
		if err != nil {
			return results, err
		}
		results = append(results, count)
	}
	return results, nil
}

// Subscriber 订阅者
type Subscriber struct {
	client *redis.Client
	pubsub *redis.PubSub
}

// NewSubscriber 创建订阅者
func (c *Client) NewSubscriber() *Subscriber {
	return &Subscriber{client: c.client}
}

// Subscribe 订阅频道
func (s *Subscriber) Subscribe(ctx context.Context, channels ...string) error {
	s.pubsub = s.client.Subscribe(ctx, channels...)
	return s.pubsub.Ping(ctx)
}

// PSubscribe 订阅模式
func (s *Subscriber) PSubscribe(ctx context.Context, channels ...string) error {
	s.pubsub = s.client.PSubscribe(ctx, channels...)
	return s.pubsub.Ping(ctx)
}

// Listen 监听消息
func (s *Subscriber) Listen(ctx context.Context, handler func(*redis.Message)) error {
	if s.pubsub == nil {
		return redis.Nil
	}

	ch := s.pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			if msg != nil {
				handler(msg)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Close 关闭订阅
func (s *Subscriber) Close() error {
	if s.pubsub == nil {
		return nil
	}
	return s.pubsub.Close()
}
