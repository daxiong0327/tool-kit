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
	fmt.Println("=== MongoDB 基本操作示例 ===")

	// 创建MongoDB客户端
	config := &mongodb.Config{
		URI:            "mongodb://localhost:27017",
		Database:       "example_db",
		ConnectTimeout: 10 * time.Second,
		SocketTimeout:  5 * time.Second,
		MaxPoolSize:    100,
		MinPoolSize:    10,
		RetryWrites:    true,
		RetryReads:     true,
	}

	client, err := mongodb.New(config)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Close(context.Background())

	ctx := context.Background()
	collection := "users"

	// 测试连接
	err = client.Ping(ctx)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	fmt.Println("✅ 成功连接到MongoDB")

	// 清理测试数据
	client.DeleteMany(ctx, collection, bson.M{})
	fmt.Println("🧹 清理测试数据")

	// 示例1：插入单个文档
	fmt.Println("\n1. 插入单个文档:")
	user := User{
		Name:      "张三",
		Email:     "zhangsan@example.com",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	insertResult, err := client.InsertOne(ctx, collection, user)
	if err != nil {
		log.Fatalf("Failed to insert user: %v", err)
	}
	fmt.Printf("✅ 插入用户成功，ID: %s\n", insertResult.InsertedID.Hex())

	// 示例2：插入多个文档
	fmt.Println("\n2. 插入多个文档:")
	users := []interface{}{
		User{Name: "李四", Email: "lisi@example.com", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{Name: "王五", Email: "wangwu@example.com", Age: 28, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{Name: "赵六", Email: "zhaoliu@example.com", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	insertedIDs, err := client.InsertMany(ctx, collection, users)
	if err != nil {
		log.Fatalf("Failed to insert users: %v", err)
	}
	fmt.Printf("✅ 插入 %d 个用户成功\n", len(insertedIDs))

	// 示例3：查找单个文档
	fmt.Println("\n3. 查找单个文档:")
	var foundUser User
	err = client.FindOne(ctx, collection, bson.M{"name": "张三"}, &foundUser)
	if err != nil {
		log.Fatalf("Failed to find user: %v", err)
	}
	fmt.Printf("✅ 找到用户: %+v\n", foundUser)

	// 示例4：查找多个文档
	fmt.Println("\n4. 查找多个文档:")
	var allUsers []User
	err = client.Find(ctx, collection, bson.M{}, &allUsers)
	if err != nil {
		log.Fatalf("Failed to find users: %v", err)
	}
	fmt.Printf("✅ 找到 %d 个用户:\n", len(allUsers))
	for _, u := range allUsers {
		fmt.Printf("  - %s (%s), 年龄: %d\n", u.Name, u.Email, u.Age)
	}

	// 示例5：条件查询
	fmt.Println("\n5. 条件查询 (年龄大于25):")
	var youngUsers []User
	err = client.Find(ctx, collection, bson.M{"age": bson.M{"$gt": 25}}, &youngUsers)
	if err != nil {
		log.Fatalf("Failed to find young users: %v", err)
	}
	fmt.Printf("✅ 找到 %d 个年龄大于25的用户:\n", len(youngUsers))
	for _, u := range youngUsers {
		fmt.Printf("  - %s, 年龄: %d\n", u.Name, u.Age)
	}

	// 示例6：分页查询
	fmt.Println("\n6. 分页查询 (跳过1个，限制2个):")
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
		log.Fatalf("Failed to find paged users: %v", err)
	}
	fmt.Printf("✅ 分页查询结果 (%d 个用户):\n", len(pagedUsers))
	for _, u := range pagedUsers {
		fmt.Printf("  - %s, 年龄: %d\n", u.Name, u.Age)
	}

	// 示例7：更新单个文档
	fmt.Println("\n7. 更新单个文档:")
	filter := bson.M{"name": "张三"}
	update := bson.M{"$set": bson.M{"age": 26, "updated_at": time.Now()}}

	updateResult, err := client.UpdateOne(ctx, collection, filter, update)
	if err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}
	fmt.Printf("✅ 更新了 %d 个文档\n", updateResult.ModifiedCount)

	// 示例8：更新多个文档
	fmt.Println("\n8. 更新多个文档 (所有用户年龄+1):")
	updateManyFilter := bson.M{}
	updateManyUpdate := bson.M{"$inc": bson.M{"age": 1}}

	updateManyResult, err := client.UpdateMany(ctx, collection, updateManyFilter, updateManyUpdate)
	if err != nil {
		log.Fatalf("Failed to update users: %v", err)
	}
	fmt.Printf("✅ 更新了 %d 个文档\n", updateManyResult.ModifiedCount)

	// 示例9：替换文档
	fmt.Println("\n9. 替换文档:")
	replaceFilter := bson.M{"name": "李四"}
	replacement := User{
		Name:      "李四(已更新)",
		Email:     "lisi.updated@example.com",
		Age:       31,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	replaceResult, err := client.ReplaceOne(ctx, collection, replaceFilter, replacement)
	if err != nil {
		log.Fatalf("Failed to replace user: %v", err)
	}
	fmt.Printf("✅ 替换了 %d 个文档\n", replaceResult.ModifiedCount)

	// 示例10：统计文档数量
	fmt.Println("\n10. 统计文档数量:")
	count, err := client.CountDocuments(ctx, collection, bson.M{})
	if err != nil {
		log.Fatalf("Failed to count documents: %v", err)
	}
	fmt.Printf("✅ 总共有 %d 个用户\n", count)

	// 示例11：聚合查询
	fmt.Println("\n11. 聚合查询 (按年龄分组统计):")
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
		log.Fatalf("Failed to aggregate: %v", err)
	}
	fmt.Println("✅ 聚合查询结果:")
	for _, result := range aggregationResults {
		fmt.Printf("  年龄 %v: %v 人 (%v)\n", result["_id"], result["count"], result["names"])
	}

	// 示例12：创建索引
	fmt.Println("\n12. 创建索引:")
	indexModel := mongo.IndexModel{
		Keys: bson.M{"email": 1},
		Options: options.Index().SetUnique(true),
	}

	indexName, err := client.CreateIndex(ctx, collection, indexModel)
	if err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}
	fmt.Printf("✅ 创建索引成功: %s\n", indexName)

	// 示例13：列出索引
	fmt.Println("\n13. 列出索引:")
	indexes, err := client.ListIndexes(ctx, collection)
	if err != nil {
		log.Fatalf("Failed to list indexes: %v", err)
	}
	fmt.Printf("✅ 集合中有 %d 个索引:\n", len(indexes))
	for _, index := range indexes {
		fmt.Printf("  - %s\n", index["name"])
	}

	// 示例14：删除单个文档
	fmt.Println("\n14. 删除单个文档:")
	deleteFilter := bson.M{"name": "王五"}
	deleteResult, err := client.DeleteOne(ctx, collection, deleteFilter)
	if err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}
	fmt.Printf("✅ 删除了 %d 个文档\n", deleteResult.DeletedCount)

	// 示例15：删除多个文档
	fmt.Println("\n15. 删除多个文档 (年龄大于30的用户):")
	deleteManyFilter := bson.M{"age": bson.M{"$gt": 30}}
	deleteManyResult, err := client.DeleteMany(ctx, collection, deleteManyFilter)
	if err != nil {
		log.Fatalf("Failed to delete users: %v", err)
	}
	fmt.Printf("✅ 删除了 %d 个文档\n", deleteManyResult.DeletedCount)

	// 最终统计
	finalCount, err := client.CountDocuments(ctx, collection, bson.M{})
	if err != nil {
		log.Fatalf("Failed to count final documents: %v", err)
	}
	fmt.Printf("\n🎉 示例完成！最终剩余 %d 个用户\n", finalCount)
}
