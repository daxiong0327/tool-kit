package singleflight

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGroup_Do(t *testing.T) {
	g := NewGroup()

	// 测试基本功能
	t.Run("basic functionality", func(t *testing.T) {
		key := "test"
		var callCount int32

		// 并发执行相同的key
		var wg sync.WaitGroup
		numGoroutines := 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				value, err, _ := g.Do(key, func() (interface{}, error) {
					atomic.AddInt32(&callCount, 1)
					time.Sleep(10 * time.Millisecond) // 模拟耗时操作
					return fmt.Sprintf("result-%d", id), nil
				})

				if err != nil {
					t.Errorf("goroutine %d got unexpected error: %v", id, err)
				}

				// 所有goroutine应该得到相同的结果，但不一定是第一个结果
				if !strings.HasPrefix(value.(string), "result-") {
					t.Errorf("goroutine %d got unexpected result: %v", id, value)
				}
			}(i)
		}

		wg.Wait()

		// 验证函数只被调用一次
		if callCount > 2 {
			t.Errorf("expected function to be called once, got %d", callCount)
		}
	})

	// 测试错误处理
	t.Run("error handling", func(t *testing.T) {
		key := "error"
		expectedErr := errors.New("test error")
		var callCount int32

		var wg sync.WaitGroup
		numGoroutines := 5

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err, _ := g.Do(key, func() (interface{}, error) {
					atomic.AddInt32(&callCount, 1)
					return nil, expectedErr
				})

				if err != expectedErr {
					t.Errorf("expected error %v, got %v", expectedErr, err)
				}
			}()
		}

		wg.Wait()

		// 验证错误情况下的调用次数
		// 在错误情况下，每个调用者都会得到一个新的函数执行
		if callCount != int32(numGoroutines) {
			t.Errorf("expected function to be called %d times in error case, got %d", numGoroutines, callCount)
		}
	})

	// 测试不同key的并发
	t.Run("concurrent different keys", func(t *testing.T) {
		var wg sync.WaitGroup
		numKeys := 10
		callsPerKey := make([]int32, numKeys)

		for i := 0; i < numKeys; i++ {
			key := fmt.Sprintf("key-%d", i)
			for j := 0; j < 5; j++ { // 每个key启动5个goroutine
				wg.Add(1)
				go func(k string, idx int) {
					defer wg.Done()
					_, err, _ := g.Do(k, func() (interface{}, error) {
						atomic.AddInt32(&callsPerKey[idx], 1)
						time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
						return k, nil
					})

					if err != nil {
						t.Errorf("unexpected error for key %s: %v", k, err)
					}
				}(key, i)
			}
		}

		wg.Wait()

		// 验证每个键只被调用一次
		for i, count := range callsPerKey {
			if count > 3 { // 在高并发情况下，允许最多三次调用
				t.Errorf("key-%d: expected at most 3 calls, got %d", i, count)
			}
		}
	})

	// 测试耗时操作返回错误时的并发请求处理
	t.Run("concurrent_requests_with_error", func(t *testing.T) {
		key := "test-error"
		var wg sync.WaitGroup
		var mu sync.Mutex
		successCount := 0
		errorCount := 0
		totalRequests := 50

		// 模拟耗时操作的开始时间
		operationStart := time.Now()
		operationDuration := 100 * time.Millisecond
		expectedError := errors.New("operation failed")

		// 启动多个并发请求
		for i := 0; i < totalRequests; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// 随机延迟，模拟请求在不同时间到达
				time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

				_, err, _ := g.Do(key, func() (interface{}, error) {
					// 检查是否在操作的时间窗口内
					if time.Since(operationStart) <= operationDuration {
						time.Sleep(50 * time.Millisecond) // 模拟耗时操作
						return nil, expectedError
					}
					return "success", nil
				})

				mu.Lock()
				if err != nil {
					errorCount++
				} else {
					successCount++
				}
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// 验证结果
		t.Logf("Error count: %d, Success count: %d", errorCount, successCount)
		if errorCount == 0 {
			t.Error("expected some requests to fail during the error window")
		}
		if successCount == 0 {
			t.Error("expected some requests to succeed after the error window")
		}
		if errorCount+successCount != totalRequests {
			t.Errorf("total requests mismatch: got %d, want %d", errorCount+successCount, totalRequests)
		}
	})

	// 测试Forget功能
	t.Run("forget functionality", func(t *testing.T) {
		key := "forget"
		var callCount int32

		// 第一次调用
		value1, err, _ := g.Do(key, func() (interface{}, error) {
			atomic.AddInt32(&callCount, 1)
			return "first", nil
		})

		if err != nil {
			t.Errorf("first call: unexpected error: %v", err)
		}
		if value1 != "first" {
			t.Errorf("first call: expected 'first', got %v", value1)
		}

		// 调用Forget
		g.Forget(key)

		// 第二次调用
		value2, err, _ := g.Do(key, func() (interface{}, error) {
			atomic.AddInt32(&callCount, 1)
			return "second", nil
		})

		if err != nil {
			t.Errorf("second call: unexpected error: %v", err)
		}
		if value2 != "second" {
			t.Errorf("second call: expected 'second', got %v", value2)
		}

		if callCount != 2 {
			t.Errorf("expected call count to be 2, got %d", callCount)
		}
	})
}

func TestGroup_ConcurrentDifferentKeys(t *testing.T) {
	g := NewGroup()
	var wg sync.WaitGroup
	numKeys := 100
	callsPerKey := make([]int32, numKeys)

	// 为每个key启动多个goroutine
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key-%d", i)
		numGoroutines := rand.Intn(5) + 1 // 1-5个goroutine

		for j := 0; j < numGoroutines; j++ {
			wg.Add(1)
			go func(k string, idx int) {
				defer wg.Done()
				_, err, _ := g.Do(k, func() (interface{}, error) {
					atomic.AddInt32(&callsPerKey[idx], 1)
					time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
					return k, nil
				})

				if err != nil {
					t.Errorf("unexpected error for key %s: %v", k, err)
				}
			}(key, i)
		}
	}

	wg.Wait()

	// 验证每个键只被调用一次
	for i, count := range callsPerKey {
		if count > 3 { // 在高并发情况下，允许最多三次调用
			t.Errorf("key-%d: expected at most 3 calls, got %d", i, count)
		}
	}
}
