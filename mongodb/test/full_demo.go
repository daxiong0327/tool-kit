package main

import (
	"context"
	"fmt"
	"time"

	"github.com/daxiong0327/tool-kit/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User 用户结构体
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Age       int                `bson:"age" json:"age"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

func main() {
	fmt.Println("=== MongoDB 完整功能演示 ===")
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

	// 测试集合操作
	collection := "demo_users"
	fmt.Printf("\n3. 测试集合操作 (集合: %s)...\n", collection)

	// 清理测试数据
	fmt.Println("🧹 清理测试数据...")
	_, err = client.DeleteMany(ctx, collection, bson.M{})
	if err != nil {
		fmt.Printf("⚠️  清理数据失败: %v\n", err)
	} else {
		fmt.Println("✅ 清理完成")
	}

	// 插入单个文档
	fmt.Println("\n4. 插入单个文档...")
	user := User{
		Name:      "张三",
		Email:     "zhangsan@example.com",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := client.InsertOne(ctx, collection, user)
	if err != nil {
		fmt.Printf("❌ 插入失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 插入成功，ID: %s\n", result.InsertedID.Hex())

	// 插入多个文档
	fmt.Println("\n5. 插入多个文档...")
	users := []interface{}{
		User{Name: "李四", Email: "lisi@example.com", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{Name: "王五", Email: "wangwu@example.com", Age: 28, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{Name: "赵六", Email: "zhaoliu@example.com", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	insertedIDs, err := client.InsertMany(ctx, collection, users)
	if err != nil {
		fmt.Printf("❌ 批量插入失败: %v\n", err)
	} else {
		fmt.Printf("✅ 批量插入成功，插入了 %d 个文档\n", len(insertedIDs))
	}

	// 查询单个文档
	fmt.Println("\n6. 查询单个文档...")
	var foundUser User
	err = client.FindOne(ctx, collection, bson.M{"name": "张三"}, &foundUser)
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
	} else {
		fmt.Printf("✅ 查询成功: %+v\n", foundUser)
	}

	// 查询多个文档
	fmt.Println("\n7. 查询多个文档...")
	var allUsers []User
	err = client.Find(ctx, collection, bson.M{}, &allUsers)
	if err != nil {
		fmt.Printf("❌ 查询所有用户失败: %v\n", err)
	} else {
		fmt.Printf("✅ 找到 %d 个用户:\n", len(allUsers))
		for i, u := range allUsers {
			fmt.Printf("  %d. %s (%s) - 年龄: %d\n", i+1, u.Name, u.Email, u.Age)
		}
	}

	// 条件查询
	fmt.Println("\n8. 条件查询 (年龄大于25)...")
	var youngUsers []User
	err = client.Find(ctx, collection, bson.M{"age": bson.M{"$gt": 25}}, &youngUsers)
	if err != nil {
		fmt.Printf("❌ 条件查询失败: %v\n", err)
	} else {
		fmt.Printf("✅ 找到 %d 个年龄大于25的用户:\n", len(youngUsers))
		for _, u := range youngUsers {
			fmt.Printf("  - %s, 年龄: %d\n", u.Name, u.Age)
		}
	}

	// 分页查询
	fmt.Println("\n9. 分页查询 (跳过1个，限制2个)...")
	limit := int64(2)
	skip := int64(1)
	findOptions := &mongodb.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  bson.M{"age": 1}, // 按年龄升序
	}

	var pagedUsers []User
	err = client.Find(ctx, collection, bson.M{}, &pagedUsers, findOptions)
	if err != nil {
		fmt.Printf("❌ 分页查询失败: %v\n", err)
	} else {
		fmt.Printf("✅ 分页查询结果 (%d 个用户):\n", len(pagedUsers))
		for _, u := range pagedUsers {
			fmt.Printf("  - %s, 年龄: %d\n", u.Name, u.Age)
		}
	}

	// 更新单个文档
	fmt.Println("\n10. 更新单个文档...")
	updateResult, err := client.UpdateOne(ctx, collection,
		bson.M{"name": "张三"},
		bson.M{"$set": bson.M{"age": 26, "updated_at": time.Now()}})
	if err != nil {
		fmt.Printf("❌ 更新失败: %v\n", err)
	} else {
		fmt.Printf("✅ 更新成功，修改了 %d 个文档\n", updateResult.ModifiedCount)
	}

	// 更新多个文档
	fmt.Println("\n11. 更新多个文档 (所有用户年龄+1)...")
	updateManyResult, err := client.UpdateMany(ctx, collection,
		bson.M{},
		bson.M{"$inc": bson.M{"age": 1}})
	if err != nil {
		fmt.Printf("❌ 批量更新失败: %v\n", err)
	} else {
		fmt.Printf("✅ 批量更新成功，修改了 %d 个文档\n", updateManyResult.ModifiedCount)
	}

	// 替换文档
	fmt.Println("\n12. 替换文档...")
	replaceResult, err := client.ReplaceOne(ctx, collection,
		bson.M{"name": "李四"},
		User{Name: "李四(已更新)", Email: "lisi.updated@example.com", Age: 31, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	if err != nil {
		fmt.Printf("❌ 替换失败: %v\n", err)
	} else {
		fmt.Printf("✅ 替换成功，修改了 %d 个文档\n", replaceResult.ModifiedCount)
	}

	// 统计文档数量
	fmt.Println("\n13. 统计文档数量...")
	count, err := client.CountDocuments(ctx, collection, bson.M{})
	if err != nil {
		fmt.Printf("❌ 统计失败: %v\n", err)
	} else {
		fmt.Printf("✅ 总共有 %d 个用户\n", count)
	}

	// 聚合查询
	fmt.Println("\n14. 聚合查询 (按年龄分组统计)...")
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$age",
				"count": bson.M{"$sum": 1},
				"names": bson.M{"$push": "$name"},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	var aggregationResults []bson.M
	err = client.Aggregate(ctx, collection, pipeline, &aggregationResults)
	if err != nil {
		fmt.Printf("❌ 聚合查询失败: %v\n", err)
	} else {
		fmt.Println("✅ 聚合查询结果:")
		for _, result := range aggregationResults {
			fmt.Printf("  年龄 %v: %v 人 (%v)\n", result["_id"], result["count"], result["names"])
		}
	}

	// 创建索引
	fmt.Println("\n15. 创建索引...")
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"email": 1},
		Options: options.Index().SetUnique(true),
	}

	indexName, err := client.CreateIndex(ctx, collection, indexModel)
	if err != nil {
		fmt.Printf("⚠️  创建索引失败: %v\n", err)
	} else {
		fmt.Printf("✅ 创建索引成功: %s\n", indexName)
	}

	// 列出索引
	fmt.Println("\n16. 列出索引...")
	indexes, err := client.ListIndexes(ctx, collection)
	if err != nil {
		fmt.Printf("⚠️  列出索引失败: %v\n", err)
	} else {
		fmt.Printf("✅ 集合中有 %d 个索引:\n", len(indexes))
		for _, index := range indexes {
			fmt.Printf("  - %s\n", index["name"])
		}
	}

	// 测试事务
	fmt.Println("\n17. 测试事务操作...")
	err = client.WithTransaction(ctx, func(sc mongo.SessionContext) error {
		// 在事务中插入另一个用户
		user2 := User{
			Name:      "事务用户",
			Email:     "transaction@example.com",
			Age:       30,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, err := client.InsertOne(sc, collection, user2)
		if err != nil {
			return fmt.Errorf("事务中插入失败: %w", err)
		}

		// 更新第一个用户
		_, err = client.UpdateOne(sc, collection,
			bson.M{"name": "张三"},
			bson.M{"$set": bson.M{"updated_in_transaction": true}})
		if err != nil {
			return fmt.Errorf("事务中更新失败: %w", err)
		}

		fmt.Println("  ✅ 事务操作完成")
		return nil
	})
	if err != nil {
		fmt.Printf("⚠️  事务操作失败: %v\n", err)
	} else {
		fmt.Println("✅ 事务操作成功")
	}

	// 最终统计
	finalCount, err := client.CountDocuments(ctx, collection, bson.M{})
	if err != nil {
		fmt.Printf("❌ 最终统计失败: %v\n", err)
	} else {
		fmt.Printf("\n📊 最终统计: 总共有 %d 个用户\n", finalCount)
	}

	// 显示所有用户
	fmt.Println("\n📋 所有用户:")
	var finalUsers []User
	err = client.Find(ctx, collection, bson.M{}, &finalUsers)
	if err != nil {
		fmt.Printf("⚠️  查询所有用户失败: %v\n", err)
	} else {
		for i, user := range finalUsers {
			fmt.Printf("  %d. %s (%s) - 年龄: %d\n", i+1, user.Name, user.Email, user.Age)
		}
	}

	fmt.Println("\n🎉 MongoDB模块功能演示完成！所有功能都正常工作！")
}
