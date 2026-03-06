package main

import "fmt"

// 演示如何使用ID.4的CAN信号定义

func main() {
	fmt.Println("========================================")
	fmt.Println("大众ID.4 CAN信号使用演示")
	fmt.Println("========================================\n")

	// 示例1: 列出所有信号
	ListAllSignals()

	// 示例2: 查看特定信号详情
	fmt.Println("\n========================================")
	fmt.Println("示例1: 查看车速信号详情")
	fmt.Println("========================================\n")
	PrintSignalInfo("Indicated_Vehicle_Speed_kph")

	// 示例3: 解码车速信号
	fmt.Println("========================================")
	fmt.Println("示例2: 解码车速信号")
	fmt.Println("========================================\n")
	speedSignal := ID4Signals["Indicated_Vehicle_Speed_kph"]
	rawSpeed := uint32(5000) // 假设从CAN数据中提取的原始值
	actualSpeed := DecodeSignal(rawSpeed, speedSignal)
	fmt.Printf("原始值: %d\n", rawSpeed)
	fmt.Printf("实际车速: %.2f %s\n\n", actualSpeed, speedSignal.Unit)

	// 示例4: 编码刹车信号
	fmt.Println("========================================")
	fmt.Println("示例3: 编码刹车踏板位置")
	fmt.Println("========================================\n")
	brakeSignal := ID4Signals["Brake_Position"]
	actualBrake := 50.0 // 50%的刹车力度
	rawBrake := EncodeSignal(actualBrake, brakeSignal)
	fmt.Printf("实际刹车位置: %.1f%s\n", actualBrake, brakeSignal.Unit)
	fmt.Printf("编码后的值: %d\n\n", rawBrake)

	// 示例5: 方向盘转角
	fmt.Println("========================================")
	fmt.Println("示例4: 解析方向盘转角")
	fmt.Println("========================================\n")
	steeringSignal := ID4Signals["Steering_Angle"]
	rawSteering := uint32(450) // 原始值
	actualSteering := DecodeSignal(rawSteering, steeringSignal)
	fmt.Printf("原始值: %d\n", rawSteering)
	fmt.Printf("方向盘转角: %.1f%s\n\n", actualSteering, steeringSignal.Unit)

	// 示例6: 档位状态
	fmt.Println("========================================")
	fmt.Println("示例5: 档位状态解析")
	fmt.Println("========================================\n")
	gearSignal := ID4Signals["Gear_Switch"]
	gearValues := map[uint32]string{
		0: "P停车",
		1: "R倒车",
		2: "N空档",
		3: "D前进",
		4: "S运动",
	}
	for rawGear, gearName := range gearValues {
		actualGear := DecodeSignal(rawGear, gearSignal)
		fmt.Printf("原始值 %d -> %.0f -> %s\n", rawGear, actualGear, gearName)
	}

	fmt.Println("\n========================================")
	fmt.Println("演示完成")
	fmt.Println("========================================")
}
