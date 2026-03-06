package can

import (
	"fmt"
	"sync"
	"time"
)

/*
========================================
CAN监听器实现
========================================

各种监听器示例：
- 速度监听器
- 仪表盘监听器
- 数据记录器
- 诊断监听器
*/

// SpeedListener 车速监听器
type SpeedListener struct {
	name         string
	lastSpeed    float64
	speedChanges uint64
	mu           sync.RWMutex
}

func NewSpeedListener(name string) *SpeedListener {
	return &SpeedListener{
		name: name,
	}
}

func (sl *SpeedListener) GetName() string {
	return sl.name
}

func (sl *SpeedListener) OnCANFrame(frame CANFrame) {
	if frame.ID == 0xFD { // 车速帧
		// 解析车速 (StartBit=32, 即第4-5字节)
		rawValue := uint32(frame.Data[4]) | uint32(frame.Data[5])<<8
		speed := DecodeSignal(rawValue, ID4Signals["Indicated_Vehicle_Speed_kph"])

		sl.mu.Lock()
		if speed != sl.lastSpeed {
			sl.speedChanges++
			sl.lastSpeed = speed
		}
		sl.mu.Unlock()
	}
}

func (sl *SpeedListener) GetSpeed() float64 {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.lastSpeed
}

func (sl *SpeedListener) PrintStats() {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	fmt.Printf("[%s] 当前车速: %.1f km/h, 速度变化次数: %d\n", sl.name, sl.lastSpeed, sl.speedChanges)
}

// DashboardListener 仪表盘监听器（监听所有主要信号）
type DashboardListener struct {
	name   string
	Speed  float64
	Gear   int
	Brake  float64
	Accel  float64
	Steer  float64
	mu     sync.RWMutex
}

func NewDashboardListener(name string) *DashboardListener {
	return &DashboardListener{
		name: name,
	}
}

func (dl *DashboardListener) GetName() string {
	return dl.name
}

func (dl *DashboardListener) OnCANFrame(frame CANFrame) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	switch frame.ID {
	case 0xFD: // 车速
		rawValue := uint32(frame.Data[4]) | uint32(frame.Data[5])<<8
		dl.Speed = DecodeSignal(rawValue, ID4Signals["Indicated_Vehicle_Speed_kph"])

	case 0xB5: // 档位
		rawValue := uint32((frame.Data[6] >> 4) & 0x07)
		dl.Gear = int(rawValue)

	case 0x3EB: // 刹车和油门
		// 油门
		accelRaw := uint32(frame.Data[2])
		dl.Accel = DecodeSignal(accelRaw, ID4Signals["Accelerator_Pedal_Position"])
		// 刹车
		brakeRaw := uint32(frame.Data[6]) | uint32(frame.Data[7]&0x03)<<8
		dl.Brake = DecodeSignal(brakeRaw, ID4Signals["Brake_Position"])

	case 0x3DA: // 转向
		angleRaw := uint32(frame.Data[5]) | uint32(frame.Data[6]&0x1F)<<8
		dl.Steer = DecodeSignal(angleRaw, ID4Signals["Steering_Angle"])
	}
}

func (dl *DashboardListener) Display() {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	gearNames := []string{"P", "R", "N", "D"}
	gearName := "?"
	if dl.Gear >= 0 && dl.Gear < len(gearNames) {
		gearName = gearNames[dl.Gear]
	}
	gearEmoji := map[string]string{"P": "🔴", "R": "🟡", "N": "⚪", "D": "🟢", "?": "⚫"}[gearName]

	fmt.Println("\n━━━━━━━━━━━━━━ 仪表盘监听器 ━━━━━━━━━━━━━━")
	fmt.Printf("🚗 车速:  %6.1f km/h\n", dl.Speed)
	fmt.Printf("⚙️  档位:  %s %s\n", gearEmoji, gearName)
	fmt.Printf("⚡ 油门:  %6.1f %%\n", dl.Accel)
	fmt.Printf("🛑 刹车:  %6.1f %%\n", dl.Brake)
	fmt.Printf("🎯 转向:  %6.1f °\n", dl.Steer)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// DataLogger 数据记录器（记录所有CAN帧）
type DataLogger struct {
	name       string
	frameLog   []CANFrame
	maxLogSize int
	mu         sync.RWMutex
}

func NewDataLogger(name string, maxLogSize int) *DataLogger {
	return &DataLogger{
		name:       name,
		frameLog:   make([]CANFrame, 0, maxLogSize),
		maxLogSize: maxLogSize,
	}
}

func (logger *DataLogger) GetName() string {
	return logger.name
}

func (logger *DataLogger) OnCANFrame(frame CANFrame) {
	logger.mu.Lock()
	defer logger.mu.Unlock()

	logger.frameLog = append(logger.frameLog, frame)
	if len(logger.frameLog) > logger.maxLogSize {
		logger.frameLog = logger.frameLog[1:] // 删除最老的帧
	}
}

func (logger *DataLogger) GetFrameCount() int {
	logger.mu.RLock()
	defer logger.mu.RUnlock()
	return len(logger.frameLog)
}

func (logger *DataLogger) PrintRecent(count int) {
	logger.mu.RLock()
	defer logger.mu.RUnlock()

	fmt.Printf("\n[%s] 最近 %d 帧:\n", logger.name, count)
	start := len(logger.frameLog) - count
	if start < 0 {
		start = 0
	}

	for i := start; i < len(logger.frameLog); i++ {
		frame := logger.frameLog[i]
		fmt.Printf("  [%s] 0x%03X [%d] %02X %02X %02X %02X %02X %02X %02X %02X\n",
			frame.Timestamp.Format("15:04:05.000"),
			frame.ID,
			frame.DLC,
			frame.Data[0], frame.Data[1], frame.Data[2], frame.Data[3],
			frame.Data[4], frame.Data[5], frame.Data[6], frame.Data[7])
	}
}

// DiagnosticListener 诊断监听器（检测异常）
type DiagnosticListener struct {
	name         string
	anomalyCount uint64
	lastFrames   map[uint32]time.Time
	mu           sync.RWMutex
}

func NewDiagnosticListener(name string) *DiagnosticListener {
	return &DiagnosticListener{
		name:       name,
		lastFrames: make(map[uint32]time.Time),
	}
}

func (diag *DiagnosticListener) GetName() string {
	return diag.name
}

func (diag *DiagnosticListener) OnCANFrame(frame CANFrame) {
	diag.mu.Lock()
	defer diag.mu.Unlock()

	// 检测帧间隔异常（超过1秒未收到同一CAN ID的帧）
	if lastTime, exists := diag.lastFrames[frame.ID]; exists {
		interval := frame.Timestamp.Sub(lastTime)
		if interval > time.Second {
			diag.anomalyCount++
			fmt.Printf("⚠️  [%s] 异常：CAN ID 0x%03X 帧间隔过长: %v\n", diag.name, frame.ID, interval)
		}
	}

	diag.lastFrames[frame.ID] = frame.Timestamp
}

func (diag *DiagnosticListener) GetAnomalyCount() uint64 {
	diag.mu.RLock()
	defer diag.mu.RUnlock()
	return diag.anomalyCount
}

// GenericListener 通用监听器（可配置）
type GenericListener struct {
	name     string
	canIDs   []uint32
	callback func(frame CANFrame)
}

func NewGenericListener(name string, canIDs []uint32, callback func(frame CANFrame)) *GenericListener {
	return &GenericListener{
		name:     name,
		canIDs:   canIDs,
		callback: callback,
	}
}

func (gl *GenericListener) GetName() string {
	return gl.name
}

func (gl *GenericListener) OnCANFrame(frame CANFrame) {
	if gl.callback != nil {
		gl.callback(frame)
	}
}
