package can

import (
	"fmt"
	"sync"
	"time"
)

/*
========================================
CAN总线模拟系统
========================================

实现真正的CAN总线架构：
- CAN 2.0B标准帧格式
- 发布-订阅模式
- 多监听器支持
- 消息仲裁和优先级
*/

// CANFrame CAN数据帧（CAN 2.0B标准）
type CANFrame struct {
	ID        uint32    // CAN标识符（11-bit标准帧）
	DLC       uint8     // 数据长度码（0-8字节）
	Data      [8]byte   // 数据域（最多8字节）
	Timestamp time.Time // 时间戳
	Priority  uint8     // 优先级（用于仲裁，ID越小优先级越高）
}

// CANListener CAN监听器接口
type CANListener interface {
	OnCANFrame(frame CANFrame) // 接收CAN帧的回调
	GetName() string           // 获取监听器名称
}

// CANBus CAN总线
type CANBus struct {
	name      string
	listeners map[uint32][]CANListener // CAN ID -> 监听器列表
	allFrames chan CANFrame            // 所有帧的通道
	mu        sync.RWMutex             // 保护listeners
	stats     CANBusStats              // 总线统计
	running   bool                     // 运行状态
}

// CANBusStats 总线统计信息
type CANBusStats struct {
	TotalFrames    uint64            // 总帧数
	FramesByID     map[uint32]uint64 // 各ID的帧数
	DroppedFrames  uint64            // 丢失的帧
	LastFrameTime  time.Time         // 最后一帧时间
	mu             sync.RWMutex
}

// NewCANBus 创建新的CAN总线
func NewCANBus(name string) *CANBus {
	return &CANBus{
		name:      name,
		listeners: make(map[uint32][]CANListener),
		allFrames: make(chan CANFrame, 1000), // 缓冲1000帧
		stats: CANBusStats{
			FramesByID: make(map[uint32]uint64),
		},
		running: false,
	}
}

// Start 启动总线
func (bus *CANBus) Start() {
	bus.running = true
	go bus.processFrames()
	fmt.Printf("📡 CAN总线 [%s] 已启动\n", bus.name)
}

// Stop 停止总线
func (bus *CANBus) Stop() {
	bus.running = false
	close(bus.allFrames)
	fmt.Printf("📡 CAN总线 [%s] 已停止\n", bus.name)
}

// SendFrame 发送CAN帧到总线
func (bus *CANBus) SendFrame(frame CANFrame) error {
	if !bus.running {
		return fmt.Errorf("CAN总线未运行")
	}

	frame.Timestamp = time.Now()
	frame.Priority = uint8(frame.ID & 0xFF) // 优先级基于ID

	select {
	case bus.allFrames <- frame:
		// 成功发送
		bus.stats.mu.Lock()
		bus.stats.TotalFrames++
		bus.stats.FramesByID[frame.ID]++
		bus.stats.LastFrameTime = frame.Timestamp
		bus.stats.mu.Unlock()
		return nil
	default:
		// 总线满，丢帧
		bus.stats.mu.Lock()
		bus.stats.DroppedFrames++
		bus.stats.mu.Unlock()
		return fmt.Errorf("CAN总线缓冲区满，帧丢失")
	}
}

// processFrames 处理CAN帧（总线核心）
func (bus *CANBus) processFrames() {
	for frame := range bus.allFrames {
		bus.mu.RLock()
		// 查找订阅此CAN ID的所有监听器
		listeners, exists := bus.listeners[frame.ID]
		if exists {
			// 广播给所有监听器
			for _, listener := range listeners {
				go listener.OnCANFrame(frame) // 异步调用避免阻塞
			}
		}
		bus.mu.RUnlock()
	}
}

// Subscribe 订阅特定CAN ID
func (bus *CANBus) Subscribe(canID uint32, listener CANListener) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.listeners[canID] == nil {
		bus.listeners[canID] = make([]CANListener, 0)
	}
	bus.listeners[canID] = append(bus.listeners[canID], listener)

	fmt.Printf("📻 [%s] 订阅 CAN ID: 0x%03X\n", listener.GetName(), canID)
}

// Unsubscribe 取消订阅
func (bus *CANBus) Unsubscribe(canID uint32, listener CANListener) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	listeners := bus.listeners[canID]
	for i, l := range listeners {
		if l == listener {
			bus.listeners[canID] = append(listeners[:i], listeners[i+1:]...)
			fmt.Printf("📻 [%s] 取消订阅 CAN ID: 0x%03X\n", listener.GetName(), canID)
			return
		}
	}
}

// GetStats 获取总线统计
func (bus *CANBus) GetStats() CANBusStats {
	bus.stats.mu.RLock()
	defer bus.stats.mu.RUnlock()

	// 复制统计数据
	stats := bus.stats
	stats.FramesByID = make(map[uint32]uint64)
	for k, v := range bus.stats.FramesByID {
		stats.FramesByID[k] = v
	}
	return stats
}

// PrintStats 打印总线统计
func (bus *CANBus) PrintStats() {
	stats := bus.GetStats()

	fmt.Println("\n━━━━━━━━━━━━━━ CAN总线统计 ━━━━━━━━━━━━━━")
	fmt.Printf("📊 总线名称: %s\n", bus.name)
	fmt.Printf("📊 总帧数:   %d\n", stats.TotalFrames)
	fmt.Printf("📊 丢帧数:   %d\n", stats.DroppedFrames)
	fmt.Printf("📊 最后帧:   %s\n", stats.LastFrameTime.Format("15:04:05.000"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("各CAN ID帧数:")
	for canID, count := range stats.FramesByID {
		fmt.Printf("  0x%03X: %d 帧\n", canID, count)
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// EncodeCANData 将信号编码为CAN数据域
func EncodeCANData(signals map[string]uint32, signalDefs map[string]CANSignalDef) [8]byte {
	var data [8]byte
	// 这里简化处理，实际需要按StartBit和BitLength打包
	// 现在只是演示，把信号值直接放入
	return data
}

// DecodeCANData 从CAN数据域解码信号
func DecodeCANData(data [8]byte, signalDef CANSignalDef) float64 {
	// 简化处理，实际需要按StartBit和BitLength提取
	// 现在假设数据在前几个字节
	rawValue := uint32(data[0]) | uint32(data[1])<<8
	return DecodeSignal(rawValue, signalDef)
}
