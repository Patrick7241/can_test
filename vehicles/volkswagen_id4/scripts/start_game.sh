#!/bin/bash
# 大众ID.4 CAN总线驾驶游戏启动脚本

echo "=========================================="
echo "  🎮 大众ID.4 CAN总线驾驶游戏"
echo "=========================================="
echo ""

# 检查Chrome是否在运行
if pgrep -x "Google Chrome" > /dev/null; then
    echo "⚠️  警告：Chrome浏览器正在运行！"
    echo ""
    echo "Chrome占用了GPU资源，游戏可能无法正常显示。"
    echo ""
    read -p "是否关闭Chrome并启动游戏？(y/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "正在关闭Chrome..."
        osascript -e 'quit app "Google Chrome"'
        sleep 2
        echo "Chrome已关闭"
    else
        echo "游戏将使用软件渲染模式（可能较慢）"
    fi
    echo ""
fi

# 切换到项目根目录
cd "$(dirname "$0")/.."

# 编译游戏（如果需要）
if [ ! -f "./game" ] || [ "main.go" -nt "./game" ]; then
    echo "正在编译游戏..."
    go build -o game .
    if [ $? -ne 0 ]; then
        echo "❌ 编译失败"
        exit 1
    fi
    echo "✅ 编译成功"
    echo ""
fi

echo "正在启动游戏..."
echo ""

# 强制使用软件渲染
export EBITEN_GRAPHICS_LIBRARY=opengl

./game

echo ""
echo "游戏已退出"

