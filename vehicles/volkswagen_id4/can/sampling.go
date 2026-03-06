package can

import (
	"fmt"
	"sync/atomic"
)

/*
========================================
采样降级策略
========================================

在高负载时只处理部分帧（降采样）
*/

// SamplingStrategy 采样策略
type SamplingStrategy int

const (
	SampleAll    SamplingStrategy = iota // 处理所有帧
	SampleHalf                            // 处理50%
	SampleQuarter                         // 处理25%
	SampleTenth                           // 处理10%
)

// SamplingCANBus 支持采样的CAN总线
type SamplingCANBus struct {
	*CANBus
	samplingStrategy atomic.Int32 // 当前采样策略
	frameCounter     atomic.Uint64
	sampledFrames    atomic.Uint64
	skippedFrames    atomic.Uint64
}

// NewSamplingCANBus 创建采样总线
func NewSamplingCANBus(name string) *SamplingCANBus {
	sbus := &SamplingCANBus{
		CANBus: NewCANBus(name),
	}
	sbus.samplingStrategy.Store(int32(SampleAll))
	return sbus
}

// SetSamplingStrategy 设置采样策略
func (sbus *SamplingCANBus) SetSamplingStrategy(strategy SamplingStrategy) {
	oldStrategy := SamplingStrategy(sbus.samplingStrategy.Load())
	sbus.samplingStrategy.Store(int32(strategy))

	strategyNames := map[SamplingStrategy]string{
		SampleAll:     "100%",
		SampleHalf:    "50%",
		SampleQuarter: "25%",
		SampleTenth:   "10%",
	}

	fmt.Printf("🔧 采样策略已调整: %s -> %s\n",
		strategyNames[oldStrategy], strategyNames[strategy])
}

// shouldSample 判断是否应该采样此帧
func (sbus *SamplingCANBus) shouldSample() bool {
	count := sbus.frameCounter.Add(1)
	strategy := SamplingStrategy(sbus.samplingStrategy.Load())

	switch strategy {
	case SampleAll:
		return true
	case SampleHalf:
		return count%2 == 0
	case SampleQuarter:
		return count%4 == 0
	case SampleTenth:
		return count%10 == 0
	default:
		return true
	}
}

// processFramesWithSampling 带采样的帧处理
func (sbus *SamplingCANBus) processFramesWithSampling() {
	for frame := range sbus.allFrames {
		if !sbus.shouldSample() {
			// 跳过此帧
			sbus.skippedFrames.Add(1)
			continue
		}

		// 处理此帧
		sbus.sampledFrames.Add(1)

		sbus.mu.RLock()
		listeners, exists := sbus.listeners[frame.ID]
		if exists {
			for _, listener := range listeners {
				go listener.OnCANFrame(frame)
			}
		}
		sbus.mu.RUnlock()
	}
}

// Start 启动（覆盖父类）
func (sbus *SamplingCANBus) Start() {
	sbus.running = true
	go sbus.processFramesWithSampling()
	fmt.Printf("📡 采样CAN总线 [%s] 已启动\n", sbus.name)
}

// PrintSamplingStats 打印采样统计
func (sbus *SamplingCANBus) PrintSamplingStats() {
	sampled := sbus.sampledFrames.Load()
	skipped := sbus.skippedFrames.Load()
	total := sampled + skipped

	samplingRate := 0.0
	if total > 0 {
		samplingRate = float64(sampled) / float64(total) * 100
	}

	fmt.Println("\n━━━━━━━━━━━━━━ 采样统计 ━━━━━━━━━━━━━━")
	fmt.Printf("📊 总帧数:     %d\n", total)
	fmt.Printf("✅ 已处理:     %d\n", sampled)
	fmt.Printf("⏭️  已跳过:     %d\n", skipped)
	fmt.Printf("📉 采样率:     %.1f%%\n", samplingRate)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// AutoAdjustSampling 根据负载自动调整采样率
func (sbus *SamplingCANBus) AutoAdjustSampling() {
	stats := sbus.GetStats()

	dropRate := 0.0
	if stats.TotalFrames > 0 {
		dropRate = float64(stats.DroppedFrames) / float64(stats.TotalFrames) * 100
	}

	bufferUsage := float64(len(sbus.allFrames)) / float64(cap(sbus.allFrames)) * 100

	currentStrategy := SamplingStrategy(sbus.samplingStrategy.Load())

	switch {
	case dropRate > 10 || bufferUsage > 95:
		// 严重负载：降到10%
		if currentStrategy != SampleTenth {
			sbus.SetSamplingStrategy(SampleTenth)
		}

	case dropRate > 5 || bufferUsage > 85:
		// 高负载：降到25%
		if currentStrategy != SampleQuarter {
			sbus.SetSamplingStrategy(SampleQuarter)
		}

	case dropRate > 1 || bufferUsage > 70:
		// 中等负载：降到50%
		if currentStrategy != SampleHalf {
			sbus.SetSamplingStrategy(SampleHalf)
		}

	case dropRate == 0 && bufferUsage < 50:
		// 正常负载：恢复100%
		if currentStrategy != SampleAll {
			sbus.SetSamplingStrategy(SampleAll)
		}
	}
}

// StartAutoSampling 启动自动采样调整
func (sbus *SamplingCANBus) StartAutoSampling() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for sbus.running {
			<-ticker.C
			sbus.AutoAdjustSampling()
		}
	}()
}
