package can

import (
	"fmt"
	"time"
)

/*
========================================
大众ID.4 CAN信号数据库
========================================

数据来源: Volkswagen-ID.4 2020-.REF
解析时间: 2026-03-06
文件格式: Racelogic CAN Data File V1a
车型: Volkswagen ID.4 (2020+)
信号总数: 12

本文件包含大众ID.4电动车的CAN总线信号定义
*/

// CANSignalDef CAN信号定义结构
type CANSignalDef struct {
	Name      string  // 信号名称
	CANID     uint32  // CAN消息ID
	Unit      string  // 单位
	StartBit  int     // 起始位位置
	BitLength int     // 数据长度（bit）
	MinValue  float64 // 最小值
	Scale     float64 // 缩放因子
	MaxValue  float64 // 最大值
	Offset    float64 // 偏移量
	Signed    bool    // 是否有符号数
	ByteOrder string  // 字节序 (Motorola=大端, Intel=小端)
	Desc      string  // 中文描述
}

// ID4Signals 大众ID.4所有CAN信号定义
var ID4Signals = map[string]CANSignalDef{
	// 信号 #1: 车速指示 (km/h)
	"Indicated_Vehicle_Speed_kph": {
		Name:      "Indicated_Vehicle_Speed_kph",
		CANID:     0xFD, // 253
		Unit:      "km/h",
		StartBit:  32,
		BitLength: 16,
		MinValue:  0,
		Scale:     0.01,
		MaxValue:  655.35,
		Offset:    655.35,
		Signed:    false,
		ByteOrder: "Intel",
		Desc:      "仪表盘显示的车辆速度(公里/小时) - 用于车速表显示和驾驶辅助系统",
	},

	// 信号 #2: 档位开关
	"Gear_Switch": {
		Name:      "Gear_Switch",
		CANID:     0xB5, // 181
		Unit:      "",
		StartBit:  52,
		BitLength: 3,
		MinValue:  0,
		Scale:     1,
		MaxValue:  7,
		Offset:    7,
		Signed:    false,
		ByteOrder: "Motorola",
		Desc:      "档位状态指示 - P/R/N/D档位信息(0-7: P停车/R倒车/N空档/D前进等)",
	},

	// 信号 #3: 横向加速度
	"Indicated_Lateral_Acceleration": {
		Name:      "Indicated_Lateral_Acceleration",
		CANID:     0x101, // 257
		Unit:      "g",
		StartBit:  40,
		BitLength: 8,
		MinValue:  -1.27,
		Scale:     0.01,
		MaxValue:  1.28,
		Offset:    -1.27,
		Signed:    false,
		ByteOrder: "Motorola",
		Desc:      "车辆横向加速度 - 用于ESP车身稳定系统，检测转弯时的侧向力",
	},

	// 信号 #4: 纵向加速度
	"Indicated_Longitudinal_Acceleration": {
		Name:      "Indicated_Longitudinal_Acceleration",
		CANID:     0x101, // 257
		Unit:      "g",
		StartBit:  24,
		BitLength: 10,
		MinValue:  -1.63154,
		Scale:     0.0031866,
		MaxValue:  1.6283518,
		Offset:    -1.63154,
		Signed:    false,
		ByteOrder: "Intel",
		Desc:      "车辆纵向加速度 - 检测加速/刹车时的前后方向加速度，用于防碰撞系统",
	},

	// 信号 #5: 横摆角速度
	"Yaw_Rate": {
		Name:      "Yaw_Rate",
		CANID:     0x101, // 257
		Unit:      "°/s",
		StartBit:  40,
		BitLength: 15,
		MinValue:  0,
		Scale:     0.01,
		MaxValue:  163.83,
		Offset:    -163.84,
		Signed:    false,
		ByteOrder: "Intel",
		Desc:      "车辆横摆角速度 - 车辆绕垂直轴旋转的速度，ESP系统用于检测转向稳定性",
	},

	// 信号 #6: 环境温度
	"Air_Temperature": {
		Name:      "Air_Temperature",
		CANID:     0x5E1, // 1505
		Unit:      "°C",
		StartBit:  56,
		BitLength: 8,
		MinValue:  -50,
		Scale:     0.5,
		MaxValue:  77.5,
		Offset:    -50,
		Signed:    false,
		ByteOrder: "Motorola",
		Desc:      "车外环境温度 - 显示在仪表盘上，也用于空调系统和电池热管理",
	},

	// 信号 #7: 方向盘转角
	"Steering_Angle": {
		Name:      "Steering_Angle",
		CANID:     0x3DA, // 986
		Unit:      "°",
		StartBit:  43,
		BitLength: 13,
		MinValue:  0,
		Scale:     0.1,
		MaxValue:  819.1,
		Offset:    0,
		Signed:    false,
		ByteOrder: "Intel",
		Desc:      "方向盘转角 - 车道保持、自动泊车等ADAS功能的核心输入信号",
	},

	// 信号 #8: 转向方向
	"Steering_Direction": {
		Name:      "Steering_Direction",
		CANID:     0x3DA, // 986
		Unit:      "",
		StartBit:  18,
		BitLength: 1,
		MinValue:  0,
		Scale:     1,
		MaxValue:  1,
		Offset:    0,
		Signed:    false,
		ByteOrder: "Motorola",
		Desc:      "转向方向标识 - 0=右转，1=左转",
	},

	// 信号 #9: 油门踏板位置
	"Accelerator_Pedal_Position": {
		Name:      "Accelerator_Pedal_Position",
		CANID:     0x3EB, // 1003
		Unit:      "%",
		StartBit:  16,
		BitLength: 8,
		MinValue:  0,
		Scale:     0.4,
		MaxValue:  102,
		Offset:    0,
		Signed:    false,
		ByteOrder: "Motorola",
		Desc:      "油门(加速)踏板位置百分比 - 电动车中控制电机输出功率的主要输入",
	},

	// 信号 #10: 刹车踏板位置
	"Brake_Position": {
		Name:      "Brake_Position",
		CANID:     0x3EB, // 1003
		Unit:      "%",
		StartBit:  54,
		BitLength: 10,
		MinValue:  -20,
		Scale:     0.2,
		MaxValue:  184.6,
		Offset:    -20,
		Signed:    false,
		ByteOrder: "Intel",
		Desc:      "刹车踏板位置百分比 - 控制刹车力度和能量回收强度的关键信号",
	},

	// 信号 #11: 车速指示 (mph)
	"Indicated_Vehicle_Speed_mph": {
		Name:      "Indicated_Vehicle_Speed_mph",
		CANID:     0xFD, // 253
		Unit:      "mph",
		StartBit:  32,
		BitLength: 16,
		MinValue:  0,
		Scale:     0.00621,
		MaxValue:  406.97235,
		Offset:    0,
		Signed:    false,
		ByteOrder: "Intel",
		Desc:      "仪表盘显示的车辆速度(英里/小时) - 与km/h信号同源，用于英制单位地区",
	},

	// 注：实际应用中Gear_Switch的值对应关系可能为:
	// 0=Park(P), 1=Reverse(R), 2=Neutral(N), 3=Drive(D), 4=Sport, 5=Manual等
	// 具体需要参考车辆技术手册
}

// DecodeSignal 通用解码函数
// rawValue: 从CAN数据中提取的原始数值
// signal: 信号定义
// 返回: 实际物理值
func DecodeSignal(rawValue uint32, signal CANSignalDef) float64 {
	return float64(rawValue)*signal.Scale + signal.Offset
}

// EncodeSignal 通用编码函数
// actualValue: 实际物理值
// signal: 信号定义
// 返回: 编码后的原始值
func EncodeSignal(actualValue float64, signal CANSignalDef) uint32 {
	rawValue := (actualValue - signal.Offset) / signal.Scale
	maxRaw := uint32((1 << uint(signal.BitLength)) - 1)

	if rawValue < 0 {
		return 0
	}
	if rawValue > float64(maxRaw) {
		return maxRaw
	}
	return uint32(rawValue)
}

// PrintSignalInfo 打印信号详细信息
func PrintSignalInfo(name string) {
	signal, exists := ID4Signals[name]
	if !exists {
		fmt.Printf("信号 '%s' 不存在\n", name)
		return
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("信号名称: %s\n", signal.Name)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("CAN ID:   0x%X (%d)\n", signal.CANID, signal.CANID)
	fmt.Printf("单位:     %s\n", signal.Unit)
	fmt.Printf("起始位:   %d\n", signal.StartBit)
	fmt.Printf("长度:     %d bit\n", signal.BitLength)
	fmt.Printf("缩放:     %.6f\n", signal.Scale)
	fmt.Printf("偏移:     %.6f\n", signal.Offset)
	fmt.Printf("值范围:   %.3f ~ %.3f %s\n", signal.MinValue, signal.MaxValue, signal.Unit)
	fmt.Printf("字节序:   %s\n", signal.ByteOrder)
	fmt.Printf("描述:     %s\n", signal.Desc)
	fmt.Println()
}

// ListAllSignals 列出所有信号
func ListAllSignals() {
	fmt.Println("========================================")
	fmt.Println("大众ID.4 CAN信号列表")
	fmt.Println("========================================\n")

	i := 1
	for name := range ID4Signals {
		fmt.Printf("%2d. %s\n", i, name)
		i++
	}
	fmt.Printf("\n共 %d 个信号\n", len(ID4Signals))
}

// VehicleState 车辆状态
type VehicleState struct {
	Speed              float64   // km/h
	Gear               int       // 0=P, 1=R, 2=N, 3=D
	BrakePosition      float64   // %
	AcceleratorPos     float64   // %
	SteeringAngle      float64   // 度
	SteeringDirection  int       // 0=右, 1=左
	LateralAccel       float64   // g
	LongitudinalAccel  float64   // g
	YawRate            float64   // °/s
	AirTemp            float64   // °C
	LastUpdate         time.Time // 最后更新时间
	TotalDistance      float64   // 总里程 km
}
