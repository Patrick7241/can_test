# 大众ID.4 CAN信号数据库

大众ID.4 (2020+) 电动车的完整CAN信号定义。

## 文件说明

- **`signals.go`** - 12个CAN信号的完整定义、编解码函数
- **`demo.go`** - 信号使用演示程序
- **`Volkswagen-ID.4 2020-.REF`** - 原始Racelogic格式信号定义文件

## 快速开始

### 运行演示程序

```bash
cd vehicles/volkswagen_id4
go run *.go
```

### 在代码中使用

```go
// 获取信号定义
speedSignal := ID4Signals["Indicated_Vehicle_Speed_kph"]

// 解码CAN数据
rawValue := uint32(5000)
actualSpeed := DecodeSignal(rawValue, speedSignal)
fmt.Printf("车速: %.2f km/h\n", actualSpeed)

// 编码为CAN数据
actualBrake := 50.0  // 50%刹车
rawBrake := EncodeSignal(actualBrake, ID4Signals["Brake_Position"])
```

## 信号列表 (12个)

### 动力与驾驶 (5个)

| 信号名称 | CAN ID | 单位 | 说明 |
|---------|--------|------|------|
| Indicated_Vehicle_Speed_kph | 0xFD (253) | km/h | 车速显示(公里) |
| Indicated_Vehicle_Speed_mph | 0xFD (253) | mph | 车速显示(英里) |
| Gear_Switch | 0xB5 (181) | - | 档位状态 P/R/N/D |
| Accelerator_Pedal_Position | 0x3EB (1003) | % | 油门踏板位置 |
| Brake_Position | 0x3EB (1003) | % | 刹车踏板位置 |

### 车辆动态 (5个)

| 信号名称 | CAN ID | 单位 | 说明 |
|---------|--------|------|------|
| Steering_Angle | 0x3DA (986) | ° | 方向盘转角 |
| Steering_Direction | 0x3DA (986) | - | 转向方向 |
| Yaw_Rate | 0x101 (257) | °/s | 横摆角速度 |
| Indicated_Lateral_Acceleration | 0x101 (257) | g | 横向加速度 |
| Indicated_Longitudinal_Acceleration | 0x101 (257) | g | 纵向加速度 |

### 环境信息 (1个)

| 信号名称 | CAN ID | 单位 | 说明 |
|---------|--------|------|------|
| Air_Temperature | 0x5E1 (1505) | °C | 车外环境温度 |

## 信号详细说明

### 1. Indicated_Vehicle_Speed_kph (车速-公里)
- **CAN ID**: 0xFD (253)
- **起始位**: 32
- **长度**: 16 bit
- **范围**: 0-655.35 km/h
- **转换**: 实际值 = 原始值 × 0.01 + 655.35
- **用途**: 仪表盘显示、ADAS系统、车速限制

### 2. Gear_Switch (档位)
- **CAN ID**: 0xB5 (181)
- **起始位**: 52
- **长度**: 3 bit (0-7)
- **可能值**:
  - 0 = P停车档
  - 1 = R倒车档
  - 2 = N空档
  - 3 = D前进档
  - 4 = S运动模式
- **用途**: 变速箱控制、驾驶模式识别

### 3. Brake_Position (刹车位置)
- **CAN ID**: 0x3EB (1003)
- **起始位**: 54
- **长度**: 10 bit
- **范围**: -20% ~ 184.6%
- **转换**: 实际值 = 原始值 × 0.2 - 20
- **用途**: 刹车力度控制、能量回收、ABS/ESP系统

### 4. Accelerator_Pedal_Position (油门位置)
- **CAN ID**: 0x3EB (1003)
- **起始位**: 16
- **长度**: 8 bit
- **范围**: 0-102%
- **转换**: 实际值 = 原始值 × 0.4
- **用途**: 电机功率控制、驾驶模式、能耗计算

### 5. Steering_Angle (方向盘转角)
- **CAN ID**: 0x3DA (986)
- **起始位**: 43
- **长度**: 13 bit
- **范围**: 0-819.1°
- **转换**: 实际值 = 原始值 × 0.1
- **用途**: 车道保持、自动泊车、转向辅助

### 6. Yaw_Rate (横摆角速度)
- **CAN ID**: 0x101 (257)
- **起始位**: 40
- **长度**: 15 bit
- **范围**: -163.84 ~ 163.83 °/s
- **转换**: 实际值 = 原始值 × 0.01 - 163.84
- **用途**: ESP稳定系统、转向稳定性检测

### 7. Indicated_Lateral_Acceleration (横向加速度)
- **CAN ID**: 0x101 (257)
- **起始位**: 40
- **长度**: 8 bit
- **范围**: -1.27 ~ 1.28 g
- **转换**: 实际值 = 原始值 × 0.01 - 1.27
- **用途**: ESP系统、转弯侧向力检测

### 8. Indicated_Longitudinal_Acceleration (纵向加速度)
- **CAN ID**: 0x101 (257)
- **起始位**: 24
- **长度**: 10 bit
- **范围**: -1.63 ~ 1.63 g
- **转换**: 实际值 = 原始值 × 0.0031866 - 1.63154
- **用途**: 防碰撞系统、加速/刹车检测

### 9. Air_Temperature (环境温度)
- **CAN ID**: 0x5E1 (1505)
- **起始位**: 56
- **长度**: 8 bit
- **范围**: -50 ~ 77.5 °C
- **转换**: 实际值 = 原始值 × 0.5 - 50
- **用途**: 仪表显示、空调控制、电池热管理

### 10. Steering_Direction (转向方向)
- **CAN ID**: 0x3DA (986)
- **起始位**: 18
- **长度**: 1 bit
- **值**: 0=右转, 1=左转
- **用途**: 转向灯控制、车道偏离警告

## API函数

### DecodeSignal()
解码CAN信号，将原始值转换为实际物理值。

```go
func DecodeSignal(rawValue uint32, signal CANSignalDef) float64
```

### EncodeSignal()
编码CAN信号，将实际物理值转换为原始值。

```go
func EncodeSignal(actualValue float64, signal CANSignalDef) uint32
```

### PrintSignalInfo()
打印信号的详细信息。

```go
func PrintSignalInfo(name string)
```

### ListAllSignals()
列出所有可用的信号名称。

```go
func ListAllSignals()
```

## 数据来源

信号定义来自 **Volkswagen-ID.4 2020-.REF** 文件，这是Racelogic格式的CAN数据库文件。

## 技术说明

### 字节序
- **Intel (小端序)**: 低字节在前，高字节在后
- **Motorola (大端序)**: 高字节在前，低字节在后

### 信号提取
从CAN数据帧中提取信号需要知道：
1. 起始位（StartBit）
2. 数据长度（BitLength）
3. 字节序（ByteOrder）

### 值转换公式
```
实际物理值 = 原始值 × Scale + Offset
原始值 = (实际物理值 - Offset) / Scale
```

## 应用场景

1. **车辆监控** - 实时监测车辆状态
2. **数据记录** - 记录行驶数据用于分析
3. **ADAS开发** - 辅助驾驶系统开发
4. **故障诊断** - 分析车辆异常行为
5. **性能调优** - 优化驾驶体验
6. **能耗分析** - 电动车能量管理

## 注意事项

1. 信号定义基于2020+年款ID.4，不同年款可能有差异
2. 某些信号可能在不同配置车型上不可用
3. 实际应用需要参考车辆官方技术文档
4. CAN ID可能在不同市场版本中有所不同

## 许可

本数据仅用于学习和研究目的。
