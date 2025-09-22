package redis

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestPipeline(t *testing.T) {
	config := DefaultConfig()
	config.Addr = "localhost:6379"

	client, err := New(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	pipeline := client.NewPipeline()

	// 测试管道操作
	key1 := "test:pipeline:key1"
	key2 := "test:pipeline:key2"
	key3 := "test:pipeline:key3"

	// 添加命令到管道
	pipeline.Set(ctx, key1, "value1", 0)
	pipeline.Set(ctx, key2, "value2", 0)
	pipeline.Set(ctx, key3, "value3", 0)
	pipeline.Get(ctx, key1)
	pipeline.Get(ctx, key2)
	pipeline.Get(ctx, key3)

	// 执行管道
	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		t.Errorf("Pipeline exec failed: %v", err)
	}

	// 检查结果
	if len(cmds) != 6 {
		t.Errorf("Expected 6 commands, got %d", len(cmds))
	}

	// 检查设置命令结果
	for i := 0; i < 3; i++ {
		cmd := cmds[i]
		if cmd.Err() != nil {
			t.Errorf("Set command %d failed: %v", i, cmd.Err())
		}
	}

	// 检查获取命令结果
	for i := 3; i < 6; i++ {
		cmd := cmds[i]
		if cmd.Err() != nil {
			t.Errorf("Get command %d failed: %v", i, cmd.Err())
		}
	}

	// 验证值
	stringOps := client.NewString()
	result1, err := stringOps.Get(ctx, key1)
	if err != nil {
		t.Errorf("Get key1 failed: %v", err)
	}
	if result1 != "value1" {
		t.Errorf("Expected 'value1', got %s", result1)
	}

	result2, err := stringOps.Get(ctx, key2)
	if err != nil {
		t.Errorf("Get key2 failed: %v", err)
	}
	if result2 != "value2" {
		t.Errorf("Expected 'value2', got %s", result2)
	}

	result3, err := stringOps.Get(ctx, key3)
	if err != nil {
		t.Errorf("Get key3 failed: %v", err)
	}
	if result3 != "value3" {
		t.Errorf("Expected 'value3', got %s", result3)
	}

	// 清理
	client.Del(ctx, key1, key2, key3)
}

func TestPipelineDiscard(t *testing.T) {
	config := DefaultConfig()
	config.Addr = "localhost:6379"

	client, err := New(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	pipeline := client.NewPipeline()

	// 添加命令到管道
	key := "test:pipeline:discard"
	pipeline.Set(ctx, key, "value", 0)
	pipeline.Get(ctx, key)

	// 丢弃管道
	pipeline.Discard()

	// 验证键不存在
	exists, err := client.Exists(ctx, key)
	if err != nil {
		t.Errorf("Exists failed: %v", err)
	}
	if exists != 0 {
		t.Error("Expected key to not exist after discard")
	}
}

func TestPipelineMixedOperations(t *testing.T) {
	config := DefaultConfig()
	config.Addr = "localhost:6379"

	client, err := New(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	pipeline := client.NewPipeline()

	// 测试混合操作
	stringKey := "test:pipeline:string"
	hashKey := "test:pipeline:hash"
	listKey := "test:pipeline:list"
	setKey := "test:pipeline:set"
	zsetKey := "test:pipeline:zset"

	// String 操作
	pipeline.Set(ctx, stringKey, "value", 0)
	pipeline.Get(ctx, stringKey)

	// Hash 操作
	pipeline.HSet(ctx, hashKey, "field1", "value1", "field2", "value2")
	pipeline.HGet(ctx, hashKey, "field1")
	pipeline.HGetAll(ctx, hashKey)

	// List 操作
	pipeline.LPush(ctx, listKey, "item1", "item2", "item3")
	pipeline.LRange(ctx, listKey, 0, -1)
	pipeline.LLen(ctx, listKey)

	// Set 操作
	pipeline.SAdd(ctx, setKey, "member1", "member2", "member3")
	pipeline.SMembers(ctx, setKey)
	pipeline.SCard(ctx, setKey)

	// ZSet 操作
	pipeline.ZAdd(ctx, zsetKey, redis.Z{Score: 1, Member: "member1"}, redis.Z{Score: 2, Member: "member2"})
	pipeline.ZRange(ctx, zsetKey, 0, -1)
	pipeline.ZCard(ctx, zsetKey)

	// 执行管道
	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		t.Errorf("Pipeline exec failed: %v", err)
	}

	// 检查命令数量
	expectedCmds := 14 // 2 string + 3 hash + 3 list + 3 set + 3 zset
	if len(cmds) != expectedCmds {
		t.Errorf("Expected %d commands, got %d", expectedCmds, len(cmds))
	}

	// 检查是否有错误
	for i, cmd := range cmds {
		if cmd.Err() != nil {
			t.Errorf("Command %d failed: %v", i, cmd.Err())
		}
	}

	// 清理
	client.Del(ctx, stringKey, hashKey, listKey, setKey, zsetKey)
}
