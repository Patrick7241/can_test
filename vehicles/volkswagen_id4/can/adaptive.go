package can

import (
	"fmt"
	"sync"
	"time"
)

/*
========================================
动态缓冲区管理
========================================

自动根据负载调整缓冲区大小
*/

// AdaptiveCANBus 自适应CAN总线（支持动态扩容）
type AdaptiveCANBus struct {
	*CANBus
	maxBufferSize   int
	resizeThreshold float64 // 触发扩容的使用率阈值
	lastResizeTime  time.Time
	resizeCooldown  time.Duration // 扩容冷却时间
	resizeMu        sync.Mutex
}

// NewAdaptiveCANBus 创建自适应总线
func NewAdaptiveCANBus(name string, initialSize, maxSize int) *AdaptiveCANBus {
	return &AdaptiveCANBus{
		CANBus:          NewCANBus(name),
		maxBufferSize:   maxSize,
		resizeThreshold: 0.8, // 80%使用率时扩容
		resizeCooldown:  5 * time.Second,
	}
}

// CheckAndResize 检查并自动扩容
func (abus *AdaptiveCANBus) CheckAndResize() bool {
	abus.resizeMu.Lock()
	defer abus.resizeMu.Unlock()

	// 检查冷却时间
	if time.Since(abus.lastResizeTime) < abus.resizeCooldown {
		return false
	}

	// 计算缓冲区使用率
	currentSize := len(abus.allFrames)
	capacity := cap(abus.allFrames)
	usage := float64(currentSize) / float64(capacity)

	// 触发扩容条件
	if usage > abus.resizeThreshold && capacity < abus.maxBufferSize {
		newCapacity := capacity * 2
		if newCapacity > abus.maxBufferSize {
			newCapacity = abus.maxBufferSize
		}

		// 创建新Channel并迁移数据
		newChannel := make(chan CANFrame, newCapacity)

		// 将旧数据复制到新Channel
		close(abus.allFrames)
		for frame := range abus.allFrames {
			newChannel <- frame
		}

		abus.allFrames = newChannel
		abus.lastResizeTime = time.Now()

		fmt.Printf("🔧 缓冲区已扩容: %d -> %d (使用率: %.1f%%)\n",
			capacity, newCapacity, usage*100)

		// 重启处理协程
		go abus.processFrames()

		return true
	}

	return false
}

// StartAutoResize 启动自动扩容监控
func (abus *AdaptiveCANBus) StartAutoResize() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for abus.running {
			<-ticker.C
			abus.CheckAndResize()
		}
	}()
}
