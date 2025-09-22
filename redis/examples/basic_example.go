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

	// 1. 字符串操作示例
	fmt.Println("\n=== 字符串操作示例 ===")
	stringOps := client.NewString()

	// 设置键值
	err = stringOps.Set(ctx, "user:name", "张三", 0)
	if err != nil {
		log.Printf("Set failed: %v", err)
	}

	// 获取值
	name, err := stringOps.Get(ctx, "user:name")
	if err != nil {
		log.Printf("Get failed: %v", err)
	} else {
		fmt.Printf("用户名: %s\n", name)
	}

	// 数值操作
	err = stringOps.Set(ctx, "user:age", "25", 0)
	if err != nil {
		log.Printf("Set failed: %v", err)
	}

	age, err := stringOps.GetInt(ctx, "user:age")
	if err != nil {
		log.Printf("GetInt failed: %v", err)
	} else {
		fmt.Printf("年龄: %d\n", age)
	}

	// 自增操作
	newAge, err := stringOps.IncrBy(ctx, "user:age", 1)
	if err != nil {
		log.Printf("IncrBy failed: %v", err)
	} else {
		fmt.Printf("年龄+1: %d\n", newAge)
	}

	// 2. 哈希操作示例
	fmt.Println("\n=== 哈希操作示例 ===")
	hashOps := client.NewHash()

	// 设置哈希字段
	_, err = hashOps.HSet(ctx, "user:profile", "name", "李四", "age", "30", "city", "北京")
	if err != nil {
		log.Printf("HSet failed: %v", err)
	}

	// 获取所有字段
	profile, err := hashOps.HGetAll(ctx, "user:profile")
	if err != nil {
		log.Printf("HGetAll failed: %v", err)
	} else {
		fmt.Printf("用户资料: %+v\n", profile)
	}

	// 获取单个字段
	city, err := hashOps.HGet(ctx, "user:profile", "city")
	if err != nil {
		log.Printf("HGet failed: %v", err)
	} else {
		fmt.Printf("城市: %s\n", city)
	}

	// 3. 列表操作示例
	fmt.Println("\n=== 列表操作示例 ===")
	listOps := client.NewList()

	// 推入元素
	_, err = listOps.LPush(ctx, "todo:list", "学习Go", "写代码", "测试")
	if err != nil {
		log.Printf("LPush failed: %v", err)
	}

	// 获取列表长度
	length, err := listOps.LLen(ctx, "todo:list")
	if err != nil {
		log.Printf("LLen failed: %v", err)
	} else {
		fmt.Printf("待办事项数量: %d\n", length)
	}

	// 获取列表内容
	items, err := listOps.LRange(ctx, "todo:list", 0, -1)
	if err != nil {
		log.Printf("LRange failed: %v", err)
	} else {
		fmt.Printf("待办事项: %+v\n", items)
	}

	// 弹出元素
	item, err := listOps.LPop(ctx, "todo:list")
	if err != nil {
		log.Printf("LPop failed: %v", err)
	} else {
		fmt.Printf("完成事项: %s\n", item)
	}

	// 4. 集合操作示例
	fmt.Println("\n=== 集合操作示例 ===")
	setOps := client.NewSet()

	// 添加成员
	_, err = setOps.SAdd(ctx, "user:tags", "程序员", "Go开发者", "开源爱好者")
	if err != nil {
		log.Printf("SAdd failed: %v", err)
	}

	// 获取所有成员
	tags, err := setOps.SMembers(ctx, "user:tags")
	if err != nil {
		log.Printf("SMembers failed: %v", err)
	} else {
		fmt.Printf("用户标签: %+v\n", tags)
	}

	// 检查成员是否存在
	isMember, err := setOps.SIsMember(ctx, "user:tags", "程序员")
	if err != nil {
		log.Printf("SIsMember failed: %v", err)
	} else {
		fmt.Printf("是程序员吗: %t\n", isMember)
	}

	// 5. 有序集合操作示例
	fmt.Println("\n=== 有序集合操作示例 ===")
	zsetOps := client.NewZSet()

	// 添加成员
	_, err = zsetOps.ZAdd(ctx, "leaderboard", redispkg.Z{Score: 100, Member: "玩家A"}, redispkg.Z{Score: 200, Member: "玩家B"}, redispkg.Z{Score: 150, Member: "玩家C"})
	if err != nil {
		log.Printf("ZAdd failed: %v", err)
	}

	// 获取排行榜（按分数排序）
	leaderboard, err := zsetOps.ZRevRangeWithScores(ctx, "leaderboard", 0, -1)
	if err != nil {
		log.Printf("ZRevRangeWithScores failed: %v", err)
	} else {
		fmt.Println("排行榜:")
		for i, member := range leaderboard {
			fmt.Printf("  %d. %s (分数: %.0f)\n", i+1, member.Member, member.Score)
		}
	}

	// 6. 管道操作示例
	fmt.Println("\n=== 管道操作示例 ===")
	pipeline := client.NewPipeline()

	// 添加多个命令到管道
	pipeline.Set(ctx, "pipeline:key1", "value1", 0)
	pipeline.Set(ctx, "pipeline:key2", "value2", 0)
	pipeline.Set(ctx, "pipeline:key3", "value3", 0)
	pipeline.Get(ctx, "pipeline:key1")
	pipeline.Get(ctx, "pipeline:key2")
	pipeline.Get(ctx, "pipeline:key3")

	// 执行管道
	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		log.Printf("Pipeline exec failed: %v", err)
	} else {
		fmt.Printf("管道执行成功，执行了 %d 个命令\n", len(cmds))
	}

	// 7. 发布订阅示例
	fmt.Println("\n=== 发布订阅示例 ===")
	publisher := client.NewPublisher()
	subscriber := client.NewSubscriber()

	// 启动订阅者
	go func() {
		err := subscriber.Subscribe(ctx, "news")
		if err != nil {
			log.Printf("Subscribe failed: %v", err)
			return
		}

		err = subscriber.Listen(ctx, func(msg *redispkg.Message) {
			fmt.Printf("收到消息: %s\n", msg.Payload)
		})
		if err != nil {
			log.Printf("Listen failed: %v", err)
		}
	}()

	// 等待订阅者准备就绪
	time.Sleep(100 * time.Millisecond)

	// 发布消息
	_, err = publisher.Publish(ctx, "news", "Hello Redis!")
	if err != nil {
		log.Printf("Publish failed: %v", err)
	} else {
		fmt.Println("消息发布成功")
	}

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 关闭订阅者
	subscriber.Close()

	// 8. 事务操作示例
	fmt.Println("\n=== 事务操作示例 ===")

	// 使用事务
	err = client.WithTransaction(ctx, func(tx *redispkg.Tx) error {
		// 在事务中执行多个操作
		tx.Set(ctx, "tx:key1", "value1", 0)
		tx.Set(ctx, "tx:key2", "value2", 0)
		tx.Set(ctx, "tx:key3", "value3", 0)
		return nil
	}, "tx:key1", "tx:key2", "tx:key3")

	if err != nil {
		log.Printf("Transaction failed: %v", err)
	} else {
		fmt.Println("事务执行成功")
	}

	// 9. 清理测试数据
	fmt.Println("\n=== 清理测试数据 ===")
	keys := []string{
		"user:name", "user:age", "user:profile", "todo:list", "user:tags",
		"leaderboard", "pipeline:key1", "pipeline:key2", "pipeline:key3",
		"tx:key1", "tx:key2", "tx:key3",
	}

	deleted, err := client.Del(ctx, keys...)
	if err != nil {
		log.Printf("Del failed: %v", err)
	} else {
		fmt.Printf("清理了 %d 个键\n", deleted)
	}

	fmt.Println("\n🎉 Redis示例执行完成!")
}
