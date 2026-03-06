package can

import (
	"fmt"
	"math"
	"time"
)

/*
========================================
车辆ECU模拟器
========================================

作为CAN总线的发送者：
- 维护车辆状态
- 定期生成CAN帧
- 发送到CAN总线
*/

// VehicleECU 车辆电子控制单元
type VehicleECU struct {
	Name       string
	State      VehicleState
	bus        *CANBus
	sendRate   time.Duration // 发送频率
	running    bool
	frameCount uint64
}

// NewVehicleECU 创建新的车辆ECU
func NewVehicleECU(name string, bus *CANBus, sendRate time.Duration) *VehicleECU {
	return &VehicleECU{
		Name: name,
		State: VehicleState{
			Speed:      0,
			Gear:       0,
			AirTemp:    22.0,
			LastUpdate: time.Now(),
		},
		bus:      bus,
		sendRate: sendRate,
		running:  false,
	}
}

// Start 启动ECU
func (ecu *VehicleECU) Start() {
	ecu.running = true
	go ecu.sendLoop()
	fmt.Printf("🚗 车辆ECU [%s] 已启动，发送频率: %v\n", ecu.Name, ecu.sendRate)
}

// Stop 停止ECU
func (ecu *VehicleECU) Stop() {
	ecu.running = false
	fmt.Printf("🚗 车辆ECU [%s] 已停止，共发送 %d 帧\n", ecu.Name, ecu.frameCount)
}

// sendLoop ECU主循环：定期发送CAN帧
func (ecu *VehicleECU) sendLoop() {
	ticker := time.NewTicker(ecu.sendRate)
	defer ticker.Stop()

	physicsUpdateRate := 100 * time.Millisecond
	physicsTicker := time.NewTicker(physicsUpdateRate)
	defer physicsTicker.Stop()

	lastPhysicsUpdate := time.Now()

	for ecu.running {
		select {
		case <-physicsTicker.C:
			// 更新物理状态
			now := time.Now()
			deltaTime := now.Sub(lastPhysicsUpdate).Seconds()
			ecu.UpdatePhysics(deltaTime)
			lastPhysicsUpdate = now

		case <-ticker.C:
			// 生成并发送CAN帧
			frames := ecu.GenerateCANFrames()
			for _, frame := range frames {
				err := ecu.bus.SendFrame(frame)
				if err != nil {
					// 丢帧，但不打印（避免刷屏）
				} else {
					ecu.frameCount++
				}
			}
		}
	}
}

// GenerateCANFrames 生成所有CAN帧
func (ecu *VehicleECU) GenerateCANFrames() []CANFrame {
	frames := make([]CANFrame, 0)

	// 0xFD - 车速
	frames = append(frames, ecu.makeSpeedFrame())

	// 0xB5 - 档位
	frames = append(frames, ecu.makeGearFrame())

	// 0x3EB - 刹车和油门
	frames = append(frames, ecu.makePedalFrame())

	// 0x3DA - 转向
	frames = append(frames, ecu.makeSteeringFrame())

	// 0x101 - 加速度和横摆
	frames = append(frames, ecu.makeDynamicsFrame())

	// 0x5E1 - 环境温度
	frames = append(frames, ecu.makeTempFrame())

	return frames
}

// makeSpeedFrame 生成车速帧 (0xFD)
func (ecu *VehicleECU) makeSpeedFrame() CANFrame {
	signal := ID4Signals["Indicated_Vehicle_Speed_kph"]
	rawValue := EncodeSignal(ecu.State.Speed, signal)

	var data [8]byte
	// StartBit=32，即第4-5字节 (小端序)
	data[4] = byte(rawValue & 0xFF)
	data[5] = byte((rawValue >> 8) & 0xFF)

	return CANFrame{
		ID:  0xFD,
		DLC: 8,
		Data: data,
	}
}

// makeGearFrame 生成档位帧 (0xB5)
func (ecu *VehicleECU) makeGearFrame() CANFrame {
	signal := ID4Signals["Gear_Switch"]
	rawValue := EncodeSignal(float64(ecu.State.Gear), signal)

	var data [8]byte
	// StartBit=52，即第6字节的bit4-6
	data[6] = byte((rawValue & 0x07) << 4)

	return CANFrame{
		ID:  0xB5,
		DLC: 8,
		Data: data,
	}
}

// makePedalFrame 生成踏板帧 (0x3EB) - 刹车和油门
func (ecu *VehicleECU) makePedalFrame() CANFrame {
	brakeSignal := ID4Signals["Brake_Position"]
	accelSignal := ID4Signals["Accelerator_Pedal_Position"]

	brakeRaw := EncodeSignal(ecu.State.BrakePosition, brakeSignal)
	accelRaw := EncodeSignal(ecu.State.AcceleratorPos, accelSignal)

	var data [8]byte
	// 油门 StartBit=16 (第2-3字节)
	data[2] = byte(accelRaw & 0xFF)
	// 刹车 StartBit=54 (第6-7字节)
	data[6] = byte(brakeRaw & 0xFF)
	data[7] = byte((brakeRaw >> 8) & 0x03)

	return CANFrame{
		ID:  0x3EB,
		DLC: 8,
		Data: data,
	}
}

// makeSteeringFrame 生成转向帧 (0x3DA)
func (ecu *VehicleECU) makeSteeringFrame() CANFrame {
	angleSignal := ID4Signals["Steering_Angle"]
	dirSignal := ID4Signals["Steering_Direction"]

	angleRaw := EncodeSignal(ecu.State.SteeringAngle, angleSignal)
	dirRaw := EncodeSignal(float64(ecu.State.SteeringDirection), dirSignal)

	var data [8]byte
	// 转向方向 StartBit=18 (第2字节bit2)
	data[2] = byte((dirRaw & 0x01) << 2)
	// 转向角 StartBit=43 (第5-6字节)
	data[5] = byte(angleRaw & 0xFF)
	data[6] = byte((angleRaw >> 8) & 0x1F)

	return CANFrame{
		ID:  0x3DA,
		DLC: 8,
		Data: data,
	}
}

// makeDynamicsFrame 生成动力学帧 (0x101) - 加速度和横摆
func (ecu *VehicleECU) makeDynamicsFrame() CANFrame {
	longSignal := ID4Signals["Indicated_Longitudinal_Acceleration"]
	latSignal := ID4Signals["Indicated_Lateral_Acceleration"]
	yawSignal := ID4Signals["Yaw_Rate"]

	longRaw := EncodeSignal(ecu.State.LongitudinalAccel, longSignal)
	latRaw := EncodeSignal(ecu.State.LateralAccel, latSignal)
	yawRaw := EncodeSignal(ecu.State.YawRate, yawSignal)

	var data [8]byte
	// 纵向加速度 StartBit=24 (第3-4字节)
	data[3] = byte(longRaw & 0xFF)
	data[4] = byte((longRaw >> 8) & 0x03)
	// 横向加速度 StartBit=40 (第5字节)
	data[5] = byte(latRaw & 0xFF)
	// 横摆角速度 StartBit=40 (第5-6字节) - 注意与横向加速度共用
	data[5] = byte(yawRaw & 0xFF)
	data[6] = byte((yawRaw >> 8) & 0x7F)

	return CANFrame{
		ID:  0x101,
		DLC: 8,
		Data: data,
	}
}

// makeTempFrame 生成温度帧 (0x5E1)
func (ecu *VehicleECU) makeTempFrame() CANFrame {
	signal := ID4Signals["Air_Temperature"]
	rawValue := EncodeSignal(ecu.State.AirTemp, signal)

	var data [8]byte
	// StartBit=56 (第7字节)
	data[7] = byte(rawValue & 0xFF)

	return CANFrame{
		ID:  0x5E1,
		DLC: 8,
		Data: data,
	}
}

// UpdatePhysics 更新物理状态 (与之前的物理模型相同)
func (ecu *VehicleECU) UpdatePhysics(deltaTime float64) {
	if ecu.State.Gear == 3 { // D档
		targetAccel := ecu.State.AcceleratorPos * 0.03
		brakeDecel := ecu.State.BrakePosition * 0.05
		ecu.State.LongitudinalAccel = targetAccel - brakeDecel
		ecu.State.Speed += ecu.State.LongitudinalAccel * deltaTime * 3.6
		resistance := ecu.State.Speed * 0.01
		ecu.State.Speed -= resistance * deltaTime
	} else if ecu.State.Gear == 1 { // R档
		if ecu.State.AcceleratorPos > 0 && ecu.State.Speed > -20 {
			ecu.State.Speed -= 0.5 * deltaTime
		}
	} else {
		if ecu.State.Speed > 0 {
			ecu.State.Speed -= 2.0 * deltaTime
		}
	}

	if ecu.State.Speed < 0 && ecu.State.Gear != 1 {
		ecu.State.Speed = 0
	}
	if ecu.State.Speed > 180 {
		ecu.State.Speed = 180
	}
	if ecu.State.Gear == 1 && ecu.State.Speed < -20 {
		ecu.State.Speed = -20
	}

	if math.Abs(ecu.State.SteeringAngle) > 0.1 && ecu.State.Speed > 0 {
		ecu.State.LateralAccel = ecu.State.SteeringAngle / 100.0 * ecu.State.Speed / 50.0
		if ecu.State.LateralAccel > 1.0 {
			ecu.State.LateralAccel = 1.0
		}
		if ecu.State.LateralAccel < -1.0 {
			ecu.State.LateralAccel = -1.0
		}
		ecu.State.YawRate = ecu.State.SteeringAngle * ecu.State.Speed / 30.0
	} else {
		ecu.State.LateralAccel *= 0.9
		ecu.State.YawRate *= 0.9
	}

	if ecu.State.SteeringAngle < -1 {
		ecu.State.SteeringDirection = 0
	} else if ecu.State.SteeringAngle > 1 {
		ecu.State.SteeringDirection = 1
	}

	if ecu.State.Speed > 0 {
		ecu.State.TotalDistance += ecu.State.Speed * deltaTime / 3600.0
	}

	ecu.State.LastUpdate = time.Now()
}

// HandleInput 处理用户输入 (与之前相同)
func (ecu *VehicleECU) HandleInput(input byte) string {
	var action string
	timestamp := time.Now().Format("15:04:05.000")

	switch input {
	case 'w', 'W':
		if ecu.State.AcceleratorPos < 100 {
			ecu.State.AcceleratorPos += 10
			if ecu.State.AcceleratorPos > 100 {
				ecu.State.AcceleratorPos = 100
			}
		}
		ecu.State.BrakePosition = 0
		action = fmt.Sprintf("[%s] 🚀 加速！油门: %.0f%%", timestamp, ecu.State.AcceleratorPos)

	case 's', 'S':
		ecu.State.AcceleratorPos = 0
		if ecu.State.BrakePosition < 100 {
			ecu.State.BrakePosition += 20
			if ecu.State.BrakePosition > 100 {
				ecu.State.BrakePosition = 100
			}
		}
		action = fmt.Sprintf("[%s] 🛑 刹车！刹车力度: %.0f%%", timestamp, ecu.State.BrakePosition)

	case 'a', 'A':
		if ecu.State.SteeringAngle < 400 {
			ecu.State.SteeringAngle += 30
		}
		action = fmt.Sprintf("[%s] ⬅️  左转！转角: %.0f°", timestamp, ecu.State.SteeringAngle)

	case 'd', 'D':
		if ecu.State.SteeringAngle > -400 {
			ecu.State.SteeringAngle -= 30
		}
		action = fmt.Sprintf("[%s] ➡️  右转！转角: %.0f°", timestamp, ecu.State.SteeringAngle)

	case 'x', 'X':
		ecu.State.SteeringAngle *= 0.5
		action = fmt.Sprintf("[%s] ↕️  方向盘回正", timestamp)

	case 'p', 'P':
		ecu.State.Gear = 0
		ecu.State.AcceleratorPos = 0
		ecu.State.BrakePosition = 0
		action = fmt.Sprintf("[%s] 🅿️  切换到 P档（驻车）", timestamp)

	case 'r', 'R':
		if ecu.State.Speed < 1 {
			ecu.State.Gear = 1
			action = fmt.Sprintf("[%s] ◀️  切换到 R档（倒车）", timestamp)
		} else {
			action = fmt.Sprintf("[%s] ⚠️  速度过快(%.1f km/h)，无法切换到R档", timestamp, ecu.State.Speed)
		}

	case 'n', 'N':
		ecu.State.Gear = 2
		ecu.State.AcceleratorPos = 0
		action = fmt.Sprintf("[%s] ⏸️  切换到 N档（空档）", timestamp)

	case 'g', 'G':
		ecu.State.Gear = 3
		action = fmt.Sprintf("[%s] ▶️  切换到 D档（前进）", timestamp)

	case ' ':
		ecu.State.AcceleratorPos *= 0.7
		ecu.State.BrakePosition *= 0.7
		action = fmt.Sprintf("[%s] 🔄 释放油门和刹车", timestamp)
	}

	return action
}
