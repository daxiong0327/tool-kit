package goroutine

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Pool 协程池
type Pool struct {
	executor    *SafeExecutor
	maxWorkers  int
	workerCount int64
	jobQueue    chan Job
	workerQueue chan chan Job
	quit        chan bool
	wg          sync.WaitGroup
	mu          sync.RWMutex
	stats       *PoolStats
}

// Job 任务接口
type Job interface {
	Execute() error
	GetID() string
	GetTimeout() time.Duration
}

// SimpleJob 简单任务实现
type SimpleJob struct {
	ID      string
	Timeout time.Duration
	Fn      func() error
}

func (j *SimpleJob) Execute() error {
	return j.Fn()
}

func (j *SimpleJob) GetID() string {
	return j.ID
}

func (j *SimpleJob) GetTimeout() time.Duration {
	return j.Timeout
}

// PoolStats 协程池统计信息
type PoolStats struct {
	TotalJobs      int64         `json:"total_jobs"`
	CompletedJobs  int64         `json:"completed_jobs"`
	FailedJobs     int64         `json:"failed_jobs"`
	ActiveWorkers  int64         `json:"active_workers"`
	QueuedJobs     int64         `json:"queued_jobs"`
	LastJobTime    time.Time     `json:"last_job_time"`
	AverageJobTime time.Duration `json:"average_job_time"`
	mu             sync.RWMutex
}

// PoolConfig 协程池配置
type PoolConfig struct {
	MaxWorkers    int           `json:"max_workers"`    // 最大工作协程数
	QueueSize     int           `json:"queue_size"`     // 任务队列大小
	WorkerTimeout time.Duration `json:"worker_timeout"` // 工作协程超时时间
	JobTimeout    time.Duration `json:"job_timeout"`    // 任务超时时间
}

// DefaultPoolConfig 默认协程池配置
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxWorkers:    10,
		QueueSize:     1000,
		WorkerTimeout: 30 * time.Minute,
		JobTimeout:    5 * time.Minute,
	}
}

// NewPool 创建协程池
func NewPool(config *PoolConfig) *Pool {
	if config == nil {
		config = DefaultPoolConfig()
	}

	pool := &Pool{
		executor:    NewSafeExecutor(),
		maxWorkers:  config.MaxWorkers,
		jobQueue:    make(chan Job, config.QueueSize),
		workerQueue: make(chan chan Job, config.MaxWorkers),
		quit:        make(chan bool),
		stats:       &PoolStats{},
	}

	// 启动调度器
	pool.startDispatcher()

	return pool
}

// startDispatcher 启动任务调度器
func (p *Pool) startDispatcher() {
	p.executor.Go(func() {
		for {
			select {
			case job := <-p.jobQueue:
				// 更新统计
				atomic.AddInt64(&p.stats.TotalJobs, 1)
				p.stats.mu.Lock()
				p.stats.LastJobTime = time.Now()
				p.stats.QueuedJobs = int64(len(p.jobQueue))
				p.stats.mu.Unlock()

				// 获取空闲工作协程
				workerJobQueue := <-p.workerQueue
				workerJobQueue <- job
			case <-p.quit:
				return
			}
		}
	})

	// 启动工作协程
	for i := 0; i < p.maxWorkers; i++ {
		p.startWorker()
	}
}

// startWorker 启动工作协程
func (p *Pool) startWorker() {
	workerJobQueue := make(chan Job)
	p.workerQueue <- workerJobQueue

	p.executor.Go(func() {
		atomic.AddInt64(&p.workerCount, 1)
		atomic.AddInt64(&p.stats.ActiveWorkers, 1)
		defer func() {
			atomic.AddInt64(&p.workerCount, -1)
			atomic.AddInt64(&p.stats.ActiveWorkers, -1)
		}()

		for {
			select {
			case job := <-workerJobQueue:
				p.executeJob(job)
			case <-p.quit:
				// 关闭工作协程的通道
				close(workerJobQueue)
				return
			}
		}
	})
}

// executeJob 执行任务
func (p *Pool) executeJob(job Job) {
	startTime := time.Now()

	// 创建任务上下文
	ctx := context.Background()
	if job.GetTimeout() > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, job.GetTimeout())
		defer cancel()
	}

	// 执行任务
	p.executor.GoWithContext(ctx, func() {
		err := job.Execute()

		// 更新统计
		jobTime := time.Since(startTime)
		p.stats.mu.Lock()
		if err != nil {
			atomic.AddInt64(&p.stats.FailedJobs, 1)
		} else {
			atomic.AddInt64(&p.stats.CompletedJobs, 1)
		}

		// 更新平均执行时间
		totalJobs := atomic.LoadInt64(&p.stats.CompletedJobs) + atomic.LoadInt64(&p.stats.FailedJobs)
		if totalJobs > 0 {
			p.stats.AverageJobTime = time.Duration(
				(int64(p.stats.AverageJobTime)*totalJobs + int64(jobTime)) / (totalJobs + 1),
			)
		}
		p.stats.mu.Unlock()
	})
}

// Submit 提交任务
func (p *Pool) Submit(job Job) error {
	select {
	case p.jobQueue <- job:
		return nil
	default:
		return ErrPoolFull
	}
}

// SubmitFunc 提交函数任务
func (p *Pool) SubmitFunc(id string, fn func() error) error {
	job := &SimpleJob{
		ID:      id,
		Timeout: 0, // 使用默认超时
		Fn:      fn,
	}
	return p.Submit(job)
}

// SubmitWithTimeout 提交带超时的任务
func (p *Pool) SubmitWithTimeout(id string, fn func() error, timeout time.Duration) error {
	job := &SimpleJob{
		ID:      id,
		Timeout: timeout,
		Fn:      fn,
	}
	return p.Submit(job)
}

// GetStats 获取协程池统计信息
func (p *Pool) GetStats() PoolStats {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()

	// 返回副本以避免锁值复制
	stats := *p.stats
	stats.QueuedJobs = int64(len(p.jobQueue))
	stats.ActiveWorkers = atomic.LoadInt64(&p.workerCount)

	return stats
}

// GetExecutorStats 获取执行器统计信息
func (p *Pool) GetExecutorStats() Stats {
	return p.executor.GetStats()
}

// SetRecoverHandler 设置崩溃恢复处理器
func (p *Pool) SetRecoverHandler(handler RecoverHandler) {
	p.executor.SetRecoverHandler(handler)
}

// SetLogger 设置日志器
func (p *Pool) SetLogger(logger Logger) {
	p.executor.SetLogger(logger)
}

// Stop 停止协程池
func (p *Pool) Stop() {
	close(p.quit)
	// 等待所有工作协程退出
	time.Sleep(200 * time.Millisecond) // 给工作协程一些时间退出
}

// StopGracefully 优雅停止协程池
func (p *Pool) StopGracefully(timeout time.Duration) {
	// 停止接收新任务
	close(p.jobQueue)

	// 等待所有任务完成或超时
	done := make(chan struct{})
	go func() {
		// 等待队列中的任务完成
		for len(p.jobQueue) > 0 {
			time.Sleep(10 * time.Millisecond)
		}
		close(done)
	}()

	select {
	case <-done:
		// 所有任务完成
	case <-time.After(timeout):
		// 超时，强制停止
	}

	// 停止工作协程
	close(p.quit)
	time.Sleep(100 * time.Millisecond) // 给工作协程一些时间退出
}

// Wait 等待所有任务完成
func (p *Pool) Wait() {
	p.wg.Wait()
}

// 错误定义
var (
	ErrPoolFull = &PoolError{Message: "pool is full"}
)

// PoolError 协程池错误
type PoolError struct {
	Message string
}

func (e *PoolError) Error() string {
	return e.Message
}
