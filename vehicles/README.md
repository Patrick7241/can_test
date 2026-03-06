# 车辆信号数据库

本目录包含各车型的CAN信号定义、演示程序和原始数据文件。

## 📂 目录结构

```
vehicles/
├── README.md              # 本文件
└── volkswagen_id4/        # 大众ID.4 (2020+)
    ├── signals.go         # 信号定义
    ├── demo.go            # 演示程序
    ├── Volkswagen-ID.4 2020-.REF  # 原始数据
    └── README.md          # 详细文档
```

## 🚗 支持的车型

### 1. 大众ID.4 (2020+)

**车型信息:**
- 类型: 纯电动SUV
- 年款: 2020年至今
- 动力: 单电机/双电机
- 续航: 402-520公里（WLTP）

**信号数量:** 12个

**主要信号:**
- 车速显示（km/h和mph）
- 档位状态（P/R/N/D）
- 油门和刹车踏板位置
- 方向盘转角和方向
- 加速度传感器（横向/纵向）
- 横摆角速度
- 环境温度

**快速开始:**
```bash
cd vehicles/volkswagen_id4
go run *.go
```

📖 [查看详细文档](volkswagen_id4/README.md)

---

## 🎯 未来计划添加的车型

### 特斯拉 Model 3
- 纯电动轿车
- 大量传感器和辅助驾驶信号
- 预计信号数: 50+

### 比亚迪 汉EV
- 国产高端电动轿车
- 刀片电池相关信号
- 预计信号数: 40+

### 蔚来 ET5
- 智能电动轿车
- ADAS和智能座舱信号
- 预计信号数: 60+

### 理想 L9
- 增程式电动SUV
- 油电混合系统信号
- 预计信号数: 70+

### 宝马 i4
- 豪华电动轿车
- 传统车企的电动化案例
- 预计信号数: 80+

---

## 📋 车型选择标准

添加新车型时，我们考虑以下因素：

1. **代表性** - 不同技术路线的代表车型
2. **流行度** - 市场保有量和关注度
3. **信号丰富度** - 包含典型和特色信号
4. **数据可获取性** - 能够获取可靠的信号定义
5. **技术价值** - 对学习和研究有帮助

---

## 🔧 车型目录规范

每个车型目录应包含：

### 必需文件

1. **signals.go** - 信号定义
   ```go
   package main

   var [Brand][Model]Signals = map[string]CANSignalDef{
       "Signal_Name": { ... },
   }
   ```

2. **demo.go** - 演示程序
   - 展示如何使用信号定义
   - 包含编码/解码示例
   - 显示典型应用场景

3. **README.md** - 详细文档
   - 车型介绍
   - 信号列表和说明
   - API文档
   - 使用示例

4. **原始数据文件**（可选）
   - .REF文件
   - .DBC文件
   - 或其他格式

### 目录命名规范

格式: `品牌_车型/`

示例:
- `volkswagen_id4/` ✅
- `tesla_model3/` ✅
- `byd_han_ev/` ✅
- `nio_et5/` ✅

避免:
- `VW-ID4/` ❌（使用小写和全称）
- `Model3/` ❌（缺少品牌）
- `tesla-model-3/` ❌（使用下划线而非横杠）

### 信号变量命名规范

格式: `[Brand][Model]Signals`

示例:
- `ID4Signals` ✅
- `Model3Signals` ✅
- `HanEVSignals` ✅

---

## 📊 信号定义标准

### 信号结构

```go
type CANSignalDef struct {
    Name      string  // 信号名称（英文，使用下划线分隔）
    CANID     uint32  // CAN消息ID（十六进制）
    Unit      string  // 单位（km/h, °C, %, g等）
    StartBit  int     // 起始位
    BitLength int     // 数据长度
    MinValue  float64 // 最小值
    Scale     float64 // 缩放因子
    MaxValue  float64 // 最大值
    Offset    float64 // 偏移量
    Signed    bool    // 是否有符号
    ByteOrder string  // Intel或Motorola
    Desc      string  // 中文描述
}
```

### 信号命名规范

**推荐格式:**
- 使用英文
- 下划线分隔单词
- 首字母大写（Go导出规则）
- 描述性强

**示例:**
- `Vehicle_Speed_kph` ✅
- `Brake_Pedal_Position` ✅
- `Steering_Wheel_Angle` ✅
- `Battery_State_of_Charge` ✅

**避免:**
- `speed` ❌（太模糊）
- `VehSpd` ❌（缩写不清晰）
- `车速` ❌（使用英文）

### 必需函数

每个车型的signals.go应实现：

1. **DecodeSignal()** - 解码函数
2. **EncodeSignal()** - 编码函数
3. **PrintSignalInfo()** - 打印信号详情
4. **ListAllSignals()** - 列出所有信号

---

## 🚀 添加新车型指南

### 步骤1: 准备数据

1. 获取车辆CAN信号定义文件（.REF, .DBC等）
2. 使用`tools/`目录下的解析器提取信号
3. 整理信号列表和参数

### 步骤2: 创建目录

```bash
mkdir -p vehicles/brand_model
cd vehicles/brand_model
```

### 步骤3: 创建文件

```bash
# 复制模板（从volkswagen_id4/）
cp ../volkswagen_id4/signals.go .
cp ../volkswagen_id4/demo.go .
cp ../volkswagen_id4/README.md .

# 或者从零开始创建
touch signals.go demo.go README.md
```

### 步骤4: 编写signals.go

```go
package main

import "fmt"

var BrandModelSignals = map[string]CANSignalDef{
    "Signal_Name_1": {
        Name:      "Signal_Name_1",
        CANID:     0x123,
        Unit:      "km/h",
        StartBit:  0,
        BitLength: 16,
        MinValue:  0,
        Scale:     0.01,
        MaxValue:  655.35,
        Offset:    0,
        Signed:    false,
        ByteOrder: "Intel",
        Desc:      "信号描述",
    },
    // 更多信号...
}

// DecodeSignal, EncodeSignal等函数...
```

### 步骤5: 编写demo.go

展示如何使用信号定义，包括：
- 列出所有信号
- 解码示例
- 编码示例
- 典型应用场景

### 步骤6: 编写README.md

包含：
- 车型介绍
- 信号列表
- 详细的信号说明
- 使用示例
- API文档

### 步骤7: 测试

```bash
# 测试编译
go build

# 运行演示
go run *.go

# 检查输出是否正确
```

### 步骤8: 更新主文档

在项目根目录的README.md中添加新车型。

---

## 📝 贡献指南

### 提交新车型

1. Fork项目
2. 创建新分支: `git checkout -b add-tesla-model3`
3. 按照规范添加车型
4. 提交更改: `git commit -am 'Add Tesla Model 3 signals'`
5. 推送分支: `git push origin add-tesla-model3`
6. 创建Pull Request

### 审核标准

Pull Request需要满足：
- ✅ 遵循目录和命名规范
- ✅ 包含所有必需文件
- ✅ 代码格式化（`go fmt`）
- ✅ 文档完整清晰
- ✅ 演示程序可运行
- ✅ 信号定义准确

---

## 📚 资源

### 学习资源
- [CAN总线协议基础](https://www.can-cia.org/)
- [车辆网络架构](https://www.autosec.se/)
- [电动车技术](https://insideevs.com/)

### 工具
- Vector CANalyzer/CANoe
- Kvaser CAN工具
- PCAN硬件和软件

### 数据来源
- 车辆维修手册
- OBD-II数据库
- 开源CAN数据库
- 逆向工程

---

## ⚠️ 免责声明

1. 本数据库仅用于学习和研究目的
2. 信号定义来自公开可获取的数据
3. 实际应用请参考车辆官方技术文档
4. 不保证数据的完整性和准确性
5. 不承担因使用本数据导致的任何责任

---

## 📞 联系方式

- 提交Issue报告问题
- 发起Discussion讨论想法
- 提交Pull Request贡献代码

---

**最后更新:** 2026-03-06
