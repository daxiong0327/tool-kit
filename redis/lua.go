package redis

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Script Lua脚本管理器
type Script struct {
	client *redis.Client
	cache  map[string]string // 脚本SHA缓存
	mutex  sync.RWMutex
}

// ScriptInfo Lua脚本信息
type ScriptInfo struct {
	Name        string        `json:"name" yaml:"name"`               // 脚本名称
	Source      string        `json:"source" yaml:"source"`           // 脚本源码
	SHA         string        `json:"sha" yaml:"sha"`                 // 脚本SHA
	Keys        []string      `json:"keys" yaml:"keys"`               // 键模式
	Args        []string      `json:"args" yaml:"args"`               // 参数模式
	Description string        `json:"description" yaml:"description"` // 脚本描述
	CreatedAt   time.Time     `json:"created_at" yaml:"created_at"`   // 创建时间
	UpdatedAt   time.Time     `json:"updated_at" yaml:"updated_at"`   // 更新时间
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`         // 执行超时
}

// ScriptResult Lua脚本执行结果
type ScriptResult struct {
	Value interface{}   `json:"value"` // 返回值
	Error error         `json:"error"` // 错误信息
	SHA   string        `json:"sha"`   // 脚本SHA
	Time  time.Duration `json:"time"`  // 执行时间
}

// ScriptOptions Lua脚本执行选项
type ScriptOptions struct {
	Keys        []string      `json:"keys" yaml:"keys"`                 // 键列表
	Args        []interface{} `json:"args" yaml:"args"`                 // 参数列表
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`           // 执行超时
	RetryCount  int           `json:"retry_count" yaml:"retry_count"`   // 重试次数
	RetryDelay  time.Duration `json:"retry_delay" yaml:"retry_delay"`   // 重试延迟
	UseCache    bool          `json:"use_cache" yaml:"use_cache"`       // 是否使用缓存
	ForceReload bool          `json:"force_reload" yaml:"force_reload"` // 强制重新加载
}

// DefaultScriptOptions 默认脚本选项
func DefaultScriptOptions() *ScriptOptions {
	return &ScriptOptions{
		Keys:        []string{},
		Args:        []interface{}{},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}
}

// NewScript 创建Lua脚本管理器
func (c *Client) NewScript() *Script {
	return &Script{
		client: c.client,
		cache:  make(map[string]string),
	}
}

// Register 注册Lua脚本
func (s *Script) Register(ctx context.Context, info *ScriptInfo) error {
	if info == nil {
		return fmt.Errorf("script info cannot be nil")
	}

	if info.Name == "" {
		return fmt.Errorf("script name cannot be empty")
	}

	if info.Source == "" {
		return fmt.Errorf("script source cannot be empty")
	}

	// 计算脚本SHA
	sha := s.calculateSHA(info.Source)
	info.SHA = sha
	info.CreatedAt = time.Now()
	info.UpdatedAt = time.Now()

	// 缓存脚本
	s.mutex.Lock()
	s.cache[info.Name] = sha
	s.mutex.Unlock()

	// 预加载脚本到Redis
	return s.loadScript(ctx, info)
}

// Execute 执行Lua脚本
func (s *Script) Execute(ctx context.Context, name string, opts *ScriptOptions) (*ScriptResult, error) {
	if opts == nil {
		opts = DefaultScriptOptions()
	}

	// 获取脚本SHA
	sha, err := s.getScriptSHA(ctx, name, opts.ForceReload)
	if err != nil {
		return nil, fmt.Errorf("failed to get script SHA: %w", err)
	}

	// 执行脚本
	start := time.Now()
	var result interface{}
	var execErr error

	// 重试逻辑
	for i := 0; i <= opts.RetryCount; i++ {
		if i > 0 {
			time.Sleep(opts.RetryDelay)
		}

		// 设置超时
		execCtx := ctx
		if opts.Timeout > 0 {
			var cancel context.CancelFunc
			execCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
			defer cancel()
		}

		// 执行脚本
		if opts.UseCache && sha != "" {
			// 使用缓存的SHA执行
			result, execErr = s.client.EvalSha(execCtx, sha, opts.Keys, opts.Args...).Result()
		} else {
			// 直接执行脚本源码
			script, err := s.getScriptSource(ctx, name)
			if err != nil {
				execErr = err
				continue
			}
			result, execErr = s.client.Eval(execCtx, script, opts.Keys, opts.Args...).Result()
		}

		// 如果脚本不存在，尝试重新加载
		if execErr != nil && isScriptNotFoundError(execErr) {
			err := s.reloadScript(ctx, name)
			if err != nil {
				continue
			}
			// 重新获取SHA
			sha, _ = s.getScriptSHA(ctx, name, false)
			continue
		}

		// 如果成功或非重试错误，跳出循环
		if execErr == nil || !isRetryableError(execErr) {
			break
		}
	}

	executionTime := time.Since(start)

	return &ScriptResult{
		Value: result,
		Error: execErr,
		SHA:   sha,
		Time:  executionTime,
	}, nil
}

// ExecuteString 执行Lua脚本（字符串形式）
func (s *Script) ExecuteString(ctx context.Context, script string, opts *ScriptOptions) (*ScriptResult, error) {
	if opts == nil {
		opts = DefaultScriptOptions()
	}

	start := time.Now()
	var result interface{}
	var execErr error

	// 重试逻辑
	for i := 0; i <= opts.RetryCount; i++ {
		if i > 0 {
			time.Sleep(opts.RetryDelay)
		}

		// 设置超时
		execCtx := ctx
		if opts.Timeout > 0 {
			var cancel context.CancelFunc
			execCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
			defer cancel()
		}

		// 执行脚本
		result, execErr = s.client.Eval(execCtx, script, opts.Keys, opts.Args...).Result()

		// 如果成功或非重试错误，跳出循环
		if execErr == nil || !isRetryableError(execErr) {
			break
		}
	}

	executionTime := time.Since(start)
	sha := s.calculateSHA(script)

	return &ScriptResult{
		Value: result,
		Error: execErr,
		SHA:   sha,
		Time:  executionTime,
	}, nil
}

// Load 加载脚本到Redis
func (s *Script) Load(ctx context.Context, script string) (string, error) {
	sha, err := s.client.ScriptLoad(ctx, script).Result()
	if err != nil {
		return "", fmt.Errorf("failed to load script: %w", err)
	}
	return sha, nil
}

// Exists 检查脚本是否存在
func (s *Script) Exists(ctx context.Context, sha string) (bool, error) {
	exists, err := s.client.ScriptExists(ctx, sha).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check script existence: %w", err)
	}
	return len(exists) > 0 && exists[0], nil
}

// Flush 清空脚本缓存
func (s *Script) Flush(ctx context.Context) error {
	err := s.client.ScriptFlush(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush scripts: %w", err)
	}

	// 清空本地缓存
	s.mutex.Lock()
	s.cache = make(map[string]string)
	s.mutex.Unlock()

	return nil
}

// GetCache 获取脚本缓存信息
func (s *Script) GetCache() map[string]string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	cache := make(map[string]string)
	for k, v := range s.cache {
		cache[k] = v
	}
	return cache
}

// ClearCache 清空本地缓存
func (s *Script) ClearCache() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cache = make(map[string]string)
}

// 私有方法

// calculateSHA 计算脚本SHA
func (s *Script) calculateSHA(script string) string {
	hash := sha1.Sum([]byte(script))
	return hex.EncodeToString(hash[:])
}

// getScriptSHA 获取脚本SHA
func (s *Script) getScriptSHA(ctx context.Context, name string, forceReload bool) (string, error) {
	if !forceReload {
		s.mutex.RLock()
		if sha, exists := s.cache[name]; exists {
			s.mutex.RUnlock()
			return sha, nil
		}
		s.mutex.RUnlock()
	}

	// 从Redis获取脚本信息
	script, err := s.getScriptSource(ctx, name)
	if err != nil {
		return "", err
	}

	sha := s.calculateSHA(script)

	// 更新缓存
	s.mutex.Lock()
	s.cache[name] = sha
	s.mutex.Unlock()

	return sha, nil
}

// getScriptSource 获取脚本源码
func (s *Script) getScriptSource(ctx context.Context, name string) (string, error) {
	// 这里可以从文件系统、数据库或其他存储中获取脚本
	// 为了简化，我们假设脚本已经通过Register方法注册
	// 在实际应用中，可以实现从文件系统或数据库加载
	return "", fmt.Errorf("script source not found: %s", name)
}

// loadScript 加载脚本到Redis
func (s *Script) loadScript(ctx context.Context, info *ScriptInfo) error {
	sha, err := s.client.ScriptLoad(ctx, info.Source).Result()
	if err != nil {
		return fmt.Errorf("failed to load script: %w", err)
	}

	// 更新SHA
	info.SHA = sha

	return nil
}

// reloadScript 重新加载脚本
func (s *Script) reloadScript(ctx context.Context, name string) error {
	// 这里可以实现重新加载脚本的逻辑
	// 例如从文件系统重新读取脚本并加载到Redis
	return fmt.Errorf("script reload not implemented: %s", name)
}

// isScriptNotFoundError 检查是否为脚本不存在错误
func isScriptNotFoundError(err error) bool {
	return err != nil && err.Error() == "NOSCRIPT No matching script. Use EVAL."
}

// isRetryableError 检查是否为可重试错误
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 网络错误、超时错误等可以重试
	errorStr := err.Error()
	retryableErrors := []string{
		"timeout",
		"connection",
		"network",
		"temporary",
		"busy",
	}

	for _, retryable := range retryableErrors {
		if contains(errorStr, retryable) {
			return true
		}
	}

	return false
}

// contains 检查字符串是否包含子字符串（忽略大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					contains(s[1:], substr)))
}
