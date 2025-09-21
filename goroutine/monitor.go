package goroutine

import (
	"runtime"
	"sync"
	"time"
)

// Monitor 协程监控器
type Monitor struct {
	executor    *SafeExecutor
	interval    time.Duration
	stopChan    chan struct{}
	mu          sync.RWMutex
	handlers    []MonitorHandler
	stats       *MonitorStats
	alertConfig *AlertConfig
}

// MonitorHandler 监控处理器
type MonitorHandler func(stats *MonitorStats)

// MonitorStats 监控统计信息
type MonitorStats struct {
	Timestamp     time.Time `json:"timestamp"`
	NumGoroutines int       `json:"num_goroutines"`
	NumCgoCalls   int64     `json:"num_cgo_calls"`
	NumGC         uint32    `json:"num_gc"`
	PauseTotalNs  uint64    `json:"pause_total_ns"`
	LastGC        time.Time `json:"last_gc"`
	HeapAlloc     uint64    `json:"heap_alloc"`
	HeapSys       uint64    `json:"heap_sys"`
	HeapIdle      uint64    `json:"heap_idle"`
	HeapInuse     uint64    `json:"heap_inuse"`
	HeapReleased  uint64    `json:"heap_released"`
	HeapObjects   uint64    `json:"heap_objects"`
	StackInuse    uint64    `json:"stack_inuse"`
	StackSys      uint64    `json:"stack_sys"`
	MSpanInuse    uint64    `json:"mspan_inuse"`
	MSpanSys      uint64    `json:"mspan_sys"`
	MCacheInuse   uint64    `json:"mcache_inuse"`
	MCacheSys     uint64    `json:"mcache_sys"`
	BuckHashSys   uint64    `json:"buck_hash_sys"`
	GCSys         uint64    `json:"gc_sys"`
	OtherSys      uint64    `json:"other_sys"`
	NextGC        uint64    `json:"next_gc"`
	GCCPUFraction float64   `json:"gc_cpu_fraction"`
	EnableGC      bool      `json:"enable_gc"`
	DebugGC       bool      `json:"debug_gc"`
	BySize        [61]struct {
		Size    uint32
		Mallocs uint64
		Frees   uint64
	} `json:"by_size"`

	// 自定义统计
	CustomStats map[string]interface{} `json:"custom_stats"`
}

// AlertConfig 告警配置
type AlertConfig struct {
	MaxGoroutines int           `json:"max_goroutines"`  // 最大协程数告警阈值
	MaxHeapAlloc  uint64        `json:"max_heap_alloc"`  // 最大堆内存告警阈值
	MaxStackInuse uint64        `json:"max_stack_inuse"` // 最大栈内存告警阈值
	CheckInterval time.Duration `json:"check_interval"`  // 检查间隔
	AlertHandler  AlertHandler  `json:"-"`               // 告警处理器
}

// AlertHandler 告警处理器
type AlertHandler func(alert *Alert)

// Alert 告警信息
type Alert struct {
	Type      string      `json:"type"`
	Message   string      `json:"message"`
	Value     interface{} `json:"value"`
	Threshold interface{} `json:"threshold"`
	Timestamp time.Time   `json:"timestamp"`
	Severity  string      `json:"severity"` // low, medium, high, critical
}

// NewMonitor 创建协程监控器
func NewMonitor(executor *SafeExecutor, interval time.Duration) *Monitor {
	return &Monitor{
		executor: executor,
		interval: interval,
		stopChan: make(chan struct{}),
		stats:    &MonitorStats{},
		alertConfig: &AlertConfig{
			MaxGoroutines: 1000,
			MaxHeapAlloc:  100 * 1024 * 1024, // 100MB
			MaxStackInuse: 10 * 1024 * 1024,  // 10MB
			CheckInterval: 30 * time.Second,
		},
	}
}

// SetAlertConfig 设置告警配置
func (m *Monitor) SetAlertConfig(config *AlertConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alertConfig = config
}

// AddHandler 添加监控处理器
func (m *Monitor) AddHandler(handler MonitorHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handler)
}

// Start 启动监控
func (m *Monitor) Start() {
	m.executor.Go(func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.collectStats()
				m.checkAlerts()
				m.notifyHandlers()
			case <-m.stopChan:
				return
			}
		}
	})
}

// Stop 停止监控
func (m *Monitor) Stop() {
	close(m.stopChan)
}

// collectStats 收集统计信息
func (m *Monitor) collectStats() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats.Timestamp = time.Now()
	m.stats.NumGoroutines = runtime.NumGoroutine()
	m.stats.NumCgoCalls = runtime.NumCgoCall()
	m.stats.NumGC = memStats.NumGC
	m.stats.PauseTotalNs = memStats.PauseTotalNs
	m.stats.LastGC = time.Unix(0, int64(memStats.LastGC))
	m.stats.HeapAlloc = memStats.HeapAlloc
	m.stats.HeapSys = memStats.HeapSys
	m.stats.HeapIdle = memStats.HeapIdle
	m.stats.HeapInuse = memStats.HeapInuse
	m.stats.HeapReleased = memStats.HeapReleased
	m.stats.HeapObjects = memStats.HeapObjects
	m.stats.StackInuse = memStats.StackInuse
	m.stats.StackSys = memStats.StackSys
	m.stats.MSpanInuse = memStats.MSpanInuse
	m.stats.MSpanSys = memStats.MSpanSys
	m.stats.MCacheInuse = memStats.MCacheInuse
	m.stats.MCacheSys = memStats.MCacheSys
	m.stats.BuckHashSys = memStats.BuckHashSys
	m.stats.GCSys = memStats.GCSys
	m.stats.OtherSys = memStats.OtherSys
	m.stats.NextGC = memStats.NextGC
	m.stats.GCCPUFraction = memStats.GCCPUFraction
	m.stats.EnableGC = memStats.EnableGC
	m.stats.DebugGC = memStats.DebugGC
	m.stats.BySize = memStats.BySize

	// 合并执行器统计
	executorStats := m.executor.GetStats()
	if m.stats.CustomStats == nil {
		m.stats.CustomStats = make(map[string]interface{})
	}
	m.stats.CustomStats["executor_total_goroutines"] = executorStats.TotalGoroutines
	m.stats.CustomStats["executor_active_goroutines"] = executorStats.ActiveGoroutines
	m.stats.CustomStats["executor_panic_count"] = executorStats.PanicCount
	m.stats.CustomStats["executor_completed_count"] = executorStats.CompletedCount
	m.stats.CustomStats["executor_last_panic_time"] = executorStats.LastPanicTime
	m.stats.CustomStats["executor_last_panic_goroutine"] = executorStats.LastPanicGoroutine
}

// checkAlerts 检查告警
func (m *Monitor) checkAlerts() {
	m.mu.RLock()
	config := m.alertConfig
	stats := m.stats
	m.mu.RUnlock()

	if config.AlertHandler == nil {
		return
	}

	// 检查协程数告警
	if stats.NumGoroutines > config.MaxGoroutines {
		alert := &Alert{
			Type:      "high_goroutines",
			Message:   "协程数量超过阈值",
			Value:     stats.NumGoroutines,
			Threshold: config.MaxGoroutines,
			Timestamp: time.Now(),
			Severity:  "high",
		}
		config.AlertHandler(alert)
	}

	// 检查堆内存告警
	if stats.HeapAlloc > config.MaxHeapAlloc {
		alert := &Alert{
			Type:      "high_heap_alloc",
			Message:   "堆内存使用超过阈值",
			Value:     stats.HeapAlloc,
			Threshold: config.MaxHeapAlloc,
			Timestamp: time.Now(),
			Severity:  "high",
		}
		config.AlertHandler(alert)
	}

	// 检查栈内存告警
	if stats.StackInuse > config.MaxStackInuse {
		alert := &Alert{
			Type:      "high_stack_inuse",
			Message:   "栈内存使用超过阈值",
			Value:     stats.StackInuse,
			Threshold: config.MaxStackInuse,
			Timestamp: time.Now(),
			Severity:  "medium",
		}
		config.AlertHandler(alert)
	}
}

// notifyHandlers 通知处理器
func (m *Monitor) notifyHandlers() {
	m.mu.RLock()
	handlers := make([]MonitorHandler, len(m.handlers))
	copy(handlers, m.handlers)
	stats := m.stats
	m.mu.RUnlock()

	for _, handler := range handlers {
		handler(stats)
	}
}

// GetStats 获取当前统计信息
func (m *Monitor) GetStats() *MonitorStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本
	stats := *m.stats
	return &stats
}

// GetGoroutineProfile 获取协程profile
func (m *Monitor) GetGoroutineProfile() []runtime.StackRecord {
	prof := make([]runtime.StackRecord, 1000)
	n, ok := runtime.GoroutineProfile(prof)
	if !ok {
		// 如果不够，重新分配更大的切片
		prof = make([]runtime.StackRecord, n)
		runtime.GoroutineProfile(prof)
	}
	return prof[:n]
}

// GetMemProfile 获取内存profile
func (m *Monitor) GetMemProfile() []runtime.MemProfileRecord {
	prof := make([]runtime.MemProfileRecord, 1000)
	n, ok := runtime.MemProfile(prof, true)
	if !ok {
		// 如果不够，重新分配更大的切片
		prof = make([]runtime.MemProfileRecord, n)
		runtime.MemProfile(prof, true)
	}
	return prof[:n]
}

// ForceGC 强制垃圾回收
func (m *Monitor) ForceGC() {
	runtime.GC()
}

// SetCustomStat 设置自定义统计
func (m *Monitor) SetCustomStat(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stats.CustomStats == nil {
		m.stats.CustomStats = make(map[string]interface{})
	}
	m.stats.CustomStats[key] = value
}

// GetCustomStat 获取自定义统计
func (m *Monitor) GetCustomStat(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.stats.CustomStats == nil {
		return nil, false
	}
	value, ok := m.stats.CustomStats[key]
	return value, ok
}
