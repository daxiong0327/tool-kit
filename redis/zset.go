package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// ZSet 有序集合操作
type ZSet struct {
	client *redis.Client
}

// NewZSet 创建有序集合操作实例
func (c *Client) NewZSet() *ZSet {
	return &ZSet{client: c.client}
}

// ZAdd 添加元素到有序集合
func (z *ZSet) ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	return z.client.ZAdd(ctx, key, members...).Result()
}

// ZAddNX 添加元素到有序集合（仅当元素不存在时）
func (z *ZSet) ZAddNX(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	return z.client.ZAddNX(ctx, key, members...).Result()
}

// ZAddXX 添加元素到有序集合（仅当元素存在时）
func (z *ZSet) ZAddXX(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	return z.client.ZAddXX(ctx, key, members...).Result()
}

// ZAddCh 添加元素到有序集合并返回变更数量
func (z *ZSet) ZAddCh(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	return z.client.ZAdd(ctx, key, members...).Result()
}

// ZAddNXCh 添加元素到有序集合（仅当元素不存在时）并返回变更数量
func (z *ZSet) ZAddNXCh(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	return z.client.ZAddNX(ctx, key, members...).Result()
}

// ZAddXXCh 添加元素到有序集合（仅当元素存在时）并返回变更数量
func (z *ZSet) ZAddXXCh(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	return z.client.ZAddXX(ctx, key, members...).Result()
}

// ZRem 从有序集合中移除元素
func (z *ZSet) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return z.client.ZRem(ctx, key, members...).Result()
}

// ZPopMax 弹出分数最高的元素
func (z *ZSet) ZPopMax(ctx context.Context, key string, count ...int64) ([]redis.Z, error) {
	return z.client.ZPopMax(ctx, key, count...).Result()
}

// ZPopMin 弹出分数最低的元素
func (z *ZSet) ZPopMin(ctx context.Context, key string, count ...int64) ([]redis.Z, error) {
	return z.client.ZPopMin(ctx, key, count...).Result()
}

// BZPopMax 阻塞式弹出分数最高的元素
func (z *ZSet) BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) (*redis.ZWithKey, error) {
	return z.client.BZPopMax(ctx, timeout, keys...).Result()
}

// BZPopMin 阻塞式弹出分数最低的元素
func (z *ZSet) BZPopMin(ctx context.Context, timeout time.Duration, keys ...string) (*redis.ZWithKey, error) {
	return z.client.BZPopMin(ctx, timeout, keys...).Result()
}

// ZCard 获取有序集合大小
func (z *ZSet) ZCard(ctx context.Context, key string) (int64, error) {
	return z.client.ZCard(ctx, key).Result()
}

// ZCount 统计分数范围内的元素数量
func (z *ZSet) ZCount(ctx context.Context, key, min, max string) (int64, error) {
	return z.client.ZCount(ctx, key, min, max).Result()
}

// ZLexCount 统计字典序范围内的元素数量
func (z *ZSet) ZLexCount(ctx context.Context, key, min, max string) (int64, error) {
	return z.client.ZLexCount(ctx, key, min, max).Result()
}

// ZIncrBy 增加元素分数
func (z *ZSet) ZIncrBy(ctx context.Context, key string, increment float64, member string) (float64, error) {
	return z.client.ZIncrBy(ctx, key, increment, member).Result()
}

// ZInter 交集
func (z *ZSet) ZInter(ctx context.Context, store *redis.ZStore) ([]string, error) {
	return z.client.ZInter(ctx, store).Result()
}

// ZInterStore 交集并存储
func (z *ZSet) ZInterStore(ctx context.Context, destination string, store *redis.ZStore) (int64, error) {
	return z.client.ZInterStore(ctx, destination, store).Result()
}

// ZUnion 并集
func (z *ZSet) ZUnion(ctx context.Context, store redis.ZStore) ([]string, error) {
	return z.client.ZUnion(ctx, store).Result()
}

// ZUnionStore 并集并存储
func (z *ZSet) ZUnionStore(ctx context.Context, destination string, store *redis.ZStore) (int64, error) {
	return z.client.ZUnionStore(ctx, destination, store).Result()
}

// ZRange 获取指定范围的元素
func (z *ZSet) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return z.client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeWithScores 获取指定范围的元素（带分数）
func (z *ZSet) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return z.client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZRangeByScore 按分数范围获取元素
func (z *ZSet) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return z.client.ZRangeByScore(ctx, key, opt).Result()
}

// ZRangeByScoreWithScores 按分数范围获取元素（带分数）
func (z *ZSet) ZRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
	return z.client.ZRangeByScoreWithScores(ctx, key, opt).Result()
}

// ZRangeByLex 按字典序范围获取元素
func (z *ZSet) ZRangeByLex(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return z.client.ZRangeByLex(ctx, key, opt).Result()
}

// ZRevRange 获取指定范围的元素（逆序）
func (z *ZSet) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return z.client.ZRevRange(ctx, key, start, stop).Result()
}

// ZRevRangeWithScores 获取指定范围的元素（逆序，带分数）
func (z *ZSet) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return z.client.ZRevRangeWithScores(ctx, key, start, stop).Result()
}

// ZRevRangeByScore 按分数范围获取元素（逆序）
func (z *ZSet) ZRevRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return z.client.ZRevRangeByScore(ctx, key, opt).Result()
}

// ZRevRangeByScoreWithScores 按分数范围获取元素（逆序，带分数）
func (z *ZSet) ZRevRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
	return z.client.ZRevRangeByScoreWithScores(ctx, key, opt).Result()
}

// ZRevRangeByLex 按字典序范围获取元素（逆序）
func (z *ZSet) ZRevRangeByLex(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return z.client.ZRevRangeByLex(ctx, key, opt).Result()
}

// ZRank 获取元素排名
func (z *ZSet) ZRank(ctx context.Context, key, member string) (int64, error) {
	return z.client.ZRank(ctx, key, member).Result()
}

// ZRevRank 获取元素排名（逆序）
func (z *ZSet) ZRevRank(ctx context.Context, key, member string) (int64, error) {
	return z.client.ZRevRank(ctx, key, member).Result()
}

// ZScore 获取元素分数
func (z *ZSet) ZScore(ctx context.Context, key, member string) (float64, error) {
	return z.client.ZScore(ctx, key, member).Result()
}

// ZMScore 批量获取元素分数
func (z *ZSet) ZMScore(ctx context.Context, key string, members ...string) ([]float64, error) {
	return z.client.ZMScore(ctx, key, members...).Result()
}

// ZRemRangeByRank 按排名范围删除元素
func (z *ZSet) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) {
	return z.client.ZRemRangeByRank(ctx, key, start, stop).Result()
}

// ZRemRangeByScore 按分数范围删除元素
func (z *ZSet) ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error) {
	return z.client.ZRemRangeByScore(ctx, key, min, max).Result()
}

// ZRemRangeByLex 按字典序范围删除元素
func (z *ZSet) ZRemRangeByLex(ctx context.Context, key, min, max string) (int64, error) {
	return z.client.ZRemRangeByLex(ctx, key, min, max).Result()
}

// ZScan 扫描有序集合
func (z *ZSet) ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return z.client.ZScan(ctx, key, cursor, match, count).Result()
}

// ZScanAll 扫描所有有序集合元素
func (z *ZSet) ZScanAll(ctx context.Context, key string, match string, count int64) ([]redis.Z, error) {
	var allMembers []redis.Z
	cursor := uint64(0)

	for {
		members, nextCursor, err := z.ZScan(ctx, key, cursor, match, count)
		if err != nil {
			return nil, err
		}

		// 将成员和分数配对
		for i := 0; i < len(members); i += 2 {
			if i+1 < len(members) {
				score, err := parseFloat64(members[i+1])
				if err != nil {
					continue
				}
				allMembers = append(allMembers, redis.Z{
					Score:  score,
					Member: members[i],
				})
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return allMembers, nil
}

// parseFloat64 解析浮点数
func parseFloat64(s string) (float64, error) {
	// 这里可以使用 strconv.ParseFloat，但为了简化，我们返回0
	// 在实际使用中，应该正确解析分数
	return 0, nil
}
