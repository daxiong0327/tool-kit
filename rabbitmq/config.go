package rabbitmq

import (
	"fmt"
	"time"
)

// Config 包含RabbitMQ连接的配置参数
type Config struct {
	// URL 是RabbitMQ服务器的连接字符串，格式为：amqp://user:pass@host:port/vhost
	URL string

	// 可选：单独指定连接参数
	Host     string
	Port     int
	Username string
	Password string
	VHost    string

	// 连接池配置
	MaxConnections int

	// 重连配置
	ReconnectDelay       time.Duration
	MaxReconnectAttempts int

	// TLS配置
	EnableTLS bool
	TLSConfig *TLSConfig

	// 心跳间隔
	Heartbeat time.Duration
}

// TLSConfig 包含TLS连接的配置参数
type TLSConfig struct {
	CertFile   string
	KeyFile    string
	CACertFile string
	Verify     bool
}

// Validate 验证配置是否有效
func (c *Config) Validate() error {
	// 如果URL为空，则需要检查其他连接参数
	if c.URL == "" {
		if c.Host == "" {
			return fmt.Errorf("必须提供URL或Host参数")
		}
		if c.Port <= 0 {
			c.Port = 5672 // 默认端口
		}
	}

	// 设置默认值
	if c.ReconnectDelay <= 0 {
		c.ReconnectDelay = 5 * time.Second
	}

	if c.MaxReconnectAttempts <= 0 {
		c.MaxReconnectAttempts = 10
	}

	if c.MaxConnections <= 0 {
		c.MaxConnections = 1
	}

	if c.Heartbeat <= 0 {
		c.Heartbeat = 10 * time.Second
	}

	return nil
}

// BuildURL 根据单独的连接参数构建URL
func (c *Config) BuildURL() string {
	if c.URL != "" {
		return c.URL
	}

	username := "guest"
	if c.Username != "" {
		username = c.Username
	}

	password := "guest"
	if c.Password != "" {
		password = c.Password
	}

	vhost := "/"
	if c.VHost != "" {
		vhost = c.VHost
	}

	protocol := "amqp"
	if c.EnableTLS {
		protocol = "amqps"
	}

	return fmt.Sprintf("%s://%s:%s@%s:%d%s", protocol, username, password, c.Host, c.Port, vhost)
}
