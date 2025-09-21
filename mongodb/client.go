package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client MongoDB客户端封装
type Client struct {
	client   *mongo.Client
	database *mongo.Database
	config   *Config
}

// Config MongoDB配置
type Config struct {
	URI            string        `json:"uri" yaml:"uri"`                         // MongoDB连接URI
	Database       string        `json:"database" yaml:"database"`               // 数据库名称
	ConnectTimeout time.Duration `json:"connect_timeout" yaml:"connect_timeout"` // 连接超时时间
	SocketTimeout  time.Duration `json:"socket_timeout" yaml:"socket_timeout"`   // Socket超时时间
	ServerTimeout  time.Duration `json:"server_timeout" yaml:"server_timeout"`   // 服务器选择超时时间
	MaxPoolSize    uint64        `json:"max_pool_size" yaml:"max_pool_size"`     // 最大连接池大小
	MinPoolSize    uint64        `json:"min_pool_size" yaml:"min_pool_size"`     // 最小连接池大小
	MaxIdleTime    time.Duration `json:"max_idle_time" yaml:"max_idle_time"`     // 最大空闲时间
	RetryWrites    bool          `json:"retry_writes" yaml:"retry_writes"`       // 是否启用重试写入
	RetryReads     bool          `json:"retry_reads" yaml:"retry_reads"`         // 是否启用重试读取
	Debug          bool          `json:"debug" yaml:"debug"`                     // 是否开启调试模式
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		URI:            "mongodb://localhost:27017",
		Database:       "test",
		ConnectTimeout: 10 * time.Second,
		SocketTimeout:  5 * time.Second,
		ServerTimeout:  5 * time.Second,
		MaxPoolSize:    100,
		MinPoolSize:    0,
		MaxIdleTime:    30 * time.Minute,
		RetryWrites:    true,
		RetryReads:     true,
		Debug:          false,
	}
}

// New 创建MongoDB客户端
func New(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	// 构建客户端选项
	clientOptions := options.Client().ApplyURI(config.URI)
	clientOptions.SetConnectTimeout(config.ConnectTimeout)
	clientOptions.SetSocketTimeout(config.SocketTimeout)
	clientOptions.SetServerSelectionTimeout(config.ServerTimeout)
	clientOptions.SetMaxPoolSize(config.MaxPoolSize)
	clientOptions.SetMinPoolSize(config.MinPoolSize)
	clientOptions.SetMaxConnIdleTime(config.MaxIdleTime)
	clientOptions.SetRetryWrites(config.RetryWrites)
	clientOptions.SetRetryReads(config.RetryReads)

	// 创建客户端
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// 测试连接
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.Database)

	return &Client{
		client:   client,
		database: database,
		config:   config,
	}, nil
}

// NewWithURI 使用URI创建客户端
func NewWithURI(uri, database string) (*Client, error) {
	config := &Config{
		URI:      uri,
		Database: database,
	}
	return New(config)
}

// Close 关闭客户端连接
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, nil)
}

// Database 获取数据库
func (c *Client) Database(name string) *mongo.Database {
	return c.client.Database(name)
}

// Collection 获取集合
func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// GetConfig 获取配置
func (c *Client) GetConfig() *Config {
	return c.config
}

// SetConfig 设置配置
func (c *Client) SetConfig(config *Config) {
	c.config = config
}

// SetDatabase 设置数据库
func (c *Client) SetDatabase(name string) {
	c.database = c.client.Database(name)
}

// GetClient 获取底层MongoDB客户端
func (c *Client) GetClient() *mongo.Client {
	return c.client
}

// GetDatabase 获取当前数据库
func (c *Client) GetDatabase() *mongo.Database {
	return c.database
}
