package main

import (
	"context"
	"fmt"
	"time"

	"github.com/daxiong0327/tool-kit/mongodb"
)

func main() {
	fmt.Println("=== MongoDB 连接测试 ===")

	// 测试不同的连接字符串
	connectionStrings := []string{
		"mongodb://localhost:27017/awakening_memory",
		"mongodb://localhost:27017",
		"mongodb://127.0.0.1:27017",
	}

	for i, uri := range connectionStrings {
		fmt.Printf("\n--- 测试连接 %d: %s ---\n", i+1, uri)

		config := &mongodb.Config{
			URI:            uri,
			Database:       "awakening_memory",
			ConnectTimeout: 5 * time.Second,
			SocketTimeout:  3 * time.Second,
			MaxPoolSize:    10,
			RetryWrites:    false,
			RetryReads:     false,
			Debug:          false,
		}

		client, err := mongodb.New(config)
		if err != nil {
			fmt.Printf("❌ 连接失败: %v\n", err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 测试连接
		err = client.Ping(ctx)
		if err != nil {
			fmt.Printf("❌ Ping失败: %v\n", err)
			client.Close(ctx)
			continue
		}
		fmt.Printf("✅ 连接成功!\n")

		// 获取数据库信息
		db := client.GetDatabase()
		fmt.Printf("✅ 数据库: %s\n", db.Name())

		// 测试基本操作（只读）
		fmt.Println("测试只读操作...")

		// 尝试列出集合（只读操作）
		collections, err := db.ListCollectionNames(ctx, nil)
		if err != nil {
			fmt.Printf("⚠️  列出集合失败: %v\n", err)
		} else {
			fmt.Printf("✅ 找到 %d 个集合: %v\n", len(collections), collections)
		}

		// 尝试统计文档（只读操作）
		if len(collections) > 0 {
			collectionName := collections[0]
			count, err := client.CountDocuments(ctx, collectionName, map[string]interface{}{})
			if err != nil {
				fmt.Printf("⚠️  统计集合 %s 失败: %v\n", collectionName, err)
			} else {
				fmt.Printf("✅ 集合 %s 有 %d 个文档\n", collectionName, count)
			}
		}

		client.Close(ctx)
		fmt.Printf("✅ 连接 %d 测试完成\n", i+1)
	}

	fmt.Println("\n🎉 连接测试完成！")
}
