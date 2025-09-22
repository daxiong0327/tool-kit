package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// List 列表操作
type List struct {
	client *redis.Client
}

// NewList 创建列表操作实例
func (c *Client) NewList() *List {
	return &List{client: c.client}
}

// LPush 从左侧推入元素
func (l *List) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return l.client.LPush(ctx, key, values...).Result()
}

// RPush 从右侧推入元素
func (l *List) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return l.client.RPush(ctx, key, values...).Result()
}

// LPushX 从左侧推入元素（仅当列表存在时）
func (l *List) LPushX(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return l.client.LPushX(ctx, key, values...).Result()
}

// RPushX 从右侧推入元素（仅当列表存在时）
func (l *List) RPushX(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return l.client.RPushX(ctx, key, values...).Result()
}

// LPop 从左侧弹出元素
func (l *List) LPop(ctx context.Context, key string) (string, error) {
	return l.client.LPop(ctx, key).Result()
}

// RPop 从右侧弹出元素
func (l *List) RPop(ctx context.Context, key string) (string, error) {
	return l.client.RPop(ctx, key).Result()
}

// LPopCount 从左侧弹出多个元素
func (l *List) LPopCount(ctx context.Context, key string, count int) ([]string, error) {
	return l.client.LPopCount(ctx, key, count).Result()
}

// RPopCount 从右侧弹出多个元素
func (l *List) RPopCount(ctx context.Context, key string, count int) ([]string, error) {
	return l.client.RPopCount(ctx, key, count).Result()
}

// BLPop 阻塞式从左侧弹出元素
func (l *List) BLPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return l.client.BLPop(ctx, timeout, keys...).Result()
}

// BRPop 阻塞式从右侧弹出元素
func (l *List) BRPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return l.client.BRPop(ctx, timeout, keys...).Result()
}

// BRPopLPush 阻塞式从右侧弹出元素并推入到另一个列表的左侧
func (l *List) BRPopLPush(ctx context.Context, source, destination string, timeout time.Duration) (string, error) {
	return l.client.BRPopLPush(ctx, source, destination, timeout).Result()
}

// LIndex 获取指定索引的元素
func (l *List) LIndex(ctx context.Context, key string, index int64) (string, error) {
	return l.client.LIndex(ctx, key, index).Result()
}

// LInsert 在指定位置插入元素
func (l *List) LInsert(ctx context.Context, key, op string, pivot, value interface{}) (int64, error) {
	return l.client.LInsert(ctx, key, op, pivot, value).Result()
}

// LInsertBefore 在指定元素前插入
func (l *List) LInsertBefore(ctx context.Context, key string, pivot, value interface{}) (int64, error) {
	return l.LInsert(ctx, key, "BEFORE", pivot, value)
}

// LInsertAfter 在指定元素后插入
func (l *List) LInsertAfter(ctx context.Context, key string, pivot, value interface{}) (int64, error) {
	return l.LInsert(ctx, key, "AFTER", pivot, value)
}

// LLen 获取列表长度
func (l *List) LLen(ctx context.Context, key string) (int64, error) {
	return l.client.LLen(ctx, key).Result()
}

// LRange 获取指定范围的元素
func (l *List) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return l.client.LRange(ctx, key, start, stop).Result()
}

// LRem 移除指定元素
func (l *List) LRem(ctx context.Context, key string, count int64, value interface{}) (int64, error) {
	return l.client.LRem(ctx, key, count, value).Result()
}

// LSet 设置指定索引的元素
func (l *List) LSet(ctx context.Context, key string, index int64, value interface{}) error {
	return l.client.LSet(ctx, key, index, value).Err()
}

// LTrim 修剪列表
func (l *List) LTrim(ctx context.Context, key string, start, stop int64) error {
	return l.client.LTrim(ctx, key, start, stop).Err()
}

// RPopLPush 从右侧弹出元素并推入到另一个列表的左侧
func (l *List) RPopLPush(ctx context.Context, source, destination string) (string, error) {
	return l.client.RPopLPush(ctx, source, destination).Result()
}

// LMove 移动元素
func (l *List) LMove(ctx context.Context, source, destination, srcpos, destpos string) (string, error) {
	return l.client.LMove(ctx, source, destination, srcpos, destpos).Result()
}

// LMoveRightLeft 从右侧弹出并推入到左侧
func (l *List) LMoveRightLeft(ctx context.Context, source, destination string) (string, error) {
	return l.LMove(ctx, source, destination, "RIGHT", "LEFT")
}

// LMoveLeftRight 从左侧弹出并推入到右侧
func (l *List) LMoveLeftRight(ctx context.Context, source, destination string) (string, error) {
	return l.LMove(ctx, source, destination, "LEFT", "RIGHT")
}

// LMoveRightRight 从右侧弹出并推入到右侧
func (l *List) LMoveRightRight(ctx context.Context, source, destination string) (string, error) {
	return l.LMove(ctx, source, destination, "RIGHT", "RIGHT")
}

// LMoveLeftLeft 从左侧弹出并推入到左侧
func (l *List) LMoveLeftLeft(ctx context.Context, source, destination string) (string, error) {
	return l.LMove(ctx, source, destination, "LEFT", "LEFT")
}

// LPos 查找元素位置
func (l *List) LPos(ctx context.Context, key string, value string, args redis.LPosArgs) (int64, error) {
	return l.client.LPos(ctx, key, value, args).Result()
}

// LPosCount 查找多个元素位置
func (l *List) LPosCount(ctx context.Context, key string, value string, count int64, args redis.LPosArgs) ([]int64, error) {
	return l.client.LPosCount(ctx, key, value, count, args).Result()
}
