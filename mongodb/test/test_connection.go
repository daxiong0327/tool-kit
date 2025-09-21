package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/daxiong0327/tool-kit/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// User 用户结构体
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Age       int                `bson:"age" json:"age"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

func main() {
	fmt.Println("=== MongoDB 连接测试 ===")
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

	// 测试基本CRUD操作
	collection := "test_users"
	fmt.Printf("\n2. 测试CRUD操作 (集合: %s)...\n", collection)

	// 清理测试数据
	client.DeleteMany(ctx, collection, bson.M{})
	fmt.Println("🧹 清理测试数据")

	// 插入测试数据
	fmt.Println("\n3. 插入测试数据...")
	user := User{
		Name:      "测试用户",
		Email:     "test@example.com",
		Age:       25,
		CreatedAt: time.Now(),
	}

	result, err := client.InsertOne(ctx, collection, user)
	if err != nil {
		log.Fatalf("❌ 插入失败: %v", err)
	}
	fmt.Printf("✅ 插入成功，ID: %s\n", result.InsertedID.Hex())

	// 查询数据
	fmt.Println("\n4. 查询数据...")
	var foundUser User
	err = client.FindOne(ctx, collection, bson.M{"name": "测试用户"}, &foundUser)
	if err != nil {
		log.Fatalf("❌ 查询失败: %v", err)
	}
	fmt.Printf("✅ 查询成功: %+v\n", foundUser)

	// 更新数据
	fmt.Println("\n5. 更新数据...")
	updateResult, err := client.UpdateOne(ctx, collection,
		bson.M{"name": "测试用户"},
		bson.M{"$set": bson.M{"age": 26}})
	if err != nil {
		log.Fatalf("❌ 更新失败: %v", err)
	}
	fmt.Printf("✅ 更新成功，修改了 %d 个文档\n", updateResult.ModifiedCount)

	// 统计文档数量
	fmt.Println("\n6. 统计文档数量...")
	count, err := client.CountDocuments(ctx, collection, bson.M{})
	if err != nil {
		log.Fatalf("❌ 统计失败: %v", err)
	}
	fmt.Printf("✅ 总共有 %d 个文档\n", count)

	// 测试聚合查询
	fmt.Println("\n7. 测试聚合查询...")
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$age",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	var aggregationResults []bson.M
	err = client.Aggregate(ctx, collection, pipeline, &aggregationResults)
	if err != nil {
		log.Fatalf("❌ 聚合查询失败: %v", err)
	}
	fmt.Println("✅ 聚合查询结果:")
	for _, result := range aggregationResults {
		fmt.Printf("  年龄 %v: %v 人\n", result["_id"], result["count"])
	}

	// 测试索引操作
	fmt.Println("\n8. 测试索引操作...")
	indexModel := bson.M{
		"key":    bson.M{"email": 1},
		"unique": true,
	}

	indexName, err := client.CreateIndex(ctx, collection, indexModel)
	if err != nil {
		log.Printf("⚠️  创建索引失败: %v", err)
	} else {
		fmt.Printf("✅ 创建索引成功: %s\n", indexName)
	}

	// 列出索引
	indexes, err := client.ListIndexes(ctx, collection)
	if err != nil {
		log.Printf("⚠️  列出索引失败: %v", err)
	} else {
		fmt.Printf("✅ 集合中有 %d 个索引:\n", len(indexes))
		for _, index := range indexes {
			fmt.Printf("  - %s\n", index["name"])
		}
	}

	// 测试事务
	fmt.Println("\n9. 测试事务操作...")
	err = client.WithTransaction(ctx, func(sc mongo.SessionContext) error {
		// 在事务中插入另一个用户
		user2 := User{
			Name:      "事务用户",
			Email:     "transaction@example.com",
			Age:       30,
			CreatedAt: time.Now(),
		}
		_, err := client.InsertOne(sc, collection, user2)
		if err != nil {
			return fmt.Errorf("事务中插入失败: %w", err)
		}

		// 更新第一个用户
		_, err = client.UpdateOne(sc, collection,
			bson.M{"name": "测试用户"},
			bson.M{"$set": bson.M{"updated_in_transaction": true}})
		if err != nil {
			return fmt.Errorf("事务中更新失败: %w", err)
		}

		fmt.Println("  ✅ 事务操作完成")
		return nil
	})
	if err != nil {
		log.Printf("⚠️  事务操作失败: %v", err)
	} else {
		fmt.Println("✅ 事务操作成功")
	}

	// 最终统计
	finalCount, err := client.CountDocuments(ctx, collection, bson.M{})
	if err != nil {
		log.Fatalf("❌ 最终统计失败: %v", err)
	}
	fmt.Printf("\n🎉 测试完成！最终有 %d 个文档\n", finalCount)

	// 显示所有用户
	fmt.Println("\n📋 所有用户:")
	var allUsers []User
	err = client.Find(ctx, collection, bson.M{}, &allUsers)
	if err != nil {
		log.Printf("⚠️  查询所有用户失败: %v", err)
	} else {
		for i, user := range allUsers {
			fmt.Printf("  %d. %s (%s) - 年龄: %d\n", i+1, user.Name, user.Email, user.Age)
		}
	}
}
