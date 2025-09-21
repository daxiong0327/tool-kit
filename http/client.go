package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"time"
)

// RetryStrategy 重试策略
type RetryStrategy int

const (
	// RetryStrategyFixed 固定延迟重试
	RetryStrategyFixed RetryStrategy = iota
	// RetryStrategyExponential 指数退避重试
	RetryStrategyExponential
	// RetryStrategyLinear 线性增长重试
	RetryStrategyLinear
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries     int           `json:"max_retries" yaml:"max_retries"`         // 最大重试次数
	BaseDelay      time.Duration `json:"base_delay" yaml:"base_delay"`           // 基础延迟时间
	MaxDelay       time.Duration `json:"max_delay" yaml:"max_delay"`             // 最大延迟时间
	Strategy       RetryStrategy `json:"strategy" yaml:"strategy"`               // 重试策略
	RetryableCodes []int         `json:"retryable_codes" yaml:"retryable_codes"` // 可重试的状态码
}

// Client HTTP客户端
type Client struct {
	client *http.Client
	config *Config
}

// Config HTTP客户端配置
type Config struct {
	BaseURL   string            `json:"base_url" yaml:"base_url"`     // 基础URL
	Timeout   time.Duration     `json:"timeout" yaml:"timeout"`       // 请求超时时间
	Headers   map[string]string `json:"headers" yaml:"headers"`       // 默认请求头
	UserAgent string            `json:"user_agent" yaml:"user_agent"` // User-Agent
	Proxy     string            `json:"proxy" yaml:"proxy"`           // 代理地址
	Insecure  bool              `json:"insecure" yaml:"insecure"`     // 是否跳过SSL验证
	Debug     bool              `json:"debug" yaml:"debug"`           // 是否开启调试模式

	// 重试配置
	Retry *RetryConfig `json:"retry" yaml:"retry"` // 重试配置

	// 连接池配置
	Pool *PoolConfig `json:"pool" yaml:"pool"` // 连接池配置

	// 向后兼容的字段（已废弃）
	RetryCount int           `json:"retry_count" yaml:"retry_count"` // 重试次数 (已废弃，使用Retry)
	RetryDelay time.Duration `json:"retry_delay" yaml:"retry_delay"` // 重试延迟 (已废弃，使用Retry)
}

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxIdleConns        int           `json:"max_idle_conns" yaml:"max_idle_conns"`                   // 最大空闲连接数
	MaxIdleConnsPerHost int           `json:"max_idle_conns_per_host" yaml:"max_idle_conns_per_host"` // 每个主机的最大空闲连接数
	MaxConnsPerHost     int           `json:"max_conns_per_host" yaml:"max_conns_per_host"`           // 每个主机的最大连接数
	IdleConnTimeout     time.Duration `json:"idle_conn_timeout" yaml:"idle_conn_timeout"`             // 空闲连接超时时间
	DisableKeepAlives   bool          `json:"disable_keep_alives" yaml:"disable_keep_alives"`         // 是否禁用Keep-Alive
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		BaseURL:   "",
		Timeout:   30 * time.Second,
		Headers:   make(map[string]string),
		UserAgent: "tool-kit-http-client/1.0.0",
		Proxy:     "",
		Insecure:  false,
		Debug:     false,
		Retry:     DefaultRetryConfig(),
		Pool:      DefaultPoolConfig(),
		// 向后兼容
		RetryCount: 3,
		RetryDelay: 1 * time.Second,
	}
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     3,
		BaseDelay:      1 * time.Second,
		MaxDelay:       30 * time.Second,
		Strategy:       RetryStrategyExponential,
		RetryableCodes: []int{500, 502, 503, 504, 408, 429},
	}
}

// DefaultPoolConfig 默认连接池配置
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     0, // 无限制
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}
}

// New 创建HTTP客户端
func New(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}

	// 确保重试和连接池配置存在
	if config.Retry == nil {
		config.Retry = DefaultRetryConfig()
	}
	if config.Pool == nil {
		config.Pool = DefaultPoolConfig()
	}

	// 创建Transport配置连接池
	transport := &http.Transport{
		MaxIdleConns:        config.Pool.MaxIdleConns,
		MaxIdleConnsPerHost: config.Pool.MaxIdleConnsPerHost,
		MaxConnsPerHost:     config.Pool.MaxConnsPerHost,
		IdleConnTimeout:     config.Pool.IdleConnTimeout,
		DisableKeepAlives:   config.Pool.DisableKeepAlives,
	}

	client := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

	return &Client{
		client: client,
		config: config,
	}
}

// NewWithRetry 创建带重试配置的HTTP客户端
func NewWithRetry(baseURL string, retryConfig *RetryConfig) *Client {
	config := &Config{
		BaseURL: baseURL,
		Retry:   retryConfig,
		Pool:    DefaultPoolConfig(),
	}
	return New(config)
}

// NewWithPool 创建带连接池配置的HTTP客户端
func NewWithPool(baseURL string, poolConfig *PoolConfig) *Client {
	config := &Config{
		BaseURL: baseURL,
		Retry:   DefaultRetryConfig(),
		Pool:    poolConfig,
	}
	return New(config)
}

// NewWithRetryAndPool 创建带重试和连接池配置的HTTP客户端
func NewWithRetryAndPool(baseURL string, retryConfig *RetryConfig, poolConfig *PoolConfig) *Client {
	config := &Config{
		BaseURL: baseURL,
		Retry:   retryConfig,
		Pool:    poolConfig,
	}
	return New(config)
}

// Request HTTP请求结构
type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
	Query   map[string]string `json:"query,omitempty"`
}

// Response HTTP响应结构
type Response struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Text       string            `json:"text"`
}

// Get 发送GET请求
func (c *Client) Get(ctx context.Context, url string, options ...Option) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodGet,
		URL:    url,
	}, options...)
}

// Post 发送POST请求
func (c *Client) Post(ctx context.Context, url string, body interface{}, options ...Option) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodPost,
		URL:    url,
		Body:   body,
	}, options...)
}

// Put 发送PUT请求
func (c *Client) Put(ctx context.Context, url string, body interface{}, options ...Option) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodPut,
		URL:    url,
		Body:   body,
	}, options...)
}

// Delete 发送DELETE请求
func (c *Client) Delete(ctx context.Context, url string, options ...Option) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodDelete,
		URL:    url,
	}, options...)
}

// Patch 发送PATCH请求
func (c *Client) Patch(ctx context.Context, url string, body interface{}, options ...Option) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodPatch,
		URL:    url,
		Body:   body,
	}, options...)
}

// calculateDelay 计算重试延迟
func (c *Client) calculateDelay(attempt int) time.Duration {
	retryConfig := c.config.Retry
	if retryConfig == nil {
		return retryConfig.BaseDelay
	}

	var delay time.Duration
	switch retryConfig.Strategy {
	case RetryStrategyFixed:
		delay = retryConfig.BaseDelay
	case RetryStrategyExponential:
		delay = time.Duration(float64(retryConfig.BaseDelay) * math.Pow(2, float64(attempt)))
	case RetryStrategyLinear:
		delay = retryConfig.BaseDelay * time.Duration(attempt+1)
	default:
		delay = retryConfig.BaseDelay
	}

	// 限制最大延迟
	if delay > retryConfig.MaxDelay {
		delay = retryConfig.MaxDelay
	}

	return delay
}

// isRetryableError 判断是否可重试的错误
func (c *Client) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 网络错误通常可以重试
	if netErr, ok := err.(net.Error); ok {
		return netErr.Temporary() || netErr.Timeout()
	}

	// 连接错误可以重试
	if opErr, ok := err.(*net.OpError); ok {
		return opErr.Temporary() || opErr.Timeout()
	}

	return false
}

// isRetryableStatusCode 判断状态码是否可重试
func (c *Client) isRetryableStatusCode(statusCode int) bool {
	retryConfig := c.config.Retry
	if retryConfig == nil || len(retryConfig.RetryableCodes) == 0 {
		// 默认可重试的状态码
		return statusCode >= 500 || statusCode == 408 || statusCode == 429
	}

	for _, code := range retryConfig.RetryableCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}

// Do 发送HTTP请求
func (c *Client) Do(ctx context.Context, req *Request, options ...Option) (*Response, error) {
	// 应用选项
	for _, option := range options {
		option(req)
	}

	var lastErr error
	retryConfig := c.config.Retry
	maxRetries := 0
	if retryConfig != nil {
		maxRetries = retryConfig.MaxRetries
	}

	// 重试循环
	for attempt := 0; attempt <= maxRetries; attempt++ {
		response, err := c.doRequest(ctx, req)

		// 如果请求成功且状态码不需要重试，直接返回
		if err == nil && !c.isRetryableStatusCode(response.StatusCode) {
			return response, nil
		}

		// 如果请求成功但状态码需要重试，或者请求失败
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("HTTP %d: %s", response.StatusCode, response.Text)
		}

		// 如果是最后一次尝试，返回当前结果
		if attempt >= maxRetries {
			if response != nil {
				return response, nil
			}
			break
		}

		// 检查是否应该重试
		shouldRetry := false
		if response != nil {
			// 检查状态码是否可重试
			shouldRetry = c.isRetryableStatusCode(response.StatusCode)
		} else {
			// 检查错误是否可重试
			shouldRetry = c.isRetryableError(err)
		}

		if !shouldRetry {
			if response != nil {
				return response, nil
			}
			break
		}

		// 计算延迟时间
		delay := c.calculateDelay(attempt)

		// 等待重试
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// 继续重试
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries+1, lastErr)
	}
	return nil, fmt.Errorf("request failed after %d attempts", maxRetries+1)
}

// doRequest 执行单次HTTP请求
func (c *Client) doRequest(ctx context.Context, req *Request) (*Response, error) {
	// 构建完整URL
	url := req.URL
	if c.config.BaseURL != "" {
		url = c.config.BaseURL + req.URL
	}

	// 创建请求体
	var body io.Reader
	if req.Body != nil {
		switch b := req.Body.(type) {
		case string:
			body = bytes.NewBufferString(b)
		case []byte:
			body = bytes.NewBuffer(b)
		case io.Reader:
			body = b
		default:
			// 尝试JSON序列化
			jsonData, err := json.Marshal(b)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			body = bytes.NewBuffer(jsonData)
		}
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置默认请求头
	for key, value := range c.config.Headers {
		httpReq.Header.Set(key, value)
	}

	// 设置请求头
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 设置查询参数
	if len(req.Query) > 0 {
		q := httpReq.URL.Query()
		for key, value := range req.Query {
			q.Add(key, value)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	// 发送请求
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 构建响应
	response := &Response{
		StatusCode: resp.StatusCode,
		Headers:    make(map[string]string),
		Body:       respBody,
		Text:       string(respBody),
	}

	// 设置响应头
	for key, values := range resp.Header {
		if len(values) > 0 {
			response.Headers[key] = values[0]
		}
	}

	return response, nil
}

// GetJSON 发送GET请求并解析JSON响应
func (c *Client) GetJSON(ctx context.Context, url string, result interface{}, options ...Option) error {
	resp, err := c.Get(ctx, url, options...)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp.Body, result)
}

// PostJSON 发送POST请求并解析JSON响应
func (c *Client) PostJSON(ctx context.Context, url string, body interface{}, result interface{}, options ...Option) error {
	resp, err := c.Post(ctx, url, body, options...)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp.Body, result)
}

// PutJSON 发送PUT请求并解析JSON响应
func (c *Client) PutJSON(ctx context.Context, url string, body interface{}, result interface{}, options ...Option) error {
	resp, err := c.Put(ctx, url, body, options...)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp.Body, result)
}

// DeleteJSON 发送DELETE请求并解析JSON响应
func (c *Client) DeleteJSON(ctx context.Context, url string, result interface{}, options ...Option) error {
	resp, err := c.Delete(ctx, url, options...)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp.Body, result)
}

// PatchJSON 发送PATCH请求并解析JSON响应
func (c *Client) PatchJSON(ctx context.Context, url string, body interface{}, result interface{}, options ...Option) error {
	resp, err := c.Patch(ctx, url, body, options...)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp.Body, result)
}

// Option 请求选项
type Option func(*Request)

// WithHeader 设置请求头
func WithHeader(key, value string) Option {
	return func(r *Request) {
		if r.Headers == nil {
			r.Headers = make(map[string]string)
		}
		r.Headers[key] = value
	}
}

// WithHeaders 设置多个请求头
func WithHeaders(headers map[string]string) Option {
	return func(r *Request) {
		if r.Headers == nil {
			r.Headers = make(map[string]string)
		}
		for key, value := range headers {
			r.Headers[key] = value
		}
	}
}

// WithQuery 设置查询参数
func WithQuery(key, value string) Option {
	return func(r *Request) {
		if r.Query == nil {
			r.Query = make(map[string]string)
		}
		r.Query[key] = value
	}
}

// WithQueries 设置多个查询参数
func WithQueries(queries map[string]string) Option {
	return func(r *Request) {
		if r.Query == nil {
			r.Query = make(map[string]string)
		}
		for key, value := range queries {
			r.Query[key] = value
		}
	}
}

// WithAuth 设置认证
func WithAuth(authType, token string) Option {
	return func(r *Request) {
		if r.Headers == nil {
			r.Headers = make(map[string]string)
		}
		r.Headers["Authorization"] = authType + " " + token
	}
}

// WithBasicAuth 设置基本认证
func WithBasicAuth(username, password string) Option {
	return func(r *Request) {
		if r.Headers == nil {
			r.Headers = make(map[string]string)
		}
		r.Headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
	}
}

// WithBearerToken 设置Bearer Token
func WithBearerToken(token string) Option {
	return func(r *Request) {
		if r.Headers == nil {
			r.Headers = make(map[string]string)
		}
		r.Headers["Authorization"] = "Bearer " + token
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(r *Request) {
		// 这个选项在当前实现中不起作用，因为超时是在客户端级别设置的
		// 但为了保持API一致性，我们保留这个选项
	}
}

// WithRetry 设置重试
func WithRetry(count int, delay time.Duration) Option {
	return func(r *Request) {
		// 这个选项在当前实现中不起作用，因为重试是在客户端级别设置的
		// 但为了保持API一致性，我们保留这个选项
	}
}

// SetBaseURL 设置基础URL
func (c *Client) SetBaseURL(baseURL string) {
	c.config.BaseURL = baseURL
}

// SetTimeout 设置超时时间
func (c *Client) SetTimeout(timeout time.Duration) {
	c.config.Timeout = timeout
	c.client.Timeout = timeout
}

// SetHeader 设置默认请求头
func (c *Client) SetHeader(key, value string) {
	c.config.Headers[key] = value
}

// SetHeaders 设置多个默认请求头
func (c *Client) SetHeaders(headers map[string]string) {
	for key, value := range headers {
		c.config.Headers[key] = value
	}
}

// SetRetry 设置重试配置
func (c *Client) SetRetry(count int, delay time.Duration) {
	c.config.RetryCount = count
	c.config.RetryDelay = delay
}

// SetProxy 设置代理
func (c *Client) SetProxy(proxyURL string) {
	c.config.Proxy = proxyURL
}

// SetInsecure 设置SSL验证
func (c *Client) SetInsecure(insecure bool) {
	c.config.Insecure = insecure
}

// SetDebug 设置调试模式
func (c *Client) SetDebug(debug bool) {
	c.config.Debug = debug
}

// SetRetryConfig 设置重试配置
func (c *Client) SetRetryConfig(retryConfig *RetryConfig) {
	c.config.Retry = retryConfig
}

// SetPoolConfig 设置连接池配置
func (c *Client) SetPoolConfig(poolConfig *PoolConfig) {
	c.config.Pool = poolConfig
	// 重新创建Transport以应用新配置
	transport := &http.Transport{
		MaxIdleConns:        poolConfig.MaxIdleConns,
		MaxIdleConnsPerHost: poolConfig.MaxIdleConnsPerHost,
		MaxConnsPerHost:     poolConfig.MaxConnsPerHost,
		IdleConnTimeout:     poolConfig.IdleConnTimeout,
		DisableKeepAlives:   poolConfig.DisableKeepAlives,
	}
	c.client.Transport = transport
}

// GetRetryConfig 获取重试配置
func (c *Client) GetRetryConfig() *RetryConfig {
	return c.config.Retry
}

// GetPoolConfig 获取连接池配置
func (c *Client) GetPoolConfig() *PoolConfig {
	return c.config.Pool
}
