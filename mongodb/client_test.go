package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultConfig()
		assert.Equal(t, "mongodb://localhost:27017", config.URI)
		assert.Equal(t, "test", config.Database)
		assert.Equal(t, 10*time.Second, config.ConnectTimeout)
		assert.Equal(t, 5*time.Second, config.SocketTimeout)
		assert.Equal(t, 5*time.Second, config.ServerTimeout)
		assert.Equal(t, uint64(100), config.MaxPoolSize)
		assert.Equal(t, uint64(0), config.MinPoolSize)
		assert.Equal(t, 30*time.Minute, config.MaxIdleTime)
		assert.True(t, config.RetryWrites)
		assert.True(t, config.RetryReads)
		assert.False(t, config.Debug)
	})

	t.Run("CustomConfig", func(t *testing.T) {
		config := &Config{
			URI:            "mongodb://localhost:27017/test",
			Database:       "custom_db",
			ConnectTimeout: 30 * time.Second,
			SocketTimeout:  10 * time.Second,
			ServerTimeout:  10 * time.Second,
			MaxPoolSize:    200,
			MinPoolSize:    10,
			MaxIdleTime:    60 * time.Minute,
			RetryWrites:    false,
			RetryReads:     false,
			Debug:          true,
		}

		assert.Equal(t, "mongodb://localhost:27017/test", config.URI)
		assert.Equal(t, "custom_db", config.Database)
		assert.Equal(t, 30*time.Second, config.ConnectTimeout)
		assert.Equal(t, 10*time.Second, config.SocketTimeout)
		assert.Equal(t, 10*time.Second, config.ServerTimeout)
		assert.Equal(t, uint64(200), config.MaxPoolSize)
		assert.Equal(t, uint64(10), config.MinPoolSize)
		assert.Equal(t, 60*time.Minute, config.MaxIdleTime)
		assert.False(t, config.RetryWrites)
		assert.False(t, config.RetryReads)
		assert.True(t, config.Debug)
	})
}

func TestClient(t *testing.T) {
	// 注意：这些测试需要MongoDB实例运行
	// 在实际环境中，应该使用测试数据库或mock
	t.Skip("需要MongoDB实例运行")

	t.Run("New with default config", func(t *testing.T) {
		client, err := New(nil)
		require.NoError(t, err)
		require.NotNil(t, client)
		defer client.Close(context.Background())

		// 测试连接
		err = client.Ping(context.Background())
		assert.NoError(t, err)
	})

	t.Run("New with custom config", func(t *testing.T) {
		config := &Config{
			URI:      "mongodb://localhost:27017",
			Database: "test_db",
		}

		client, err := New(config)
		require.NoError(t, err)
		require.NotNil(t, client)
		defer client.Close(context.Background())

		assert.Equal(t, "test_db", client.config.Database)
	})

	t.Run("NewWithURI", func(t *testing.T) {
		client, err := NewWithURI("mongodb://localhost:27017", "test_db")
		require.NoError(t, err)
		require.NotNil(t, client)
		defer client.Close(context.Background())

		assert.Equal(t, "mongodb://localhost:27017", client.config.URI)
		assert.Equal(t, "test_db", client.config.Database)
	})

	t.Run("SetDatabase", func(t *testing.T) {
		client, err := New(nil)
		require.NoError(t, err)
		defer client.Close(context.Background())

		client.SetDatabase("new_database")
		assert.Equal(t, "new_database", client.database.Name())
	})

	t.Run("GetConfig", func(t *testing.T) {
		config := &Config{
			URI:      "mongodb://localhost:27017",
			Database: "test_db",
		}

		client, err := New(config)
		require.NoError(t, err)
		defer client.Close(context.Background())

		retrievedConfig := client.GetConfig()
		assert.Equal(t, config, retrievedConfig)
	})

	t.Run("SetConfig", func(t *testing.T) {
		client, err := New(nil)
		require.NoError(t, err)
		defer client.Close(context.Background())

		newConfig := &Config{
			URI:      "mongodb://localhost:27017",
			Database: "new_test_db",
		}

		client.SetConfig(newConfig)
		assert.Equal(t, newConfig, client.GetConfig())
	})
}

func TestCRUD(t *testing.T) {
	// 注意：这些测试需要MongoDB实例运行
	t.Skip("需要MongoDB实例运行")

	client, err := New(nil)
	require.NoError(t, err)
	defer client.Close(context.Background())

	collection := "test_collection"
	ctx := context.Background()

	// 清理测试数据
	client.DeleteMany(ctx, collection, bson.M{})

	t.Run("InsertOne", func(t *testing.T) {
		document := bson.M{
			"name":  "test",
			"value": 123,
		}

		result, err := client.InsertOne(ctx, collection, document)
		require.NoError(t, err)
		assert.NotEqual(t, primitive.NilObjectID, result.InsertedID)
	})

	t.Run("InsertMany", func(t *testing.T) {
		documents := []interface{}{
			bson.M{"name": "test1", "value": 1},
			bson.M{"name": "test2", "value": 2},
			bson.M{"name": "test3", "value": 3},
		}

		insertedIDs, err := client.InsertMany(ctx, collection, documents)
		require.NoError(t, err)
		assert.Len(t, insertedIDs, 3)
	})

	t.Run("FindOne", func(t *testing.T) {
		var result bson.M
		err := client.FindOne(ctx, collection, bson.M{"name": "test"}, &result)
		require.NoError(t, err)
		assert.Equal(t, "test", result["name"])
		assert.Equal(t, int32(123), result["value"])
	})

	t.Run("Find", func(t *testing.T) {
		var results []bson.M
		err := client.Find(ctx, collection, bson.M{"name": bson.M{"$regex": "test"}}, &results)
		require.NoError(t, err)
		assert.Len(t, results, 4) // 1个InsertOne + 3个InsertMany
	})

	t.Run("UpdateOne", func(t *testing.T) {
		filter := bson.M{"name": "test"}
		update := bson.M{"$set": bson.M{"value": 456}}

		result, err := client.UpdateOne(ctx, collection, filter, update)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.MatchedCount)
		assert.Equal(t, int64(1), result.ModifiedCount)
	})

	t.Run("UpdateMany", func(t *testing.T) {
		filter := bson.M{"name": bson.M{"$regex": "test"}}
		update := bson.M{"$set": bson.M{"updated": true}}

		result, err := client.UpdateMany(ctx, collection, filter, update)
		require.NoError(t, err)
		assert.Equal(t, int64(4), result.MatchedCount)
		assert.Equal(t, int64(4), result.ModifiedCount)
	})

	t.Run("ReplaceOne", func(t *testing.T) {
		filter := bson.M{"name": "test"}
		replacement := bson.M{"name": "replaced", "value": 789}

		result, err := client.ReplaceOne(ctx, collection, filter, replacement)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.MatchedCount)
		assert.Equal(t, int64(1), result.ModifiedCount)
	})

	t.Run("CountDocuments", func(t *testing.T) {
		count, err := client.CountDocuments(ctx, collection, bson.M{})
		require.NoError(t, err)
		assert.Equal(t, int64(4), count) // 3个InsertMany + 1个replaced
	})

	t.Run("EstimatedDocumentCount", func(t *testing.T) {
		count, err := client.EstimatedDocumentCount(ctx, collection)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(0))
	})

	t.Run("DeleteOne", func(t *testing.T) {
		filter := bson.M{"name": "replaced"}

		result, err := client.DeleteOne(ctx, collection, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.DeletedCount)
	})

	t.Run("DeleteMany", func(t *testing.T) {
		filter := bson.M{"name": bson.M{"$regex": "test"}}

		result, err := client.DeleteMany(ctx, collection, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.DeletedCount)
	})
}

func TestFindOptions(t *testing.T) {
	t.Run("Create FindOptions", func(t *testing.T) {
		limit := int64(10)
		skip := int64(5)
		sort := bson.M{"name": 1}
		projection := bson.M{"name": 1, "value": 1}

		opts := &FindOptions{
			Limit:      &limit,
			Skip:       &skip,
			Sort:       sort,
			Projection: projection,
		}

		assert.Equal(t, int64(10), *opts.Limit)
		assert.Equal(t, int64(5), *opts.Skip)
		assert.Equal(t, bson.M{"name": 1}, opts.Sort)
		assert.Equal(t, bson.M{"name": 1, "value": 1}, opts.Projection)
	})
}

func TestTransactionOptions(t *testing.T) {
	t.Run("Create TransactionOptions", func(t *testing.T) {
		maxTime := int64(5000)
		opts := NewTransactionOptions().
			SetMaxCommitTime(maxTime)

		assert.Equal(t, int64(5000), *opts.MaxCommitTime)
	})

	t.Run("ToMongoOptions", func(t *testing.T) {
		maxTime := int64(5000)
		opts := NewTransactionOptions().
			SetMaxCommitTime(maxTime)

		mongoOpts := opts.ToMongoOptions()
		assert.NotNil(t, mongoOpts)
	})
}

func TestBulkWriteOptions(t *testing.T) {
	t.Run("Create BulkWriteOptions", func(t *testing.T) {
		ordered := true
		bypass := false

		opts := NewBulkWriteOptions().
			SetOrdered(ordered).
			SetBypassDocumentValidation(bypass)

		assert.True(t, *opts.Ordered)
		assert.False(t, *opts.BypassDocumentValidation)
	})

	t.Run("ToMongoOptions", func(t *testing.T) {
		ordered := true
		opts := NewBulkWriteOptions().
			SetOrdered(ordered)

		mongoOpts := opts.ToMongoOptions()
		assert.NotNil(t, mongoOpts)
	})
}

func TestBulkOperations(t *testing.T) {
	t.Run("BulkInsert", func(t *testing.T) {
		bi := &BulkInsert{
			Documents: []interface{}{
				bson.M{"name": "bulk1", "value": 1},
				bson.M{"name": "bulk2", "value": 2},
			},
		}

		assert.Len(t, bi.Documents, 2)
	})

	t.Run("BulkUpdate", func(t *testing.T) {
		bu := &BulkUpdate{
			Filter: bson.M{"name": "test"},
			Update: bson.M{"$set": bson.M{"value": 999}},
			Upsert: false,
		}

		assert.Equal(t, bson.M{"name": "test"}, bu.Filter)
		assert.Equal(t, bson.M{"$set": bson.M{"value": 999}}, bu.Update)
		assert.False(t, bu.Upsert)
	})

	t.Run("BulkReplace", func(t *testing.T) {
		br := &BulkReplace{
			Filter:      bson.M{"name": "test"},
			Replacement: bson.M{"name": "replaced", "value": 888},
			Upsert:      true,
		}

		assert.Equal(t, bson.M{"name": "test"}, br.Filter)
		assert.Equal(t, bson.M{"name": "replaced", "value": 888}, br.Replacement)
		assert.True(t, br.Upsert)
	})

	t.Run("BulkDelete", func(t *testing.T) {
		bd := &BulkDelete{
			Filter: bson.M{"name": "test"},
		}

		assert.Equal(t, bson.M{"name": "test"}, bd.Filter)
	})
}
