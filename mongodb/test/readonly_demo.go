package main

import (
	"context"
	"fmt"
	"time"

	"github.com/daxiong0327/tool-kit/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	fmt.Println("=== MongoDB 只读功能演示 ===")
	fmt.Println("连接字符串: mongodb://localhost:27017/awakening_memory")

	// 创建MongoDB客户端
	config := &mongodb.Config{
		URI:            "mongodb://localhost:27017/awakening_memory",
		Database:       "awakening_memory",
		ConnectTimeout: 10 * time.Second,
		SocketTimeout:  5 * time.Second,
		MaxPoolSize:    100,
		MinPoolSize:    10,
		RetryWrites:    false, // 禁用写操作
		RetryReads:     true,
		Debug:          false,
	}

	client, err := mongodb.New(config)
	if err != nil {
		fmt.Printf("❌ 连接MongoDB失败: %v\n", err)
		return
	}
	defer client.Close(context.Background())

	ctx := context.Background()

	// 测试连接
	fmt.Println("\n1. 测试连接...")
	err = client.Ping(ctx)
	if err != nil {
		fmt.Printf("❌ Ping失败: %v\n", err)
		return
	}
	fmt.Println("✅ 连接成功!")

	// 获取数据库信息
	fmt.Println("\n2. 获取数据库信息...")
	db := client.GetDatabase()
	fmt.Printf("✅ 当前数据库: %s\n", db.Name())

	// 列出集合
	fmt.Println("\n3. 列出数据库中的集合...")
	collections, err := db.ListCollectionNames(ctx, nil)
	if err != nil {
		fmt.Printf("⚠️  列出集合失败: %v\n", err)
	} else {
		fmt.Printf("✅ 找到 %d 个集合:\n", len(collections))
		for i, collection := range collections {
			fmt.Printf("  %d. %s\n", i+1, collection)
		}
	}

	// 测试每个集合
	if len(collections) > 0 {
		for i, collectionName := range collections {
			fmt.Printf("\n--- 测试集合 %d: %s ---\n", i+1, collectionName)

			// 统计文档数量
			count, err := client.CountDocuments(ctx, collectionName, bson.M{})
			if err != nil {
				fmt.Printf("⚠️  统计集合 %s 失败: %v\n", collectionName, err)
				continue
			}
			fmt.Printf("✅ 集合 %s 有 %d 个文档\n", collectionName, count)

			// 查询前几个文档
			if count > 0 {
				fmt.Printf("📋 查询集合 %s 的前几个文档:\n", collectionName)
				var docs []bson.M
				limit := int64(3)
				findOptions := &mongodb.FindOptions{
					Limit: &limit,
				}

				err = client.Find(ctx, collectionName, bson.M{}, &docs, findOptions)
				if err != nil {
					fmt.Printf("⚠️  查询集合 %s 失败: %v\n", collectionName, err)
				} else {
					fmt.Printf("✅ 找到 %d 个文档:\n", len(docs))
					for j, doc := range docs {
						fmt.Printf("  %d. %+v\n", j+1, doc)
					}
				}
			}

			// 测试聚合查询
			fmt.Printf("📊 测试集合 %s 的聚合查询:\n", collectionName)
			pipeline := []bson.M{
				{
					"$group": bson.M{
						"_id":   nil,
						"count": bson.M{"$sum": 1},
					},
				},
			}

			var aggregationResults []bson.M
			err = client.Aggregate(ctx, collectionName, pipeline, &aggregationResults)
			if err != nil {
				fmt.Printf("⚠️  聚合查询失败: %v\n", err)
			} else {
				fmt.Printf("✅ 聚合查询结果: %+v\n", aggregationResults)
			}
		}
	} else {
		fmt.Println("\n📝 数据库中没有集合，让我们测试创建集合的能力...")

		// 尝试创建一个测试集合（只读模式可能失败）
		testCollection := "test_readonly"
		fmt.Printf("尝试在集合 %s 中插入测试文档...\n", testCollection)

		testDoc := bson.M{
			"name":      "只读测试",
			"message":   "这是一个只读测试文档",
			"timestamp": time.Now(),
		}

		result, err := client.InsertOne(ctx, testCollection, testDoc)
		if err != nil {
			fmt.Printf("⚠️  插入失败 (预期): %v\n", err)
		} else {
			fmt.Printf("✅ 插入成功，ID: %s\n", result.InsertedID.Hex())
		}
	}

	// 测试配置获取
	fmt.Println("\n4. 测试配置获取...")
	config = client.GetConfig()
	fmt.Printf("✅ 客户端配置:\n")
	fmt.Printf("  - URI: %s\n", config.URI)
	fmt.Printf("  - Database: %s\n", config.Database)
	fmt.Printf("  - ConnectTimeout: %v\n", config.ConnectTimeout)
	fmt.Printf("  - SocketTimeout: %v\n", config.SocketTimeout)
	fmt.Printf("  - MaxPoolSize: %d\n", config.MaxPoolSize)
	fmt.Printf("  - MinPoolSize: %d\n", config.MinPoolSize)
	fmt.Printf("  - RetryWrites: %t\n", config.RetryWrites)
	fmt.Printf("  - RetryReads: %t\n", config.RetryReads)
	fmt.Printf("  - Debug: %t\n", config.Debug)

	// 测试连接池信息
	fmt.Println("\n5. 测试连接池信息...")
	fmt.Printf("✅ 连接池配置:\n")
	fmt.Printf("  - 最大连接数: %d\n", config.MaxPoolSize)
	fmt.Printf("  - 最小连接数: %d\n", config.MinPoolSize)
	fmt.Printf("  - 连接超时: %v\n", config.ConnectTimeout)
	fmt.Printf("  - Socket超时: %v\n", config.SocketTimeout)

	fmt.Println("\n🎉 MongoDB只读功能演示完成！")
	fmt.Println("📝 注意: 由于MongoDB需要认证才能进行写操作，所以只演示了只读功能")
	fmt.Println("💡 要测试完整的CRUD功能，请配置MongoDB认证或使用无认证的MongoDB实例")
}
