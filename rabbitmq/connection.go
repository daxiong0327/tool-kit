package rabbitmq

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ConnectionManager 管理RabbitMQ连接
type ConnectionManager struct {
	config Config
	conn   *amqp.Connection
	mutex  sync.Mutex

	closed        bool
	notifyCloseCh chan *amqp.Error

	// 连接状态回调
	onConnected    func()
	onDisconnected func(err error)
}

// NewConnectionManager 创建一个新的连接管理器
func NewConnectionManager(config Config) (*ConnectionManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	cm := &ConnectionManager{
		config: config,
		closed: false,
	}

	// 建立初始连接
	if err := cm.connect(); err != nil {
		return nil, fmt.Errorf("initial connection failed: %w", err)
	}

	// 启动重连监控
	go cm.reconnectMonitor()

	return cm, nil
}

// SetCallbacks 设置连接状态回调函数
func (cm *ConnectionManager) SetCallbacks(onConnected func(), onDisconnected func(err error)) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.onConnected = onConnected
	cm.onDisconnected = onDisconnected
}

// connect 建立RabbitMQ连接
func (cm *ConnectionManager) connect() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.closed {
		return fmt.Errorf("connection manager is closed")
	}

	// 配置连接参数
	url := cm.config.BuildURL()
	config := amqp.Config{
		Heartbeat: cm.config.Heartbeat,
	}

	// 配置TLS
	if cm.config.EnableTLS && cm.config.TLSConfig != nil {
		tlsConfig, err := cm.createTLSConfig()
		if err != nil {
			return fmt.Errorf("TLS configuration error: %w", err)
		}
		config.TLSClientConfig = tlsConfig
	}

	// 建立连接
	conn, err := amqp.DialConfig(url, config)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// 设置关闭通知通道
	cm.notifyCloseCh = make(chan *amqp.Error, 1)
	cm.conn = conn
	cm.conn.NotifyClose(cm.notifyCloseCh)

	// 触发连接回调
	if cm.onConnected != nil {
		cm.onConnected()
	}

	return nil
}

// reconnectMonitor 监控连接并在断开时重连
func (cm *ConnectionManager) reconnectMonitor() {
	for {
		// 等待连接关闭通知
		err, ok := <-cm.notifyCloseCh
		if !ok {
			// 通道已关闭，退出监控
			return
		}

		// 触发断开连接回调
		if cm.onDisconnected != nil {
			cm.onDisconnected(err)
		}

		// 检查是否已手动关闭
		cm.mutex.Lock()
		if cm.closed {
			cm.mutex.Unlock()
			return
		}
		cm.mutex.Unlock()

		// 尝试重连
		log.Printf("RabbitMQ connection closed: %v, trying to reconnect...", err)

		attempts := 0
		for {
			attempts++
			if attempts > cm.config.MaxReconnectAttempts {
				log.Printf("reached maximum reconnect attempts (%d), giving up", cm.config.MaxReconnectAttempts)
				return
			}

			time.Sleep(cm.config.ReconnectDelay)
			log.Printf("trying to reconnect #%d...", attempts)

			err := cm.connect()
			if err == nil {
				log.Printf("reconnect successful")
				break
			}

			log.Printf("reconnect failed: %v", err)
		}
	}
}

// GetConnection 获取当前的RabbitMQ连接
func (cm *ConnectionManager) GetConnection() (*amqp.Connection, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.closed {
		return nil, fmt.Errorf("connection manager is closed")
	}

	if cm.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	return cm.conn, nil
}

// CreateChannel 创建一个新的通道
func (cm *ConnectionManager) CreateChannel() (*amqp.Channel, error) {
	conn, err := cm.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return ch, nil
}

// Close 关闭连接管理器
func (cm *ConnectionManager) Close() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.closed {
		return nil
	}

	cm.closed = true

	if cm.conn != nil {
		err := cm.conn.Close()
		if err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	return nil
}

// createTLSConfig 创建TLS配置
func (cm *ConnectionManager) createTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !cm.config.TLSConfig.Verify,
	}

	// 加载客户端证书
	if cm.config.TLSConfig.CertFile != "" && cm.config.TLSConfig.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cm.config.TLSConfig.CertFile, cm.config.TLSConfig.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// 加载CA证书
	if cm.config.TLSConfig.CACertFile != "" {
		caCert, err := ioutil.ReadFile(cm.config.TLSConfig.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}
