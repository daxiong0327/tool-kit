package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BulkWriteResult 批量写入结果
type BulkWriteResult struct {
	InsertedCount int64                   `json:"inserted_count" bson:"inserted_count"`
	MatchedCount  int64                   `json:"matched_count" bson:"matched_count"`
	ModifiedCount int64                   `json:"modified_count" bson:"modified_count"`
	DeletedCount  int64                   `json:"deleted_count" bson:"deleted_count"`
	UpsertedCount int64                   `json:"upserted_count" bson:"upserted_count"`
	UpsertedIDs   map[int64]interface{}   `json:"upserted_ids" bson:"upserted_ids"`
}

// BulkOperation 批量操作接口
type BulkOperation interface {
	Execute(ctx context.Context, collection string) (*BulkWriteResult, error)
}

// BulkInsert 批量插入操作
type BulkInsert struct {
	Documents []interface{} `json:"documents" bson:"documents"`
}

// BulkUpdate 批量更新操作
type BulkUpdate struct {
	Filter bson.M `json:"filter" bson:"filter"`
	Update bson.M `json:"update" bson:"update"`
	Upsert bool   `json:"upsert" bson:"upsert"`
}

// BulkReplace 批量替换操作
type BulkReplace struct {
	Filter      bson.M       `json:"filter" bson:"filter"`
	Replacement interface{}  `json:"replacement" bson:"replacement"`
	Upsert      bool         `json:"upsert" bson:"upsert"`
}

// BulkDelete 批量删除操作
type BulkDelete struct {
	Filter bson.M `json:"filter" bson:"filter"`
}

// Execute 执行批量插入
func (bi *BulkInsert) Execute(ctx context.Context, collection string) (*BulkWriteResult, error) {
	client := &Client{} // 这里需要从上下文获取客户端
	coll := client.database.Collection(collection)
	
	operations := make([]mongo.WriteModel, len(bi.Documents))
	for i, doc := range bi.Documents {
		operations[i] = mongo.NewInsertOneModel().SetDocument(doc)
	}

	result, err := coll.BulkWrite(ctx, operations)
	if err != nil {
		return nil, fmt.Errorf("failed to execute bulk insert: %w", err)
	}

	return &BulkWriteResult{
		InsertedCount: result.InsertedCount,
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		DeletedCount:  result.DeletedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedIDs:   result.UpsertedIDs,
	}, nil
}

// Execute 执行批量更新
func (bu *BulkUpdate) Execute(ctx context.Context, collection string) (*BulkWriteResult, error) {
	client := &Client{} // 这里需要从上下文获取客户端
	coll := client.database.Collection(collection)
	
	operation := mongo.NewUpdateOneModel().
		SetFilter(bu.Filter).
		SetUpdate(bu.Update).
		SetUpsert(bu.Upsert)

	result, err := coll.BulkWrite(ctx, []mongo.WriteModel{operation})
	if err != nil {
		return nil, fmt.Errorf("failed to execute bulk update: %w", err)
	}

	return &BulkWriteResult{
		InsertedCount: result.InsertedCount,
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		DeletedCount:  result.DeletedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedIDs:   result.UpsertedIDs,
	}, nil
}

// Execute 执行批量替换
func (br *BulkReplace) Execute(ctx context.Context, collection string) (*BulkWriteResult, error) {
	client := &Client{} // 这里需要从上下文获取客户端
	coll := client.database.Collection(collection)
	
	operation := mongo.NewReplaceOneModel().
		SetFilter(br.Filter).
		SetReplacement(br.Replacement).
		SetUpsert(br.Upsert)

	result, err := coll.BulkWrite(ctx, []mongo.WriteModel{operation})
	if err != nil {
		return nil, fmt.Errorf("failed to execute bulk replace: %w", err)
	}

	return &BulkWriteResult{
		InsertedCount: result.InsertedCount,
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		DeletedCount:  result.DeletedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedIDs:   result.UpsertedIDs,
	}, nil
}

// Execute 执行批量删除
func (bd *BulkDelete) Execute(ctx context.Context, collection string) (*BulkWriteResult, error) {
	client := &Client{} // 这里需要从上下文获取客户端
	coll := client.database.Collection(collection)
	
	operation := mongo.NewDeleteOneModel().SetFilter(bd.Filter)

	result, err := coll.BulkWrite(ctx, []mongo.WriteModel{operation})
	if err != nil {
		return nil, fmt.Errorf("failed to execute bulk delete: %w", err)
	}

	return &BulkWriteResult{
		InsertedCount: result.InsertedCount,
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		DeletedCount:  result.DeletedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedIDs:   result.UpsertedIDs,
	}, nil
}

// BulkWrite 批量写入操作
func (c *Client) BulkWrite(ctx context.Context, collection string, operations []BulkOperation) (*BulkWriteResult, error) {
	coll := c.database.Collection(collection)
	
	writeModels := make([]mongo.WriteModel, 0, len(operations))
	
	for _, op := range operations {
		switch v := op.(type) {
		case *BulkInsert:
			for _, doc := range v.Documents {
				writeModels = append(writeModels, mongo.NewInsertOneModel().SetDocument(doc))
			}
		case *BulkUpdate:
			writeModels = append(writeModels, mongo.NewUpdateOneModel().
				SetFilter(v.Filter).
				SetUpdate(v.Update).
				SetUpsert(v.Upsert))
		case *BulkReplace:
			writeModels = append(writeModels, mongo.NewReplaceOneModel().
				SetFilter(v.Filter).
				SetReplacement(v.Replacement).
				SetUpsert(v.Upsert))
		case *BulkDelete:
			writeModels = append(writeModels, mongo.NewDeleteOneModel().SetFilter(v.Filter))
		default:
			return nil, fmt.Errorf("unsupported bulk operation type: %T", op)
		}
	}

	result, err := coll.BulkWrite(ctx, writeModels)
	if err != nil {
		return nil, fmt.Errorf("failed to execute bulk write: %w", err)
	}

	return &BulkWriteResult{
		InsertedCount: result.InsertedCount,
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		DeletedCount:  result.DeletedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedIDs:   result.UpsertedIDs,
	}, nil
}

// BulkWriteOptions 批量写入选项
type BulkWriteOptions struct {
	Ordered                  *bool        `json:"ordered" bson:"ordered"`
	BypassDocumentValidation *bool        `json:"bypass_document_validation" bson:"bypass_document_validation"`
	WriteConcern             interface{}  `json:"write_concern" bson:"write_concern"`
}

// NewBulkWriteOptions 创建批量写入选项
func NewBulkWriteOptions() *BulkWriteOptions {
	return &BulkWriteOptions{}
}

// SetOrdered 设置是否有序执行
func (bwo *BulkWriteOptions) SetOrdered(ordered bool) *BulkWriteOptions {
	bwo.Ordered = &ordered
	return bwo
}

// SetBypassDocumentValidation 设置是否绕过文档验证
func (bwo *BulkWriteOptions) SetBypassDocumentValidation(bypass bool) *BulkWriteOptions {
	bwo.BypassDocumentValidation = &bypass
	return bwo
}

// SetWriteConcern 设置写关注
func (bwo *BulkWriteOptions) SetWriteConcern(wc interface{}) *BulkWriteOptions {
	bwo.WriteConcern = wc
	return bwo
}

// ToMongoOptions 转换为MongoDB选项
func (bwo *BulkWriteOptions) ToMongoOptions() *options.BulkWriteOptions {
	opts := options.BulkWrite()
	
	if bwo.Ordered != nil {
		opts.SetOrdered(*bwo.Ordered)
	}
	if bwo.BypassDocumentValidation != nil {
		opts.SetBypassDocumentValidation(*bwo.BypassDocumentValidation)
	}

	return opts
}
