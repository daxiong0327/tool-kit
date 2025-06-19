package singleflight

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
)

// 测试缓存
func TestCachedGroup_Basic(t *testing.T) {
	g := NewCachedGroup(time.Minute, time.Minute)

	// Test basic functionality
	v, found := g.Get("key")
	if found {
		t.Error("expected key not found")
	}
	if v != nil {
		t.Error("expected nil value")
	}

	g.Set("key", "value", time.Minute)
	v, found = g.Get("key")
	if !found {
		t.Error("expected key found")
	}
	if v != "value" {
		t.Errorf("expected value, got %v", v)
	}

	g.Delete("key")
	v, found = g.Get("key")
	if found {
		t.Error("expected key not found after delete")
	}
}

func TestCachedGroup_Do(t *testing.T) {
	g := NewCachedGroup(time.Minute, time.Minute)
	var callCount int32

	fn := func() (interface{}, error) {
		atomic.AddInt32(&callCount, 1)
		return "result", nil
	}

	// First call
	result := g.Do("key", time.Minute, fn)
	if result.Err != nil {
		t.Errorf("unexpected error: %v", result.Err)
	}
	if result.Value != "result" {
		t.Errorf("expected result, got %v", result.Value)
	}
	if result.FromCache {
		t.Error("should not be from cache")
	}
	if count := atomic.LoadInt32(&callCount); count != 1 {
		t.Errorf("expected 1 call, got %d", count)
	}

	// Second call should use cache
	result = g.Do("key", time.Minute, fn)
	if !result.FromCache {
		t.Error("should be from cache")
	}
	if count := atomic.LoadInt32(&callCount); count != 1 {
		t.Errorf("expected 1 call, got %d", count)
	}
}

// 测试并发调用
func TestCachedGroup_DoConcurrent(t *testing.T) {
	g := NewCachedGroup(time.Minute, time.Minute)
	var callCount int32

	fn := func() (interface{}, error) {
		time.Sleep(100 * time.Millisecond) // Simulate work
		atomic.AddInt32(&callCount, 1)
		return "result", nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := g.Do("key", time.Minute, fn)
			if result.Err != nil {
				t.Errorf("unexpected error: %v", result.Err)
			}
			if result.Value != "result" {
				t.Errorf("expected result, got %v", result.Value)
			}
		}()
	}
	wg.Wait()

	if count := atomic.LoadInt32(&callCount); count != 1 {
		t.Errorf("expected 1 call, got %d", count)
	}
}

// 测试错误
func TestCachedGroup_DoError(t *testing.T) {
	g := NewCachedGroup(time.Minute, time.Minute)
	expectedErr := errors.New("test error")

	fn := func() (interface{}, error) {
		return nil, expectedErr
	}

	result := g.Do("key", time.Minute, fn)
	if result.Err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, result.Err)
	}

	// Error results should not be cached
	v, found := g.Get("key")
	if found {
		t.Error("error result should not be cached")
	}
	if v != nil {
		t.Error("expected nil value")
	}
}

func TestCachedGroup_DoWithFallback(t *testing.T) {
	g := NewCachedGroup(time.Minute, time.Minute)
	expectedErr := errors.New("test error")

	fn := func() (interface{}, error) {
		return nil, expectedErr
	}

	fallback := "fallback"
	result := g.DoWithFallback("key", time.Minute, fn, fallback)
	if result != fallback {
		t.Errorf("expected fallback value, got %v", result)
	}

	// Success case
	fn = func() (interface{}, error) {
		return "success", nil
	}
	result = g.DoWithFallback("key2", time.Minute, fn, fallback)
	if result != "success" {
		t.Errorf("expected success value, got %v", result)
	}
}

func TestCachedGroup_Expiration(t *testing.T) {
	g := NewCachedGroup(time.Minute, time.Minute)
	var callCount int32

	fn := func() (interface{}, error) {
		atomic.AddInt32(&callCount, 1)
		return "result", nil
	}

	// First call
	g.Do("key", 100*time.Millisecond, fn)
	if count := atomic.LoadInt32(&callCount); count != 1 {
		t.Errorf("expected 1 call, got %d", count)
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should call again after expiration
	g.Do("key", time.Minute, fn)
	if count := atomic.LoadInt32(&callCount); count != 2 {
		t.Errorf("expected 2 calls, got %d", count)
	}
}

func TestCachedGroup_Forget(t *testing.T) {
	g := NewCachedGroup(time.Minute, time.Minute)
	var callCount int32

	fn := func() (interface{}, error) {
		atomic.AddInt32(&callCount, 1)
		return "result", nil
	}

	// First call
	g.Do("key", time.Minute, fn)
	if count := atomic.LoadInt32(&callCount); count != 1 {
		t.Errorf("expected 1 call, got %d", count)
	}

	// Forget the key
	g.Forget("key")

	// Should call again after forgetting
	g.Do("key", time.Minute, fn)
	if count := atomic.LoadInt32(&callCount); count != 2 {
		t.Errorf("expected 2 calls, got %d", count)
	}
}

func TestCachedGroup_FlushCache(t *testing.T) {
	g := NewCachedGroup(time.Minute, time.Minute)
	var callCount int32

	fn := func() (interface{}, error) {
		atomic.AddInt32(&callCount, 1)
		return "result", nil
	}

	// Set up multiple keys
	g.Do("key1", time.Minute, fn)
	g.Do("key2", time.Minute, fn)
	if count := atomic.LoadInt32(&callCount); count != 2 {
		t.Errorf("expected 2 calls, got %d", count)
	}

	// Flush cache
	g.FlushCache()

	// Should call again for both keys
	g.Do("key1", time.Minute, fn)
	g.Do("key2", time.Minute, fn)
	if count := atomic.LoadInt32(&callCount); count != 4 {
		t.Errorf("expected 4 calls, got %d", count)
	}
}

func TestCachedGroup_GetCacheStats(t *testing.T) {
	g := NewCachedGroup(time.Minute, time.Minute)

	// Add some items
	g.Set("key1", "value1", time.Minute)
	g.Set("key2", "value2", time.Minute)

	items, _, _ := g.GetCacheStats()
	if items != 2 {
		t.Errorf("expected 2 items, got %d", items)
	}

	// Remove an item
	g.Delete("key1")
	items, _, _ = g.GetCacheStats()
	if items != 1 {
		t.Errorf("expected 1 item, got %d", items)
	}

	// Flush cache
	g.FlushCache()
	items, _, _ = g.GetCacheStats()
	if items != 0 {
		t.Errorf("expected 0 items, got %d", items)
	}
}

// 测试shared标志
func TestCachedGroup_Shared(t *testing.T) {
	g := NewCachedGroup(time.Minute, time.Minute)
	key := "test_key"
	var sharedCount int32
	var fromCacheCount int32
	var doneCh = make(chan struct{})
	var firstWg sync.WaitGroup

	// 启动第一个请求，但让它在执行函数时等待
	firstWg.Add(1)
	go func() {
		defer firstWg.Done()
		result := g.Do(key, time.Minute, func() (interface{}, error) {
			// 通知其他goroutine可以开始了
			close(doneCh)
			// 等待一段时间让其他请求进入
			time.Sleep(50 * time.Millisecond)
			return "value", nil
		})

		if result.Err != nil {
			t.Errorf("unexpected error: %v", result.Err)
		}
		if result.Value != "value" {
			t.Errorf("unexpected value: %v", result.Value)
		}
		if result.FromCache {
			t.Error("First request should not be from cache")
		}
	}()

	// 等待第一个请求开始执行
	<-doneCh

	// 并发执行多个请求
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := g.Do(key, time.Minute, func() (interface{}, error) {
				return "unexpected", nil // 这个函数不应该被调用
			})

			if result.Err != nil {
				t.Errorf("unexpected error: %v", result.Err)
			}
			if result.Value != "value" {
				t.Errorf("unexpected value: %v", result.Value)
			}
			if result.Shared {
				atomic.AddInt32(&sharedCount, 1)
			}
			if result.FromCache {
				atomic.AddInt32(&fromCacheCount, 1)
			}
		}()
	}

	// 等待所有请求完成
	wg.Wait()
	firstWg.Wait()

	// 检查结果
	// 1. 如果请求在第一个请求完成之前进入，它们会共享结果
	// 2. 如果请求在第一个请求完成之后进入，它们会从缓存获取结果
	totalCount := atomic.LoadInt32(&sharedCount) + atomic.LoadInt32(&fromCacheCount)
	if totalCount != 3 {
		t.Errorf("Expected 3 results (either shared or from cache), got shared=%d, fromCache=%d",
			atomic.LoadInt32(&sharedCount), atomic.LoadInt32(&fromCacheCount))
	}

	// 验证每个请求要么是共享的，要么是从缓存获取的
	for i := 0; i < 3; i++ {
		result := g.Do(key, time.Minute, func() (interface{}, error) {
			t.Error("Function should not be called for cached value")
			return nil, nil
		})
		if !result.FromCache {
			t.Error("Subsequent requests should get result from cache")
		}
	}
}

func TestCachedGroup_DefaultExpiration(t *testing.T) {
	// 使用较短的默认过期时间
	g := NewCachedGroup(100*time.Millisecond, time.Minute)

	// 设置一个值但不指定过期时间（使用默认过期时间）
	g.Set("key", "value", cache.DefaultExpiration)

	// 立即获取，应该能获取到
	v, found := g.Get("key")
	if !found {
		t.Error("expected to find key immediately")
	}
	if v != "value" {
		t.Errorf("expected value, got %v", v)
	}

	// 等待超过默认过期时间
	time.Sleep(150 * time.Millisecond)

	// 再次获取，应该已经过期
	v, found = g.Get("key")
	if found {
		t.Error("expected key to be expired")
	}
}

func TestCachedGroup_CleanupInterval(t *testing.T) {
	// 使用较短的清理间隔
	g := NewCachedGroup(100*time.Millisecond, 200*time.Millisecond)

	// 记录初始状态
	initialItems, _, _ := g.GetCacheStats()

	// 添加一些带过期时间的项目
	g.Set("key1", "value1", 100*time.Millisecond)
	g.Set("key2", "value2", 100*time.Millisecond)
	g.Set("key3", "value3", time.Minute) // 这个不会很快过期

	// 等待足够的时间让前两个键过期，并让清理程序运行
	time.Sleep(400 * time.Millisecond)

	// 检查缓存统计
	currentItems, _, _ := g.GetCacheStats()
	if currentItems != initialItems+1 {
		t.Errorf("expected %d items after cleanup, got %d", initialItems+1, currentItems)
	}

	// 确认具体的键是否按预期过期和清理
	_, found1 := g.Get("key1")
	_, found2 := g.Get("key2")
	_, found3 := g.Get("key3")

	if found1 || found2 {
		t.Error("expected key1 and key2 to be cleaned up")
	}
	if !found3 {
		t.Error("expected key3 to still exist")
	}
}

func TestCachedGroup_ExpirationAndCleanup(t *testing.T) {
	// 使用短的过期时间和长的清理间隔
	g := NewCachedGroup(100*time.Millisecond, 10*time.Minute)

	// 设置一个值
	g.Set("key", "value", cache.DefaultExpiration) // 使用默认过期时间（100ms）

	// 立即获取，应该能获取到
	v, found := g.Get("key")
	if !found {
		t.Error("expected to find key immediately")
	}
	if v != "value" {
		t.Errorf("expected value, got %v", v)
	}

	// 等待过期时间过后，但在清理间隔之前
	time.Sleep(150 * time.Millisecond)

	// 尝试获取，即使还没到清理时间，也应该获取不到
	v, found = g.Get("key")
	if found {
		t.Error("expected key to be expired even before cleanup")
	}

	// 检查缓存统计，此时项目虽然过期，但因为还没被清理，所以还在计数中
	items, _, _ := g.GetCacheStats()
	if items == 0 {
		t.Error("item should still be counted before cleanup")
	}

	// 手动触发清理
	g.FlushCache()

	// 再次检查缓存统计，这时项目应该被清理掉了
	items, _, _ = g.GetCacheStats()
	if items != 0 {
		t.Error("item should be removed after flush")
	}
}
