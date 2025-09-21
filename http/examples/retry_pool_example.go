package main

import (
	"context"
	"fmt"
	"time"

	"github.com/daxiong0327/tool-kit/http"
)

func main() {
	fmt.Println("=== HTTP 重试和连接池功能演示 ===")

	// 示例1：基本重试配置
	fmt.Println("\n1. 基本重试配置:")
	retryConfig := &http.RetryConfig{
		MaxRetries:     3,
		BaseDelay:      1 * time.Second,
		MaxDelay:       10 * time.Second,
		Strategy:       http.RetryStrategyExponential,
		RetryableCodes: []int{500, 502, 503, 504, 408, 429},
	}

	client := http.NewWithRetry("https://httpbin.org", retryConfig)
	ctx := context.Background()

	// 这个请求可能会失败，但会重试
	resp, err := client.Get(ctx, "/status/500")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
	} else {
		fmt.Printf("状态码: %d\n", resp.StatusCode)
	}

	// 示例2：连接池配置
	fmt.Println("\n2. 连接池配置:")
	poolConfig := &http.PoolConfig{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     20,
		IdleConnTimeout:     60 * time.Second,
		DisableKeepAlives:   false,
	}

	poolClient := http.NewWithPool("https://httpbin.org", poolConfig)
	fmt.Printf("连接池配置: %+v\n", poolClient.GetPoolConfig())

	// 示例3：组合配置
	fmt.Println("\n3. 重试 + 连接池组合配置:")
	combinedClient := http.NewWithRetryAndPool("https://httpbin.org", retryConfig, poolConfig)
	fmt.Printf("重试配置: %+v\n", combinedClient.GetRetryConfig())
	fmt.Printf("连接池配置: %+v\n", combinedClient.GetPoolConfig())

	// 示例4：不同重试策略
	fmt.Println("\n4. 不同重试策略演示:")

	// 固定延迟策略
	fixedRetryConfig := &http.RetryConfig{
		MaxRetries:     2,
		BaseDelay:      500 * time.Millisecond,
		MaxDelay:       5 * time.Second,
		Strategy:       http.RetryStrategyFixed,
		RetryableCodes: []int{500, 502, 503, 504},
	}

	// 线性增长策略
	linearRetryConfig := &http.RetryConfig{
		MaxRetries:     2,
		BaseDelay:      500 * time.Millisecond,
		MaxDelay:       5 * time.Second,
		Strategy:       http.RetryStrategyLinear,
		RetryableCodes: []int{500, 502, 503, 504},
	}

	// 指数退避策略
	exponentialRetryConfig := &http.RetryConfig{
		MaxRetries:     2,
		BaseDelay:      500 * time.Millisecond,
		MaxDelay:       5 * time.Second,
		Strategy:       http.RetryStrategyExponential,
		RetryableCodes: []int{500, 502, 503, 504},
	}

	strategies := []struct {
		name   string
		config *http.RetryConfig
	}{
		{"固定延迟", fixedRetryConfig},
		{"线性增长", linearRetryConfig},
		{"指数退避", exponentialRetryConfig},
	}

	for _, strategy := range strategies {
		fmt.Printf("\n%s策略延迟计算:\n", strategy.name)
		client := http.NewWithRetry("", strategy.config)

		for attempt := 0; attempt < 3; attempt++ {
			delay := calculateDelay(strategy.config, attempt)
			fmt.Printf("  尝试 %d: %v\n", attempt+1, delay)
		}
	}

	// 示例5：动态配置修改
	fmt.Println("\n5. 动态配置修改:")
	client = http.New(nil)

	// 修改重试配置
	newRetryConfig := &http.RetryConfig{
		MaxRetries:     5,
		BaseDelay:      2 * time.Second,
		MaxDelay:       30 * time.Second,
		Strategy:       http.RetryStrategyExponential,
		RetryableCodes: []int{500, 502, 503, 504, 408, 429},
	}
	client.SetRetryConfig(newRetryConfig)
	fmt.Printf("新的重试配置: %+v\n", client.GetRetryConfig())

	// 修改连接池配置
	newPoolConfig := &http.PoolConfig{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     120 * time.Second,
		DisableKeepAlives:   false,
	}
	client.SetPoolConfig(newPoolConfig)
	fmt.Printf("新的连接池配置: %+v\n", client.GetPoolConfig())

	fmt.Println("\n=== 演示完成 ===")
}

// calculateDelay 计算延迟时间（用于演示）
func calculateDelay(config *http.RetryConfig, attempt int) time.Duration {
	var delay time.Duration
	switch config.Strategy {
	case http.RetryStrategyFixed:
		delay = config.BaseDelay
	case http.RetryStrategyExponential:
		delay = time.Duration(float64(config.BaseDelay) * float64(1<<attempt))
	case http.RetryStrategyLinear:
		delay = config.BaseDelay * time.Duration(attempt+1)
	default:
		delay = config.BaseDelay
	}

	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}

	return delay
}
