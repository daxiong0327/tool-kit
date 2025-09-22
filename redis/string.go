package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// String 字符串操作
type String struct {
	client *redis.Client
}

// NewString 创建字符串操作实例
func (c *Client) NewString() *String {
	return &String{client: c.client}
}

// Set 设置键值
func (s *String) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return s.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (s *String) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

// GetInt 获取整数值
func (s *String) GetInt(ctx context.Context, key string) (int64, error) {
	return s.client.Get(ctx, key).Int64()
}

// GetFloat 获取浮点数值
func (s *String) GetFloat(ctx context.Context, key string) (float64, error) {
	return s.client.Get(ctx, key).Float64()
}

// GetBool 获取布尔值
func (s *String) GetBool(ctx context.Context, key string) (bool, error) {
	return s.client.Get(ctx, key).Bool()
}

// SetNX 设置键值（仅当键不存在时）
func (s *String) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, value, expiration).Result()
}

// SetXX 设置键值（仅当键存在时）
func (s *String) SetXX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return s.client.SetXX(ctx, key, value, expiration).Result()
}

// SetEX 设置键值并指定过期时间
func (s *String) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return s.client.SetEx(ctx, key, value, expiration).Err()
}

// SetRange 设置字符串的指定范围
func (s *String) SetRange(ctx context.Context, key string, offset int64, value string) (int64, error) {
	return s.client.SetRange(ctx, key, offset, value).Result()
}

// GetRange 获取字符串的指定范围
func (s *String) GetRange(ctx context.Context, key string, start, end int64) (string, error) {
	return s.client.GetRange(ctx, key, start, end).Result()
}

// Append 追加字符串
func (s *String) Append(ctx context.Context, key, value string) (int64, error) {
	return s.client.Append(ctx, key, value).Result()
}

// StrLen 获取字符串长度
func (s *String) StrLen(ctx context.Context, key string) (int64, error) {
	return s.client.StrLen(ctx, key).Result()
}

// Incr 自增1
func (s *String) Incr(ctx context.Context, key string) (int64, error) {
	return s.client.Incr(ctx, key).Result()
}

// IncrBy 自增指定值
func (s *String) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return s.client.IncrBy(ctx, key, value).Result()
}

// IncrByFloat 自增浮点数值
func (s *String) IncrByFloat(ctx context.Context, key string, value float64) (float64, error) {
	return s.client.IncrByFloat(ctx, key, value).Result()
}

// Decr 自减1
func (s *String) Decr(ctx context.Context, key string) (int64, error) {
	return s.client.Decr(ctx, key).Result()
}

// DecrBy 自减指定值
func (s *String) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return s.client.DecrBy(ctx, key, value).Result()
}

// MGet 批量获取
func (s *String) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return s.client.MGet(ctx, keys...).Result()
}

// MSet 批量设置
func (s *String) MSet(ctx context.Context, pairs ...interface{}) error {
	return s.client.MSet(ctx, pairs...).Err()
}

// MSetNX 批量设置（仅当所有键都不存在时）
func (s *String) MSetNX(ctx context.Context, pairs ...interface{}) (bool, error) {
	return s.client.MSetNX(ctx, pairs...).Result()
}

// GetSet 获取旧值并设置新值
func (s *String) GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	return s.client.GetSet(ctx, key, value).Result()
}

// BitCount 计算字符串中设置位的数量
func (s *String) BitCount(ctx context.Context, key string, bitCount *redis.BitCount) (int64, error) {
	return s.client.BitCount(ctx, key, bitCount).Result()
}

// BitOpAnd 对多个字符串执行按位AND操作
func (s *String) BitOpAnd(ctx context.Context, destKey string, keys ...string) (int64, error) {
	return s.client.BitOpAnd(ctx, destKey, keys...).Result()
}

// BitOpOr 对多个字符串执行按位OR操作
func (s *String) BitOpOr(ctx context.Context, destKey string, keys ...string) (int64, error) {
	return s.client.BitOpOr(ctx, destKey, keys...).Result()
}

// BitOpXor 对多个字符串执行按位XOR操作
func (s *String) BitOpXor(ctx context.Context, destKey string, keys ...string) (int64, error) {
	return s.client.BitOpXor(ctx, destKey, keys...).Result()
}

// BitOpNot 对字符串执行按位NOT操作
func (s *String) BitOpNot(ctx context.Context, destKey, key string) (int64, error) {
	return s.client.BitOpNot(ctx, destKey, key).Result()
}

// BitPos 查找第一个设置位或清除位的位置
func (s *String) BitPos(ctx context.Context, key string, bit int64, pos ...int64) (int64, error) {
	return s.client.BitPos(ctx, key, bit, pos...).Result()
}

// BitField 对字符串执行位域操作
func (s *String) BitField(ctx context.Context, key string, args ...interface{}) ([]int64, error) {
	return s.client.BitField(ctx, key, args...).Result()
}

// SetBit 设置位
func (s *String) SetBit(ctx context.Context, key string, offset int64, value int) (int64, error) {
	return s.client.SetBit(ctx, key, offset, value).Result()
}

// GetBit 获取位
func (s *String) GetBit(ctx context.Context, key string, offset int64) (int64, error) {
	return s.client.GetBit(ctx, key, offset).Result()
}
