package redis

import (
	"context"
	"testing"
)

func TestClient(t *testing.T) {
	// 使用默认配置创建客户端
	config := DefaultConfig()
	config.Addr = "localhost:6379"
	config.Password = ""
	config.DB = 0

	client, err := New(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 测试连接
	err = client.Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}

	// 测试基本操作
	key := "test:key"
	value := "test:value"

	// 设置键值
	stringOps := client.NewString()
	err = stringOps.Set(ctx, key, value, 0)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	// 获取值
	result, err := stringOps.Get(ctx, key)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}

	// 检查键是否存在
	exists, err := client.Exists(ctx, key)
	if err != nil {
		t.Errorf("Exists failed: %v", err)
	}
	if exists != 1 {
		t.Errorf("Expected key to exist, got %d", exists)
	}

	// 删除键
	deleted, err := client.Del(ctx, key)
	if err != nil {
		t.Errorf("Del failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("Expected 1 key deleted, got %d", deleted)
	}
}

func TestClientFromURL(t *testing.T) {
	// 测试从URL创建客户端
	url := "redis://localhost:6379/0"
	client, err := NewFromURL(url)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 测试连接
	err = client.Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestClientConfig(t *testing.T) {
	config := DefaultConfig()
	config.Addr = "localhost:6379"

	client, err := New(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	// 测试获取配置
	clientConfig := client.GetConfig()
	if clientConfig.Addr != config.Addr {
		t.Errorf("Expected addr %s, got %s", config.Addr, clientConfig.Addr)
	}

	// 测试获取统计信息
	stats := client.GetStats()
	if stats == nil {
		t.Error("Expected stats to be non-nil")
	}
}

func TestClientSetConfig(t *testing.T) {
	config := DefaultConfig()
	config.Addr = "localhost:6379"

	client, err := New(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	// 测试更新配置
	newConfig := DefaultConfig()
	newConfig.Addr = "localhost:6379"
	newConfig.DB = 1

	err = client.SetConfig(newConfig)
	if err != nil {
		t.Errorf("SetConfig failed: %v", err)
	}

	// 验证配置已更新
	clientConfig := client.GetConfig()
	if clientConfig.DB != newConfig.DB {
		t.Errorf("Expected DB %d, got %d", newConfig.DB, clientConfig.DB)
	}
}
