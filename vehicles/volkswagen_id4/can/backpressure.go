package can

import (
	"fmt"
	"sync/atomic"
	"time"
)

/*
========================================
背压(Backpressure)机制
========================================

当消费者处理不过来时，通知生产者降速
*/

// BackpressureSignal 背压信号
type BackpressureSignal struct {
	Level      int       // 背压等级 (0=正常, 1=警告, 2=严重)
	BufferUsage float64   // 缓冲区使用率
	DropRate   float64   // 丢帧率
	Timestamp  time.Time
}

// BackpressureManager 背压管理器
type BackpressureManager struct {
	signalChan chan BackpressureSignal
	level      atomic.Int32 // 当前背压等级
}

// NewBackpressureManager 创建背压管理器
func NewBackpressureManager() *BackpressureManager {
	return &BackpressureManager{
		signalChan: make(chan BackpressureSignal, 10),
	}
}

// GetSignalChannel 获取背压信号通道
func (bpm *BackpressureManager) GetSignalChannel() <-chan BackpressureSignal {
	return bpm.signalChan
}

// SendSignal 发送背压信号
func (bpm *BackpressureManager) SendSignal(signal BackpressureSignal) {
	select {
	case bpm.signalChan <- signal:
		bpm.level.Store(int32(signal.Level))
	default:
		// 信号通道满了，跳过
	}
}

// GetLevel 获取当前背压等级
func (bpm *BackpressureManager) GetLevel() int {
	return int(bpm.level.Load())
}

// BackpressureCANBus 支持背压的CAN总线
type BackpressureCANBus struct {
	*CANBus
	backpressure *BackpressureManager
}

// NewBackpressureCANBus 创建支持背压的总线
func NewBackpressureCANBus(name string) *BackpressureCANBus {
	return &BackpressureCANBus{
		CANBus:       NewCANBus(name),
		backpressure: NewBackpressureManager(),
	}
}

// MonitorBackpressure 监控背压
func (bbus *BackpressureCANBus) MonitorBackpressure() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for bbus.running {
		<-ticker.C

		// 计算指标
		bufferUsage := float64(len(bbus.allFrames)) / float64(cap(bbus.allFrames)) * 100
		stats := bbus.GetStats()

		dropRate := 0.0
		if stats.TotalFrames > 0 {
			dropRate = float64(stats.DroppedFrames) / float64(stats.TotalFrames) * 100
		}

		// 判断背压等级
		var level int
		switch {
		case bufferUsage > 90 || dropRate > 5.0:
			level = 2 // 严重
		case bufferUsage > 70 || dropRate > 1.0:
			level = 1 // 警告
		default:
			level = 0 // 正常
		}

		// 发送背压信号
		signal := BackpressureSignal{
			Level:       level,
			BufferUsage: bufferUsage,
			DropRate:    dropRate,
			Timestamp:   time.Now(),
		}

		bbus.backpressure.SendSignal(signal)

		if level > 0 {
			fmt.Printf("⚠️  背压警告 [等级:%d] 缓冲区:%.1f%% 丢帧率:%.2f%%\n",
				level, bufferUsage, dropRate)
		}
	}
}

// BackpressureECU 响应背压的ECU
type BackpressureECU struct {
	*VehicleECU
	backpressureMgr *BackpressureManager
	normalRate      time.Duration
	reducedRate     time.Duration
}

// NewBackpressureECU 创建响应背压的ECU
func NewBackpressureECU(name string, bus *CANBus, bpm *BackpressureManager) *BackpressureECU {
	return &BackpressureECU{
		VehicleECU:      NewVehicleECU(name, bus, 100*time.Millisecond),
		backpressureMgr: bpm,
		normalRate:      100 * time.Millisecond,
		reducedRate:     300 * time.Millisecond,
	}
}

// ListenToBackpressure 监听背压信号并调整
func (becu *BackpressureECU) ListenToBackpressure() {
	go func() {
		for signal := range becu.backpressureMgr.GetSignalChannel() {
			switch signal.Level {
			case 2: // 严重
				becu.sendRate = becu.reducedRate * 2 // 降到原来的1/6
				fmt.Printf("🚨 ECU响应严重背压: 发送频率降低到 %v\n", becu.sendRate)

			case 1: // 警告
				becu.sendRate = becu.reducedRate // 降到原来的1/3
				fmt.Printf("⚠️  ECU响应背压警告: 发送频率降低到 %v\n", becu.sendRate)

			case 0: // 正常
				becu.sendRate = becu.normalRate
			}
		}
	}()
}
