package redis

import (
	"context"
	"testing"
	"time"
)

func TestLuaScript(t *testing.T) {
	config := DefaultConfig()
	config.Addr = "localhost:6379"

	client, err := New(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	script := client.NewScript()

	// 测试脚本注册
	scriptInfo := &ScriptInfo{
		Name:        "test_script",
		Source:      `return "Hello, " .. ARGV[1]`,
		Keys:        []string{},
		Args:        []string{"name"},
		Description: "测试脚本",
		Timeout:     5 * time.Second,
	}

	err = script.Register(ctx, scriptInfo)
	if err != nil {
		t.Errorf("Register script failed: %v", err)
	}

	// 测试脚本执行
	opts := &ScriptOptions{
		Keys:        []string{},
		Args:        []interface{}{"World"},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	result, err := script.Execute(ctx, "test_script", opts)
	if err != nil {
		t.Errorf("Execute script failed: %v", err)
	}

	if result.Value != "Hello, World" {
		t.Errorf("Expected 'Hello, World', got %v", result.Value)
	}

	if result.SHA == "" {
		t.Error("Expected SHA to be set")
	}

	if result.Time <= 0 {
		t.Error("Expected execution time to be positive")
	}
}

func TestLuaScriptString(t *testing.T) {
	config := DefaultConfig()
	config.Addr = "localhost:6379"

	client, err := New(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	script := client.NewScript()

	// 测试直接执行脚本字符串
	scriptSource := `return "Hello, " .. ARGV[1]`
	opts := &ScriptOptions{
		Keys:        []string{},
		Args:        []interface{}{"Lua"},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    false,
		ForceReload: false,
	}

	result, err := script.ExecuteString(ctx, scriptSource, opts)
	if err != nil {
		t.Errorf("Execute string script failed: %v", err)
	}

	if result.Value != "Hello, Lua" {
		t.Errorf("Expected 'Hello, Lua', got %v", result.Value)
	}

	if result.SHA == "" {
		t.Error("Expected SHA to be set")
	}
}

func TestLuaScriptTemplates(t *testing.T) {
	config := DefaultConfig()
	config.Addr = "localhost:6379"

	client, err := New(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	script := client.NewScript()
	templates := script.NewScriptTemplates()

	// 测试注册常用脚本
	err = templates.RegisterCommonScripts(ctx)
	if err != nil {
		t.Errorf("Register common scripts failed: %v", err)
	}

	// 测试分布式锁脚本
	lockOpts := &ScriptOptions{
		Keys:        []string{"test:lock"},
		Args:        []interface{}{"lock_value", 10},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	result, err := script.Execute(ctx, "distributed_lock", lockOpts)
	if err != nil {
		t.Errorf("Execute distributed lock script failed: %v", err)
	}

	if result.Value != int64(1) {
		t.Errorf("Expected 1, got %v", result.Value)
	}

	// 测试分布式锁释放脚本
	unlockOpts := &ScriptOptions{
		Keys:        []string{"test:lock"},
		Args:        []interface{}{"lock_value"},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	result, err = script.Execute(ctx, "distributed_unlock", unlockOpts)
	if err != nil {
		t.Errorf("Execute distributed unlock script failed: %v", err)
	}

	if result.Value != int64(1) {
		t.Errorf("Expected 1, got %v", result.Value)
	}

	// 清理
	client.Del(ctx, "test:lock")
}

func TestLuaScriptManager(t *testing.T) {
	config := DefaultConfig()
	config.Addr = "localhost:6379"

	client, err := New(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	script := client.NewScript()
	manager := script.NewScriptManager()

	// 测试注册脚本
	scriptInfo := &ScriptInfo{
		Name:        "manager_test_script",
		Source:      `return "Manager test: " .. ARGV[1]`,
		Keys:        []string{},
		Args:        []string{"name"},
		Description: "管理器测试脚本",
		Timeout:     5 * time.Second,
	}

	err = manager.RegisterScript(ctx, scriptInfo)
	if err != nil {
		t.Errorf("Register script failed: %v", err)
	}

	// 测试执行脚本
	opts := &ScriptOptions{
		Keys:        []string{},
		Args:        []interface{}{"Success"},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	result, err := manager.ExecuteScript(ctx, "manager_test_script", opts)
	if err != nil {
		t.Errorf("Execute script failed: %v", err)
	}

	if result.Value != "Manager test: Success" {
		t.Errorf("Expected 'Manager test: Success', got %v", result.Value)
	}

	// 测试获取脚本信息
	info, exists := manager.GetScriptInfo("manager_test_script")
	if !exists {
		t.Error("Expected script info to exist")
	}

	if info.Name != "manager_test_script" {
		t.Errorf("Expected name 'manager_test_script', got %s", info.Name)
	}

	// 测试获取统计信息
	stats, exists := manager.GetScriptStats("manager_test_script")
	if !exists {
		t.Error("Expected script stats to exist")
	}

	if stats.Executions != 1 {
		t.Errorf("Expected 1 execution, got %d", stats.Executions)
	}

	if stats.SuccessRate != 100.0 {
		t.Errorf("Expected 100%% success rate, got %.2f", stats.SuccessRate)
	}

	// 测试重置统计
	manager.ResetStats("manager_test_script")
	stats, _ = manager.GetScriptStats("manager_test_script")
	if stats.Executions != 0 {
		t.Errorf("Expected 0 executions after reset, got %d", stats.Executions)
	}
}

func TestLuaScriptErrorHandling(t *testing.T) {
	config := DefaultConfig()
	config.Addr = "localhost:6379"

	client, err := New(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	script := client.NewScript()

	// 测试执行不存在的脚本
	opts := &ScriptOptions{
		Keys:        []string{},
		Args:        []interface{}{"test"},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	_, err = script.Execute(ctx, "nonexistent_script", opts)
	if err == nil {
		t.Error("Expected error for nonexistent script")
	}

	// 测试执行语法错误的脚本
	badScript := `return "Hello, " .. ARGV[1] .. " -- syntax error"`
	opts.UseCache = false

	result, err := script.ExecuteString(ctx, badScript, opts)
	if err != nil {
		t.Errorf("Execute bad script failed: %v", err)
	}

	// 脚本应该执行成功，因为语法是正确的
	if result.Value != "Hello, test -- syntax error" {
		t.Errorf("Expected 'Hello, test -- syntax error', got %v", result.Value)
	}
}

func TestLuaScriptCache(t *testing.T) {
	config := DefaultConfig()
	config.Addr = "localhost:6379"

	client, err := New(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	script := client.NewScript()

	// 测试脚本缓存
	scriptInfo := &ScriptInfo{
		Name:        "cache_test_script",
		Source:      `return "Cache test: " .. ARGV[1]`,
		Keys:        []string{},
		Args:        []string{"name"},
		Description: "缓存测试脚本",
		Timeout:     5 * time.Second,
	}

	err = script.Register(ctx, scriptInfo)
	if err != nil {
		t.Errorf("Register script failed: %v", err)
	}

	// 获取缓存信息
	cache := script.GetCache()
	if len(cache) == 0 {
		t.Error("Expected cache to have entries")
	}

	if cache["cache_test_script"] == "" {
		t.Error("Expected cache to contain script SHA")
	}

	// 清空缓存
	script.ClearCache()
	cache = script.GetCache()
	if len(cache) != 0 {
		t.Error("Expected cache to be empty after clear")
	}
}

func TestLuaScriptLoadAndExists(t *testing.T) {
	config := DefaultConfig()
	config.Addr = "localhost:6379"

	client, err := New(config)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	script := client.NewScript()

	// 测试加载脚本
	scriptSource := `return "Load test: " .. ARGV[1]`
	sha, err := script.Load(ctx, scriptSource)
	if err != nil {
		t.Errorf("Load script failed: %v", err)
	}

	if sha == "" {
		t.Error("Expected SHA to be set")
	}

	// 测试检查脚本是否存在
	exists, err := script.Exists(ctx, sha)
	if err != nil {
		t.Errorf("Check script existence failed: %v", err)
	}

	if !exists {
		t.Error("Expected script to exist")
	}

	// 测试检查不存在的脚本
	exists, err = script.Exists(ctx, "nonexistent_sha")
	if err != nil {
		t.Errorf("Check nonexistent script failed: %v", err)
	}

	if exists {
		t.Error("Expected nonexistent script to not exist")
	}
}
