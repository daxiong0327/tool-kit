package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Transaction 事务操作
type Transaction struct {
	client *redis.Client
	tx     redis.Pipeliner
}

// NewTransaction 创建事务实例
func (c *Client) NewTransaction() *Transaction {
	return &Transaction{client: c.client}
}

// Watch 监视键
func (t *Transaction) Watch(ctx context.Context, keys ...string) error {
	t.tx = t.client.TxPipeline()
	return t.client.Watch(ctx, func(tx *redis.Tx) error {
		// 这里可以添加事务逻辑
		return nil
	}, keys...)
}

// Exec 执行事务
func (t *Transaction) Exec(ctx context.Context, fn func(*redis.Tx) error, keys []string) error {
	return t.client.Watch(ctx, fn, keys...)
}

// ExecWithRetry 带重试的事务执行
func (t *Transaction) ExecWithRetry(ctx context.Context, fn func(*redis.Tx) error, keys []string, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		err := t.client.Watch(ctx, fn, keys...)
		if err == nil {
			return nil
		}
		if err == redis.TxFailedErr {
			continue // 重试
		}
		return err
	}
	return redis.TxFailedErr
}

// Multi 开始事务
func (t *Transaction) Multi(ctx context.Context) error {
	t.tx = t.client.TxPipeline()
	return nil
}

// ExecPipeline 执行管道事务
func (t *Transaction) ExecPipeline(ctx context.Context) ([]redis.Cmder, error) {
	if t.tx == nil {
		return nil, redis.Nil
	}
	return t.tx.Exec(ctx)
}

// Discard 丢弃事务
func (t *Transaction) Discard() {
	if t.tx != nil {
		t.tx.Discard()
	}
}

// String 操作
func (t *Transaction) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	if t.tx == nil {
		return t.client.Set(ctx, key, value, expiration)
	}
	return t.tx.Set(ctx, key, value, expiration)
}

func (t *Transaction) Get(ctx context.Context, key string) *redis.StringCmd {
	if t.tx == nil {
		return t.client.Get(ctx, key)
	}
	return t.tx.Get(ctx, key)
}

func (t *Transaction) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	if t.tx == nil {
		return t.client.Del(ctx, keys...)
	}
	return t.tx.Del(ctx, keys...)
}

func (t *Transaction) Incr(ctx context.Context, key string) *redis.IntCmd {
	if t.tx == nil {
		return t.client.Incr(ctx, key)
	}
	return t.tx.Incr(ctx, key)
}

func (t *Transaction) IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd {
	if t.tx == nil {
		return t.client.IncrBy(ctx, key, value)
	}
	return t.tx.IncrBy(ctx, key, value)
}

// Hash 操作
func (t *Transaction) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	if t.tx == nil {
		return t.client.HSet(ctx, key, values...)
	}
	return t.tx.HSet(ctx, key, values...)
}

func (t *Transaction) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	if t.tx == nil {
		return t.client.HGet(ctx, key, field)
	}
	return t.tx.HGet(ctx, key, field)
}

func (t *Transaction) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	if t.tx == nil {
		return t.client.HDel(ctx, key, fields...)
	}
	return t.tx.HDel(ctx, key, fields...)
}

// List 操作
func (t *Transaction) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	if t.tx == nil {
		return t.client.LPush(ctx, key, values...)
	}
	return t.tx.LPush(ctx, key, values...)
}

func (t *Transaction) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	if t.tx == nil {
		return t.client.RPush(ctx, key, values...)
	}
	return t.tx.RPush(ctx, key, values...)
}

func (t *Transaction) LPop(ctx context.Context, key string) *redis.StringCmd {
	if t.tx == nil {
		return t.client.LPop(ctx, key)
	}
	return t.tx.LPop(ctx, key)
}

func (t *Transaction) RPop(ctx context.Context, key string) *redis.StringCmd {
	if t.tx == nil {
		return t.client.RPop(ctx, key)
	}
	return t.tx.RPop(ctx, key)
}

// Set 操作
func (t *Transaction) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	if t.tx == nil {
		return t.client.SAdd(ctx, key, members...)
	}
	return t.tx.SAdd(ctx, key, members...)
}

func (t *Transaction) SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	if t.tx == nil {
		return t.client.SRem(ctx, key, members...)
	}
	return t.tx.SRem(ctx, key, members...)
}

// ZSet 操作
func (t *Transaction) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	if t.tx == nil {
		return t.client.ZAdd(ctx, key, members...)
	}
	return t.tx.ZAdd(ctx, key, members...)
}

func (t *Transaction) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	if t.tx == nil {
		return t.client.ZRem(ctx, key, members...)
	}
	return t.tx.ZRem(ctx, key, members...)
}

// 通用操作
func (t *Transaction) Ping(ctx context.Context) *redis.StatusCmd {
	if t.tx == nil {
		return t.client.Ping(ctx)
	}
	return t.tx.Ping(ctx)
}

// TransactionOptions 事务选项
type TransactionOptions struct {
	MaxRetries int           `json:"max_retries" yaml:"max_retries"` // 最大重试次数
	Timeout    time.Duration `json:"timeout" yaml:"timeout"`         // 超时时间
}

// DefaultTransactionOptions 默认事务选项
func DefaultTransactionOptions() *TransactionOptions {
	return &TransactionOptions{
		MaxRetries: 3,
		Timeout:    5 * time.Second,
	}
}

// WithTransaction 执行事务
func (c *Client) WithTransaction(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error {
	return c.client.Watch(ctx, fn, keys...)
}

// WithTransactionOptions 带选项的事务执行
func (c *Client) WithTransactionOptions(ctx context.Context, fn func(*redis.Tx) error, keys []string, opts *TransactionOptions) error {
	if opts == nil {
		opts = DefaultTransactionOptions()
	}

	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	for i := 0; i < opts.MaxRetries; i++ {
		err := c.client.Watch(ctx, fn, keys...)
		if err == nil {
			return nil
		}
		if err == redis.TxFailedErr {
			continue // 重试
		}
		return err
	}
	return redis.TxFailedErr
}
