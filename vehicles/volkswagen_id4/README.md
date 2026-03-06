# 大众ID.4 CAN总线驾驶游戏

带真实游戏画面的CAN总线驾驶模拟器。

## 📁 项目结构

```
volkswagen_id4/
├── main.go          # 游戏主入口（唯一的main函数）
├── game.go          # 游戏核心逻辑
├── signals.go       # CAN信号定义
├── canbus.go        # CAN总线实现
├── ecu.go           # 车辆ECU
├── listeners.go     # 监听器实现
├── start_game.sh    # 启动脚本
├── docs/            # 文档目录
│   ├── GAME.md
│   ├── CANBUS.md
│   └── TROUBLESHOOTING.md
└── README.md
```

## 🎮 快速开始

### 方法1：使用启动脚本（推荐）

```bash
cd vehicles/volkswagen_id4
./start_game.sh
```

启动脚本会自动：
- 检测Chrome是否运行并询问是否关闭
- 编译游戏（如果需要）
- 启动游戏

### 方法2：手动编译运行

```bash
cd vehicles/volkswagen_id4

# 编译
go build -o game *.go

# 运行
./game
```

### 方法3：直接运行源码

```bash
go run *.go
```

## 🎯 游戏特性

### 真实游戏画面
- ✅ 俯视图驾驶视角
- ✅ 三车道系统，躲避障碍物
- ✅ 碰撞检测，撞到就输
- ✅ 难度随时间递增
- ✅ 流畅的60 FPS

### CAN总线集成
- ✅ 实时CAN信号生成和传输
- ✅ 车辆ECU每100ms发送CAN帧
- ✅ 监听器实时接收和解析信号
- ✅ HUD显示车速、档位、油门、刹车等

### 三种障碍物
- 🚗 红色车辆
- 🔶 橙色锥桶
- ⚠️  灰色路障

## 🎮 操作说明

**主菜单**:
- `空格` - 开始游戏

**游戏中**:
- `W` - 加速（油门增加）
- `S` - 刹车（刹车增加）
- `A` - 切换到左车道
- `D` - 切换到右车道
- `空格` - 松开油门和刹车
- `P` - 暂停/继续

**游戏结束**:
- `R` - 重新开始
- `Q` - 退出游戏

## ⚠️ 故障排查

如果游戏窗口黑屏或报Metal错误：

1. **关闭Chrome浏览器**（占用大量GPU资源）
2. **关闭其他GPU密集型应用**
3. **重启电脑**（清理GPU缓存）
4. 查看 [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) 获取详细解决方案

游戏已优化为 **480x360** 分辨率以减少GPU占用。

## 📖 详细文档

- [游戏完整教程](docs/GAME.md) - 玩法、技巧、技术细节
- [CAN总线文档](docs/CANBUS.md) - CAN总线系统架构
- [故障排查](docs/TROUBLESHOOTING.md) - 常见问题解决

## 📊 CAN信号

游戏使用了大众ID.4的12个真实CAN信号：

- 车速 (0xFD)
- 档位 (0xB5)
- 刹车/油门 (0x3EB)
- 转向角 (0x3DA)
- 加速度/横摆 (0x101)
- 环境温度 (0x5E1)

详细信号定义请查看 [docs/CANBUS.md](docs/CANBUS.md)

---

**创建时间**: 2026-03-06
**引擎**: Ebiten v2.9.8
**车型**: 大众ID.4 (2020+)
**许可**: 仅供学习和研究使用


