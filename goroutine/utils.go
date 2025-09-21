package goroutine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// GlobalExecutor 全局协程执行器
var GlobalExecutor = NewSafeExecutor()

// Go 使用全局执行器启动协程
func Go(fn func()) {
	GlobalExecutor.Go(fn)
}

// GoWithContext 使用全局执行器带上下文启动协程
func GoWithContext(ctx context.Context, fn func()) {
	GlobalExecutor.GoWithContext(ctx, fn)
}

// GoWithTimeout 使用全局执行器带超时启动协程
func GoWithTimeout(fn func(), timeout time.Duration) {
	GlobalExecutor.GoWithTimeout(fn, timeout)
}

// GoWithDelay 使用全局执行器延迟启动协程
func GoWithDelay(fn func(), delay time.Duration) {
	GlobalExecutor.GoWithDelay(fn, delay)
}

// GoWithInterval 使用全局执行器定时执行协程
func GoWithInterval(fn func(), interval time.Duration) chan<- struct{} {
	return GlobalExecutor.GoWithInterval(fn, interval)
}

// SetGlobalRecoverHandler 设置全局崩溃恢复处理器
func SetGlobalRecoverHandler(handler RecoverHandler) {
	GlobalExecutor.SetRecoverHandler(handler)
}

// SetGlobalLogger 设置全局日志器
func SetGlobalLogger(logger Logger) {
	GlobalExecutor.SetLogger(logger)
}

// GetGlobalStats 获取全局统计信息
func GetGlobalStats() Stats {
	return GlobalExecutor.GetStats()
}

// Batch 批量执行函数
type Batch struct {
	executor *SafeExecutor
	wg       *WaitGroup
	mu       sync.Mutex
	results  []BatchResult
	errors   []error
}

// BatchResult 批量执行结果
type BatchResult struct {
	Index int
	Value interface{}
	Error error
}

// NewBatch 创建批量执行器
func NewBatch(executor *SafeExecutor) *Batch {
	if executor == nil {
		executor = GlobalExecutor
	}
	return &Batch{
		executor: executor,
		wg:       executor.NewWaitGroup(),
		results:  make([]BatchResult, 0),
		errors:   make([]error, 0),
	}
}

// Add 添加任务到批量执行器
func (b *Batch) Add(fn func() (interface{}, error)) {
	index := len(b.results)
	b.wg.Go(func() {
		value, err := fn()
		b.mu.Lock()
		defer b.mu.Unlock()
		b.results = append(b.results, BatchResult{
			Index: index,
			Value: value,
			Error: err,
		})
		if err != nil {
			b.errors = append(b.errors, err)
		}
	})
}

// Wait 等待所有任务完成
func (b *Batch) Wait() ([]BatchResult, []error) {
	b.wg.Wait()
	return b.results, b.errors
}

// WaitForResults 等待结果并返回
func (b *Batch) WaitForResults() []BatchResult {
	results, _ := b.Wait()
	return results
}

// WaitForErrors 等待错误并返回
func (b *Batch) WaitForErrors() []error {
	_, errors := b.Wait()
	return errors
}

// HasErrors 检查是否有错误
func (b *Batch) HasErrors() bool {
	_, errors := b.Wait()
	return len(errors) > 0
}

// Retry 重试执行器
type Retry struct {
	executor    *SafeExecutor
	maxAttempts int
	delay       time.Duration
	backoff     BackoffStrategy
}

// BackoffStrategy 退避策略
type BackoffStrategy interface {
	GetDelay(attempt int) time.Duration
}

// FixedBackoff 固定延迟退避
type FixedBackoff struct {
	Delay time.Duration
}

func (b *FixedBackoff) GetDelay(attempt int) time.Duration {
	return b.Delay
}

// ExponentialBackoff 指数退避
type ExponentialBackoff struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

func (b *ExponentialBackoff) GetDelay(attempt int) time.Duration {
	delay := b.BaseDelay
	for i := 0; i < attempt; i++ {
		delay *= 2
		if delay > b.MaxDelay {
			delay = b.MaxDelay
		}
	}
	return delay
}

// LinearBackoff 线性退避
type LinearBackoff struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

func (b *LinearBackoff) GetDelay(attempt int) time.Duration {
	delay := b.BaseDelay * time.Duration(attempt+1)
	if delay > b.MaxDelay {
		delay = b.MaxDelay
	}
	return delay
}

// NewRetry 创建重试执行器
func NewRetry(executor *SafeExecutor, maxAttempts int, delay time.Duration, backoff BackoffStrategy) *Retry {
	if executor == nil {
		executor = GlobalExecutor
	}
	if backoff == nil {
		backoff = &FixedBackoff{Delay: delay}
	}
	return &Retry{
		executor:    executor,
		maxAttempts: maxAttempts,
		delay:       delay,
		backoff:     backoff,
	}
}

// Execute 执行重试
func (r *Retry) Execute(fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < r.maxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// 如果不是最后一次尝试，等待后重试
		if attempt < r.maxAttempts-1 {
			delay := r.backoff.GetDelay(attempt)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("retry failed after %d attempts: %w", r.maxAttempts, lastErr)
}

// ExecuteAsync 异步执行重试
func (r *Retry) ExecuteAsync(fn func() error) <-chan error {
	result := make(chan error, 1)

	r.executor.Go(func() {
		result <- r.Execute(fn)
	})

	return result
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	executor        *SafeExecutor
	failureCount    int64
	successCount    int64
	maxFailures     int64
	resetTimeout    time.Duration
	state           CircuitState
	lastFailureTime time.Time
	mu              sync.RWMutex
}

// CircuitState 熔断器状态
type CircuitState int

const (
	StateClosed   CircuitState = iota // 关闭状态
	StateOpen                         // 开启状态
	StateHalfOpen                     // 半开状态
)

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(executor *SafeExecutor, maxFailures int64, resetTimeout time.Duration) *CircuitBreaker {
	if executor == nil {
		executor = GlobalExecutor
	}
	return &CircuitBreaker{
		executor:     executor,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// Execute 执行熔断器保护的方法
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.RLock()
	state := cb.state
	cb.mu.RUnlock()

	switch state {
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.mu.Unlock()
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	case StateHalfOpen:
		// 允许一次尝试
	case StateClosed:
		// 正常执行
	}

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		if cb.failureCount >= cb.maxFailures {
			cb.state = StateOpen
		}
	} else {
		cb.successCount++
		if cb.state == StateHalfOpen {
			cb.state = StateClosed
			cb.failureCount = 0
		}
	}

	return err
}

// GetState 获取熔断器状态
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats 获取熔断器统计
func (cb *CircuitBreaker) GetStats() (failureCount, successCount int64) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failureCount, cb.successCount
}

// Reset 重置熔断器
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount = 0
	cb.successCount = 0
	cb.state = StateClosed
}

// RateLimiter 限流器
type RateLimiter struct {
	executor *SafeExecutor
	limit    int
	interval time.Duration
	tokens   chan struct{}
	ticker   *time.Ticker
	mu       sync.Mutex
}

// NewRateLimiter 创建限流器
func NewRateLimiter(executor *SafeExecutor, limit int, interval time.Duration) *RateLimiter {
	if executor == nil {
		executor = GlobalExecutor
	}

	rl := &RateLimiter{
		executor: executor,
		limit:    limit,
		interval: interval,
		tokens:   make(chan struct{}, limit),
	}

	// 填充初始令牌
	for i := 0; i < limit; i++ {
		rl.tokens <- struct{}{}
	}

	// 启动令牌补充协程
	rl.ticker = time.NewTicker(interval / time.Duration(limit))
	executor.Go(func() {
		for range rl.ticker.C {
			select {
			case rl.tokens <- struct{}{}:
			default:
				// 令牌已满，跳过
			}
		}
	})

	return rl
}

// Allow 检查是否允许执行
func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// Wait 等待令牌可用
func (rl *RateLimiter) Wait() {
	<-rl.tokens
}

// Stop 停止限流器
func (rl *RateLimiter) Stop() {
	if rl.ticker != nil {
		rl.ticker.Stop()
	}
}

// GetAvailableTokens 获取可用令牌数
func (rl *RateLimiter) GetAvailableTokens() int {
	return len(rl.tokens)
}
