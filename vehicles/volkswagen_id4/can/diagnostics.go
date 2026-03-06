package can

import (
	"fmt"
	"time"
)

/*
========================================
CAN总线诊断工具
========================================

用于检测和分析丢帧问题
*/

// DropFrameReason 丢帧原因
type DropFrameReason int

const (
	ReasonUnknown DropFrameReason = iota
	ReasonSlowConsumer                // 消费者处理慢
	ReasonBurstTraffic                // 突发流量
	ReasonDeadlockDetected            // 检测到死锁
	ReasonHighPriorityStarvation      // 高优先级饥饿
)

// DiagnosticInfo 诊断信息
type DiagnosticInfo struct {
	Timestamp          time.Time
	TotalFrames        uint64
	DroppedFrames      uint64
	DropRate           float64       // 丢帧率 %
	BufferUsage        float64       // 缓冲区使用率 %
	ConsumptionRate    float64       // 消费速率 (帧/秒)
	ProductionRate     float64       // 生产速率 (帧/秒)
	SlowestListener    string        // 最慢的监听器
	Recommendation     string        // 建议措施
	Reason             DropFrameReason
}

// Diagnose 诊断CAN总线状态
func (bus *CANBus) Diagnose() DiagnosticInfo {
	stats := bus.GetStats()

	// 计算丢帧率
	dropRate := 0.0
	if stats.TotalFrames > 0 {
		dropRate = float64(stats.DroppedFrames) / float64(stats.TotalFrames) * 100
	}

	// 估算缓冲区使用率（通过Channel长度）
	bufferUsage := float64(len(bus.allFrames)) / 1000.0 * 100

	info := DiagnosticInfo{
		Timestamp:     time.Now(),
		TotalFrames:   stats.TotalFrames,
		DroppedFrames: stats.DroppedFrames,
		DropRate:      dropRate,
		BufferUsage:   bufferUsage,
	}

	// 分析原因并给出建议
	info.analyzeAndRecommend()

	return info
}

// analyzeAndRecommend 分析原因并给出建议
func (info *DiagnosticInfo) analyzeAndRecommend() {
	switch {
	case info.DropRate > 10.0:
		// 严重丢帧（>10%）
		info.Reason = ReasonSlowConsumer
		info.Recommendation = "严重丢帧！建议: 1) 优化监听器代码 2) 增大缓冲区到10000 3) 降低ECU发送频率"

	case info.DropRate > 1.0:
		// 中度丢帧（1-10%）
		info.Reason = ReasonBurstTraffic
		info.Recommendation = "中度丢帧。建议: 1) 检查是否有突发流量 2) 考虑增大缓冲区到5000"

	case info.DropRate > 0.1:
		// 轻微丢帧（0.1-1%）
		info.Reason = ReasonUnknown
		info.Recommendation = "轻微丢帧，可接受范围。建议: 监控是否恶化"

	case info.BufferUsage > 80.0:
		// 缓冲区使用率高
		info.Reason = ReasonSlowConsumer
		info.Recommendation = "缓冲区使用率过高！建议: 预防性增大缓冲区或优化监听器"

	default:
		info.Recommendation = "系统运行正常 ✓"
	}
}

// PrintDiagnostic 打印诊断报告
func (info DiagnosticInfo) Print() {
	fmt.Println("\n━━━━━━━━━━━━━━ CAN总线诊断报告 ━━━━━━━━━━━━━━")
	fmt.Printf("⏰ 时间:         %s\n", info.Timestamp.Format("15:04:05.000"))
	fmt.Printf("📊 总帧数:       %d\n", info.TotalFrames)
	fmt.Printf("❌ 丢帧数:       %d\n", info.DroppedFrames)
	fmt.Printf("📉 丢帧率:       %.2f%%\n", info.DropRate)
	fmt.Printf("📦 缓冲区使用率: %.1f%%\n", info.BufferUsage)

	// 健康状态
	if info.DropRate > 5.0 {
		fmt.Println("🔴 健康状态:     严重异常")
	} else if info.DropRate > 1.0 {
		fmt.Println("🟡 健康状态:     警告")
	} else if info.DropRate > 0 {
		fmt.Println("🟢 健康状态:     良好（有轻微丢帧）")
	} else {
		fmt.Println("🟢 健康状态:     完美")
	}

	fmt.Printf("💡 建议措施:     %s\n", info.Recommendation)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
