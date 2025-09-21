package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 日志接口
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
	Sync() error
}

// OutputConfig 输出配置
type OutputConfig struct {
	Type        string `json:"type" yaml:"type"`                 // 输出类型: stdout, file
	File        string `json:"file" yaml:"file"`                 // 文件路径（当Type为file时）
	MaxSize     int    `json:"max_size" yaml:"max_size"`         // 日志文件最大大小(MB)
	MaxBackups  int    `json:"max_backups" yaml:"max_backups"`   // 最大备份文件数
	MaxAge      int    `json:"max_age" yaml:"max_age"`           // 最大保存天数
	Compress    bool   `json:"compress" yaml:"compress"`         // 是否压缩备份文件
	TimeFormat  string `json:"time_format" yaml:"time_format"`   // 时间格式，如 "2006.01.02_15:04:05.000"
	UseRotation bool   `json:"use_rotation" yaml:"use_rotation"` // 是否启用日志轮转
}

// Config 日志配置
type Config struct {
	Level   string         `json:"level" yaml:"level"`     // 日志级别: debug, info, warn, error, fatal
	Format  string         `json:"format" yaml:"format"`   // 日志格式: json, console
	Outputs []OutputConfig `json:"outputs" yaml:"outputs"` // 输出配置列表

	// 向后兼容的字段
	Output     string `json:"output" yaml:"output"`           // 输出方式: stdout, file (已废弃，使用Outputs)
	File       string `json:"file" yaml:"file"`               // 日志文件路径 (已废弃，使用Outputs)
	MaxSize    int    `json:"max_size" yaml:"max_size"`       // 日志文件最大大小(MB) (已废弃，使用Outputs)
	MaxBackups int    `json:"max_backups" yaml:"max_backups"` // 最大备份文件数 (已废弃，使用Outputs)
	MaxAge     int    `json:"max_age" yaml:"max_age"`         // 最大保存天数 (已废弃，使用Outputs)
	Compress   bool   `json:"compress" yaml:"compress"`       // 是否压缩备份文件 (已废弃，使用Outputs)
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:  "info",
		Format: "json",
		Outputs: []OutputConfig{
			{
				Type:       "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   true,
			},
		},
		// 向后兼容的默认值
		Output:     "stdout",
		File:       "logs/app.log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}
}

// MultiWriter 多输出写入器
type MultiWriter struct {
	writers []io.Writer
	mu      sync.RWMutex
}

// NewMultiWriter 创建多输出写入器
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

// Write 实现io.Writer接口
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	// 写入到所有writer
	for _, w := range mw.writers {
		if n, err = w.Write(p); err != nil {
			return n, err
		}
	}
	return len(p), nil
}

// AddWriter 添加新的写入器
func (mw *MultiWriter) AddWriter(w io.Writer) {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	mw.writers = append(mw.writers, w)
}

// logger 日志实现
type logger struct {
	zap *zap.Logger
}

// New 创建新的日志实例
func New(config *Config) (Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 解析日志级别
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return nil, err
	}

	// 创建编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 选择编码器
	var encoder zapcore.Encoder
	switch config.Format {
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 处理多输出
	var writeSyncer zapcore.WriteSyncer
	if len(config.Outputs) > 0 {
		// 使用新的多输出配置
		writers := make([]io.Writer, 0, len(config.Outputs))
		for _, output := range config.Outputs {
			writer, err := createWriter(output)
			if err != nil {
				return nil, err
			}
			writers = append(writers, writer)
		}

		if len(writers) == 1 {
			writeSyncer = zapcore.AddSync(writers[0])
		} else {
			multiWriter := NewMultiWriter(writers...)
			writeSyncer = zapcore.AddSync(multiWriter)
		}
	} else {
		// 向后兼容：使用旧的单输出配置
		writeSyncer, err = createLegacyWriteSyncer(config)
		if err != nil {
			return nil, err
		}
	}

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &logger{zap: zapLogger}, nil
}

// GenerateTimeBasedFileName 生成基于时间的文件名
func GenerateTimeBasedFileName(basePath, timeFormat string) string {
	if timeFormat == "" {
		timeFormat = "2006.01.02_15:04:05.000"
	}

	now := time.Now()
	timeStr := now.Format(timeFormat)

	// 如果basePath包含目录，则保持目录结构
	dir := filepath.Dir(basePath)
	ext := filepath.Ext(basePath)
	name := filepath.Base(basePath)
	if ext != "" {
		name = name[:len(name)-len(ext)]
	}

	// 生成新文件名: app_2025.09.21_20:23:03.000.log
	newName := fmt.Sprintf("%s_%s%s", name, timeStr, ext)

	if dir != "." {
		return filepath.Join(dir, newName)
	}
	return newName
}

// createWriter 根据OutputConfig创建Writer
func createWriter(output OutputConfig) (io.Writer, error) {
	switch output.Type {
	case "file":
		if output.File == "" {
			return nil, fmt.Errorf("file path is required for file output")
		}

		// 确保目录存在
		if err := os.MkdirAll(filepath.Dir(output.File), 0755); err != nil {
			return nil, err
		}

		// 如果启用了轮转，使用lumberjack
		if output.UseRotation {
			// 生成基于时间的文件名
			timeBasedFile := GenerateTimeBasedFileName(output.File, output.TimeFormat)

			// 设置lumberjack配置
			lj := &lumberjack.Logger{
				Filename:   timeBasedFile,
				MaxSize:    output.MaxSize, // MB
				MaxBackups: output.MaxBackups,
				MaxAge:     output.MaxAge, // days
				Compress:   output.Compress,
			}
			return lj, nil
		} else {
			// 普通文件写入
			file, err := os.OpenFile(output.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				return nil, err
			}
			return file, nil
		}
	case "stdout":
		return os.Stdout, nil
	default:
		return os.Stdout, nil
	}
}

// createLegacyWriteSyncer 向后兼容的WriteSyncer创建
func createLegacyWriteSyncer(config *Config) (zapcore.WriteSyncer, error) {
	switch config.Output {
	case "file":
		if err := os.MkdirAll(filepath.Dir(config.File), 0755); err != nil {
			return nil, err
		}
		file, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		return zapcore.AddSync(file), nil
	case "stdout":
		return zapcore.AddSync(os.Stdout), nil
	default:
		return zapcore.AddSync(os.Stdout), nil
	}
}

// NewWithWriter 使用指定的Writer创建日志实例
func NewWithWriter(level, format string, writer io.Writer) (Logger, error) {
	config := &Config{
		Level:  level,
		Format: format,
		Output: "custom",
	}

	// 解析日志级别
	zapLevel, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return nil, err
	}

	// 创建编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 选择编码器
	var encoder zapcore.Encoder
	switch config.Format {
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 创建核心
	core := zapcore.NewCore(encoder, zapcore.AddSync(writer), zapLevel)

	// 创建logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &logger{zap: zapLogger}, nil
}

// Debug 调试日志
func (l *logger) Debug(args ...interface{}) {
	l.zap.Sugar().Debug(args...)
}

// Debugf 格式化调试日志
func (l *logger) Debugf(format string, args ...interface{}) {
	l.zap.Sugar().Debugf(format, args...)
}

// Info 信息日志
func (l *logger) Info(args ...interface{}) {
	l.zap.Sugar().Info(args...)
}

// Infof 格式化信息日志
func (l *logger) Infof(format string, args ...interface{}) {
	l.zap.Sugar().Infof(format, args...)
}

// Warn 警告日志
func (l *logger) Warn(args ...interface{}) {
	l.zap.Sugar().Warn(args...)
}

// Warnf 格式化警告日志
func (l *logger) Warnf(format string, args ...interface{}) {
	l.zap.Sugar().Warnf(format, args...)
}

// Error 错误日志
func (l *logger) Error(args ...interface{}) {
	l.zap.Sugar().Error(args...)
}

// Errorf 格式化错误日志
func (l *logger) Errorf(format string, args ...interface{}) {
	l.zap.Sugar().Errorf(format, args...)
}

// Fatal 致命错误日志
func (l *logger) Fatal(args ...interface{}) {
	l.zap.Sugar().Fatal(args...)
}

// Fatalf 格式化致命错误日志
func (l *logger) Fatalf(format string, args ...interface{}) {
	l.zap.Sugar().Fatalf(format, args...)
}

// WithField 添加字段
func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{zap: l.zap.With(zap.Any(key, value))}
}

// WithFields 添加多个字段
func (l *logger) WithFields(fields map[string]interface{}) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &logger{zap: l.zap.With(zapFields...)}
}

// WithError 添加错误字段
func (l *logger) WithError(err error) Logger {
	return &logger{zap: l.zap.With(zap.Error(err))}
}

// Sync 同步日志
func (l *logger) Sync() error {
	return l.zap.Sync()
}

// GetZapLogger 获取底层zap logger
func (l *logger) GetZapLogger() *zap.Logger {
	return l.zap
}

// NewDevelopment 创建开发环境日志
func NewDevelopment() (Logger, error) {
	config := &Config{
		Level:  "debug",
		Format: "console",
		Output: "stdout",
	}
	return New(config)
}

// NewProduction 创建生产环境日志
func NewProduction() (Logger, error) {
	config := &Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	return New(config)
}

// NewTest 创建测试环境日志
func NewTest() (Logger, error) {
	config := &Config{
		Level:  "error",
		Format: "console",
		Output: "stdout",
	}
	return New(config)
}

// NewMultiOutput 创建多输出日志实例
func NewMultiOutput(level, format string, outputs ...OutputConfig) (Logger, error) {
	config := &Config{
		Level:   level,
		Format:  format,
		Outputs: outputs,
	}
	return New(config)
}

// NewFileAndConsole 创建同时输出到文件和控制台的日志实例
func NewFileAndConsole(level, format, filePath string) (Logger, error) {
	config := &Config{
		Level:  level,
		Format: format,
		Outputs: []OutputConfig{
			{Type: "stdout"},
			{Type: "file", File: filePath},
		},
	}
	return New(config)
}

// NewRotatingFile 创建带轮转功能的文件日志实例
func NewRotatingFile(level, format, filePath string, maxSize, maxBackups, maxAge int) (Logger, error) {
	config := &Config{
		Level:  level,
		Format: format,
		Outputs: []OutputConfig{
			{
				Type:        "file",
				File:        filePath,
				MaxSize:     maxSize,
				MaxBackups:  maxBackups,
				MaxAge:      maxAge,
				Compress:    true,
				UseRotation: true,
				TimeFormat:  "2006.01.02_15:04:05.000",
			},
		},
	}
	return New(config)
}

// NewRotatingFileAndConsole 创建同时输出到轮转文件和控制台的日志实例
func NewRotatingFileAndConsole(level, format, filePath string, maxSize, maxBackups, maxAge int) (Logger, error) {
	config := &Config{
		Level:  level,
		Format: format,
		Outputs: []OutputConfig{
			{Type: "stdout"},
			{
				Type:        "file",
				File:        filePath,
				MaxSize:     maxSize,
				MaxBackups:  maxBackups,
				MaxAge:      maxAge,
				Compress:    true,
				UseRotation: true,
				TimeFormat:  "2006.01.02_15:04:05.000",
			},
		},
	}
	return New(config)
}
