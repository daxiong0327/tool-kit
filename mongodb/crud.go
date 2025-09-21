package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertResult 插入结果
type InsertResult struct {
	InsertedID primitive.ObjectID `json:"inserted_id" bson:"inserted_id"`
}

// UpdateResult 更新结果
type UpdateResult struct {
	MatchedCount  int64       `json:"matched_count" bson:"matched_count"`
	ModifiedCount int64       `json:"modified_count" bson:"modified_count"`
	UpsertedCount int64       `json:"upserted_count" bson:"upserted_count"`
	UpsertedID    interface{} `json:"upserted_id" bson:"upserted_id"`
}

// DeleteResult 删除结果
type DeleteResult struct {
	DeletedCount int64 `json:"deleted_count" bson:"deleted_count"`
}

// FindOptions 查询选项
type FindOptions struct {
	Limit           *int64             `json:"limit" bson:"limit"`
	Skip            *int64             `json:"skip" bson:"skip"`
	Sort            bson.M             `json:"sort" bson:"sort"`
	Projection      bson.M             `json:"projection" bson:"projection"`
	Collation       *options.Collation `json:"collation" bson:"collation"`
	Hint            interface{}        `json:"hint" bson:"hint"`
	Max             bson.M             `json:"max" bson:"max"`
	Min             bson.M             `json:"min" bson:"min"`
	MaxTime         *int64             `json:"max_time" bson:"max_time"`
	NoCursorTimeout *bool              `json:"no_cursor_timeout" bson:"no_cursor_timeout"`
	OplogReplay     *bool              `json:"oplog_replay" bson:"oplog_replay"`
	ReturnKey       *bool              `json:"return_key" bson:"return_key"`
	ShowRecordID    *bool              `json:"show_record_id" bson:"show_record_id"`
	Snapshot        *bool              `json:"snapshot" bson:"snapshot"`
	BatchSize       *int32             `json:"batch_size" bson:"batch_size"`
}

// InsertOne 插入单个文档
func (c *Client) InsertOne(ctx context.Context, collection string, document interface{}) (*InsertResult, error) {
	coll := c.database.Collection(collection)

	result, err := coll.InsertOne(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid inserted ID type")
	}

	return &InsertResult{
		InsertedID: insertedID,
	}, nil
}

// InsertMany 插入多个文档
func (c *Client) InsertMany(ctx context.Context, collection string, documents []interface{}) ([]primitive.ObjectID, error) {
	coll := c.database.Collection(collection)

	result, err := coll.InsertMany(ctx, documents)
	if err != nil {
		return nil, fmt.Errorf("failed to insert documents: %w", err)
	}

	var insertedIDs []primitive.ObjectID
	for _, id := range result.InsertedIDs {
		if objectID, ok := id.(primitive.ObjectID); ok {
			insertedIDs = append(insertedIDs, objectID)
		}
	}

	return insertedIDs, nil
}

// FindOne 查找单个文档
func (c *Client) FindOne(ctx context.Context, collection string, filter bson.M, result interface{}) error {
	coll := c.database.Collection(collection)

	err := coll.FindOne(ctx, filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("no documents found")
		}
		return fmt.Errorf("failed to find document: %w", err)
	}

	return nil
}

// Find 查找多个文档
func (c *Client) Find(ctx context.Context, collection string, filter bson.M, results interface{}, opts ...*FindOptions) error {
	coll := c.database.Collection(collection)

	// 构建查询选项
	findOptions := options.Find()
	if len(opts) > 0 && opts[0] != nil {
		opt := opts[0]
		if opt.Limit != nil {
			findOptions.SetLimit(*opt.Limit)
		}
		if opt.Skip != nil {
			findOptions.SetSkip(*opt.Skip)
		}
		if opt.Sort != nil {
			findOptions.SetSort(opt.Sort)
		}
		if opt.Projection != nil {
			findOptions.SetProjection(opt.Projection)
		}
		if opt.Collation != nil {
			findOptions.SetCollation(opt.Collation)
		}
		if opt.Hint != nil {
			findOptions.SetHint(opt.Hint)
		}
		if opt.Max != nil {
			findOptions.SetMax(opt.Max)
		}
		if opt.Min != nil {
			findOptions.SetMin(opt.Min)
		}
		if opt.MaxTime != nil {
			findOptions.SetMaxTime(time.Duration(*opt.MaxTime) * time.Millisecond)
		}
		if opt.NoCursorTimeout != nil {
			findOptions.SetNoCursorTimeout(*opt.NoCursorTimeout)
		}
		if opt.OplogReplay != nil {
			findOptions.SetOplogReplay(*opt.OplogReplay)
		}
		if opt.ReturnKey != nil {
			findOptions.SetReturnKey(*opt.ReturnKey)
		}
		if opt.ShowRecordID != nil {
			findOptions.SetShowRecordID(*opt.ShowRecordID)
		}
		if opt.Snapshot != nil {
			findOptions.SetSnapshot(*opt.Snapshot)
		}
		if opt.BatchSize != nil {
			findOptions.SetBatchSize(*opt.BatchSize)
		}
	}

	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, results)
	if err != nil {
		return fmt.Errorf("failed to decode documents: %w", err)
	}

	return nil
}

// UpdateOne 更新单个文档
func (c *Client) UpdateOne(ctx context.Context, collection string, filter bson.M, update bson.M, upsert ...bool) (*UpdateResult, error) {
	coll := c.database.Collection(collection)

	updateOptions := options.Update()
	if len(upsert) > 0 {
		updateOptions.SetUpsert(upsert[0])
	}

	result, err := coll.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return &UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedID:    result.UpsertedID,
	}, nil
}

// UpdateMany 更新多个文档
func (c *Client) UpdateMany(ctx context.Context, collection string, filter bson.M, update bson.M, upsert ...bool) (*UpdateResult, error) {
	coll := c.database.Collection(collection)

	updateOptions := options.Update()
	if len(upsert) > 0 {
		updateOptions.SetUpsert(upsert[0])
	}

	result, err := coll.UpdateMany(ctx, filter, update, updateOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to update documents: %w", err)
	}

	return &UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedID:    result.UpsertedID,
	}, nil
}

// ReplaceOne 替换单个文档
func (c *Client) ReplaceOne(ctx context.Context, collection string, filter bson.M, replacement interface{}, upsert ...bool) (*UpdateResult, error) {
	coll := c.database.Collection(collection)

	replaceOptions := options.Replace()
	if len(upsert) > 0 {
		replaceOptions.SetUpsert(upsert[0])
	}

	result, err := coll.ReplaceOne(ctx, filter, replacement, replaceOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to replace document: %w", err)
	}

	return &UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedID:    result.UpsertedID,
	}, nil
}

// DeleteOne 删除单个文档
func (c *Client) DeleteOne(ctx context.Context, collection string, filter bson.M) (*DeleteResult, error) {
	coll := c.database.Collection(collection)

	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to delete document: %w", err)
	}

	return &DeleteResult{
		DeletedCount: result.DeletedCount,
	}, nil
}

// DeleteMany 删除多个文档
func (c *Client) DeleteMany(ctx context.Context, collection string, filter bson.M) (*DeleteResult, error) {
	coll := c.database.Collection(collection)

	result, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to delete documents: %w", err)
	}

	return &DeleteResult{
		DeletedCount: result.DeletedCount,
	}, nil
}

// CountDocuments 统计文档数量
func (c *Client) CountDocuments(ctx context.Context, collection string, filter bson.M) (int64, error) {
	coll := c.database.Collection(collection)

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

// EstimatedDocumentCount 估算文档数量
func (c *Client) EstimatedDocumentCount(ctx context.Context, collection string) (int64, error) {
	coll := c.database.Collection(collection)

	count, err := coll.EstimatedDocumentCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate document count: %w", err)
	}

	return count, nil
}

// Aggregate 聚合查询
func (c *Client) Aggregate(ctx context.Context, collection string, pipeline []bson.M, results interface{}) error {
	coll := c.database.Collection(collection)

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("failed to aggregate: %w", err)
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, results)
	if err != nil {
		return fmt.Errorf("failed to decode aggregation results: %w", err)
	}

	return nil
}

// CreateIndex 创建索引
func (c *Client) CreateIndex(ctx context.Context, collection string, model mongo.IndexModel) (string, error) {
	coll := c.database.Collection(collection)

	indexName, err := coll.Indexes().CreateOne(ctx, model)
	if err != nil {
		return "", fmt.Errorf("failed to create index: %w", err)
	}

	return indexName, nil
}

// CreateIndexes 创建多个索引
func (c *Client) CreateIndexes(ctx context.Context, collection string, models []mongo.IndexModel) ([]string, error) {
	coll := c.database.Collection(collection)

	indexNames, err := coll.Indexes().CreateMany(ctx, models)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return indexNames, nil
}

// ListIndexes 列出索引
func (c *Client) ListIndexes(ctx context.Context, collection string) ([]bson.M, error) {
	coll := c.database.Collection(collection)

	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	err = cursor.All(ctx, &indexes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode indexes: %w", err)
	}

	return indexes, nil
}

// DropIndex 删除索引
func (c *Client) DropIndex(ctx context.Context, collection string, name string) error {
	coll := c.database.Collection(collection)

	_, err := coll.Indexes().DropOne(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to drop index: %w", err)
	}

	return nil
}

// DropIndexes 删除所有索引
func (c *Client) DropIndexes(ctx context.Context, collection string) error {
	coll := c.database.Collection(collection)

	_, err := coll.Indexes().DropAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to drop indexes: %w", err)
	}

	return nil
}
