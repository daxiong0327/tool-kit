package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/daxiong0327/tool-kit/goroutine"
)

// 模拟一个可能失败的服务
type Service struct {
	name        string
	successRate float64
	mu          sync.RWMutex
}

func NewService(name string, successRate float64) *Service {
	return &Service{
		name:        name,
		successRate: successRate,
	}
}

func (s *Service) Call() error {
	s.mu.RLock()
	successRate := s.successRate
	s.mu.RUnlock()

	if rand.Float64() < successRate {
		return nil
	}
	return fmt.Errorf("服务 %s 调用失败", s.name)
}

func (s *Service) SetSuccessRate(rate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.successRate = rate
}

func main() {
	fmt.Println("=== 协程管理模块高级示例 ===")

	// 示例1：微服务架构中的协程管理
	fmt.Println("\n1. 微服务架构中的协程管理:")
	microserviceExample()

	// 示例2：高并发任务处理
	fmt.Println("\n2. 高并发任务处理:")
	highConcurrencyExample()

	// 示例3：分布式任务调度
	fmt.Println("\n3. 分布式任务调度:")
	distributedSchedulingExample()

	// 示例4：实时监控和告警
	fmt.Println("\n4. 实时监控和告警:")
	realtimeMonitoringExample()

	// 示例5：优雅关闭
	fmt.Println("\n5. 优雅关闭:")
	gracefulShutdownExample()

	fmt.Println("\n🎉 高级示例完成！")
}

// microserviceExample 微服务架构中的协程管理示例
func microserviceExample() {
	// 创建服务
	userService := NewService("用户服务", 0.8)
	orderService := NewService("订单服务", 0.7)
	paymentService := NewService("支付服务", 0.9)

	// 创建协程池
	config := &goroutine.PoolConfig{
		MaxWorkers: 10,
		QueueSize:  100,
		JobTimeout: 5 * time.Second,
	}
	pool := goroutine.NewPool(config)
	defer pool.Stop()

	// 设置崩溃恢复处理器
	pool.SetRecoverHandler(func(panicValue interface{}, stack []byte, goroutineID string) {
		fmt.Printf("  🚨 微服务协程崩溃: %v (协程ID: %s)\n", panicValue, goroutineID)
	})

	// 模拟用户请求处理
	for i := 0; i < 20; i++ {
		requestID := fmt.Sprintf("req-%d", i+1)
		pool.SubmitFunc(requestID, func() error {
			return processUserRequest(requestID, userService, orderService, paymentService)
		})
	}

	time.Sleep(2 * time.Second)

	// 显示统计信息
	stats := pool.GetStats()
	fmt.Printf("  📊 微服务统计: 总请求=%d, 成功=%d, 失败=%d\n",
		stats.TotalJobs, stats.CompletedJobs, stats.FailedJobs)
}

func processUserRequest(requestID string, userService, orderService, paymentService *Service) error {
	fmt.Printf("  🔄 处理请求 %s\n", requestID)

	// 调用用户服务
	if err := userService.Call(); err != nil {
		return fmt.Errorf("用户服务调用失败: %w", err)
	}

	// 调用订单服务
	if err := orderService.Call(); err != nil {
		return fmt.Errorf("订单服务调用失败: %w", err)
	}

	// 调用支付服务
	if err := paymentService.Call(); err != nil {
		return fmt.Errorf("支付服务调用失败: %w", err)
	}

	fmt.Printf("  ✅ 请求 %s 处理成功\n", requestID)
	return nil
}

// highConcurrencyExample 高并发任务处理示例
func highConcurrencyExample() {
	executor := goroutine.NewSafeExecutor()

	// 创建限流器
	rateLimiter := goroutine.NewRateLimiter(executor, 5, 1*time.Second)
	defer rateLimiter.Stop()

	// 创建熔断器
	circuitBreaker := goroutine.NewCircuitBreaker(executor, 10, 2*time.Second)

	// 创建重试器
	retry := goroutine.NewRetry(executor, 3, 100*time.Millisecond, &goroutine.ExponentialBackoff{
		BaseDelay: 50 * time.Millisecond,
		MaxDelay:  1 * time.Second,
	})

	// 模拟高并发任务
	var wg sync.WaitGroup
	numTasks := 50

	wg.Add(numTasks)
	for i := 0; i < numTasks; i++ {
		taskID := i + 1
		executor.Go(func() {
			defer wg.Done()
			processHighConcurrencyTask(taskID, rateLimiter, circuitBreaker, retry)
		})
	}

	wg.Wait()
	fmt.Println("  ✅ 高并发任务处理完成")
}

func processHighConcurrencyTask(taskID int, rl *goroutine.RateLimiter, cb *goroutine.CircuitBreaker, retry *goroutine.Retry) {
	// 限流检查
	if !rl.Allow() {
		fmt.Printf("  ⏳ 任务 %d 被限流\n", taskID)
		return
	}

	// 使用熔断器执行任务
	err := cb.Execute(func() error {
		return retry.Execute(func() error {
			// 模拟任务执行
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

			// 模拟失败
			if rand.Float64() < 0.3 {
				return fmt.Errorf("任务 %d 执行失败", taskID)
			}

			fmt.Printf("  ✅ 任务 %d 执行成功\n", taskID)
			return nil
		})
	})

	if err != nil {
		fmt.Printf("  ❌ 任务 %d 最终失败: %v\n", taskID, err)
	}
}

// distributedSchedulingExample 分布式任务调度示例
func distributedSchedulingExample() {
	executor := goroutine.NewSafeExecutor()

	// 创建任务调度器
	scheduler := NewTaskScheduler(executor)
	defer scheduler.Stop()

	// 添加任务
	scheduler.AddTask("daily-report", func() error {
		fmt.Println("  📊 生成日报")
		time.Sleep(100 * time.Millisecond)
		return nil
	}, 1*time.Second)

	scheduler.AddTask("data-backup", func() error {
		fmt.Println("  💾 数据备份")
		time.Sleep(200 * time.Millisecond)
		return nil
	}, 2*time.Second)

	scheduler.AddTask("health-check", func() error {
		fmt.Println("  🏥 健康检查")
		time.Sleep(50 * time.Millisecond)
		return nil
	}, 500*time.Millisecond)

	// 运行调度器
	scheduler.Start()

	// 运行一段时间
	time.Sleep(5 * time.Second)

	// 显示统计信息
	stats := scheduler.GetStats()
	fmt.Printf("  📈 调度器统计: 总任务=%d, 成功=%d, 失败=%d\n",
		stats.TotalTasks, stats.SuccessfulTasks, stats.FailedTasks)
}

// TaskScheduler 任务调度器
type TaskScheduler struct {
	executor *goroutine.SafeExecutor
	tasks    map[string]*ScheduledTask
	stopChan chan struct{}
	mu       sync.RWMutex
	stats    *SchedulerStats
}

type ScheduledTask struct {
	ID       string
	Fn       func() error
	Interval time.Duration
	StopChan chan struct{}
}

type SchedulerStats struct {
	TotalTasks      int64 `json:"total_tasks"`
	SuccessfulTasks int64 `json:"successful_tasks"`
	FailedTasks     int64 `json:"failed_tasks"`
	mu              sync.RWMutex
}

func NewTaskScheduler(executor *goroutine.SafeExecutor) *TaskScheduler {
	return &TaskScheduler{
		executor: executor,
		tasks:    make(map[string]*ScheduledTask),
		stopChan: make(chan struct{}),
		stats:    &SchedulerStats{},
	}
}

func (s *TaskScheduler) AddTask(id string, fn func() error, interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stopChan := make(chan struct{})
	task := &ScheduledTask{
		ID:       id,
		Fn:       fn,
		Interval: interval,
		StopChan: stopChan,
	}

	s.tasks[id] = task

	// 启动任务
	s.executor.Go(func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.executor.Go(func() {
					s.executeTask(task)
				})
			case <-stopChan:
				return
			}
		}
	})
}

func (s *TaskScheduler) executeTask(task *ScheduledTask) {
	s.stats.mu.Lock()
	s.stats.TotalTasks++
	s.stats.mu.Unlock()

	err := task.Fn()

	s.stats.mu.Lock()
	if err != nil {
		s.stats.FailedTasks++
	} else {
		s.stats.SuccessfulTasks++
	}
	s.stats.mu.Unlock()
}

func (s *TaskScheduler) Start() {
	// 调度器已经在AddTask时启动
}

func (s *TaskScheduler) Stop() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, task := range s.tasks {
		close(task.StopChan)
	}
	close(s.stopChan)
}

func (s *TaskScheduler) GetStats() SchedulerStats {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()

	// 返回副本以避免锁值复制
	stats := *s.stats
	return stats
}

// realtimeMonitoringExample 实时监控和告警示例
func realtimeMonitoringExample() {
	executor := goroutine.NewSafeExecutor()

	// 创建监控器
	monitor := goroutine.NewMonitor(executor, 500*time.Millisecond)

	// 设置告警配置
	monitor.SetAlertConfig(&goroutine.AlertConfig{
		MaxGoroutines: 20,
		MaxHeapAlloc:  100 * 1024 * 1024, // 100MB
		AlertHandler: func(alert *goroutine.Alert) {
			fmt.Printf("  🚨 告警 [%s]: %s (当前值: %v, 阈值: %v)\n",
				alert.Severity, alert.Message, alert.Value, alert.Threshold)
		},
	})

	// 添加监控处理器
	monitor.AddHandler(func(stats *goroutine.MonitorStats) {
		fmt.Printf("  📊 实时监控: 协程数=%d, 堆内存=%d KB, GC次数=%d\n",
			stats.NumGoroutines, stats.HeapAlloc/1024, stats.NumGC)
	})

	// 启动监控
	monitor.Start()
	defer monitor.Stop()

	// 模拟高负载
	for i := 0; i < 30; i++ {
		executor.Go(func() {
			// 模拟内存分配
			data := make([]byte, 1024*1024) // 1MB
			_ = data
			time.Sleep(100 * time.Millisecond)
		})
	}

	// 运行一段时间
	time.Sleep(3 * time.Second)

	// 强制GC
	monitor.ForceGC()
	fmt.Println("  🗑️ 强制垃圾回收完成")
}

// gracefulShutdownExample 优雅关闭示例
func gracefulShutdownExample() {
	executor := goroutine.NewSafeExecutor()

	// 创建协程池
	config := &goroutine.PoolConfig{
		MaxWorkers: 5,
		QueueSize:  50,
		JobTimeout: 2 * time.Second,
	}
	pool := goroutine.NewPool(config)

	// 设置优雅关闭信号
	shutdown := make(chan struct{})

	// 启动优雅关闭处理器
	executor.Go(func() {
		<-shutdown
		fmt.Println("  🛑 开始优雅关闭...")

		// 停止接收新任务
		pool.StopGracefully(5 * time.Second)

		fmt.Println("  ✅ 优雅关闭完成")
	})

	// 提交一些长时间运行的任务
	for i := 0; i < 10; i++ {
		taskID := fmt.Sprintf("long-task-%d", i+1)
		pool.SubmitFunc(taskID, func() error {
			fmt.Printf("  🔄 执行长时间任务 %s\n", taskID)
			time.Sleep(2 * time.Second)
			fmt.Printf("  ✅ 长时间任务 %s 完成\n", taskID)
			return nil
		})
	}

	// 等待一段时间
	time.Sleep(1 * time.Second)

	// 触发优雅关闭
	close(shutdown)

	// 等待关闭完成
	time.Sleep(6 * time.Second)
}
