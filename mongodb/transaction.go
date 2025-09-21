package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Session MongoDB会话
type Session struct {
	session mongo.Session
	client  *Client
}

// NewSession 创建新会话
func (c *Client) NewSession() (*Session, error) {
	session, err := c.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	return &Session{
		session: session,
		client:  c,
	}, nil
}

// WithTransaction 在事务中执行操作
func (s *Session) WithTransaction(ctx context.Context, fn func(mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := mongo.WithSession(ctx, s.session, func(sc mongo.SessionContext) error {
		var err error
		result, err = s.session.WithTransaction(sc, func(sc mongo.SessionContext) (interface{}, error) {
			return fn(sc)
		})
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	return result, nil
}

// StartTransaction 开始事务
func (s *Session) StartTransaction(opts ...*options.TransactionOptions) error {
	err := s.session.StartTransaction(opts...)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	return nil
}

// CommitTransaction 提交事务
func (s *Session) CommitTransaction(ctx context.Context) error {
	err := s.session.CommitTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// AbortTransaction 中止事务
func (s *Session) AbortTransaction(ctx context.Context) error {
	err := s.session.AbortTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to abort transaction: %w", err)
	}
	return nil
}

// EndSession 结束会话
func (s *Session) EndSession(ctx context.Context) {
	s.session.EndSession(ctx)
}

// GetSession 获取底层会话
func (s *Session) GetSession() mongo.Session {
	return s.session
}

// TransactionOptions 事务选项
type TransactionOptions struct {
	ReadConcern    interface{} `json:"read_concern" bson:"read_concern"`
	WriteConcern   interface{} `json:"write_concern" bson:"write_concern"`
	ReadPreference interface{} `json:"read_preference" bson:"read_preference"`
	MaxCommitTime  *int64      `json:"max_commit_time" bson:"max_commit_time"`
}

// NewTransactionOptions 创建事务选项
func NewTransactionOptions() *TransactionOptions {
	return &TransactionOptions{}
}

// SetReadConcern 设置读关注
func (to *TransactionOptions) SetReadConcern(rc interface{}) *TransactionOptions {
	to.ReadConcern = rc
	return to
}

// SetWriteConcern 设置写关注
func (to *TransactionOptions) SetWriteConcern(wc interface{}) *TransactionOptions {
	to.WriteConcern = wc
	return to
}

// SetReadPreference 设置读偏好
func (to *TransactionOptions) SetReadPreference(rp interface{}) *TransactionOptions {
	to.ReadPreference = rp
	return to
}

// SetMaxCommitTime 设置最大提交时间
func (to *TransactionOptions) SetMaxCommitTime(maxTime int64) *TransactionOptions {
	to.MaxCommitTime = &maxTime
	return to
}

// ToMongoOptions 转换为MongoDB选项
func (to *TransactionOptions) ToMongoOptions() *options.TransactionOptions {
	opts := options.Transaction()

	if to.MaxCommitTime != nil {
		maxTime := time.Duration(*to.MaxCommitTime) * time.Millisecond
		opts.SetMaxCommitTime(&maxTime)
	}

	return opts
}

// TransactionFunc 事务函数类型
type TransactionFunc func(mongo.SessionContext) error

// WithTransaction 在事务中执行操作（便捷方法）
func (c *Client) WithTransaction(ctx context.Context, fn TransactionFunc, opts ...*TransactionOptions) error {
	session, err := c.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	var transactionOpts *options.TransactionOptions
	if len(opts) > 0 && opts[0] != nil {
		transactionOpts = opts[0].ToMongoOptions()
	}

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		return session.session.WithTransaction(sc, func(sc mongo.SessionContext) (interface{}, error) {
			err := fn(sc)
			return nil, err
		}, transactionOpts)
	})
	return err
}
