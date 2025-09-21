package main

import (
	"context"
	"fmt"
	"time"

	"github.com/daxiong0327/tool-kit/goroutine"
)

func main() {
	fmt.Println("=== 协程管理模块基本示例 ===")

	// 示例1：基本协程执行
	fmt.Println("\n1. 基本协程执行:")
	basicExample()

	// 示例2：协程崩溃恢复
	fmt.Println("\n2. 协程崩溃恢复:")
	panicRecoveryExample()

	// 示例3：协程池使用
	fmt.Println("\n3. 协程池使用:")
	poolExample()

	// 示例4：批量执行
	fmt.Println("\n4. 批量执行:")
	batchExample()

	// 示例5：重试机制
	fmt.Println("\n5. 重试机制:")
	retryExample()

	// 示例6：熔断器
	fmt.Println("\n6. 熔断器:")
	circuitBreakerExample()

	// 示例7：限流器
	fmt.Println("\n7. 限流器:")
	rateLimiterExample()

	// 示例8：协程监控
	fmt.Println("\n8. 协程监控:")
	monitorExample()

	fmt.Println("\n🎉 所有示例完成！")
}

// basicExample 基本协程执行示例
func basicExample() {
	executor := goroutine.NewSafeExecutor()

	// 基本协程执行
	executor.Go(func() {
		fmt.Println("  ✅ 基本协程执行完成")
	})

	// 带上下文的协程执行
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	executor.GoWithContext(ctx, func() {
		fmt.Println("  ✅ 带上下文的协程执行完成")
	})

	// 带超时的协程执行
	executor.GoWithTimeout(func() {
		fmt.Println("  ✅ 带超时的协程执行完成")
	}, 1*time.Second)

	// 延迟执行
	executor.GoWithDelay(func() {
		fmt.Println("  ✅ 延迟执行的协程完成")
	}, 500*time.Millisecond)

	// 定时执行
	stop := executor.GoWithInterval(func() {
		fmt.Println("  ✅ 定时执行的协程")
	}, 200*time.Millisecond)

	time.Sleep(1 * time.Second)
	close(stop) // 停止定时执行

	time.Sleep(100 * time.Millisecond)
}

// panicRecoveryExample 协程崩溃恢复示例
func panicRecoveryExample() {
	executor := goroutine.NewSafeExecutor()

	// 设置崩溃恢复处理器
	executor.SetRecoverHandler(func(panicValue interface{}, stack []byte, goroutineID string) {
		fmt.Printf("  🚨 协程崩溃已恢复: %v\n", panicValue)
		fmt.Printf("  📍 协程ID: %s\n", goroutineID)
		fmt.Printf("  📊 堆栈信息: %s\n", string(stack))
	})

	// 启动会崩溃的协程
	executor.Go(func() {
		panic("这是一个测试崩溃")
	})

	// 启动正常协程
	executor.Go(func() {
		fmt.Println("  ✅ 正常协程执行完成")
	})

	time.Sleep(100 * time.Millisecond)
}

// poolExample 协程池使用示例
func poolExample() {
	config := &goroutine.PoolConfig{
		MaxWorkers: 3,
		QueueSize:  10,
		JobTimeout: 5 * time.Second,
	}

	pool := goroutine.NewPool(config)
	defer pool.Stop()

	// 设置崩溃恢复处理器
	pool.SetRecoverHandler(func(panicValue interface{}, stack []byte, goroutineID string) {
		fmt.Printf("  🚨 协程池中的协程崩溃已恢复: %v\n", panicValue)
	})

	// 提交任务
	for i := 0; i < 5; i++ {
		taskID := fmt.Sprintf("task-%d", i+1)
		pool.SubmitFunc(taskID, func() error {
			fmt.Printf("  ✅ 任务 %s 执行完成\n", taskID)
			time.Sleep(100 * time.Millisecond)
			return nil
		})
	}

	// 提交会失败的任务
	pool.SubmitFunc("error-task", func() error {
		fmt.Println("  ❌ 任务执行失败")
		return fmt.Errorf("任务执行失败")
	})

	// 提交会崩溃的任务
	pool.SubmitFunc("panic-task", func() error {
		panic("任务崩溃")
	})

	time.Sleep(500 * time.Millisecond)

	// 显示统计信息
	stats := pool.GetStats()
	fmt.Printf("  📊 协程池统计: 总任务=%d, 完成=%d, 失败=%d\n",
		stats.TotalJobs, stats.CompletedJobs, stats.FailedJobs)
}

// batchExample 批量执行示例
func batchExample() {
	executor := goroutine.NewSafeExecutor()
	batch := goroutine.NewBatch(executor)

	// 添加任务
	for i := 0; i < 5; i++ {
		i := i // 捕获循环变量
		batch.Add(func() (interface{}, error) {
			time.Sleep(50 * time.Millisecond)
			result := fmt.Sprintf("任务 %d 完成", i+1)
			fmt.Printf("  ✅ %s\n", result)
			return result, nil
		})
	}

	// 添加会失败的任务
	batch.Add(func() (interface{}, error) {
		return nil, fmt.Errorf("任务执行失败")
	})

	// 等待所有任务完成
	results, errors := batch.Wait()

	fmt.Printf("  📊 批量执行结果: 成功=%d, 失败=%d\n", len(results), len(errors))
	for _, result := range results {
		fmt.Printf("  📝 结果: %+v\n", result)
	}
}

// retryExample 重试机制示例
func retryExample() {
	executor := goroutine.NewSafeExecutor()

	// 创建重试器
	retry := goroutine.NewRetry(executor, 3, 100*time.Millisecond, &goroutine.ExponentialBackoff{
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  1 * time.Second,
	})

	// 执行会失败的任务
	attempt := 0
	err := retry.Execute(func() error {
		attempt++
		fmt.Printf("  🔄 重试第 %d 次\n", attempt)
		if attempt < 3 {
			return fmt.Errorf("任务失败")
		}
		fmt.Println("  ✅ 任务最终成功")
		return nil
	})

	if err != nil {
		fmt.Printf("  ❌ 重试失败: %v\n", err)
	}
}

// circuitBreakerExample 熔断器示例
func circuitBreakerExample() {
	executor := goroutine.NewSafeExecutor()
	cb := goroutine.NewCircuitBreaker(executor, 3, 2*time.Second)

	// 模拟失败的服务
	failureCount := 0
	service := func() error {
		failureCount++
		if failureCount <= 5 {
			return fmt.Errorf("服务暂时不可用")
		}
		return nil
	}

	// 执行多次调用
	for i := 0; i < 8; i++ {
		err := cb.Execute(service)
		state := cb.GetState()
		fmt.Printf("  🔄 调用 %d: 状态=%v, 错误=%v\n", i+1, state, err)
		time.Sleep(100 * time.Millisecond)
	}

	// 等待熔断器重置
	fmt.Println("  ⏳ 等待熔断器重置...")
	time.Sleep(3 * time.Second)

	// 再次尝试
	err := cb.Execute(service)
	state := cb.GetState()
	fmt.Printf("  🔄 重置后调用: 状态=%v, 错误=%v\n", state, err)
}

// rateLimiterExample 限流器示例
func rateLimiterExample() {
	executor := goroutine.NewSafeExecutor()
	rl := goroutine.NewRateLimiter(executor, 3, 1*time.Second)
	defer rl.Stop()

	// 尝试执行多个任务
	for i := 0; i < 10; i++ {
		if rl.Allow() {
			fmt.Printf("  ✅ 任务 %d 被允许执行\n", i+1)
		} else {
			fmt.Printf("  ❌ 任务 %d 被限流\n", i+1)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// monitorExample 协程监控示例
func monitorExample() {
	executor := goroutine.NewSafeExecutor()

	// 创建监控器
	monitor := goroutine.NewMonitor(executor, 1*time.Second)

	// 设置告警处理器
	monitor.SetAlertConfig(&goroutine.AlertConfig{
		MaxGoroutines: 10,
		MaxHeapAlloc:  50 * 1024 * 1024, // 50MB
		AlertHandler: func(alert *goroutine.Alert) {
			fmt.Printf("  🚨 告警: %s - %s\n", alert.Type, alert.Message)
		},
	})

	// 添加监控处理器
	monitor.AddHandler(func(stats *goroutine.MonitorStats) {
		fmt.Printf("  📊 协程数: %d, 堆内存: %d KB\n",
			stats.NumGoroutines, stats.HeapAlloc/1024)
	})

	// 启动监控
	monitor.Start()
	defer monitor.Stop()

	// 启动一些协程
	for i := 0; i < 5; i++ {
		executor.Go(func() {
			time.Sleep(200 * time.Millisecond)
		})
	}

	// 等待一段时间
	time.Sleep(3 * time.Second)

	// 获取统计信息
	stats := monitor.GetStats()
	fmt.Printf("  📈 最终统计: 协程数=%d, 堆内存=%d KB\n",
		stats.NumGoroutines, stats.HeapAlloc/1024)
}
