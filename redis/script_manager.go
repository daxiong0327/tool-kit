package redis

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ScriptManager 脚本管理器
type ScriptManager struct {
	script      *Script
	scripts     map[string]*ScriptInfo
	stats       map[string]*ScriptStats
	mutex       sync.RWMutex
	monitor     *ScriptMonitor
	fileWatcher *ScriptFileWatcher
}

// ScriptStats 脚本统计信息
type ScriptStats struct {
	Name          string        `json:"name"`           // 脚本名称
	Executions    int64         `json:"executions"`     // 执行次数
	TotalTime     time.Duration `json:"total_time"`     // 总执行时间
	AverageTime   time.Duration `json:"average_time"`   // 平均执行时间
	MaxTime       time.Duration `json:"max_time"`       // 最大执行时间
	MinTime       time.Duration `json:"min_time"`       // 最小执行时间
	Errors        int64         `json:"errors"`         // 错误次数
	LastExecution time.Time     `json:"last_execution"` // 最后执行时间
	LastError     time.Time     `json:"last_error"`     // 最后错误时间
	LastErrorMsg  string        `json:"last_error_msg"` // 最后错误信息
	SuccessRate   float64       `json:"success_rate"`   // 成功率
	CacheHits     int64         `json:"cache_hits"`     // 缓存命中次数
	CacheMisses   int64         `json:"cache_misses"`   // 缓存未命中次数
	Reloads       int64         `json:"reloads"`        // 重新加载次数
}

// ScriptMonitor 脚本监控器
type ScriptMonitor struct {
	manager    *ScriptManager
	interval   time.Duration
	stopChan   chan struct{}
	alertRules []*AlertRule
	mutex      sync.RWMutex
}

// AlertRule 告警规则
type AlertRule struct {
	Name        string        `json:"name"`         // 规则名称
	ScriptName  string        `json:"script_name"`  // 脚本名称
	Condition   string        `json:"condition"`    // 条件表达式
	Threshold   float64       `json:"threshold"`    // 阈值
	Duration    time.Duration `json:"duration"`     // 持续时间
	Enabled     bool          `json:"enabled"`      // 是否启用
	LastTrigger time.Time     `json:"last_trigger"` // 最后触发时间
}

// ScriptFileWatcher 脚本文件监控器
type ScriptFileWatcher struct {
	manager    *ScriptManager
	watchPaths []string
	interval   time.Duration
	stopChan   chan struct{}
	mutex      sync.RWMutex
}

// NewScriptManager 创建脚本管理器
func (s *Script) NewScriptManager() *ScriptManager {
	manager := &ScriptManager{
		script:  s,
		scripts: make(map[string]*ScriptInfo),
		stats:   make(map[string]*ScriptStats),
	}

	// 创建监控器
	manager.monitor = &ScriptMonitor{
		manager:  manager,
		interval: 30 * time.Second,
		stopChan: make(chan struct{}),
	}

	// 创建文件监控器
	manager.fileWatcher = &ScriptFileWatcher{
		manager:    manager,
		watchPaths: []string{},
		interval:   10 * time.Second,
		stopChan:   make(chan struct{}),
	}

	return manager
}

// RegisterScript 注册脚本
func (sm *ScriptManager) RegisterScript(ctx context.Context, info *ScriptInfo) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 注册脚本
	if err := sm.script.Register(ctx, info); err != nil {
		return err
	}

	// 保存脚本信息
	sm.scripts[info.Name] = info

	// 初始化统计信息
	sm.stats[info.Name] = &ScriptStats{
		Name:          info.Name,
		MinTime:       time.Hour, // 初始化为很大的值
		LastExecution: time.Now(),
	}

	return nil
}

// ExecuteScript 执行脚本
func (sm *ScriptManager) ExecuteScript(ctx context.Context, name string, opts *ScriptOptions) (*ScriptResult, error) {
	start := time.Now()

	// 执行脚本
	result, err := sm.script.Execute(ctx, name, opts)

	// 更新统计信息
	sm.updateStats(name, result, err, time.Since(start))

	return result, err
}

// GetScriptInfo 获取脚本信息
func (sm *ScriptManager) GetScriptInfo(name string) (*ScriptInfo, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	info, exists := sm.scripts[name]
	return info, exists
}

// GetScriptStats 获取脚本统计信息
func (sm *ScriptManager) GetScriptStats(name string) (*ScriptStats, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats, exists := sm.stats[name]
	return stats, exists
}

// GetAllScriptStats 获取所有脚本统计信息
func (sm *ScriptManager) GetAllScriptStats() map[string]*ScriptStats {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats := make(map[string]*ScriptStats)
	for k, v := range sm.stats {
		stats[k] = v
	}
	return stats
}

// ResetStats 重置统计信息
func (sm *ScriptManager) ResetStats(name string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if stats, exists := sm.stats[name]; exists {
		stats.Executions = 0
		stats.TotalTime = 0
		stats.AverageTime = 0
		stats.MaxTime = 0
		stats.MinTime = time.Hour
		stats.Errors = 0
		stats.SuccessRate = 0
		stats.CacheHits = 0
		stats.CacheMisses = 0
		stats.Reloads = 0
	}
}

// StartMonitor 启动监控
func (sm *ScriptManager) StartMonitor() {
	go sm.monitor.start()
}

// StopMonitor 停止监控
func (sm *ScriptManager) StopMonitor() {
	sm.monitor.stop()
}

// StartFileWatcher 启动文件监控
func (sm *ScriptManager) StartFileWatcher(watchPaths []string) {
	sm.fileWatcher.setWatchPaths(watchPaths)
	go sm.fileWatcher.start()
}

// StopFileWatcher 停止文件监控
func (sm *ScriptManager) StopFileWatcher() {
	sm.fileWatcher.stop()
}

// AddAlertRule 添加告警规则
func (sm *ScriptManager) AddAlertRule(rule *AlertRule) {
	sm.monitor.addAlertRule(rule)
}

// RemoveAlertRule 移除告警规则
func (sm *ScriptManager) RemoveAlertRule(name string) {
	sm.monitor.removeAlertRule(name)
}

// GetAlertRules 获取告警规则
func (sm *ScriptManager) GetAlertRules() []*AlertRule {
	return sm.monitor.getAlertRules()
}

// 私有方法

// updateStats 更新统计信息
func (sm *ScriptManager) updateStats(name string, result *ScriptResult, err error, duration time.Duration) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	stats, exists := sm.stats[name]
	if !exists {
		stats = &ScriptStats{
			Name:          name,
			MinTime:       time.Hour,
			LastExecution: time.Now(),
		}
		sm.stats[name] = stats
	}

	// 更新基本统计
	stats.Executions++
	stats.TotalTime += duration
	stats.AverageTime = stats.TotalTime / time.Duration(stats.Executions)
	stats.LastExecution = time.Now()

	// 更新时间统计
	if duration > stats.MaxTime {
		stats.MaxTime = duration
	}
	if duration < stats.MinTime {
		stats.MinTime = duration
	}

	// 更新错误统计
	if err != nil {
		stats.Errors++
		stats.LastError = time.Now()
		stats.LastErrorMsg = err.Error()
	}

	// 计算成功率
	if stats.Executions > 0 {
		stats.SuccessRate = float64(stats.Executions-stats.Errors) / float64(stats.Executions) * 100
	}

	// 更新缓存统计
	if result != nil && result.SHA != "" {
		stats.CacheHits++
	} else {
		stats.CacheMisses++
	}
}

// ScriptMonitor 方法

// start 启动监控
func (sm *ScriptMonitor) start() {
	ticker := time.NewTicker(sm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.checkAlerts()
		case <-sm.stopChan:
			return
		}
	}
}

// stop 停止监控
func (sm *ScriptMonitor) stop() {
	close(sm.stopChan)
}

// addAlertRule 添加告警规则
func (sm *ScriptMonitor) addAlertRule(rule *AlertRule) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.alertRules = append(sm.alertRules, rule)
}

// removeAlertRule 移除告警规则
func (sm *ScriptMonitor) removeAlertRule(name string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for i, rule := range sm.alertRules {
		if rule.Name == name {
			sm.alertRules = append(sm.alertRules[:i], sm.alertRules[i+1:]...)
			break
		}
	}
}

// getAlertRules 获取告警规则
func (sm *ScriptMonitor) getAlertRules() []*AlertRule {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	rules := make([]*AlertRule, len(sm.alertRules))
	copy(rules, sm.alertRules)
	return rules
}

// checkAlerts 检查告警
func (sm *ScriptMonitor) checkAlerts() {
	sm.mutex.RLock()
	rules := make([]*AlertRule, len(sm.alertRules))
	copy(rules, sm.alertRules)
	sm.mutex.RUnlock()

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		stats, exists := sm.manager.GetScriptStats(rule.ScriptName)
		if !exists {
			continue
		}

		// 检查告警条件
		if sm.evaluateCondition(rule, stats) {
			sm.triggerAlert(rule, stats)
		}
	}
}

// evaluateCondition 评估告警条件
func (sm *ScriptMonitor) evaluateCondition(rule *AlertRule, stats *ScriptStats) bool {
	switch rule.Condition {
	case "error_rate":
		return stats.SuccessRate < rule.Threshold
	case "execution_time":
		return float64(stats.AverageTime) > rule.Threshold
	case "execution_count":
		return float64(stats.Executions) > rule.Threshold
	case "cache_miss_rate":
		total := stats.CacheHits + stats.CacheMisses
		if total == 0 {
			return false
		}
		missRate := float64(stats.CacheMisses) / float64(total) * 100
		return missRate > rule.Threshold
	default:
		return false
	}
}

// triggerAlert 触发告警
func (sm *ScriptMonitor) triggerAlert(rule *AlertRule, stats *ScriptStats) {
	// 检查是否在冷却期内
	if time.Since(rule.LastTrigger) < rule.Duration {
		return
	}

	// 触发告警
	rule.LastTrigger = time.Now()

	// 这里可以实现告警通知逻辑
	// 例如：发送邮件、短信、Slack通知等
	fmt.Printf("ALERT: %s - Script: %s, Condition: %s, Threshold: %.2f\n",
		rule.Name, rule.ScriptName, rule.Condition, rule.Threshold)
}

// ScriptFileWatcher 方法

// setWatchPaths 设置监控路径
func (sfw *ScriptFileWatcher) setWatchPaths(paths []string) {
	sfw.mutex.Lock()
	defer sfw.mutex.Unlock()
	sfw.watchPaths = paths
}

// start 启动文件监控
func (sfw *ScriptFileWatcher) start() {
	ticker := time.NewTicker(sfw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sfw.checkFiles()
		case <-sfw.stopChan:
			return
		}
	}
}

// stop 停止文件监控
func (sfw *ScriptFileWatcher) stop() {
	close(sfw.stopChan)
}

// checkFiles 检查文件变化
func (sfw *ScriptFileWatcher) checkFiles() {
	sfw.mutex.RLock()
	paths := make([]string, len(sfw.watchPaths))
	copy(paths, sfw.watchPaths)
	sfw.mutex.RUnlock()

	// 这里可以实现文件变化检测逻辑
	// 例如：检查文件修改时间、内容哈希等
	// 当检测到变化时，重新加载脚本
	for _, path := range paths {
		// 实现文件监控逻辑
		_ = path
	}
}
