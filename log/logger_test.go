package log

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Run("New with default config", func(t *testing.T) {
		logger, err := New(nil)
		require.NoError(t, err)
		assert.NotNil(t, logger)

		logger.Info("test message")
		logger.Infof("test message with format: %s", "value")
	})

	t.Run("New with custom config", func(t *testing.T) {
		config := &Config{
			Level:  "debug",
			Format: "console",
			Output: "stdout",
		}

		logger, err := New(config)
		require.NoError(t, err)
		assert.NotNil(t, logger)

		logger.Debug("debug message")
		logger.Info("info message")
		logger.Warn("warn message")
		logger.Error("error message")
	})

	t.Run("NewWithWriter", func(t *testing.T) {
		var buf bytes.Buffer
		logger, err := NewWithWriter("info", "json", &buf)
		require.NoError(t, err)
		assert.NotNil(t, logger)

		logger.Info("test message")
		assert.Contains(t, buf.String(), "test message")
	})

	t.Run("WithField", func(t *testing.T) {
		var buf bytes.Buffer
		logger, err := NewWithWriter("info", "json", &buf)
		require.NoError(t, err)

		logger.WithField("key", "value").Info("test message")
		assert.Contains(t, buf.String(), "key")
		assert.Contains(t, buf.String(), "value")
	})

	t.Run("WithFields", func(t *testing.T) {
		var buf bytes.Buffer
		logger, err := NewWithWriter("info", "json", &buf)
		require.NoError(t, err)

		fields := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}
		logger.WithFields(fields).Info("test message")
		assert.Contains(t, buf.String(), "key1")
		assert.Contains(t, buf.String(), "value1")
		assert.Contains(t, buf.String(), "key2")
		assert.Contains(t, buf.String(), "value2")
	})

	t.Run("WithError", func(t *testing.T) {
		var buf bytes.Buffer
		logger, err := NewWithWriter("info", "json", &buf)
		require.NoError(t, err)

		testErr := assert.AnError
		logger.WithError(testErr).Error("test error")
		assert.Contains(t, buf.String(), "error")
	})

	t.Run("Development logger", func(t *testing.T) {
		logger, err := NewDevelopment()
		require.NoError(t, err)
		assert.NotNil(t, logger)

		logger.Debug("debug message")
		logger.Info("info message")
	})

	t.Run("Production logger", func(t *testing.T) {
		logger, err := NewProduction()
		require.NoError(t, err)
		assert.NotNil(t, logger)

		logger.Info("info message")
	})

	t.Run("Test logger", func(t *testing.T) {
		logger, err := NewTest()
		require.NoError(t, err)
		assert.NotNil(t, logger)

		logger.Error("error message")
	})
}

func TestConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultConfig()
		assert.Equal(t, "info", config.Level)
		assert.Equal(t, "json", config.Format)
		assert.Equal(t, "stdout", config.Output)
		assert.Equal(t, "logs/app.log", config.File)
		assert.Equal(t, 100, config.MaxSize)
		assert.Equal(t, 3, config.MaxBackups)
		assert.Equal(t, 7, config.MaxAge)
		assert.True(t, config.Compress)
	})
}

func TestMultiOutput(t *testing.T) {
	t.Run("NewMultiOutput with stdout and file", func(t *testing.T) {
		outputs := []OutputConfig{
			{Type: "stdout"},
			{Type: "file", File: "test.log"},
		}

		logger, err := NewMultiOutput("info", "json", outputs...)
		require.NoError(t, err)
		assert.NotNil(t, logger)

		logger.Info("multi output test")
		logger.Sync()

		// 清理测试文件
		os.Remove("test.log")
	})

	t.Run("NewFileAndConsole", func(t *testing.T) {
		logger, err := NewFileAndConsole("info", "console", "test_console.log")
		require.NoError(t, err)
		assert.NotNil(t, logger)

		logger.Info("file and console test")
		logger.Sync()

		// 清理测试文件
		os.Remove("test_console.log")
	})

	t.Run("MultiWriter functionality", func(t *testing.T) {
		var buf1, buf2 bytes.Buffer

		multiWriter := NewMultiWriter(&buf1, &buf2)
		testMessage := "test message"

		n, err := multiWriter.Write([]byte(testMessage))
		require.NoError(t, err)
		assert.Equal(t, len(testMessage), n)
		assert.Equal(t, testMessage, buf1.String())
		assert.Equal(t, testMessage, buf2.String())
	})

	t.Run("MultiWriter with dynamic writer addition", func(t *testing.T) {
		var buf1, buf2 bytes.Buffer

		multiWriter := NewMultiWriter(&buf1)
		testMessage1 := "first message"
		testMessage2 := "second message"

		// 写入第一个消息
		multiWriter.Write([]byte(testMessage1))
		assert.Equal(t, testMessage1, buf1.String())
		assert.Empty(t, buf2.String())

		// 添加新的writer
		multiWriter.AddWriter(&buf2)

		// 写入第二个消息
		multiWriter.Write([]byte(testMessage2))
		assert.Equal(t, testMessage1+testMessage2, buf1.String())
		assert.Equal(t, testMessage2, buf2.String())
	})

	t.Run("Backward compatibility", func(t *testing.T) {
		// 测试旧的配置方式仍然工作
		config := &Config{
			Level:  "info",
			Format: "json",
			Output: "stdout", // 使用旧的字段
		}

		logger, err := New(config)
		require.NoError(t, err)
		assert.NotNil(t, logger)

		logger.Info("backward compatibility test")
	})
}

func TestLogRotation(t *testing.T) {
	t.Run("NewRotatingFile", func(t *testing.T) {
		logger, err := NewRotatingFile("info", "json", "logs/rotating_test.log", 1, 3, 7)
		require.NoError(t, err)
		assert.NotNil(t, logger)

		logger.Info("rotating file test")
		logger.Sync()

		// 清理测试文件
		os.Remove("logs/rotating_test.log")
	})

	t.Run("NewRotatingFileAndConsole", func(t *testing.T) {
		logger, err := NewRotatingFileAndConsole("info", "console", "logs/rotating_console_test.log", 1, 3, 7)
		require.NoError(t, err)
		assert.NotNil(t, logger)

		logger.Info("rotating file and console test")
		logger.Sync()

		// 清理测试文件
		os.Remove("logs/rotating_console_test.log")
	})

	t.Run("Time-based filename generation", func(t *testing.T) {
		// 测试文件名生成
		basePath := "logs/app.log"
		timeFormat := "2006.01.02_15:04:05.000"

		// 由于时间会变化，我们只测试格式是否正确
		fileName := GenerateTimeBasedFileName(basePath, timeFormat)
		assert.Contains(t, fileName, "app_")
		assert.Contains(t, fileName, ".log")
		assert.Contains(t, fileName, "logs/")
	})

	t.Run("Custom time format", func(t *testing.T) {
		basePath := "test.log"
		timeFormat := "2006-01-02_15-04-05"

		fileName := GenerateTimeBasedFileName(basePath, timeFormat)
		assert.Contains(t, fileName, "test_")
		assert.Contains(t, fileName, ".log")
	})

	t.Run("Rotation with custom config", func(t *testing.T) {
		outputs := []OutputConfig{
			{
				Type:        "file",
				File:        "logs/custom_rotation.log",
				MaxSize:     1, // 1MB
				MaxBackups:  2,
				MaxAge:      3, // 3 days
				Compress:    true,
				UseRotation: true,
				TimeFormat:  "2006.01.02_15:04:05.000",
			},
		}

		logger, err := NewMultiOutput("info", "json", outputs...)
		require.NoError(t, err)
		assert.NotNil(t, logger)

		logger.Info("custom rotation test")
		logger.Sync()

		// 清理测试文件
		os.Remove("logs/custom_rotation.log")
	})
}

func BenchmarkLogger(b *testing.B) {
	var buf bytes.Buffer
	logger, err := NewWithWriter("info", "json", &buf)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}
