package can

import (
	"fmt"
	"time"
)

/*
========================================
自适应发送频率控制
========================================

根据总线负载动态调整ECU发送频率
*/

// AdaptiveECU 自适应ECU（可动态调整发送频率）
type AdaptiveECU struct {
	*VehicleECU
	minSendRate     time.Duration // 最快频率
	maxSendRate     time.Duration // 最慢频率
	currentSendRate time.Duration // 当前频率
	bus             *CANBus
}

// NewAdaptiveECU 创建自适应ECU
func NewAdaptiveECU(name string, bus *CANBus, initialRate time.Duration) *AdaptiveECU {
	return &AdaptiveECU{
		VehicleECU:      NewVehicleECU(name, bus, initialRate),
		minSendRate:     50 * time.Millisecond,  // 最快20Hz
		maxSendRate:     500 * time.Millisecond, // 最慢2Hz
		currentSendRate: initialRate,
		bus:             bus,
	}
}

// AdjustSendRate 根据总线状态调整发送频率
func (aecu *AdaptiveECU) AdjustSendRate() {
	stats := aecu.bus.GetStats()

	// 计算丢帧率
	dropRate := 0.0
	if stats.TotalFrames > 0 {
		dropRate = float64(stats.DroppedFrames) / float64(stats.TotalFrames) * 100
	}

	oldRate := aecu.currentSendRate

	switch {
	case dropRate > 5.0:
		// 严重丢帧：降低到最慢频率
		aecu.currentSendRate = aecu.maxSendRate
		fmt.Printf("⚠️  严重丢帧(%.1f%%)，降低发送频率: %v -> %v\n",
			dropRate, oldRate, aecu.currentSendRate)

	case dropRate > 1.0:
		// 中度丢帧：降低50%
		newRate := aecu.currentSendRate * 3 / 2
		if newRate > aecu.maxSendRate {
			newRate = aecu.maxSendRate
		}
		aecu.currentSendRate = newRate
		fmt.Printf("⚠️  中度丢帧(%.1f%%)，降低发送频率: %v -> %v\n",
			dropRate, oldRate, aecu.currentSendRate)

	case dropRate == 0 && aecu.currentSendRate > aecu.minSendRate:
		// 无丢帧：尝试提高频率
		newRate := aecu.currentSendRate * 4 / 5
		if newRate < aecu.minSendRate {
			newRate = aecu.minSendRate
		}
		aecu.currentSendRate = newRate
		fmt.Printf("✅ 无丢帧，提高发送频率: %v -> %v\n",
			oldRate, aecu.currentSendRate)
	}

	// 更新ECU的实际发送频率
	aecu.VehicleECU.sendRate = aecu.currentSendRate
}

// StartAutoAdjust 启动自动调整
func (aecu *AdaptiveECU) StartAutoAdjust() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for aecu.running {
			<-ticker.C
			aecu.AdjustSendRate()
		}
	}()
}
