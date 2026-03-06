package main

import (
	"fmt"
	"math/rand"
	"time"

	"can_test/vehicles/volkswagen_id4/gameplay"

	"github.com/hajimehoshi/ebiten/v2"
)

/*
========================================
大众ID.4 CAN总线驾驶游戏 - 主入口
========================================

真实的游戏画面 + CAN总线通信
*/

func main() {
	rand.Seed(time.Now().UnixNano())

	g := gameplay.New()

	ebiten.SetWindowSize(gameplay.ScreenWidth, gameplay.ScreenHeight)
	ebiten.SetWindowTitle("🚗 大众ID.4 CAN总线驾驶游戏")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetMaxTPS(60) // 限制更新频率

	if err := ebiten.RunGame(g); err != nil {
		if err.Error() != "quit game" {
			panic(err)
		}
	}

	// 清理资源
	g.Cleanup()
	fmt.Println("游戏已退出")
}
