package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/daxiong0327/tool-kit/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	fmt.Println("=== MongoDB 简单连接测试 ===")
	fmt.Println("连接字符串: mongodb://localhost:27017/awakening_memory")

	// 创建MongoDB客户端
	config := &mongodb.Config{
		URI:            "mongodb://localhost:27017/awakening_memory",
		Database:       "awakening_memory",
		ConnectTimeout: 10 * time.Second,
		SocketTimeout:  5 * time.Second,
		MaxPoolSize:    100,
		MinPoolSize:    10,
		RetryWrites:    true,
		RetryReads:     true,
		Debug:          true,
	}

	client, err := mongodb.New(config)
	if err != nil {
		log.Fatalf("❌ 连接MongoDB失败: %v", err)
	}
	defer client.Close(context.Background())

	ctx := context.Background()

	// 测试连接
	fmt.Println("\n1. 测试连接...")
	err = client.Ping(ctx)
	if err != nil {
		log.Fatalf("❌ Ping失败: %v", err)
	}
	fmt.Println("✅ 连接成功!")

	// 测试数据库信息
	fmt.Println("\n2. 获取数据库信息...")
	db := client.GetDatabase()
	fmt.Printf("✅ 当前数据库: %s\n", db.Name())

	// 测试集合操作
	collection := "test_collection"
	fmt.Printf("\n3. 测试集合操作 (集合: %s)...\n", collection)

	// 清理测试数据
	fmt.Println("🧹 清理测试数据...")
	_, err = client.DeleteMany(ctx, collection, bson.M{})
	if err != nil {
		log.Printf("⚠️  清理数据失败: %v", err)
	} else {
		fmt.Println("✅ 清理完成")
	}

	// 插入测试数据
	fmt.Println("\n4. 插入测试数据...")
	document := bson.M{
		"name":      "测试文档",
		"message":   "Hello MongoDB!",
		"timestamp": time.Now(),
		"number":    42,
	}

	result, err := client.InsertOne(ctx, collection, document)
	if err != nil {
		log.Fatalf("❌ 插入失败: %v", err)
	}
	fmt.Printf("✅ 插入成功，ID: %s\n", result.InsertedID.Hex())

	// 查询数据
	fmt.Println("\n5. 查询数据...")
	var foundDoc bson.M
	err = client.FindOne(ctx, collection, bson.M{"name": "测试文档"}, &foundDoc)
	if err != nil {
		log.Fatalf("❌ 查询失败: %v", err)
	}
	fmt.Printf("✅ 查询成功: %+v\n", foundDoc)

	// 更新数据
	fmt.Println("\n6. 更新数据...")
	updateResult, err := client.UpdateOne(ctx, collection, 
		bson.M{"name": "测试文档"}, 
		bson.M{"$set": bson.M{"number": 100, "updated": true}})
	if err != nil {
		log.Fatalf("❌ 更新失败: %v", err)
	}
	fmt.Printf("✅ 更新成功，修改了 %d 个文档\n", updateResult.ModifiedCount)

	// 统计文档数量
	fmt.Println("\n7. 统计文档数量...")
	count, err := client.CountDocuments(ctx, collection, bson.M{})
	if err != nil {
		log.Fatalf("❌ 统计失败: %v", err)
	}
	fmt.Printf("✅ 总共有 %d 个文档\n", count)

	// 查询所有文档
	fmt.Println("\n8. 查询所有文档...")
	var allDocs []bson.M
	err = client.Find(ctx, collection, bson.M{}, &allDocs)
	if err != nil {
		log.Fatalf("❌ 查询所有文档失败: %v", err)
	}
	fmt.Printf("✅ 找到 %d 个文档:\n", len(allDocs))
	for i, doc := range allDocs {
		fmt.Printf("  %d. %+v\n", i+1, doc)
	}

	// 测试聚合查询
	fmt.Println("\n9. 测试聚合查询...")
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$name",
				"count": bson.M{"$sum": 1},
				"max_number": bson.M{"$max": "$number"},
			},
		},
	}

	var aggregationResults []bson.M
	err = client.Aggregate(ctx, collection, pipeline, &aggregationResults)
	if err != nil {
		log.Printf("⚠️  聚合查询失败: %v", err)
	} else {
		fmt.Println("✅ 聚合查询结果:")
		for _, result := range aggregationResults {
			fmt.Printf("  名称: %v, 数量: %v, 最大数字: %v\n", 
				result["_id"], result["count"], result["max_number"])
		}
	}

	// 删除测试数据
	fmt.Println("\n10. 删除测试数据...")
	deleteResult, err := client.DeleteMany(ctx, collection, bson.M{})
	if err != nil {
		log.Printf("⚠️  删除失败: %v", err)
	} else {
		fmt.Printf("✅ 删除了 %d 个文档\n", deleteResult.DeletedCount)
	}

	fmt.Println("\n🎉 所有测试完成！MongoDB模块工作正常！")
}
