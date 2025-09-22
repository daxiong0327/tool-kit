package redis

import (
	"context"
	"testing"
	"time"
)

func TestString(t *testing.T) {
	config := DefaultConfig()
	config.Addr = "localhost:6379"

	client, err := New(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	stringOps := client.NewString()

	// 测试基本设置和获取
	key := "test:string:key"
	value := "test:string:value"

	err = stringOps.Set(ctx, key, value, 0)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	result, err := stringOps.Get(ctx, key)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}

	// 测试过期时间
	expireKey := "test:string:expire"
	err = stringOps.Set(ctx, expireKey, value, 100*time.Millisecond)
	if err != nil {
		t.Errorf("Set with expiration failed: %v", err)
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	_, err = stringOps.Get(ctx, expireKey)
	if err == nil {
		t.Error("Expected key to be expired")
	}

	// 测试SetNX
	nxKey := "test:string:nx"
	success, err := stringOps.SetNX(ctx, nxKey, value, 0)
	if err != nil {
		t.Errorf("SetNX failed: %v", err)
	}
	if !success {
		t.Error("Expected SetNX to succeed")
	}

	// 再次尝试SetNX应该失败
	success, err = stringOps.SetNX(ctx, nxKey, "newvalue", 0)
	if err != nil {
		t.Errorf("SetNX failed: %v", err)
	}
	if success {
		t.Error("Expected SetNX to fail on existing key")
	}

	// 测试SetXX
	xxKey := "test:string:xx"
	success, err = stringOps.SetXX(ctx, xxKey, value, 0)
	if err != nil {
		t.Errorf("SetXX failed: %v", err)
	}
	if success {
		t.Error("Expected SetXX to fail on non-existing key")
	}

	// 先设置键，再测试SetXX
	err = stringOps.Set(ctx, xxKey, "initial", 0)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	success, err = stringOps.SetXX(ctx, xxKey, value, 0)
	if err != nil {
		t.Errorf("SetXX failed: %v", err)
	}
	if !success {
		t.Error("Expected SetXX to succeed on existing key")
	}

	// 测试数值操作
	intKey := "test:string:int"
	err = stringOps.Set(ctx, intKey, "10", 0)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	// 测试Incr
	count, err := stringOps.Incr(ctx, intKey)
	if err != nil {
		t.Errorf("Incr failed: %v", err)
	}
	if count != 11 {
		t.Errorf("Expected 11, got %d", count)
	}

	// 测试IncrBy
	count, err = stringOps.IncrBy(ctx, intKey, 5)
	if err != nil {
		t.Errorf("IncrBy failed: %v", err)
	}
	if count != 16 {
		t.Errorf("Expected 16, got %d", count)
	}

	// 测试Decr
	count, err = stringOps.Decr(ctx, intKey)
	if err != nil {
		t.Errorf("Decr failed: %v", err)
	}
	if count != 15 {
		t.Errorf("Expected 15, got %d", count)
	}

	// 测试DecrBy
	count, err = stringOps.DecrBy(ctx, intKey, 5)
	if err != nil {
		t.Errorf("DecrBy failed: %v", err)
	}
	if count != 10 {
		t.Errorf("Expected 10, got %d", count)
	}

	// 测试Append
	appendKey := "test:string:append"
	err = stringOps.Set(ctx, appendKey, "hello", 0)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	length, err := stringOps.Append(ctx, appendKey, " world")
	if err != nil {
		t.Errorf("Append failed: %v", err)
	}
	if length != 11 {
		t.Errorf("Expected length 11, got %d", length)
	}

	result, err = stringOps.Get(ctx, appendKey)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if result != "hello world" {
		t.Errorf("Expected 'hello world', got %s", result)
	}

	// 测试StrLen
	length, err = stringOps.StrLen(ctx, appendKey)
	if err != nil {
		t.Errorf("StrLen failed: %v", err)
	}
	if length != 11 {
		t.Errorf("Expected length 11, got %d", length)
	}

	// 测试GetRange
	rangeResult, err := stringOps.GetRange(ctx, appendKey, 0, 4)
	if err != nil {
		t.Errorf("GetRange failed: %v", err)
	}
	if rangeResult != "hello" {
		t.Errorf("Expected 'hello', got %s", rangeResult)
	}

	// 测试SetRange
	_, err = stringOps.SetRange(ctx, appendKey, 6, "Redis")
	if err != nil {
		t.Errorf("SetRange failed: %v", err)
	}

	result, err = stringOps.Get(ctx, appendKey)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if result != "hello Redis" {
		t.Errorf("Expected 'hello Redis', got %s", result)
	}

	// 清理
	client.Del(ctx, key, nxKey, xxKey, intKey, appendKey)
}
