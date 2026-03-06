package can

import (
	"fmt"
	"time"
)

/*
========================================
组合策略示例
========================================

演示如何组合多种策略处理丢帧问题
*/

// ComboStrategy 组合策略
type ComboStrategy struct {
	bus              *BackpressureCANBus
	ecu              *BackpressureECU
	monitor          *ListenerMonitor
	backpressureMgr  *BackpressureManager
}

// NewComboStrategy 创建组合策略
func NewComboStrategy(name string) *ComboStrategy {
	// 1. 创建支持背压的总线
	bus := NewBackpressureCANBus(name)

	// 2. 创建背压管理器
	bpm := bus.backpressure

	// 3. 创建响应背压的ECU
	ecu := NewBackpressureECU("ComboECU", bus.CANBus, bpm)

	// 4. 创建监听器监控器
	monitor := NewListenerMonitor()

	return &ComboStrategy{
		bus:             bus,
		ecu:             ecu,
		monitor:         monitor,
		backpressureMgr: bpm,
	}
}

// Start 启动组合策略
func (cs *ComboStrategy) Start() {
	// 启动总线
	cs.bus.Start()

	// 启动背压监控
	go cs.bus.MonitorBackpressure()

	// 启动ECU并监听背压
	cs.ecu.Start()
	cs.ecu.ListenToBackpressure()

	// 启动监听器性能报告（每30秒）
	cs.monitor.StartPeriodicReport(30 * time.Second)

	fmt.Println("✅ 组合策略已启动:")
	fmt.Println("   - 背压监控: 自动检测总线压力")
	fmt.Println("   - 动态降频: ECU根据背压调整发送频率")
	fmt.Println("   - 性能监控: 定期报告监听器性能")
}

// AddMonitoredListener 添加被监控的监听器
func (cs *ComboStrategy) AddMonitoredListener(listener CANListener, canIDs []uint32) {
	// 包装为监控监听器（超过5ms算慢）
	ml := NewMonitoredListener(listener, 5*time.Millisecond)

	// 订阅CAN ID
	for _, canID := range canIDs {
		cs.bus.Subscribe(canID, ml)
	}

	// 添加到监控器
	cs.monitor.AddListener(ml)
}

// GetDiagnostic 获取诊断信息
func (cs *ComboStrategy) GetDiagnostic() DiagnosticInfo {
	return cs.bus.Diagnose()
}

// ========================================
// 使用示例
// ========================================

func ExampleUsage() {
	fmt.Println("━━━━━━━━━━━━━━ 丢帧处理策略演示 ━━━━━━━━━━━━━━")

	// 创建组合策略
	strategy := NewComboStrategy("VW-ID4-Adaptive")
	strategy.Start()

	// 添加仪表盘监听器（监控）
	dashboard := NewDashboardListener("智能仪表盘")
	strategy.AddMonitoredListener(dashboard, []uint32{0xFD, 0xB5, 0x3EB, 0x3DA})

	// 添加速度监听器（监控）
	speedListener := NewSpeedListener("速度监控")
	strategy.AddMonitoredListener(speedListener, []uint32{0xFD})

	// 模拟运行
	fmt.Println("\n🚀 系统运行中...")
	time.Sleep(5 * time.Second)

	// 检查诊断信息
	diag := strategy.GetDiagnostic()
	diag.Print()

	// 模拟重负载场景
	fmt.Println("\n⚡ 模拟高负载场景...")
	for i := 0; i < 5000; i++ {
		frame := CANFrame{
			ID:  0xFD,
			DLC: 8,
		}
		strategy.bus.SendFrame(frame)
	}

	time.Sleep(3 * time.Second)

	// 再次检查
	diag = strategy.GetDiagnostic()
	diag.Print()

	fmt.Println("\n✅ 演示完成")
}

// ========================================
// 决策树：选择合适的策略
// ========================================

func ChooseStrategy(dropRate float64, bufferUsage float64) []string {
	recommendations := make([]string, 0)

	// 1. 轻微丢帧（0.1-1%）
	if dropRate > 0.1 && dropRate <= 1.0 {
		recommendations = append(recommendations,
			"✅ 启用监听器性能监控，找出慢监听器")
		return recommendations
	}

	// 2. 中度丢帧（1-5%）
	if dropRate > 1.0 && dropRate <= 5.0 {
		recommendations = append(recommendations,
			"⚠️  优先级1: 启用背压机制",
			"⚠️  优先级2: 优化最慢的监听器",
			"⚠️  优先级3: 考虑增大缓冲区到5000")
		return recommendations
	}

	// 3. 严重丢帧（>5%）
	if dropRate > 5.0 {
		recommendations = append(recommendations,
			"🚨 紧急措施1: 立即启用采样降级（降到25%）",
			"🚨 紧急措施2: ECU降低发送频率50%",
			"🚨 紧急措施3: 使用优先级队列，丢弃低优先级帧",
			"🚨 长期方案: 重构监听器架构，使用异步处理")
		return recommendations
	}

	// 4. 缓冲区使用率高但无丢帧
	if bufferUsage > 80 && dropRate == 0 {
		recommendations = append(recommendations,
			"⚠️  预防性措施: 启用自适应缓冲区扩容",
			"⚠️  预防性措施: 启用背压监控")
		return recommendations
	}

	recommendations = append(recommendations, "✅ 系统运行正常，无需调整")
	return recommendations
}

// PrintStrategyComparison 打印策略对比
func PrintStrategyComparison() {
	fmt.Println("\n━━━━━━━━━━━━━━ 策略对比表 ━━━━━━━━━━━━━━\n")

	strategies := []struct {
		name        string
		complexity  string
		effectiveness string
		latency     string
		memoryUsage string
		useCases    string
	}{
		{"动态扩容缓冲区", "⭐⭐⭐", "⭐⭐⭐⭐⭐", "低", "高", "突发流量"},
		{"动态调整发送频率", "⭐⭐⭐", "⭐⭐⭐⭐", "中", "低", "持续高负载"},
		{"优先级丢帧", "⭐⭐⭐⭐", "⭐⭐⭐⭐⭐", "低", "中", "混合负载"},
		{"背压机制", "⭐⭐⭐⭐", "⭐⭐⭐⭐", "中", "低", "生产-消费不平衡"},
		{"监听器监控", "⭐⭐", "⭐⭐⭐⭐", "低", "低", "性能调优"},
		{"采样降级", "⭐⭐", "⭐⭐⭐", "低", "低", "极端高负载"},
	}

	fmt.Printf("%-20s %-10s %-12s %-8s %-12s %s\n",
		"策略", "复杂度", "有效性", "延迟", "内存占用", "适用场景")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, s := range strategies {
		fmt.Printf("%-20s %-10s %-12s %-8s %-12s %s\n",
			s.name, s.complexity, s.effectiveness, s.latency, s.memoryUsage, s.useCases)
	}

	fmt.Println("\n推荐组合:")
	fmt.Println("  🥇 最佳实践: 背压机制 + 监听器监控 + 动态调频")
	fmt.Println("  🥈 经济方案: 动态扩容 + 监听器监控")
	fmt.Println("  🥉 极简方案: 仅增大缓冲区到10000")
}
