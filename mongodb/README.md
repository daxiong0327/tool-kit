# MongoDB 客户端封装

基于 `mongo-go-driver` 的 MongoDB 客户端封装，提供简洁易用的 API 和丰富的功能。

## 特性

- 🚀 **简洁易用**: 提供直观的 API 接口
- 🔧 **灵活配置**: 支持连接池、超时、重试等配置
- 🔒 **事务支持**: 完整的事务操作支持
- 📦 **批量操作**: 高效的批量写入和更新
- 🏗️ **索引管理**: 便捷的索引创建和管理
- 🔍 **聚合查询**: 支持复杂的聚合操作
- 📊 **连接池**: 内置连接池管理，提升性能
- 🛡️ **错误处理**: 完善的错误处理和类型安全

## 安装

```bash
go get github.com/daxiong0327/tool-kit/mongodb
```

## 快速开始

### 基本连接

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/daxiong0327/tool-kit/mongodb"
)

func main() {
    // 使用默认配置
    client, err := mongodb.New(nil)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close(context.Background())
    
    // 测试连接
    err = client.Ping(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    
    println("Connected to MongoDB!")
}
```

### 自定义配置

```go
package main

import (
    "github.com/daxiong0327/tool-kit/mongodb"
    "time"
)

func main() {
    config := &mongodb.Config{
        URI:            "mongodb://localhost:27017",
        Database:       "myapp",
        ConnectTimeout: 10 * time.Second,
        SocketTimeout:  5 * time.Second,
        MaxPoolSize:    100,
        MinPoolSize:    10,
        RetryWrites:    true,
        RetryReads:     true,
    }
    
    client, err := mongodb.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close(context.Background())
}
```

## 基本操作

### 插入文档

```go
// 插入单个文档
document := bson.M{
    "name":  "张三",
    "email": "zhangsan@example.com",
    "age":   25,
}

result, err := client.InsertOne(ctx, "users", document)
if err != nil {
    log.Fatal(err)
}
println("Inserted ID:", result.InsertedID)

// 插入多个文档
documents := []interface{}{
    bson.M{"name": "李四", "email": "lisi@example.com", "age": 30},
    bson.M{"name": "王五", "email": "wangwu@example.com", "age": 28},
}

insertedIDs, err := client.InsertMany(ctx, "users", documents)
if err != nil {
    log.Fatal(err)
}
println("Inserted", len(insertedIDs), "documents")
```

### 查询文档

```go
// 查找单个文档
var user bson.M
err := client.FindOne(ctx, "users", bson.M{"name": "张三"}, &user)
if err != nil {
    log.Fatal(err)
}
println("Found user:", user)

// 查找多个文档
var users []bson.M
err = client.Find(ctx, "users", bson.M{}, &users)
if err != nil {
    log.Fatal(err)
}
println("Found", len(users), "users")

// 条件查询
var youngUsers []bson.M
err = client.Find(ctx, "users", bson.M{"age": bson.M{"$gt": 25}}, &youngUsers)
if err != nil {
    log.Fatal(err)
}
```

### 分页查询

```go
// 分页查询
limit := int64(10)
skip := int64(0)
findOptions := &mongodb.FindOptions{
    Limit: &limit,
    Skip:  &skip,
    Sort:  bson.M{"age": 1}, // 按年龄升序
}

var users []bson.M
err := client.Find(ctx, "users", bson.M{}, &users, findOptions)
if err != nil {
    log.Fatal(err)
}
```

### 更新文档

```go
// 更新单个文档
filter := bson.M{"name": "张三"}
update := bson.M{"$set": bson.M{"age": 26}}

result, err := client.UpdateOne(ctx, "users", filter, update)
if err != nil {
    log.Fatal(err)
}
println("Updated", result.ModifiedCount, "documents")

// 更新多个文档
filter = bson.M{"age": bson.M{"$lt": 30}}
update = bson.M{"$inc": bson.M{"age": 1}}

result, err = client.UpdateMany(ctx, "users", filter, update)
if err != nil {
    log.Fatal(err)
}
println("Updated", result.ModifiedCount, "documents")

// 替换文档
filter = bson.M{"name": "张三"}
replacement := bson.M{
    "name":  "张三(已更新)",
    "email": "zhangsan.updated@example.com",
    "age":   26,
}

result, err = client.ReplaceOne(ctx, "users", filter, replacement)
if err != nil {
    log.Fatal(err)
}
```

### 删除文档

```go
// 删除单个文档
filter := bson.M{"name": "张三"}
result, err := client.DeleteOne(ctx, "users", filter)
if err != nil {
    log.Fatal(err)
}
println("Deleted", result.DeletedCount, "documents")

// 删除多个文档
filter = bson.M{"age": bson.M{"$lt": 25}}
result, err = client.DeleteMany(ctx, "users", filter)
if err != nil {
    log.Fatal(err)
}
println("Deleted", result.DeletedCount, "documents")
```

### 统计文档

```go
// 统计文档数量
count, err := client.CountDocuments(ctx, "users", bson.M{})
if err != nil {
    log.Fatal(err)
}
println("Total users:", count)

// 估算文档数量
estimatedCount, err := client.EstimatedDocumentCount(ctx, "users")
if err != nil {
    log.Fatal(err)
}
println("Estimated users:", estimatedCount)
```

## 聚合查询

```go
// 聚合查询示例
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

var results []bson.M
err := client.Aggregate(ctx, "users", pipeline, &results)
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    println("Age", result["_id"], ":", result["count"], "users")
}
```

## 索引管理

```go
// 创建单个索引
indexModel := bson.M{
    "key":    bson.M{"email": 1},
    "unique": true,
}

indexName, err := client.CreateIndex(ctx, "users", indexModel)
if err != nil {
    log.Fatal(err)
}
println("Created index:", indexName)

// 创建多个索引
indexModels := []bson.M{
    {"key": bson.M{"name": 1}},
    {"key": bson.M{"age": 1}},
    {"key": bson.M{"email": 1}, "unique": true},
}

indexNames, err := client.CreateIndexes(ctx, "users", indexModels)
if err != nil {
    log.Fatal(err)
}
println("Created indexes:", indexNames)

// 列出索引
indexes, err := client.ListIndexes(ctx, "users")
if err != nil {
    log.Fatal(err)
}
for _, index := range indexes {
    println("Index:", index["name"])
}

// 删除索引
err = client.DropIndex(ctx, "users", "email_1")
if err != nil {
    log.Fatal(err)
}

// 删除所有索引
err = client.DropIndexes(ctx, "users")
if err != nil {
    log.Fatal(err)
}
```

## 事务操作

### 基本事务

```go
// 使用事务执行多个操作
err := client.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
    // 在事务中执行操作
    _, err := client.InsertOne(sc, "users", bson.M{"name": "张三"})
    if err != nil {
        return nil, err
    }
    
    _, err = client.InsertOne(sc, "users", bson.M{"name": "李四"})
    if err != nil {
        return nil, err
    }
    
    return nil, nil
})
if err != nil {
    log.Fatal(err)
}
```

### 带选项的事务

```go
// 配置事务选项
transactionOpts := mongodb.NewTransactionOptions().
    SetMaxCommitTime(5000) // 5秒超时

// 设置读关注和写关注
readConcern := options.ReadConcern().SetLevel("majority")
writeConcern := options.WriteConcern().SetW("majority").SetJournal(true)

transactionOpts.SetReadConcern(readConcern)
transactionOpts.SetWriteConcern(writeConcern)

err := client.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
    // 事务操作
    return nil, nil
}, transactionOpts)
```

### 手动事务管理

```go
// 创建会话
session, err := client.NewSession()
if err != nil {
    log.Fatal(err)
}
defer session.EndSession(ctx)

// 开始事务
err = session.StartTransaction()
if err != nil {
    log.Fatal(err)
}

// 执行操作
_, err = client.InsertOne(ctx, "users", bson.M{"name": "张三"})
if err != nil {
    session.AbortTransaction(ctx)
    log.Fatal(err)
}

// 提交事务
err = session.CommitTransaction(ctx)
if err != nil {
    log.Fatal(err)
}
```

## 批量操作

```go
// 批量写入
operations := []mongodb.BulkOperation{
    &mongodb.BulkInsert{
        Documents: []interface{}{
            bson.M{"name": "张三", "age": 25},
            bson.M{"name": "李四", "age": 30},
        },
    },
    &mongodb.BulkUpdate{
        Filter: bson.M{"name": "张三"},
        Update: bson.M{"$set": bson.M{"age": 26}},
    },
    &mongodb.BulkDelete{
        Filter: bson.M{"age": bson.M{"$lt": 25}},
    },
}

result, err := client.BulkWrite(ctx, "users", operations)
if err != nil {
    log.Fatal(err)
}
println("Bulk write completed:", result.InsertedCount, "inserted")
```

## 配置选项

### 基本配置

| 选项 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| URI | string | "mongodb://localhost:27017" | MongoDB连接URI |
| Database | string | "test" | 数据库名称 |
| ConnectTimeout | time.Duration | 10s | 连接超时时间 |
| SocketTimeout | time.Duration | 5s | Socket超时时间 |
| ServerTimeout | time.Duration | 5s | 服务器选择超时时间 |
| MaxPoolSize | uint64 | 100 | 最大连接池大小 |
| MinPoolSize | uint64 | 0 | 最小连接池大小 |
| MaxIdleTime | time.Duration | 30m | 最大空闲时间 |
| RetryWrites | bool | true | 是否启用重试写入 |
| RetryReads | bool | true | 是否启用重试读取 |
| Debug | bool | false | 是否开启调试模式 |

### 查询选项 (FindOptions)

| 选项 | 类型 | 描述 |
|------|------|------|
| Limit | *int64 | 限制返回文档数量 |
| Skip | *int64 | 跳过文档数量 |
| Sort | bson.M | 排序规则 |
| Projection | bson.M | 字段投影 |
| Collation | *options.Collation | 排序规则 |
| Hint | interface{} | 索引提示 |
| Max | bson.M | 最大值 |
| Min | bson.M | 最小值 |
| MaxTime | *int64 | 最大执行时间 |
| NoCursorTimeout | *bool | 是否禁用游标超时 |
| OplogReplay | *bool | 是否重放oplog |
| ReturnKey | *bool | 是否返回键 |
| ShowRecordID | *bool | 是否显示记录ID |
| Snapshot | *bool | 是否快照 |
| BatchSize | *int32 | 批处理大小 |

### 事务选项 (TransactionOptions)

| 选项 | 类型 | 描述 |
|------|------|------|
| ReadConcern | *options.ReadConcern | 读关注级别 |
| WriteConcern | *options.WriteConcern | 写关注级别 |
| ReadPreference | *options.ReadPreference | 读偏好 |
| MaxCommitTime | *int64 | 最大提交时间(毫秒) |

### 批量写入选项 (BulkWriteOptions)

| 选项 | 类型 | 描述 |
|------|------|------|
| Ordered | *bool | 是否有序执行 |
| BypassDocumentValidation | *bool | 是否绕过文档验证 |
| WriteConcern | *options.WriteConcern | 写关注级别 |

## 错误处理

```go
result, err := client.InsertOne(ctx, "users", document)
if err != nil {
    switch {
    case err == mongo.ErrNoDocuments:
        println("No documents found")
    case err == mongo.ErrClientDisconnected:
        println("Client disconnected")
    default:
        log.Printf("MongoDB error: %v", err)
    }
}
```

## 性能优化

### 连接池配置

```go
config := &mongodb.Config{
    MaxPoolSize:    200,              // 增加最大连接数
    MinPoolSize:    20,               // 设置最小连接数
    MaxIdleTime:    60 * time.Minute, // 增加空闲时间
    ConnectTimeout: 30 * time.Second, // 增加连接超时
}
```

### 批量操作

```go
// 使用批量操作提高性能
operations := make([]mongodb.BulkOperation, 0, 1000)
for i := 0; i < 1000; i++ {
    operations = append(operations, &mongodb.BulkInsert{
        Documents: []interface{}{bson.M{"index": i}},
    })
}

result, err := client.BulkWrite(ctx, "collection", operations)
```

### 索引优化

```go
// 创建复合索引
indexModel := bson.M{
    "key": bson.M{
        "name": 1,
        "age":  1,
    },
    "name": "name_age_idx",
}

client.CreateIndex(ctx, "users", indexModel)
```

## 示例

查看 `examples/` 目录中的完整示例：

- `basic_example.go` - 基本CRUD操作示例
- `transaction_example.go` - 事务操作示例

## 许可证

MIT License
