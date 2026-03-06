# Metal分配失败问题解决方案

## 问题原因

`[CAMetalLayer nextDrawable] returning nil because allocation failed` 表示GPU无法分配图形缓冲区。

可能原因：
1. 其他应用占用了太多GPU资源
2. GPU驱动问题
3. macOS版本兼容性（macOS 15 Sequoia的Metal API变化）

## 解决方案

### 方案1: 减少GPU占用（推荐）

1. **关闭其他GPU密集型应用**：
   - 关闭Chrome浏览器（占用大量GPU）
   - 关闭视频播放器
   - 关闭其他游戏

2. **使用已优化的版本**：
```bash
# 我已经优化了窗口大小和渲染
./game_demo
```

### 方案2: 使用环境变量

```bash
# 设置Metal环境变量
export METAL_DEVICE_WRAPPER_TYPE=1
export METAL_ERROR_MODE=0
./game_demo
```

或使用启动脚本：
```bash
./start_game.sh
```

### 方案3: 降低分辨率

如果还是不行，可以进一步降低分辨率。编辑 `game_demo.go`：

```go
const (
	screenWidth  = 400  // 原来是600
	screenHeight = 320  // 原来是480
	roadWidth    = 200.0  // 原来是300.0
	laneWidth    = roadWidth / 3.0
)
```

然后重新编译：
```bash
go build -o game_demo signals.go canbus.go vehicle_ecu.go can_listeners.go game_demo.go
```

### 方案4: 使用命令行版本

如果GPU问题无法解决，使用原来的命令行版本：

```bash
# CAN总线系统（无图形界面）
go run signals.go canbus.go vehicle_ecu.go can_listeners.go canbus_demo.go

# 或简化版模拟器
go run signals.go simple_demo.go
```

### 方案5: 检查系统GPU

```bash
# 查看GPU使用情况
system_profiler SPDisplaysDataType

# 查看Metal支持
system_profiler SPDisplaysDataType | grep -i metal
```

## 临时解决方案

如果以上都不行，可以：

1. **重启电脑** - 清理GPU缓存
2. **更新macOS** - 确保有最新的Metal驱动
3. **使用活动监视器** - 强制退出占用GPU的进程

## 已知问题

macOS 15 Sequoia中，Apple弃用了一些Metal API，ebiten可能需要更新。

如果问题持续，可以：
- 等待ebiten更新到支持新API的版本
- 或者使用命令行版本的CAN总线系统
