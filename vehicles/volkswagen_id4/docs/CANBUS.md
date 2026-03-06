# 🚌 CAN总线模拟系统

真正的CAN总线架构：车辆ECU向总线发送帧，监听器订阅并接收帧。

## ✨ 系统架构

```
┌─────────────┐
│  车辆ECU     │  (发送者)
│  VehicleECU  │  定期发送CAN帧
└──────┬──────┘
       │ 📡 发送CAN帧
       ↓
┌─────────────────────────────────────┐
│          CAN总线 (CANBus)           │  ← 中心消息总线
│   - 路由CAN帧                       │
│   - 管理订阅                        │
│   - 统计信息                        │
└─┬────┬────┬────┬────┬────┬────┬────┘
  │    │    │    │    │    │    │
  │    │    │    │    │    │    └──→ 📻 自定义监听器
  │    │    │    │    │    └─────→ 📻 诊断监听器
  │    │    │    │    └──────────→ 📻 数据记录器
  │    │    │    └───────────────→ 📻 仪表盘监听器
  │    │    └────────────────────→ 📻 速度监听器
  ↓    ↓                           (接收者)
  更多监听器...
```

## 🎯 核心概念

### 1. CANFrame - CAN帧

符合CAN 2.0B标准的数据帧：

```go
type CANFrame struct {
    ID        uint32    // CAN标识符(11-bit)
    DLC       uint8     // 数据长度码(0-8字节)
    Data      [8]byte   // 数据域(最多8字节)
    Timestamp time.Time // 时间戳
    Priority  uint8     // 优先级(ID越小优先级越高)
}
```

### 2. CANBus - CAN总线

中心消息总线，负责：
- **接收帧**: 从发送者接收CAN帧
- **路由**: 根据CAN ID路由到订阅者
- **统计**: 记录帧数、丢帧数等

```go
canBus := NewCANBus("VW-ID4-CAN")
canBus.Start()
```

### 3. VehicleECU - 车辆ECU

作为发送者：
- 维护车辆状态
- 定期(100ms)生成CAN帧
- 发送到CAN总线

```go
vehicleECU := NewVehicleECU("ID4-ECU", canBus, 100*time.Millisecond)
vehicleECU.Start()
```

### 4. CANListener - 监听器

作为接收者，实现接口：

```go
type CANListener interface {
    OnCANFrame(frame CANFrame) // 接收CAN帧的回调
    GetName() string           // 获取监听器名称
}
```

## 📡 CAN帧映射

车辆状态 → CAN帧的映射关系：

| CAN ID | 信号内容 | 更新频率 |
|--------|---------|---------|
| 0xFD | 车速 | 100ms |
| 0xB5 | 档位 | 100ms |
| 0x3EB | 刹车位置 + 油门位置 | 100ms |
| 0x3DA | 转向角 + 转向方向 | 100ms |
| 0x101 | 纵向加速度 + 横向加速度 + 横摆角速度 | 100ms |
| 0x5E1 | 环境温度 | 100ms |

## 🎮 快速开始

### 运行CAN总线演示

```bash
cd vehicles/volkswagen_id4
go run signals.go canbus.go vehicle_ecu.go can_listeners.go canbus_demo.go
```

或者编译后运行：

```bash
go build -o canbus_demo signals.go canbus.go vehicle_ecu.go can_listeners.go canbus_demo.go
./canbus_demo
```

### 操作说明

**驾驶控制**:
- `W` 加速 → 增加油门 → 发送 0x3EB 帧
- `S` 刹车 → 增加刹车 → 发送 0x3EB 帧
- `A` 左转 → 转向左 → 发送 0x3DA 帧
- `D` 右转 → 转向右 → 发送 0x3DA 帧
- `空格` 松开油门/刹车

**档位切换**:
- `P` P档 → 发送 0xB5 帧
- `R` R档 → 发送 0xB5 帧
- `N` N档 → 发送 0xB5 帧
- `G` D档 → 发送 0xB5 帧

**查看信息**:
- `I` 显示仪表盘（从监听器读取）
- `B` 显示总线统计
- `L` 显示最近CAN帧日志

## 🔍 监听器详解

### 1. SpeedListener - 速度监听器

监听车速信号 (0xFD)：

```go
speedListener := NewSpeedListener("速度监听器")
canBus.Subscribe(0xFD, speedListener)
```

功能：
- 实时接收车速CAN帧
- 记录速度变化次数
- 提供当前车速查询

### 2. DashboardListener - 仪表盘监听器

监听多个信号 (0xFD, 0xB5, 0x3EB, 0x3DA)：

```go
dashboardListener := NewDashboardListener("仪表盘")
canBus.Subscribe(0xFD, dashboardListener)  // 车速
canBus.Subscribe(0xB5, dashboardListener)  // 档位
canBus.Subscribe(0x3EB, dashboardListener) // 刹车油门
canBus.Subscribe(0x3DA, dashboardListener) // 转向
```

功能：
- 从CAN总线读取并解析信号
- 维护仪表盘状态
- 提供可视化显示

### 3. DataLogger - 数据记录器

记录所有CAN帧：

```go
dataLogger := NewDataLogger("数据记录器", 1000)
canBus.Subscribe(0xFD, dataLogger)
canBus.Subscribe(0xB5, dataLogger)
// ... 订阅所有CAN ID
```

功能：
- 记录所有接收到的CAN帧
- 保留最近N帧（循环缓冲）
- 支持查询和导出

### 4. DiagnosticListener - 诊断监听器

监控总线健康：

```go
diagListener := NewDiagnosticListener("诊断系统")
canBus.Subscribe(0xFD, diagListener)
```

功能：
- 检测帧间隔异常
- 识别通信故障
- 统计异常次数

### 5. GenericListener - 通用监听器

自定义逻辑：

```go
customListener := NewGenericListener("自定义", []uint32{0x101},
    func(frame CANFrame) {
        // 自定义处理逻辑
        fmt.Printf("收到动力学帧: 0x%03X\n", frame.ID)
    })
canBus.Subscribe(0x101, customListener)
```

## 💡 使用示例

### 示例1: 创建自定义监听器

```go
type MyListener struct {
    name string
}

func (ml *MyListener) GetName() string {
    return ml.name
}

func (ml *MyListener) OnCANFrame(frame CANFrame) {
    fmt.Printf("[%s] 收到帧: 0x%03X\n", ml.name, frame.ID)
    // 解析数据...
}

// 使用
myListener := &MyListener{name: "我的监听器"}
canBus.Subscribe(0xFD, myListener)
```

### 示例2: 监听特定信号

```go
// 只监听车速变化
speedListener := NewGenericListener("车速监控", []uint32{0xFD},
    func(frame CANFrame) {
        rawValue := uint32(frame.Data[4]) | uint32(frame.Data[5])<<8
        speed := DecodeSignal(rawValue, ID4Signals["Indicated_Vehicle_Speed_kph"])
        if speed > 100 {
            fmt.Printf("⚠️  超速警告: %.1f km/h\n", speed)
        }
    })
canBus.Subscribe(0xFD, speedListener)
```

### 示例3: 记录到文件

```go
type FileLogger struct {
    name string
    file *os.File
}

func (fl *FileLogger) OnCANFrame(frame CANFrame) {
    fmt.Fprintf(fl.file, "%s,0x%03X,%d,%02X%02X%02X%02X%02X%02X%02X%02X\n",
        frame.Timestamp.Format("15:04:05.000"),
        frame.ID, frame.DLC,
        frame.Data[0], frame.Data[1], frame.Data[2], frame.Data[3],
        frame.Data[4], frame.Data[5], frame.Data[6], frame.Data[7])
}
```

## 📊 总线统计

查看总线运行状态：

```bash
# 按 'B' 键查看
━━━━━━━━━━━━━━ CAN总线统计 ━━━━━━━━━━━━━━
📊 总线名称: VW-ID4-CAN
📊 总帧数:   15420
📊 丢帧数:   0
📊 最后帧:   15:23:45.123
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
各CAN ID帧数:
  0x0FD: 2570 帧  (车速)
  0x0B5: 2570 帧  (档位)
  0x3EB: 2570 帧  (刹车油门)
  0x3DA: 2570 帧  (转向)
  0x101: 2570 帧  (动力学)
  0x5E1: 2570 帧  (温度)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## 🔧 技术细节

### CAN帧打包

车辆状态需要打包成8字节CAN帧：

```go
// 示例: 车速帧 (0xFD)
func makeSpeedFrame(speed float64) CANFrame {
    signal := ID4Signals["Indicated_Vehicle_Speed_kph"]
    rawValue := EncodeSignal(speed, signal)

    var data [8]byte
    // StartBit=32, 即第4-5字节 (小端序)
    data[4] = byte(rawValue & 0xFF)
    data[5] = byte((rawValue >> 8) & 0xFF)

    return CANFrame{
        ID:  0xFD,
        DLC: 8,
        Data: data,
    }
}
```

### 并发安全

所有监听器回调都是异步调用：

```go
for _, listener := range listeners {
    go listener.OnCANFrame(frame) // 异步避免阻塞
}
```

监听器内部需要使用互斥锁保护共享数据：

```go
type MyListener struct {
    data map[string]float64
    mu   sync.RWMutex
}

func (ml *MyListener) OnCANFrame(frame CANFrame) {
    ml.mu.Lock()
    defer ml.mu.Unlock()
    // 安全地修改data
}
```

## 🎯 与传统方式对比

### 传统方式 (simple_demo.go)

```
用户输入 → 直接修改车辆状态 → 生成CAN信号 → 打印显示
```

缺点：
- 信号只是编码/解码演示
- 没有真正的"总线"概念
- 无法模拟多ECU通信

### CAN总线方式 (canbus_demo.go)

```
用户输入 → 修改ECU状态 → 生成CAN帧 → 发送到总线 → 广播给所有监听器
```

优点：
- ✅ 真正的发布-订阅模式
- ✅ 模拟真实CAN总线通信
- ✅ 支持多个ECU和监听器
- ✅ 解耦发送者和接收者
- ✅ 便于添加新的监听器

## 🚀 扩展功能

### 添加新的ECU

```go
// 创建第二个ECU (例如：电池管理系统)
batteryECU := NewBatteryECU("BMS-ECU", canBus, 500*time.Millisecond)
batteryECU.Start()
```

### 添加CAN过滤器

```go
// 只订阅高优先级帧 (ID < 0x100)
type HighPriorityListener struct { ... }

func (hpl *HighPriorityListener) OnCANFrame(frame CANFrame) {
    if frame.ID < 0x100 {
        // 处理高优先级帧
    }
}
```

### 实现CAN仲裁

```go
// 当多个ECU同时发送时，优先发送ID小的帧
type ArbitrationCANBus struct {
    *CANBus
    pendingFrames []CANFrame
}

func (acb *ArbitrationCANBus) SendWithArbitration(frame CANFrame) {
    // 按ID排序，优先发送
}
```

## 📖 学习资源

- **CAN 2.0B规范**: ISO 11898标准
- **车辆网络**: 了解汽车内部网络架构
- **实时系统**: 学习实时通信和时序要求

## ❓ 常见问题

**Q: 为什么需要CAN总线？**
A: 真实车辆中，各个ECU通过CAN总线通信。使用总线架构可以更真实地模拟车辆通信。

**Q: 监听器会丢失帧吗？**
A: 如果监听器处理太慢，可能会丢失帧。建议在监听器中快速处理或使用缓冲。

**Q: 可以监听所有CAN ID吗？**
A: 可以，只需订阅所有CAN ID即可。DataLogger就是这样做的。

**Q: 如何添加新的信号？**
A: 在signals.go中添加信号定义，然后在VehicleECU中生成相应的CAN帧。

## 🎉 总结

CAN总线模拟系统提供了：
- ✅ 真实的CAN总线架构
- ✅ 发布-订阅通信模式
- ✅ 多监听器支持
- ✅ 实时信号处理
- ✅ 可扩展的架构

开始使用CAN总线，体验真正的车辆网络通信！

---

**创建时间**: 2026-03-06
**车型**: 大众ID.4 (2020+)
**CAN标准**: CAN 2.0B
