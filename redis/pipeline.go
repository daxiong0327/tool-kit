package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Pipeline 管道操作
type Pipeline struct {
	pipe redis.Pipeliner
}

// NewPipeline 创建管道实例
func (c *Client) NewPipeline() *Pipeline {
	return &Pipeline{pipe: c.client.Pipeline()}
}

// Exec 执行管道命令
func (p *Pipeline) Exec(ctx context.Context) ([]redis.Cmder, error) {
	return p.pipe.Exec(ctx)
}

// Discard 丢弃管道命令
func (p *Pipeline) Discard() {
	p.pipe.Discard()
}

// String 操作
func (p *Pipeline) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return p.pipe.Set(ctx, key, value, expiration)
}

func (p *Pipeline) Get(ctx context.Context, key string) *redis.StringCmd {
	return p.pipe.Get(ctx, key)
}

func (p *Pipeline) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return p.pipe.Del(ctx, keys...)
}

func (p *Pipeline) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	return p.pipe.Exists(ctx, keys...)
}

func (p *Pipeline) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return p.pipe.Expire(ctx, key, expiration)
}

func (p *Pipeline) TTL(ctx context.Context, key string) *redis.DurationCmd {
	return p.pipe.TTL(ctx, key)
}

func (p *Pipeline) Incr(ctx context.Context, key string) *redis.IntCmd {
	return p.pipe.Incr(ctx, key)
}

func (p *Pipeline) IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd {
	return p.pipe.IncrBy(ctx, key, value)
}

func (p *Pipeline) Decr(ctx context.Context, key string) *redis.IntCmd {
	return p.pipe.Decr(ctx, key)
}

func (p *Pipeline) DecrBy(ctx context.Context, key string, value int64) *redis.IntCmd {
	return p.pipe.DecrBy(ctx, key, value)
}

func (p *Pipeline) MGet(ctx context.Context, keys ...string) *redis.SliceCmd {
	return p.pipe.MGet(ctx, keys...)
}

func (p *Pipeline) MSet(ctx context.Context, pairs ...interface{}) *redis.StatusCmd {
	return p.pipe.MSet(ctx, pairs...)
}

// Hash 操作
func (p *Pipeline) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return p.pipe.HSet(ctx, key, values...)
}

func (p *Pipeline) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return p.pipe.HGet(ctx, key, field)
}

func (p *Pipeline) HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd {
	return p.pipe.HGetAll(ctx, key)
}

func (p *Pipeline) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	return p.pipe.HDel(ctx, key, fields...)
}

func (p *Pipeline) HExists(ctx context.Context, key, field string) *redis.BoolCmd {
	return p.pipe.HExists(ctx, key, field)
}

func (p *Pipeline) HLen(ctx context.Context, key string) *redis.IntCmd {
	return p.pipe.HLen(ctx, key)
}

func (p *Pipeline) HKeys(ctx context.Context, key string) *redis.StringSliceCmd {
	return p.pipe.HKeys(ctx, key)
}

func (p *Pipeline) HVals(ctx context.Context, key string) *redis.StringSliceCmd {
	return p.pipe.HVals(ctx, key)
}

func (p *Pipeline) HMGet(ctx context.Context, key string, fields ...string) *redis.SliceCmd {
	return p.pipe.HMGet(ctx, key, fields...)
}

func (p *Pipeline) HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd {
	return p.pipe.HMSet(ctx, key, values...)
}

func (p *Pipeline) HIncrBy(ctx context.Context, key, field string, incr int64) *redis.IntCmd {
	return p.pipe.HIncrBy(ctx, key, field, incr)
}

func (p *Pipeline) HIncrByFloat(ctx context.Context, key, field string, incr float64) *redis.FloatCmd {
	return p.pipe.HIncrByFloat(ctx, key, field, incr)
}

// List 操作
func (p *Pipeline) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return p.pipe.LPush(ctx, key, values...)
}

func (p *Pipeline) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return p.pipe.RPush(ctx, key, values...)
}

func (p *Pipeline) LPop(ctx context.Context, key string) *redis.StringCmd {
	return p.pipe.LPop(ctx, key)
}

func (p *Pipeline) RPop(ctx context.Context, key string) *redis.StringCmd {
	return p.pipe.RPop(ctx, key)
}

func (p *Pipeline) LLen(ctx context.Context, key string) *redis.IntCmd {
	return p.pipe.LLen(ctx, key)
}

func (p *Pipeline) LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return p.pipe.LRange(ctx, key, start, stop)
}

func (p *Pipeline) LIndex(ctx context.Context, key string, index int64) *redis.StringCmd {
	return p.pipe.LIndex(ctx, key, index)
}

func (p *Pipeline) LSet(ctx context.Context, key string, index int64, value interface{}) *redis.StatusCmd {
	return p.pipe.LSet(ctx, key, index, value)
}

func (p *Pipeline) LRem(ctx context.Context, key string, count int64, value interface{}) *redis.IntCmd {
	return p.pipe.LRem(ctx, key, count, value)
}

func (p *Pipeline) LTrim(ctx context.Context, key string, start, stop int64) *redis.StatusCmd {
	return p.pipe.LTrim(ctx, key, start, stop)
}

// Set 操作
func (p *Pipeline) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return p.pipe.SAdd(ctx, key, members...)
}

func (p *Pipeline) SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return p.pipe.SRem(ctx, key, members...)
}

func (p *Pipeline) SPop(ctx context.Context, key string) *redis.StringCmd {
	return p.pipe.SPop(ctx, key)
}

func (p *Pipeline) SCard(ctx context.Context, key string) *redis.IntCmd {
	return p.pipe.SCard(ctx, key)
}

func (p *Pipeline) SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	return p.pipe.SIsMember(ctx, key, member)
}

func (p *Pipeline) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	return p.pipe.SMembers(ctx, key)
}

func (p *Pipeline) SUnion(ctx context.Context, keys ...string) *redis.StringSliceCmd {
	return p.pipe.SUnion(ctx, keys...)
}

func (p *Pipeline) SInter(ctx context.Context, keys ...string) *redis.StringSliceCmd {
	return p.pipe.SInter(ctx, keys...)
}

func (p *Pipeline) SDiff(ctx context.Context, keys ...string) *redis.StringSliceCmd {
	return p.pipe.SDiff(ctx, keys...)
}

// ZSet 操作
func (p *Pipeline) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	return p.pipe.ZAdd(ctx, key, members...)
}

func (p *Pipeline) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return p.pipe.ZRem(ctx, key, members...)
}

func (p *Pipeline) ZCard(ctx context.Context, key string) *redis.IntCmd {
	return p.pipe.ZCard(ctx, key)
}

func (p *Pipeline) ZCount(ctx context.Context, key, min, max string) *redis.IntCmd {
	return p.pipe.ZCount(ctx, key, min, max)
}

func (p *Pipeline) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return p.pipe.ZRange(ctx, key, start, stop)
}

func (p *Pipeline) ZRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	return p.pipe.ZRangeWithScores(ctx, key, start, stop)
}

func (p *Pipeline) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	return p.pipe.ZRangeByScore(ctx, key, opt)
}

func (p *Pipeline) ZRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.ZSliceCmd {
	return p.pipe.ZRangeByScoreWithScores(ctx, key, opt)
}

func (p *Pipeline) ZRevRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return p.pipe.ZRevRange(ctx, key, start, stop)
}

func (p *Pipeline) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	return p.pipe.ZRevRangeWithScores(ctx, key, start, stop)
}

func (p *Pipeline) ZRank(ctx context.Context, key, member string) *redis.IntCmd {
	return p.pipe.ZRank(ctx, key, member)
}

func (p *Pipeline) ZRevRank(ctx context.Context, key, member string) *redis.IntCmd {
	return p.pipe.ZRevRank(ctx, key, member)
}

func (p *Pipeline) ZScore(ctx context.Context, key, member string) *redis.FloatCmd {
	return p.pipe.ZScore(ctx, key, member)
}

func (p *Pipeline) ZIncrBy(ctx context.Context, key string, increment float64, member string) *redis.FloatCmd {
	return p.pipe.ZIncrBy(ctx, key, increment, member)
}

// 通用操作
func (p *Pipeline) Ping(ctx context.Context) *redis.StatusCmd {
	return p.pipe.Ping(ctx)
}

func (p *Pipeline) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	return p.pipe.Keys(ctx, pattern)
}

func (p *Pipeline) FlushDB(ctx context.Context) *redis.StatusCmd {
	return p.pipe.FlushDB(ctx)
}

func (p *Pipeline) FlushAll(ctx context.Context) *redis.StatusCmd {
	return p.pipe.FlushAll(ctx)
}
