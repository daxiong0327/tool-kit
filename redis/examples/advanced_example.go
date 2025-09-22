package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/daxiong0327/tool-kit/redis"
	redispkg "github.com/redis/go-redis/v9"
)

func main() {
	// 创建Redis客户端
	config := redis.DefaultConfig()
	config.Addr = "localhost:6379"
	config.Password = ""
	config.DB = 0
	config.PoolSize = 20
	config.MinIdleConns = 5

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

	// 1. 缓存示例
	fmt.Println("\n=== 缓存示例 ===")
	cacheExample(ctx, client)

	// 2. 计数器示例
	fmt.Println("\n=== 计数器示例 ===")
	counterExample(ctx, client)

	// 3. 分布式锁示例
	fmt.Println("\n=== 分布式锁示例 ===")
	distributedLockExample(ctx, client)

	// 4. 消息队列示例
	fmt.Println("\n=== 消息队列示例 ===")
	messageQueueExample(ctx, client)

	// 5. 排行榜示例
	fmt.Println("\n=== 排行榜示例 ===")
	leaderboardExample(ctx, client)

	// 6. 会话管理示例
	fmt.Println("\n=== 会话管理示例 ===")
	sessionExample(ctx, client)

	// 7. 限流示例
	fmt.Println("\n=== 限流示例 ===")
	rateLimitExample(ctx, client)

	// 8. 清理测试数据
	fmt.Println("\n=== 清理测试数据 ===")
	cleanupExample(ctx, client)

	fmt.Println("\n🎉 Redis高级示例执行完成!")
}

// 缓存示例
func cacheExample(ctx context.Context, client *redis.Client) {
	stringOps := client.NewString()

	// 模拟缓存数据
	cacheKey := "cache:user:123"
	userData := `{"id": 123, "name": "张三", "email": "zhangsan@example.com"}`

	// 设置缓存，过期时间5分钟
	err := stringOps.Set(ctx, cacheKey, userData, 5*time.Minute)
	if err != nil {
		log.Printf("Set cache failed: %v", err)
		return
	}

	// 获取缓存
	cachedData, err := stringOps.Get(ctx, cacheKey)
	if err != nil {
		log.Printf("Get cache failed: %v", err)
		return
	}

	fmt.Printf("缓存数据: %s\n", cachedData)

	// 检查缓存是否存在
	exists, err := client.Exists(ctx, cacheKey)
	if err != nil {
		log.Printf("Exists failed: %v", err)
		return
	}

	fmt.Printf("缓存存在: %t\n", exists == 1)

	// 获取TTL
	ttl, err := client.TTL(ctx, cacheKey)
	if err != nil {
		log.Printf("TTL failed: %v", err)
		return
	}

	fmt.Printf("缓存TTL: %v\n", ttl)
}

// 计数器示例
func counterExample(ctx context.Context, client *redis.Client) {
	stringOps := client.NewString()

	counterKey := "counter:page:views"

	// 页面访问计数
	for i := 0; i < 5; i++ {
		count, err := stringOps.Incr(ctx, counterKey)
		if err != nil {
			log.Printf("Incr failed: %v", err)
			return
		}
		fmt.Printf("页面访问次数: %d\n", count)
		time.Sleep(100 * time.Millisecond)
	}

	// 获取最终计数
	finalCount, err := stringOps.GetInt(ctx, counterKey)
	if err != nil {
		log.Printf("GetInt failed: %v", err)
		return
	}

	fmt.Printf("最终页面访问次数: %d\n", finalCount)

	// 重置计数器
	err = stringOps.Set(ctx, counterKey, "0", 0)
	if err != nil {
		log.Printf("Set failed: %v", err)
		return
	}

	fmt.Println("计数器已重置")
}

// 分布式锁示例
func distributedLockExample(ctx context.Context, client *redis.Client) {
	stringOps := client.NewString()

	lockKey := "lock:resource:123"
	lockValue := "client-001"
	lockTimeout := 10 * time.Second

	// 尝试获取锁
	success, err := stringOps.SetNX(ctx, lockKey, lockValue, lockTimeout)
	if err != nil {
		log.Printf("SetNX failed: %v", err)
		return
	}

	if success {
		fmt.Println("✅ 成功获取分布式锁")

		// 模拟业务处理
		fmt.Println("执行业务逻辑...")
		time.Sleep(2 * time.Second)

		// 释放锁
		client.Del(ctx, lockKey)
		fmt.Println("✅ 分布式锁已释放")
	} else {
		fmt.Println("❌ 获取分布式锁失败，资源被占用")
	}
}

// 消息队列示例
func messageQueueExample(ctx context.Context, client *redis.Client) {
	listOps := client.NewList()
	publisher := client.NewPublisher()
	subscriber := client.NewSubscriber()

	queueKey := "queue:messages"

	// 启动消费者
	go func() {
		err := subscriber.Subscribe(ctx, "queue:notify")
		if err != nil {
			log.Printf("Subscribe failed: %v", err)
			return
		}

		err = subscriber.Listen(ctx, func(msg *redispkg.Message) {
			fmt.Printf("收到通知: %s\n", msg.Payload)
		})
		if err != nil {
			log.Printf("Listen failed: %v", err)
		}
	}()

	// 等待订阅者准备就绪
	time.Sleep(100 * time.Millisecond)

	// 生产者：添加消息到队列
	messages := []string{"消息1", "消息2", "消息3", "消息4", "消息5"}
	for _, msg := range messages {
		_, err := listOps.LPush(ctx, queueKey, msg)
		if err != nil {
			log.Printf("LPush failed: %v", err)
			continue
		}
		fmt.Printf("发送消息: %s\n", msg)
	}

	// 消费者：处理队列中的消息
	fmt.Println("开始处理队列消息...")
	for i := 0; i < len(messages); i++ {
		msg, err := listOps.RPop(ctx, queueKey)
		if err != nil {
			log.Printf("RPop failed: %v", err)
			continue
		}
		fmt.Printf("处理消息: %s\n", msg)

		// 发送处理完成通知
		publisher.Publish(ctx, "queue:notify", fmt.Sprintf("消息 %s 处理完成", msg))
		time.Sleep(200 * time.Millisecond)
	}

	// 关闭订阅者
	subscriber.Close()
}

// 排行榜示例
func leaderboardExample(ctx context.Context, client *redis.Client) {
	zsetOps := client.NewZSet()

	leaderboardKey := "leaderboard:game"

	// 添加玩家分数
	players := []struct {
		name  string
		score float64
	}{
		{"玩家A", 1000},
		{"玩家B", 1500},
		{"玩家C", 800},
		{"玩家D", 2000},
		{"玩家E", 1200},
	}

	for _, player := range players {
		_, err := zsetOps.ZAdd(ctx, leaderboardKey, redispkg.Z{Score: player.score, Member: player.name})
		if err != nil {
			log.Printf("ZAdd failed: %v", err)
			continue
		}
		fmt.Printf("添加玩家: %s (分数: %.0f)\n", player.name, player.score)
	}

	// 获取排行榜前3名
	topPlayers, err := zsetOps.ZRevRangeWithScores(ctx, leaderboardKey, 0, 2)
	if err != nil {
		log.Printf("ZRevRangeWithScores failed: %v", err)
		return
	}

	fmt.Println("排行榜前3名:")
	for i, player := range topPlayers {
		fmt.Printf("  %d. %s (分数: %.0f)\n", i+1, player.Member, player.Score)
	}

	// 获取玩家排名
	rank, err := zsetOps.ZRevRank(ctx, leaderboardKey, "玩家B")
	if err != nil {
		log.Printf("ZRevRank failed: %v", err)
		return
	}

	fmt.Printf("玩家B的排名: %d\n", rank+1)

	// 更新玩家分数
	newScore, err := zsetOps.ZIncrBy(ctx, leaderboardKey, 500, "玩家B")
	if err != nil {
		log.Printf("ZIncrBy failed: %v", err)
		return
	}

	fmt.Printf("玩家B新分数: %.0f\n", newScore)
}

// 会话管理示例
func sessionExample(ctx context.Context, client *redis.Client) {
	hashOps := client.NewHash()

	sessionKey := "session:user:123"
	sessionTimeout := 30 * time.Minute

	// 创建会话
	sessionData := map[string]interface{}{
		"user_id":    "123",
		"username":   "张三",
		"login_time": time.Now().Unix(),
		"ip_address": "192.168.1.100",
		"user_agent": "Mozilla/5.0...",
	}

	_, err := hashOps.HSet(ctx, sessionKey, sessionData)
	if err != nil {
		log.Printf("HSet failed: %v", err)
		return
	}

	// 设置会话过期时间
	client.Expire(ctx, sessionKey, sessionTimeout)

	fmt.Println("✅ 会话创建成功")

	// 获取会话信息
	session, err := hashOps.HGetAll(ctx, sessionKey)
	if err != nil {
		log.Printf("HGetAll failed: %v", err)
		return
	}

	fmt.Printf("会话信息: %+v\n", session)

	// 更新会话活动时间
	hashOps.HSet(ctx, sessionKey, "last_activity", time.Now().Unix())

	// 检查会话是否存在
	exists, err := hashOps.HExists(ctx, sessionKey, "user_id")
	if err != nil {
		log.Printf("HExists failed: %v", err)
		return
	}

	fmt.Printf("会话存在: %t\n", exists)

	// 获取会话TTL
	ttl, err := client.TTL(ctx, sessionKey)
	if err != nil {
		log.Printf("TTL failed: %v", err)
		return
	}

	fmt.Printf("会话TTL: %v\n", ttl)
}

// 限流示例
func rateLimitExample(ctx context.Context, client *redis.Client) {
	stringOps := client.NewString()

	rateLimitKey := "rate_limit:api:user:123"
	limit := 10  // 限制10次请求
	window := 60 // 时间窗口60秒
	windowKey := fmt.Sprintf("%s:%d", rateLimitKey, time.Now().Unix()/int64(window))

	// 检查当前请求数
	current, err := stringOps.GetInt(ctx, windowKey)
	if err != nil && err != redispkg.Nil {
		log.Printf("GetInt failed: %v", err)
		return
	}

	if current >= int64(limit) {
		fmt.Printf("❌ 请求被限流，当前请求数: %d，限制: %d\n", current, limit)
		return
	}

	// 增加请求计数
	newCount, err := stringOps.Incr(ctx, windowKey)
	if err != nil {
		log.Printf("Incr failed: %v", err)
		return
	}

	// 设置过期时间
	if newCount == 1 {
		client.Expire(ctx, windowKey, time.Duration(window)*time.Second)
	}

	fmt.Printf("✅ 请求通过，当前请求数: %d/%d\n", newCount, limit)

	// 模拟多次请求
	for i := 0; i < 5; i++ {
		current, err := stringOps.GetInt(ctx, windowKey)
		if err != nil && err != redispkg.Nil {
			log.Printf("GetInt failed: %v", err)
			continue
		}

		if current >= int64(limit) {
			fmt.Printf("❌ 请求被限流，当前请求数: %d\n", current)
			break
		}

		newCount, err := stringOps.Incr(ctx, windowKey)
		if err != nil {
			log.Printf("Incr failed: %v", err)
			continue
		}

		fmt.Printf("✅ 请求通过，当前请求数: %d/%d\n", newCount, limit)
		time.Sleep(100 * time.Millisecond)
	}
}

// 清理示例
func cleanupExample(ctx context.Context, client *redis.Client) {
	// 获取所有测试键
	keys, err := client.Keys(ctx, "cache:*")
	if err != nil {
		log.Printf("Keys failed: %v", err)
		return
	}

	keys = append(keys, "counter:*", "lock:*", "queue:*", "leaderboard:*", "session:*", "rate_limit:*")

	// 删除所有测试键
	for _, pattern := range keys {
		patternKeys, err := client.Keys(ctx, pattern)
		if err != nil {
			log.Printf("Keys failed: %v", err)
			continue
		}

		if len(patternKeys) > 0 {
			deleted, err := client.Del(ctx, patternKeys...)
			if err != nil {
				log.Printf("Del failed: %v", err)
				continue
			}
			fmt.Printf("清理了 %d 个 %s 键\n", deleted, pattern)
		}
	}

	fmt.Println("✅ 测试数据清理完成")
}
