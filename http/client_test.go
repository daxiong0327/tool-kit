package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/get":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"query":  r.URL.Query().Get("key"),
			})
		case "/post":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"body":   "received",
			})
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		}
	}))
	defer server.Close()

	ctx := context.Background()

	t.Run("New with default config", func(t *testing.T) {
		client := New(nil)
		assert.NotNil(t, client)
		assert.NotNil(t, client.client)
		assert.NotNil(t, client.config)
	})

	t.Run("New with custom config", func(t *testing.T) {
		config := &Config{
			BaseURL:    server.URL,
			Timeout:    10 * time.Second,
			RetryCount: 2,
			RetryDelay: 100 * time.Millisecond,
			Headers: map[string]string{
				"X-Custom-Header": "test-value",
			},
		}

		client := New(config)
		assert.NotNil(t, client)
		assert.Equal(t, server.URL, client.config.BaseURL)
	})

	t.Run("GET request", func(t *testing.T) {
		client := New(&Config{BaseURL: server.URL})

		resp, err := client.Get(ctx, "/get")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Text, "get")
	})

	t.Run("GET request with query", func(t *testing.T) {
		client := New(&Config{BaseURL: server.URL})

		resp, err := client.Get(ctx, "/get", WithQuery("key", "value"))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Text, "value")
	})

	t.Run("POST request", func(t *testing.T) {
		client := New(&Config{BaseURL: server.URL})

		body := map[string]string{"message": "hello"}
		resp, err := client.Post(ctx, "/post", body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Text, "received")
	})

	t.Run("PUT request", func(t *testing.T) {
		client := New(&Config{BaseURL: server.URL})

		body := map[string]string{"message": "hello"}
		resp, err := client.Put(ctx, "/post", body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Text, "received")
	})

	t.Run("DELETE request", func(t *testing.T) {
		client := New(&Config{BaseURL: server.URL})

		resp, err := client.Delete(ctx, "/get")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("PATCH request", func(t *testing.T) {
		client := New(&Config{BaseURL: server.URL})

		body := map[string]string{"message": "hello"}
		resp, err := client.Patch(ctx, "/post", body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Text, "received")
	})

	t.Run("GET JSON", func(t *testing.T) {
		client := New(&Config{BaseURL: server.URL})

		var result map[string]interface{}
		err := client.GetJSON(ctx, "/get", &result)
		require.NoError(t, err)
		assert.Equal(t, "GET", result["method"])
		assert.Equal(t, "/get", result["path"])
	})

	t.Run("POST JSON", func(t *testing.T) {
		client := New(&Config{BaseURL: server.URL})

		body := map[string]string{"message": "hello"}
		var result map[string]interface{}
		err := client.PostJSON(ctx, "/post", body, &result)
		require.NoError(t, err)
		assert.Equal(t, "POST", result["method"])
		assert.Equal(t, "/post", result["path"])
	})

	t.Run("With headers", func(t *testing.T) {
		client := New(&Config{BaseURL: server.URL})

		resp, err := client.Get(ctx, "/get", WithHeader("X-Test-Header", "test-value"))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("With multiple headers", func(t *testing.T) {
		client := New(&Config{BaseURL: server.URL})

		headers := map[string]string{
			"X-Test-Header-1": "value1",
			"X-Test-Header-2": "value2",
		}
		resp, err := client.Get(ctx, "/get", WithHeaders(headers))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("With basic auth", func(t *testing.T) {
		client := New(&Config{BaseURL: server.URL})

		resp, err := client.Get(ctx, "/get", WithBasicAuth("username", "password"))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("With bearer token", func(t *testing.T) {
		client := New(&Config{BaseURL: server.URL})

		resp, err := client.Get(ctx, "/get", WithBearerToken("test-token"))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("With timeout", func(t *testing.T) {
		client := New(&Config{BaseURL: server.URL})

		resp, err := client.Get(ctx, "/get", WithTimeout(5*time.Second))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Error handling", func(t *testing.T) {
		client := New(&Config{BaseURL: server.URL})

		resp, err := client.Get(ctx, "/error")
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Contains(t, resp.Text, "internal server error")
	})
}

func TestConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultConfig()
		assert.Equal(t, "", config.BaseURL)
		assert.Equal(t, 30*time.Second, config.Timeout)
		assert.Equal(t, 3, config.RetryCount)
		assert.Equal(t, 1*time.Second, config.RetryDelay)
		assert.Equal(t, "tool-kit-http-client/1.0.0", config.UserAgent)
		assert.False(t, config.Insecure)
		assert.False(t, config.Debug)
	})
}

func TestClientMethods(t *testing.T) {
	client := New(nil)

	t.Run("SetBaseURL", func(t *testing.T) {
		client.SetBaseURL("https://api.example.com")
		assert.Equal(t, "https://api.example.com", client.config.BaseURL)
	})

	t.Run("SetTimeout", func(t *testing.T) {
		client.SetTimeout(10 * time.Second)
		assert.Equal(t, 10*time.Second, client.config.Timeout)
	})

	t.Run("SetHeader", func(t *testing.T) {
		client.SetHeader("X-Test", "value")
		assert.Equal(t, "value", client.config.Headers["X-Test"])
	})

	t.Run("SetHeaders", func(t *testing.T) {
		headers := map[string]string{
			"X-Test-1": "value1",
			"X-Test-2": "value2",
		}
		client.SetHeaders(headers)
		assert.Equal(t, "value1", client.config.Headers["X-Test-1"])
		assert.Equal(t, "value2", client.config.Headers["X-Test-2"])
	})

	t.Run("SetRetry", func(t *testing.T) {
		client.SetRetry(5, 2*time.Second)
		assert.Equal(t, 5, client.config.RetryCount)
		assert.Equal(t, 2*time.Second, client.config.RetryDelay)
	})

	t.Run("SetProxy", func(t *testing.T) {
		client.SetProxy("http://proxy.example.com:8080")
		assert.Equal(t, "http://proxy.example.com:8080", client.config.Proxy)
	})

	t.Run("SetInsecure", func(t *testing.T) {
		client.SetInsecure(true)
		assert.True(t, client.config.Insecure)
	})

	t.Run("SetDebug", func(t *testing.T) {
		client.SetDebug(true)
		assert.True(t, client.config.Debug)
	})
}

func TestRetry(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}))
	defer server.Close()

	t.Run("Retry with exponential backoff", func(t *testing.T) {
		attemptCount = 0
		retryConfig := &RetryConfig{
			MaxRetries:     3,
			BaseDelay:      100 * time.Millisecond,
			MaxDelay:       1 * time.Second,
			Strategy:       RetryStrategyExponential,
			RetryableCodes: []int{500, 502, 503, 504},
		}

		client := NewWithRetry(server.URL, retryConfig)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/test")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Text, "success")
		assert.Equal(t, 3, attemptCount) // 应该重试2次，总共3次尝试
	})

	t.Run("Retry with fixed delay", func(t *testing.T) {
		attemptCount = 0
		retryConfig := &RetryConfig{
			MaxRetries:     2,
			BaseDelay:      50 * time.Millisecond,
			MaxDelay:       1 * time.Second,
			Strategy:       RetryStrategyFixed,
			RetryableCodes: []int{500, 502, 503, 504},
		}

		client := NewWithRetry(server.URL, retryConfig)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/test")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Text, "success")
		assert.Equal(t, 3, attemptCount) // 应该重试2次，总共3次尝试
	})

	t.Run("Retry with linear delay", func(t *testing.T) {
		attemptCount = 0
		retryConfig := &RetryConfig{
			MaxRetries:     2,
			BaseDelay:      50 * time.Millisecond,
			MaxDelay:       1 * time.Second,
			Strategy:       RetryStrategyLinear,
			RetryableCodes: []int{500, 502, 503, 504},
		}

		client := NewWithRetry(server.URL, retryConfig)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/test")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Text, "success")
		assert.Equal(t, 3, attemptCount) // 应该重试2次，总共3次尝试
	})

	t.Run("No retry for non-retryable status code", func(t *testing.T) {
		attemptCount = 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			w.WriteHeader(http.StatusBadRequest) // 400不是可重试的状态码
			w.Write([]byte("bad request"))
		}))
		defer server.Close()

		retryConfig := &RetryConfig{
			MaxRetries:     3,
			BaseDelay:      50 * time.Millisecond,
			Strategy:       RetryStrategyFixed,
			RetryableCodes: []int{500, 502, 503, 504},
		}

		client := NewWithRetry(server.URL, retryConfig)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/test")
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, 1, attemptCount) // 不应该重试
	})

	t.Run("Max retries exceeded", func(t *testing.T) {
		attemptCount = 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		}))
		defer server.Close()

		retryConfig := &RetryConfig{
			MaxRetries:     2,
			BaseDelay:      50 * time.Millisecond,
			Strategy:       RetryStrategyFixed,
			RetryableCodes: []int{500, 502, 503, 504},
		}

		client := NewWithRetry(server.URL, retryConfig)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/test")
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Equal(t, 3, attemptCount) // 应该重试2次，总共3次尝试
	})
}

func TestPoolConfig(t *testing.T) {
	t.Run("Default pool config", func(t *testing.T) {
		poolConfig := DefaultPoolConfig()
		assert.Equal(t, 100, poolConfig.MaxIdleConns)
		assert.Equal(t, 10, poolConfig.MaxIdleConnsPerHost)
		assert.Equal(t, 0, poolConfig.MaxConnsPerHost)
		assert.Equal(t, 90*time.Second, poolConfig.IdleConnTimeout)
		assert.False(t, poolConfig.DisableKeepAlives)
	})

	t.Run("Custom pool config", func(t *testing.T) {
		poolConfig := &PoolConfig{
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 5,
			MaxConnsPerHost:     20,
			IdleConnTimeout:     60 * time.Second,
			DisableKeepAlives:   true,
		}

		client := NewWithPool("https://api.example.com", poolConfig)
		assert.Equal(t, poolConfig, client.GetPoolConfig())
	})

	t.Run("Set pool config", func(t *testing.T) {
		client := New(nil)
		poolConfig := &PoolConfig{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 20,
			MaxConnsPerHost:     50,
			IdleConnTimeout:     120 * time.Second,
			DisableKeepAlives:   false,
		}

		client.SetPoolConfig(poolConfig)
		assert.Equal(t, poolConfig, client.GetPoolConfig())
	})
}

func TestRetryConfig(t *testing.T) {
	t.Run("Default retry config", func(t *testing.T) {
		retryConfig := DefaultRetryConfig()
		assert.Equal(t, 3, retryConfig.MaxRetries)
		assert.Equal(t, 1*time.Second, retryConfig.BaseDelay)
		assert.Equal(t, 30*time.Second, retryConfig.MaxDelay)
		assert.Equal(t, RetryStrategyExponential, retryConfig.Strategy)
		assert.Contains(t, retryConfig.RetryableCodes, 500)
		assert.Contains(t, retryConfig.RetryableCodes, 502)
		assert.Contains(t, retryConfig.RetryableCodes, 503)
		assert.Contains(t, retryConfig.RetryableCodes, 504)
	})

	t.Run("Set retry config", func(t *testing.T) {
		client := New(nil)
		retryConfig := &RetryConfig{
			MaxRetries:     5,
			BaseDelay:      2 * time.Second,
			MaxDelay:       60 * time.Second,
			Strategy:       RetryStrategyLinear,
			RetryableCodes: []int{500, 502, 503, 504, 408, 429},
		}

		client.SetRetryConfig(retryConfig)
		assert.Equal(t, retryConfig, client.GetRetryConfig())
	})
}

func BenchmarkClient(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := New(&Config{BaseURL: server.URL})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Get(ctx, "/")
		if err != nil {
			b.Fatal(err)
		}
	}
}
