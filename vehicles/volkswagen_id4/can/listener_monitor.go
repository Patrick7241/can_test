package can

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

/*
========================================
监听器性能监控
========================================

监控每个监听器的处理时间，找出性能瓶颈
*/

// ListenerStats 监听器统计
type ListenerStats struct {
	Name              string
	ProcessedFrames   uint64
	TotalProcessTime  time.Duration
	AvgProcessTime    time.Duration
	MaxProcessTime    time.Duration
	LastProcessTime   time.Duration
	SlowFrameCount    uint64 // 处理超过阈值的帧数
}

// MonitoredListener 被监控的监听器包装
type MonitoredListener struct {
	listener      CANListener
	stats         ListenerStats
	mu            sync.RWMutex
	slowThreshold time.Duration // 慢处理阈值
}

// NewMonitoredListener 创建监控包装
func NewMonitoredListener(listener CANListener, slowThreshold time.Duration) *MonitoredListener {
	return &MonitoredListener{
		listener: listener,
		stats: ListenerStats{
			Name: listener.GetName(),
		},
		slowThreshold: slowThreshold,
	}
}

// OnCANFrame 拦截并监控处理时间
func (ml *MonitoredListener) OnCANFrame(frame CANFrame) {
	startTime := time.Now()

	// 调用实际监听器
	ml.listener.OnCANFrame(frame)

	processingTime := time.Since(startTime)

	// 更新统计
	ml.mu.Lock()
	ml.stats.ProcessedFrames++
	ml.stats.TotalProcessTime += processingTime
	ml.stats.AvgProcessTime = ml.stats.TotalProcessTime / time.Duration(ml.stats.ProcessedFrames)
	ml.stats.LastProcessTime = processingTime

	if processingTime > ml.stats.MaxProcessTime {
		ml.stats.MaxProcessTime = processingTime
	}

	if processingTime > ml.slowThreshold {
		ml.stats.SlowFrameCount++
		fmt.Printf("⚠️  慢监听器: [%s] 处理耗时 %v (CAN ID: 0x%03X)\n",
			ml.stats.Name, processingTime, frame.ID)
	}
	ml.mu.Unlock()
}

func (ml *MonitoredListener) GetName() string {
	return ml.listener.GetName()
}

// GetStats 获取统计数据
func (ml *MonitoredListener) GetStats() ListenerStats {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	return ml.stats
}

// ListenerMonitor 监听器监控管理器
type ListenerMonitor struct {
	monitoredListeners []*MonitoredListener
	mu                 sync.RWMutex
}

// NewListenerMonitor 创建监听器监控器
func NewListenerMonitor() *ListenerMonitor {
	return &ListenerMonitor{
		monitoredListeners: make([]*MonitoredListener, 0),
	}
}

// AddListener 添加被监控的监听器
func (lm *ListenerMonitor) AddListener(ml *MonitoredListener) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.monitoredListeners = append(lm.monitoredListeners, ml)
}

// PrintReport 打印性能报告
func (lm *ListenerMonitor) PrintReport() {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	fmt.Println("\n━━━━━━━━━━━━━━ 监听器性能报告 ━━━━━━━━━━━━━━")

	// 按平均处理时间排序（慢的在前）
	sortedListeners := make([]*MonitoredListener, len(lm.monitoredListeners))
	copy(sortedListeners, lm.monitoredListeners)

	sort.Slice(sortedListeners, func(i, j int) bool {
		return sortedListeners[i].stats.AvgProcessTime > sortedListeners[j].stats.AvgProcessTime
	})

	for i, ml := range sortedListeners {
		stats := ml.GetStats()

		// 健康状态指示器
		healthIcon := "🟢"
		if stats.AvgProcessTime > 10*time.Millisecond {
			healthIcon = "🔴" // 严重
		} else if stats.AvgProcessTime > 1*time.Millisecond {
			healthIcon = "🟡" // 警告
		}

		fmt.Printf("\n%s [#%d] %s\n", healthIcon, i+1, stats.Name)
		fmt.Printf("   处理帧数:     %d\n", stats.ProcessedFrames)
		fmt.Printf("   平均耗时:     %v\n", stats.AvgProcessTime)
		fmt.Printf("   最大耗时:     %v\n", stats.MaxProcessTime)
		fmt.Printf("   最近耗时:     %v\n", stats.LastProcessTime)
		fmt.Printf("   慢帧数:       %d (%.2f%%)\n",
			stats.SlowFrameCount,
			float64(stats.SlowFrameCount)/float64(stats.ProcessedFrames)*100)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 给出优化建议
	if len(sortedListeners) > 0 {
		slowest := sortedListeners[0].GetStats()
		if slowest.AvgProcessTime > 10*time.Millisecond {
			fmt.Printf("\n💡 建议: 优化 [%s]，平均耗时 %v 过高！\n",
				slowest.Name, slowest.AvgProcessTime)
			fmt.Println("   1. 检查是否有阻塞操作（IO、Sleep、锁竞争）")
			fmt.Println("   2. 考虑使用异步处理或缓存")
			fmt.Println("   3. 减少复杂计算")
		}
	}
}

// StartPeriodicReport 定期打印报告
func (lm *ListenerMonitor) StartPeriodicReport(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			lm.PrintReport()
		}
	}()
}
