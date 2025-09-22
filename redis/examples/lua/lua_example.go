package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/daxiong0327/tool-kit/redis"
)

func main() {
	// 创建Redis客户端
	config := redis.DefaultConfig()
	config.Addr = "localhost:6379"
	config.Password = ""
	config.DB = 0

	client, err := redis.New(config)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 测试连接
	err = client.Ping(ctx)
	if err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}

	fmt.Println("✅ Redis连接成功!")

	// 1. 基本Lua脚本使用
	fmt.Println("\n=== 基本Lua脚本使用 ===")
	basicLuaExample(ctx, client)

	// 2. 脚本模板使用
	fmt.Println("\n=== 脚本模板使用 ===")
	templateLuaExample(ctx, client)

	// 3. 脚本管理器使用
	fmt.Println("\n=== 脚本管理器使用 ===")
	managerLuaExample(ctx, client)

	// 4. 实际应用场景
	fmt.Println("\n=== 实际应用场景 ===")
	realWorldLuaExample(ctx, client)

	// 5. 清理测试数据
	fmt.Println("\n=== 清理测试数据 ===")
	cleanupLuaExample(ctx, client)

	fmt.Println("\n🎉 Lua脚本示例执行完成!")
}

// 基本Lua脚本使用示例
func basicLuaExample(ctx context.Context, client *redis.Client) {
	script := client.NewScript()

	// 注册一个简单的脚本
	scriptInfo := &redis.ScriptInfo{
		Name:        "hello_script",
		Source:      `return "Hello, " .. ARGV[1] .. "! Current time: " .. redis.call('TIME')[1]`,
		Keys:        []string{},
		Args:        []string{"name"},
		Description: "问候脚本",
		Timeout:     5 * time.Second,
	}

	err := script.Register(ctx, scriptInfo)
	if err != nil {
		log.Printf("Register script failed: %v", err)
		return
	}

	fmt.Println("✅ 脚本注册成功")

	// 执行脚本
	opts := &redis.ScriptOptions{
		Keys:        []string{},
		Args:        []interface{}{"Lua"},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	result, err := script.Execute(ctx, "hello_script", opts)
	if err != nil {
		log.Printf("Execute script failed: %v", err)
		return
	}

	fmt.Printf("脚本执行结果: %v\n", result.Value)
	fmt.Printf("执行时间: %v\n", result.Time)
	fmt.Printf("脚本SHA: %s\n", result.SHA)

	// 直接执行脚本字符串
	scriptSource := `return "Direct execution: " .. ARGV[1]`
	result, err = script.ExecuteString(ctx, scriptSource, opts)
	if err != nil {
		log.Printf("Execute string script failed: %v", err)
		return
	}

	fmt.Printf("直接执行结果: %v\n", result.Value)
}

// 脚本模板使用示例
func templateLuaExample(ctx context.Context, client *redis.Client) {
	script := client.NewScript()
	templates := script.NewScriptTemplates()

	// 注册常用脚本模板
	err := templates.RegisterCommonScripts(ctx)
	if err != nil {
		log.Printf("Register common scripts failed: %v", err)
		return
	}

	fmt.Println("✅ 常用脚本模板注册成功")

	// 测试分布式锁
	fmt.Println("\n--- 分布式锁测试 ---")
	lockKey := "test:distributed:lock"
	lockValue := "client-001"
	lockTTL := 10

	lockOpts := &redis.ScriptOptions{
		Keys:        []string{lockKey},
		Args:        []interface{}{lockValue, lockTTL},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	result, err := script.Execute(ctx, "distributed_lock", lockOpts)
	if err != nil {
		log.Printf("Execute distributed lock failed: %v", err)
		return
	}

	if result.Value == int64(1) {
		fmt.Println("✅ 成功获取分布式锁")
		
		// 模拟业务处理
		time.Sleep(2 * time.Second)
		
		// 释放锁
		unlockOpts := &redis.ScriptOptions{
			Keys:        []string{lockKey},
			Args:        []interface{}{lockValue},
			Timeout:     5 * time.Second,
			RetryCount:  3,
			RetryDelay:  100 * time.Millisecond,
			UseCache:    true,
			ForceReload: false,
		}

		result, err = script.Execute(ctx, "distributed_unlock", unlockOpts)
		if err != nil {
			log.Printf("Execute distributed unlock failed: %v", err)
			return
		}

		if result.Value == int64(1) {
			fmt.Println("✅ 成功释放分布式锁")
		}
	} else {
		fmt.Println("❌ 获取分布式锁失败")
	}

	// 测试限流
	fmt.Println("\n--- 限流测试 ---")
	rateLimitKey := "test:rate_limit"
	limit := 5
	window := 60

	rateLimitOpts := &redis.ScriptOptions{
		Keys:        []string{rateLimitKey},
		Args:        []interface{}{limit, window},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	for i := 0; i < 7; i++ {
		result, err := script.Execute(ctx, "rate_limit", rateLimitOpts)
		if err != nil {
			log.Printf("Execute rate limit failed: %v", err)
			continue
		}

		results := result.Value.([]interface{})
		allowed := results[0].(int64)
		remaining := results[1].(int64)

		if allowed == 1 {
			fmt.Printf("✅ 请求 %d 通过，剩余次数: %d\n", i+1, remaining)
		} else {
			fmt.Printf("❌ 请求 %d 被限流，剩余次数: %d\n", i+1, remaining)
		}

		time.Sleep(100 * time.Millisecond)
	}

	// 测试计数器
	fmt.Println("\n--- 计数器测试 ---")
	counterKey := "test:counter"
	counterTTL := 30

	counterOpts := &redis.ScriptOptions{
		Keys:        []string{counterKey},
		Args:        []interface{}{1, counterTTL},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	for i := 0; i < 5; i++ {
		result, err := script.Execute(ctx, "counter", counterOpts)
		if err != nil {
			log.Printf("Execute counter failed: %v", err)
			continue
		}

		fmt.Printf("计数器值: %v\n", result.Value)
		time.Sleep(100 * time.Millisecond)
	}
}

// 脚本管理器使用示例
func managerLuaExample(ctx context.Context, client *redis.Client) {
	script := client.NewScript()
	manager := script.NewScriptManager()

	// 注册脚本
	scriptInfo := &redis.ScriptInfo{
		Name:        "manager_test_script",
		Source:      `return "Manager test: " .. ARGV[1] .. " (execution #" .. redis.call('INCR', 'test:execution_count') .. ")"`,
		Keys:        []string{},
		Args:        []string{"name"},
		Description: "管理器测试脚本",
		Timeout:     5 * time.Second,
	}

	err := manager.RegisterScript(ctx, scriptInfo)
	if err != nil {
		log.Printf("Register script failed: %v", err)
		return
	}

	fmt.Println("✅ 脚本注册成功")

	// 执行脚本多次
	opts := &redis.ScriptOptions{
		Keys:        []string{},
		Args:        []interface{}{"Lua Manager"},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	for i := 0; i < 5; i++ {
		result, err := manager.ExecuteScript(ctx, "manager_test_script", opts)
		if err != nil {
			log.Printf("Execute script failed: %v", err)
			continue
		}

		fmt.Printf("执行结果 %d: %v\n", i+1, result.Value)
		time.Sleep(100 * time.Millisecond)
	}

	// 获取统计信息
	stats, exists := manager.GetScriptStats("manager_test_script")
	if exists {
		fmt.Printf("\n脚本统计信息:\n")
		fmt.Printf("  执行次数: %d\n", stats.Executions)
		fmt.Printf("  平均执行时间: %v\n", stats.AverageTime)
		fmt.Printf("  最大执行时间: %v\n", stats.MaxTime)
		fmt.Printf("  最小执行时间: %v\n", stats.MinTime)
		fmt.Printf("  错误次数: %d\n", stats.Errors)
		fmt.Printf("  成功率: %.2f%%\n", stats.SuccessRate)
		fmt.Printf("  缓存命中次数: %d\n", stats.CacheHits)
		fmt.Printf("  缓存未命中次数: %d\n", stats.CacheMisses)
	}

	// 添加告警规则
	alertRule := &redis.AlertRule{
		Name:        "high_error_rate",
		ScriptName:  "manager_test_script",
		Condition:   "error_rate",
		Threshold:   50.0,
		Duration:    1 * time.Minute,
		Enabled:     true,
	}

	manager.AddAlertRule(alertRule)
	fmt.Println("✅ 告警规则添加成功")

	// 启动监控
	manager.StartMonitor()
	fmt.Println("✅ 脚本监控启动成功")

	// 等待一段时间让监控运行
	time.Sleep(2 * time.Second)

	// 停止监控
	manager.StopMonitor()
	fmt.Println("✅ 脚本监控停止")
}

// 实际应用场景示例
func realWorldLuaExample(ctx context.Context, client *redis.Client) {
	script := client.NewScript()
	templates := script.NewScriptTemplates()

	// 注册常用脚本
	err := templates.RegisterCommonScripts(ctx)
	if err != nil {
		log.Printf("Register common scripts failed: %v", err)
		return
	}

	// 1. 用户会话管理
	fmt.Println("\n--- 用户会话管理 ---")
	sessionKey := "user:session:123"
	sessionData := map[string]interface{}{
		"user_id":    "123",
		"username":   "张三",
		"login_time": time.Now().Unix(),
		"ip_address": "192.168.1.100",
	}

	// 使用原子哈希设置脚本
	hashSetOpts := &redis.ScriptOptions{
		Keys:        []string{sessionKey},
		Args:        []interface{}{"user_id", sessionData["user_id"], 300},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	result, err := script.Execute(ctx, "atomic_hash_set", hashSetOpts)
	if err != nil {
		log.Printf("Execute hash set failed: %v", err)
		return
	}

	fmt.Printf("会话设置结果: %v\n", result.Value)

	// 2. 商品库存管理
	fmt.Println("\n--- 商品库存管理 ---")
	productKey := "product:stock:1001"
	initialStock := 100

	// 设置初始库存
	stringOps := client.NewString()
	err = stringOps.Set(ctx, productKey, initialStock, 0)
	if err != nil {
		log.Printf("Set initial stock failed: %v", err)
		return
	}

	// 使用原子自增脚本扣减库存
	atomicIncrementOpts := &redis.ScriptOptions{
		Keys:        []string{productKey},
		Args:        []interface{}{-1, 0}, // 扣减1，最小值为0
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	for i := 0; i < 5; i++ {
		result, err := script.Execute(ctx, "atomic_increment", atomicIncrementOpts)
		if err != nil {
			log.Printf("Execute atomic increment failed: %v", err)
			continue
		}

		results := result.Value.([]interface{})
		newStock := results[0].(int64)
		success := results[1].(int64)

		if success == 1 {
			fmt.Printf("扣减库存成功，当前库存: %d\n", newStock)
		} else {
			fmt.Printf("扣减库存失败，当前库存: %d\n", newStock)
		}

		time.Sleep(100 * time.Millisecond)
	}

	// 3. 排行榜管理
	fmt.Println("\n--- 排行榜管理 ---")
	leaderboardKey := "game:leaderboard"

	// 添加玩家分数
	zsetAddOpts := &redis.ScriptOptions{
		Keys:        []string{leaderboardKey},
		Args:        []interface{}{10, 1000, "玩家A", 1500, "玩家B", 800, "玩家C"},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	result, err = script.Execute(ctx, "atomic_zset_add", zsetAddOpts)
	if err != nil {
		log.Printf("Execute zset add failed: %v", err)
		return
	}

	fmt.Printf("排行榜添加结果: %v\n", result.Value)

	// 获取排行榜
	zsetRangeOpts := &redis.ScriptOptions{
		Keys:        []string{leaderboardKey},
		Args:        []interface{}{0, 2, "true"}, // 前3名，带分数
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	result, err = script.Execute(ctx, "atomic_zset_range", zsetRangeOpts)
	if err != nil {
		log.Printf("Execute zset range failed: %v", err)
		return
	}

	fmt.Printf("排行榜前3名: %v\n", result.Value)

	// 4. 消息队列管理
	fmt.Println("\n--- 消息队列管理 ---")
	queueKey := "message:queue"

	// 添加消息到队列
	listPushOpts := &redis.ScriptOptions{
		Keys:        []string{queueKey},
		Args:        []interface{}{"right", 100, "消息1", "消息2", "消息3"},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	result, err = script.Execute(ctx, "atomic_list_push", listPushOpts)
	if err != nil {
		log.Printf("Execute list push failed: %v", err)
		return
	}

	fmt.Printf("消息队列添加结果: %v\n", result.Value)

	// 从队列获取消息
	listPopOpts := &redis.ScriptOptions{
		Keys:        []string{queueKey},
		Args:        []interface{}{"left", 1},
		Timeout:     5 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		UseCache:    true,
		ForceReload: false,
	}

	for i := 0; i < 3; i++ {
		result, err := script.Execute(ctx, "atomic_list_pop", listPopOpts)
		if err != nil {
			log.Printf("Execute list pop failed: %v", err)
			continue
		}

		results := result.Value.([]interface{})
		message := results[0]
		success := results[1].(int64)

		if success == 1 {
			fmt.Printf("获取消息: %v\n", message)
		} else {
			fmt.Printf("队列为空\n")
			break
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// 清理测试数据
func cleanupLuaExample(ctx context.Context, client *redis.Client) {
	keys := []string{
		"test:distributed:lock",
		"test:rate_limit",
		"test:counter",
		"user:session:123",
		"product:stock:1001",
		"game:leaderboard",
		"message:queue",
		"test:execution_count",
	}

	deleted, err := client.Del(ctx, keys...)
	if err != nil {
		log.Printf("Del failed: %v", err)
		return
	}

	fmt.Printf("清理了 %d 个测试键\n", deleted)
}
