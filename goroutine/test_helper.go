package goroutine

import (
	"context"
	"testing"
	"time"
)

// RunWithTimeout 运行测试并设置超时
func RunWithTimeout(t *testing.T, timeout time.Duration, fn func()) {
	done := make(chan struct{})

	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
		// 测试完成
	case <-time.After(timeout):
		t.Fatalf("Test timed out after %v", timeout)
	}
}

// RunTestWithContext 使用上下文运行测试
func RunTestWithContext(t *testing.T, timeout time.Duration, fn func(context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})

	go func() {
		defer close(done)
		fn(ctx)
	}()

	select {
	case <-done:
		// 测试完成
	case <-ctx.Done():
		t.Fatalf("Test timed out after %v", timeout)
	}
}
