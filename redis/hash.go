package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Hash 哈希操作
type Hash struct {
	client *redis.Client
}

// NewHash 创建哈希操作实例
func (c *Client) NewHash() *Hash {
	return &Hash{client: c.client}
}

// HSet 设置哈希字段
func (h *Hash) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return h.client.HSet(ctx, key, values...).Result()
}

// HGet 获取哈希字段值
func (h *Hash) HGet(ctx context.Context, key, field string) (string, error) {
	return h.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段和值
func (h *Hash) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return h.client.HGetAll(ctx, key).Result()
}

// HExists 检查哈希字段是否存在
func (h *Hash) HExists(ctx context.Context, key, field string) (bool, error) {
	return h.client.HExists(ctx, key, field).Result()
}

// HDel 删除哈希字段
func (h *Hash) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return h.client.HDel(ctx, key, fields...).Result()
}

// HLen 获取哈希字段数量
func (h *Hash) HLen(ctx context.Context, key string) (int64, error) {
	return h.client.HLen(ctx, key).Result()
}

// HKeys 获取所有哈希字段名
func (h *Hash) HKeys(ctx context.Context, key string) ([]string, error) {
	return h.client.HKeys(ctx, key).Result()
}

// HVals 获取所有哈希字段值
func (h *Hash) HVals(ctx context.Context, key string) ([]string, error) {
	return h.client.HVals(ctx, key).Result()
}

// HMGet 批量获取哈希字段值
func (h *Hash) HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, error) {
	return h.client.HMGet(ctx, key, fields...).Result()
}

// HMSet 批量设置哈希字段
func (h *Hash) HMSet(ctx context.Context, key string, values ...interface{}) error {
	return h.client.HMSet(ctx, key, values...).Err()
}

// HIncrBy 哈希字段自增
func (h *Hash) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	return h.client.HIncrBy(ctx, key, field, incr).Result()
}

// HIncrByFloat 哈希字段自增浮点数
func (h *Hash) HIncrByFloat(ctx context.Context, key, field string, incr float64) (float64, error) {
	return h.client.HIncrByFloat(ctx, key, field, incr).Result()
}

// HSetNX 设置哈希字段（仅当字段不存在时）
func (h *Hash) HSetNX(ctx context.Context, key, field string, value interface{}) (bool, error) {
	return h.client.HSetNX(ctx, key, field, value).Result()
}

// HStrLen 获取哈希字段值的字符串长度
func (h *Hash) HStrLen(ctx context.Context, key, field string) (int64, error) {
	return h.client.HStrLen(ctx, key, field).Result()
}

// HScan 扫描哈希字段
func (h *Hash) HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return h.client.HScan(ctx, key, cursor, match, count).Result()
}

// HScanAll 扫描所有哈希字段
func (h *Hash) HScanAll(ctx context.Context, key string, match string, count int64) (map[string]string, error) {
	result := make(map[string]string)
	cursor := uint64(0)

	for {
		fields, nextCursor, err := h.HScan(ctx, key, cursor, match, count)
		if err != nil {
			return nil, err
		}

		// 将字段和值配对
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				result[fields[i]] = fields[i+1]
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return result, nil
}

// HExpire 设置哈希键的过期时间
func (h *Hash) HExpire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return h.client.Expire(ctx, key, expiration).Result()
}

// HTTL 获取哈希键的剩余生存时间
func (h *Hash) HTTL(ctx context.Context, key string) (time.Duration, error) {
	return h.client.TTL(ctx, key).Result()
}

// HPersist 移除哈希键的过期时间
func (h *Hash) HPersist(ctx context.Context, key string) (bool, error) {
	return h.client.Persist(ctx, key).Result()
}
