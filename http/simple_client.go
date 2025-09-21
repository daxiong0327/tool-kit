package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SimpleClient 简化的HTTP客户端
type SimpleClient struct {
	client *http.Client
	config *Config
}

// NewSimpleClient 创建简化的HTTP客户端
func NewSimpleClient(config *Config) *SimpleClient {
	if config == nil {
		config = DefaultConfig()
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &SimpleClient{
		client: client,
		config: config,
	}
}

// Get 发送GET请求
func (c *SimpleClient) Get(ctx context.Context, url string, options ...Option) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodGet,
		URL:    url,
	}, options...)
}

// Post 发送POST请求
func (c *SimpleClient) Post(ctx context.Context, url string, body interface{}, options ...Option) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodPost,
		URL:    url,
		Body:   body,
	}, options...)
}

// Put 发送PUT请求
func (c *SimpleClient) Put(ctx context.Context, url string, body interface{}, options ...Option) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodPut,
		URL:    url,
		Body:   body,
	}, options...)
}

// Delete 发送DELETE请求
func (c *SimpleClient) Delete(ctx context.Context, url string, options ...Option) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodDelete,
		URL:    url,
	}, options...)
}

// Patch 发送PATCH请求
func (c *SimpleClient) Patch(ctx context.Context, url string, body interface{}, options ...Option) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodPatch,
		URL:    url,
		Body:   body,
	}, options...)
}

// Do 发送HTTP请求
func (c *SimpleClient) Do(ctx context.Context, req *Request, options ...Option) (*Response, error) {
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
func (c *SimpleClient) GetJSON(ctx context.Context, url string, result interface{}, options ...Option) error {
	resp, err := c.Get(ctx, url, options...)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp.Body, result)
}

// PostJSON 发送POST请求并解析JSON响应
func (c *SimpleClient) PostJSON(ctx context.Context, url string, body interface{}, result interface{}, options ...Option) error {
	resp, err := c.Post(ctx, url, body, options...)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp.Body, result)
}

// PutJSON 发送PUT请求并解析JSON响应
func (c *SimpleClient) PutJSON(ctx context.Context, url string, body interface{}, result interface{}, options ...Option) error {
	resp, err := c.Put(ctx, url, body, options...)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp.Body, result)
}

// DeleteJSON 发送DELETE请求并解析JSON响应
func (c *SimpleClient) DeleteJSON(ctx context.Context, url string, result interface{}, options ...Option) error {
	resp, err := c.Delete(ctx, url, options...)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp.Body, result)
}

// PatchJSON 发送PATCH请求并解析JSON响应
func (c *SimpleClient) PatchJSON(ctx context.Context, url string, body interface{}, result interface{}, options ...Option) error {
	resp, err := c.Patch(ctx, url, body, options...)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp.Body, result)
}

// SetBaseURL 设置基础URL
func (c *SimpleClient) SetBaseURL(baseURL string) {
	c.config.BaseURL = baseURL
}

// SetTimeout 设置超时时间
func (c *SimpleClient) SetTimeout(timeout time.Duration) {
	c.config.Timeout = timeout
	c.client.Timeout = timeout
}

// SetHeader 设置默认请求头
func (c *SimpleClient) SetHeader(key, value string) {
	c.config.Headers[key] = value
}

// SetHeaders 设置多个默认请求头
func (c *SimpleClient) SetHeaders(headers map[string]string) {
	for key, value := range headers {
		c.config.Headers[key] = value
	}
}
