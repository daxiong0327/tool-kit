package goroutine

import (
	"context"
	"log"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// SafeExecutor 协程安全执行器
type SafeExecutor struct {
	recoverHandler RecoverHandler
	logger         Logger
	stats          *Stats
	mu             sync.RWMutex
}

// RecoverHandler 崩溃恢复处理器
type RecoverHandler func(panicValue interface{}, stack []byte, goroutineID string)

// Logger 日志接口
type Logger interface {
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

// DefaultLogger 默认日志实现
type DefaultLogger struct{}

func (l *DefaultLogger) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

func (l *DefaultLogger) Warnf(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

func (l *DefaultLogger) Infof(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

func (l *DefaultLogger) Debugf(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

// Stats 协程统计信息
type Stats struct {
	TotalGoroutines    int64     `json:"total_goroutines"`
	ActiveGoroutines   int64     `json:"active_goroutines"`
	PanicCount         int64     `json:"panic_count"`
	CompletedCount     int64     `json:"completed_count"`
	LastPanicTime      time.Time `json:"last_panic_time"`
	LastPanicGoroutine string    `json:"last_panic_goroutine"`
	mu                 sync.RWMutex
}

// NewSafeExecutor 创建协程安全执行器
func NewSafeExecutor() *SafeExecutor {
	return &SafeExecutor{
		recoverHandler: defaultRecoverHandler,
		logger:         &DefaultLogger{},
		stats:          &Stats{},
	}
}

// SetRecoverHandler 设置崩溃恢复处理器
func (se *SafeExecutor) SetRecoverHandler(handler RecoverHandler) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.recoverHandler = handler
}

// SetLogger 设置日志器
func (se *SafeExecutor) SetLogger(logger Logger) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.logger = logger
}

// GetStats 获取统计信息
func (se *SafeExecutor) GetStats() Stats {
	se.stats.mu.RLock()
	defer se.stats.mu.RUnlock()

	// 返回副本以避免锁值复制
	stats := *se.stats
	return stats
}

// ResetStats 重置统计信息
func (se *SafeExecutor) ResetStats() {
	se.stats.mu.Lock()
	defer se.stats.mu.Unlock()
	se.stats = &Stats{}
}

// Go 安全启动协程
func (se *SafeExecutor) Go(fn func()) {
	se.GoWithContext(context.Background(), fn)
}

// GoWithContext 带上下文的安全启动协程
func (se *SafeExecutor) GoWithContext(ctx context.Context, fn func()) {
	se.mu.RLock()
	recoverHandler := se.recoverHandler
	logger := se.logger
	se.mu.RUnlock()

	// 更新统计信息
	se.stats.mu.Lock()
	se.stats.TotalGoroutines++
	se.stats.ActiveGoroutines++
	se.stats.mu.Unlock()

	go func() {
		defer func() {
			// 更新活跃协程数
			se.stats.mu.Lock()
			se.stats.ActiveGoroutines--
			se.stats.mu.Unlock()

			// 处理panic
			if r := recover(); r != nil {
				se.handlePanic(r, recoverHandler, logger)
			} else {
				// 正常完成
				se.stats.mu.Lock()
				se.stats.CompletedCount++
				se.stats.mu.Unlock()
			}
		}()

		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			logger.Warnf("Goroutine cancelled due to context: %v", ctx.Err())
			return
		default:
		}

		// 执行函数
		fn()
	}()
}

// GoWithTimeout 带超时的安全启动协程
func (se *SafeExecutor) GoWithTimeout(fn func(), timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	se.GoWithContext(ctx, fn)
}

// GoWithDelay 延迟启动协程
func (se *SafeExecutor) GoWithDelay(fn func(), delay time.Duration) {
	se.Go(func() {
		time.Sleep(delay)
		fn()
	})
}

// GoWithInterval 定时执行协程
func (se *SafeExecutor) GoWithInterval(fn func(), interval time.Duration) chan<- struct{} {
	stop := make(chan struct{})

	se.Go(func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				se.Go(fn) // 每次执行都启动新的安全协程
			case <-stop:
				return
			}
		}
	})

	return stop
}

// handlePanic 处理panic
func (se *SafeExecutor) handlePanic(panicValue interface{}, recoverHandler RecoverHandler, logger Logger) {
	// 更新panic统计
	se.stats.mu.Lock()
	se.stats.PanicCount++
	se.stats.LastPanicTime = time.Now()
	se.stats.LastPanicGoroutine = getGoroutineID()
	se.stats.mu.Unlock()

	// 获取堆栈信息
	stack := debug.Stack()
	goroutineID := getGoroutineID()

	// 记录panic信息
	logger.Errorf("Goroutine panic recovered: %v", panicValue)
	logger.Errorf("Goroutine ID: %s", goroutineID)
	logger.Errorf("Stack trace:\n%s", string(stack))

	// 调用恢复处理器
	if recoverHandler != nil {
		recoverHandler(panicValue, stack, goroutineID)
	}
}

// getGoroutineID 获取当前协程ID
func getGoroutineID() string {
	buf := make([]byte, 64)
	n := runtime.Stack(buf, false)
	idField := string(buf[:n])
	// 提取协程ID
	for i := 0; i < len(idField); i++ {
		if idField[i] == ' ' {
			return idField[:i]
		}
	}
	return "unknown"
}

// defaultRecoverHandler 默认崩溃恢复处理器
func defaultRecoverHandler(panicValue interface{}, stack []byte, goroutineID string) {
	// 默认处理器只记录日志，不进行额外处理
}

// WaitGroup 安全的WaitGroup包装器
type WaitGroup struct {
	wg   sync.WaitGroup
	se   *SafeExecutor
	mu   sync.Mutex
	done bool
}

// NewWaitGroup 创建安全的WaitGroup
func (se *SafeExecutor) NewWaitGroup() *WaitGroup {
	return &WaitGroup{
		se: se,
	}
}

// Add 添加协程计数
func (wg *WaitGroup) Add(delta int) {
	wg.mu.Lock()
	defer wg.mu.Unlock()
	if wg.done {
		panic("WaitGroup: Add called after Wait")
	}
	wg.wg.Add(delta)
}

// Done 完成一个协程
func (wg *WaitGroup) Done() {
	wg.wg.Done()
}

// Wait 等待所有协程完成
func (wg *WaitGroup) Wait() {
	wg.mu.Lock()
	wg.done = true
	wg.mu.Unlock()
	wg.wg.Wait()
}

// Go 在WaitGroup中安全启动协程
func (wg *WaitGroup) Go(fn func()) {
	wg.Add(1)
	wg.se.Go(func() {
		defer wg.Done()
		fn()
	})
}

// GoWithContext 在WaitGroup中带上下文安全启动协程
func (wg *WaitGroup) GoWithContext(ctx context.Context, fn func()) {
	wg.Add(1)
	wg.se.GoWithContext(ctx, func() {
		defer wg.Done()
		fn()
	})
}
