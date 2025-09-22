package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client Redis客户端封装
type Client struct {
	client *redis.Client
	config *Config
}

// Config Redis配置
type Config struct {
	// 基本配置
	Addr     string `json:"addr" yaml:"addr"`         // Redis地址，格式：host:port
	Password string `json:"password" yaml:"password"` // 密码
	DB       int    `json:"db" yaml:"db"`             // 数据库编号

	// 连接池配置
	PoolSize        int           `json:"pool_size" yaml:"pool_size"`                   // 连接池大小
	MinIdleConns    int           `json:"min_idle_conns" yaml:"min_idle_conns"`         // 最小空闲连接数
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`         // 最大空闲连接数
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" yaml:"conn_max_idle_time"` // 连接最大空闲时间
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`   // 连接最大生存时间

	// 超时配置
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`   // 连接超时
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`   // 读取超时
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"` // 写入超时

	// 重试配置
	MaxRetries      int           `json:"max_retries" yaml:"max_retries"`             // 最大重试次数
	MinRetryBackoff time.Duration `json:"min_retry_backoff" yaml:"min_retry_backoff"` // 最小重试间隔
	MaxRetryBackoff time.Duration `json:"max_retry_backoff" yaml:"max_retry_backoff"` // 最大重试间隔

	// 其他配置
	Username     string `json:"username" yaml:"username"`           // 用户名
	Protocol     int    `json:"protocol" yaml:"protocol"`           // 协议版本 (2 或 3)
	DisableAuth  bool   `json:"disable_auth" yaml:"disable_auth"`   // 禁用认证
	DisableAuth2 bool   `json:"disable_auth2" yaml:"disable_auth2"` // 禁用认证2 (兼容性)
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Addr:            "localhost:6379",
		Password:        "",
		DB:              0,
		PoolSize:        10,
		MinIdleConns:    5,
		MaxIdleConns:    10,
		ConnMaxIdleTime: 30 * time.Minute,
		ConnMaxLifetime: 0, // 不限制连接生存时间
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
		Username:        "",
		Protocol:        3,
		DisableAuth:     false,
		DisableAuth2:    false,
	}
}

// New 创建新的Redis客户端
func New(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 创建Redis客户端选项
	opts := &redis.Options{
		Addr:            config.Addr,
		Password:        config.Password,
		DB:              config.DB,
		PoolSize:        config.PoolSize,
		MinIdleConns:    config.MinIdleConns,
		MaxIdleConns:    config.MaxIdleConns,
		ConnMaxIdleTime: config.ConnMaxIdleTime,
		ConnMaxLifetime: config.ConnMaxLifetime,
		DialTimeout:     config.DialTimeout,
		ReadTimeout:     config.ReadTimeout,
		WriteTimeout:    config.WriteTimeout,
		MaxRetries:      config.MaxRetries,
		MinRetryBackoff: config.MinRetryBackoff,
		MaxRetryBackoff: config.MaxRetryBackoff,
		Username:        config.Username,
		Protocol:        config.Protocol,
	}

	// 处理认证禁用选项（兼容性处理）
	if config.DisableAuth || config.DisableAuth2 {
		// go-redis v9 不支持 DisableAuth，通过设置空密码实现
		opts.Password = ""
	}

	// 创建Redis客户端
	client := redis.NewClient(opts)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), config.DialTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{
		client: client,
		config: config,
	}, nil
}

// NewFromURL 从URL创建Redis客户端
func NewFromURL(url string) (*Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{
		client: client,
		config: &Config{
			Addr:     opts.Addr,
			Password: opts.Password,
			DB:       opts.DB,
		},
	}, nil
}

// GetClient 获取原始Redis客户端
func (c *Client) GetClient() *redis.Client {
	return c.client
}

// GetConfig 获取配置
func (c *Client) GetConfig() *Config {
	return c.config
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.client.Close()
}

// SetConfig 更新配置
func (c *Client) SetConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 关闭旧连接
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("failed to close old connection: %w", err)
	}

	// 创建新客户端
	newClient, err := New(config)
	if err != nil {
		return fmt.Errorf("failed to create new client: %w", err)
	}

	// 更新客户端
	c.client = newClient.client
	c.config = newClient.config

	return nil
}

// GetStats 获取连接池统计信息
func (c *Client) GetStats() *redis.PoolStats {
	return c.client.PoolStats()
}

// FlushDB 清空当前数据库
func (c *Client) FlushDB(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// FlushAll 清空所有数据库
func (c *Client) FlushAll(ctx context.Context) error {
	return c.client.FlushAll(ctx).Err()
}

// Info 获取Redis信息
func (c *Client) Info(ctx context.Context, section ...string) (string, error) {
	return c.client.Info(ctx, section...).Result()
}

// Keys 获取匹配的键
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Del(ctx, keys...).Result()
}

// Expire 设置键的过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.client.Expire(ctx, key, expiration).Result()
}

// TTL 获取键的剩余生存时间
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Type 获取键的类型
func (c *Client) Type(ctx context.Context, key string) (string, error) {
	return c.client.Type(ctx, key).Result()
}
