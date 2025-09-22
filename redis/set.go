package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Set 集合操作
type Set struct {
	client *redis.Client
}

// NewSet 创建集合操作实例
func (c *Client) NewSet() *Set {
	return &Set{client: c.client}
}

// SAdd 添加元素到集合
func (s *Set) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return s.client.SAdd(ctx, key, members...).Result()
}

// SRem 从集合中移除元素
func (s *Set) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return s.client.SRem(ctx, key, members...).Result()
}

// SPop 随机弹出元素
func (s *Set) SPop(ctx context.Context, key string) (string, error) {
	return s.client.SPop(ctx, key).Result()
}

// SPopN 随机弹出多个元素
func (s *Set) SPopN(ctx context.Context, key string, count int64) ([]string, error) {
	return s.client.SPopN(ctx, key, count).Result()
}

// SRandMember 随机获取元素
func (s *Set) SRandMember(ctx context.Context, key string) (string, error) {
	return s.client.SRandMember(ctx, key).Result()
}

// SRandMemberN 随机获取多个元素
func (s *Set) SRandMemberN(ctx context.Context, key string, count int64) ([]string, error) {
	return s.client.SRandMemberN(ctx, key, count).Result()
}

// SMove 移动元素到另一个集合
func (s *Set) SMove(ctx context.Context, source, destination string, member interface{}) (bool, error) {
	return s.client.SMove(ctx, source, destination, member).Result()
}

// SCard 获取集合大小
func (s *Set) SCard(ctx context.Context, key string) (int64, error) {
	return s.client.SCard(ctx, key).Result()
}

// SIsMember 检查元素是否在集合中
func (s *Set) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return s.client.SIsMember(ctx, key, member).Result()
}

// SMIsMember 批量检查元素是否在集合中
func (s *Set) SMIsMember(ctx context.Context, key string, members ...interface{}) ([]bool, error) {
	return s.client.SMIsMember(ctx, key, members...).Result()
}

// SMembers 获取所有成员
func (s *Set) SMembers(ctx context.Context, key string) ([]string, error) {
	return s.client.SMembers(ctx, key).Result()
}

// SScan 扫描集合成员
func (s *Set) SScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return s.client.SScan(ctx, key, cursor, match, count).Result()
}

// SScanAll 扫描所有集合成员
func (s *Set) SScanAll(ctx context.Context, key string, match string, count int64) ([]string, error) {
	var allMembers []string
	cursor := uint64(0)

	for {
		members, nextCursor, err := s.SScan(ctx, key, cursor, match, count)
		if err != nil {
			return nil, err
		}

		allMembers = append(allMembers, members...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return allMembers, nil
}

// SUnion 并集
func (s *Set) SUnion(ctx context.Context, keys ...string) ([]string, error) {
	return s.client.SUnion(ctx, keys...).Result()
}

// SUnionStore 并集并存储
func (s *Set) SUnionStore(ctx context.Context, destination string, keys ...string) (int64, error) {
	return s.client.SUnionStore(ctx, destination, keys...).Result()
}

// SInter 交集
func (s *Set) SInter(ctx context.Context, keys ...string) ([]string, error) {
	return s.client.SInter(ctx, keys...).Result()
}

// SInterStore 交集并存储
func (s *Set) SInterStore(ctx context.Context, destination string, keys ...string) (int64, error) {
	return s.client.SInterStore(ctx, destination, keys...).Result()
}

// SDiff 差集
func (s *Set) SDiff(ctx context.Context, keys ...string) ([]string, error) {
	return s.client.SDiff(ctx, keys...).Result()
}

// SDiffStore 差集并存储
func (s *Set) SDiffStore(ctx context.Context, destination string, keys ...string) (int64, error) {
	return s.client.SDiffStore(ctx, destination, keys...).Result()
}

// SInterCard 交集基数
func (s *Set) SInterCard(ctx context.Context, limit int64, keys ...string) (int64, error) {
	return s.client.SInterCard(ctx, limit, keys...).Result()
}
