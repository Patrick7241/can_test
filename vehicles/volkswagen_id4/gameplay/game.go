package gameplay

import (
	"can_test/vehicles/volkswagen_id4/can"
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

/*
========================================
大众ID.4 CAN总线驾驶游戏 - 游戏逻辑
========================================

游戏核心逻辑：
- 游戏状态管理
- 碰撞检测
- 渲染系统
- 输入处理
*/

const (
	ScreenWidth  = 480
	ScreenHeight = 360
	roadWidth    = 200.0
	laneWidth    = roadWidth / 3.0
)

// GameState 游戏状态
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StateGameOver
)

// GameObject 游戏对象基类
type GameObject struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	VelX   float64
	VelY   float64
	Color  color.RGBA
}

// Bounds 获取边界矩形
func (obj *GameObject) Bounds() (x1, y1, x2, y2 float64) {
	return obj.X, obj.Y, obj.X + obj.Width, obj.Y + obj.Height
}

// Intersects 检测碰撞
func (obj *GameObject) Intersects(other *GameObject) bool {
	x1, y1, x2, y2 := obj.Bounds()
	ox1, oy1, ox2, oy2 := other.Bounds()

	return x1 < ox2 && x2 > ox1 && y1 < oy2 && y2 > oy1
}

// Player 玩家车辆
type Player struct {
	GameObject
	Lane int // 当前车道 (0, 1, 2)
}

// Obstacle 障碍物
type Obstacle struct {
	GameObject
	Type string // "car", "cone", "barrier"
}

// Game 游戏主结构
type Game struct {
	state       GameState
	player      *Player
	obstacles   []*Obstacle
	score       int
	gameTime    float64
	lastSpawn   float64
	spawnRate   float64
	canBus      *can.CANBus
	vehicleECU  *can.VehicleECU
	dashboard   *can.DashboardListener
	roadOffset  float64
	isPaused    bool
}

// New 创建游戏
func New() *Game {
	// 初始化CAN总线
	canBus := can.NewCANBus("VW-ID4-CAN")
	canBus.Start()

	// 创建车辆ECU
	vehicleECU := can.NewVehicleECU("ID4-ECU", canBus, 100*time.Millisecond)

	// 创建仪表盘监听器
	dashboard := can.NewDashboardListener("仪表盘")
	canBus.Subscribe(0xFD, dashboard)  // 车速
	canBus.Subscribe(0xB5, dashboard)  // 档位
	canBus.Subscribe(0x3EB, dashboard) // 刹车油门
	canBus.Subscribe(0x3DA, dashboard) // 转向

	// 启动ECU
	vehicleECU.Start()

	// 创建玩家
	player := &Player{
		GameObject: GameObject{
			X:      ScreenWidth/2 - 20,
			Y:      ScreenHeight - 120,
			Width:  40,
			Height: 70,
			Color:  color.RGBA{0, 150, 255, 255}, // 蓝色车
		},
		Lane: 1, // 中间车道
	}

	game := &Game{
		state:      StateMenu,
		player:     player,
		obstacles:  make([]*Obstacle, 0),
		score:      0,
		gameTime:   0,
		lastSpawn:  0,
		spawnRate:  1.5, // 每1.5秒生成一个障碍物
		canBus:     canBus,
		vehicleECU: vehicleECU,
		dashboard:  dashboard,
		roadOffset: 0,
		isPaused:   false,
	}

	return game
}

// Cleanup 清理资源
func (g *Game) Cleanup() {
	if g.vehicleECU != nil {
		g.vehicleECU.Stop()
	}
	if g.canBus != nil {
		g.canBus.Stop()
	}
}

// Update 游戏更新逻辑
func (g *Game) Update() error {
	switch g.state {
	case StateMenu:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = StatePlaying
			g.vehicleECU.State.Gear = 3 // D档
		}

	case StatePlaying:
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			g.isPaused = !g.isPaused
		}

		if g.isPaused {
			return nil
		}

		// 更新游戏时间
		g.gameTime += 1.0 / 60.0

		// 处理输入
		g.handleInput()

		// 更新玩家位置
		g.updatePlayer()

		// 生成障碍物
		if g.gameTime-g.lastSpawn > g.spawnRate {
			g.spawnObstacle()
			g.lastSpawn = g.gameTime
			// 随着时间增加难度
			if g.spawnRate > 0.5 {
				g.spawnRate -= 0.02
			}
		}

		// 更新障碍物
		g.updateObstacles()

		// 碰撞检测
		if g.checkCollisions() {
			g.state = StateGameOver
			g.vehicleECU.State.Gear = 0 // P档
		}

		// 更新分数
		g.score = int(g.gameTime * 10)

		// 更新道路滚动效果
		speed := g.vehicleECU.State.Speed
		if speed > 0 {
			g.roadOffset += speed * 0.5
			if g.roadOffset > 100 {
				g.roadOffset = 0
			}
		}

	case StateGameOver:
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			// 重新开始
			*g = *New()
			g.state = StatePlaying
			g.vehicleECU.State.Gear = 3
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
			return fmt.Errorf("quit game")
		}
	}

	return nil
}

// handleInput 处理输入
func (g *Game) handleInput() {
	// 加速
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.vehicleECU.HandleInput('w')
	}

	// 刹车
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.vehicleECU.HandleInput('s')
	}

	// 左转（切换车道）
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		if g.player.Lane > 0 {
			g.player.Lane--
			g.vehicleECU.HandleInput('a')
		}
	}

	// 右转（切换车道）
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		if g.player.Lane < 2 {
			g.player.Lane++
			g.vehicleECU.HandleInput('d')
		}
	}

	// 松开油门和刹车
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.vehicleECU.HandleInput(' ')
	}
}

// updatePlayer 更新玩家位置
func (g *Game) updatePlayer() {
	// 计算目标X位置
	roadLeft := ScreenWidth/2 - roadWidth/2
	targetX := roadLeft + float64(g.player.Lane)*laneWidth + laneWidth/2 - g.player.Width/2

	// 平滑移动到目标车道
	g.player.X += (targetX - g.player.X) * 0.2
}

// spawnObstacle 生成障碍物
func (g *Game) spawnObstacle() {
	lane := rand.Intn(3)
	roadLeft := ScreenWidth/2 - roadWidth/2
	x := roadLeft + float64(lane)*laneWidth + laneWidth/2 - 20

	obstacleType := []string{"car", "cone", "barrier"}[rand.Intn(3)]

	var width, height float64
	var clr color.RGBA

	switch obstacleType {
	case "car":
		width, height = 40, 70
		clr = color.RGBA{255, 50, 50, 255} // 红色车
	case "cone":
		width, height = 30, 30
		clr = color.RGBA{255, 165, 0, 255} // 橙色锥桶
	case "barrier":
		width, height = 100, 20
		clr = color.RGBA{150, 150, 150, 255} // 灰色路障
	}

	obstacle := &Obstacle{
		GameObject: GameObject{
			X:      x,
			Y:      -height,
			Width:  width,
			Height: height,
			VelY:   3.0 + g.gameTime*0.05, // 速度随时间增加
			Color:  clr,
		},
		Type: obstacleType,
	}

	g.obstacles = append(g.obstacles, obstacle)
}

// updateObstacles 更新障碍物
func (g *Game) updateObstacles() {
	// 移动障碍物
	for i := len(g.obstacles) - 1; i >= 0; i-- {
		obs := g.obstacles[i]
		obs.Y += obs.VelY

		// 移除屏幕外的障碍物
		if obs.Y > ScreenHeight {
			g.obstacles = append(g.obstacles[:i], g.obstacles[i+1:]...)
		}
	}
}

// checkCollisions 碰撞检测
func (g *Game) checkCollisions() bool {
	playerObj := &g.player.GameObject
	for _, obs := range g.obstacles {
		if playerObj.Intersects(&obs.GameObject) {
			return true
		}
	}
	return false
}

// Draw 绘制游戏
func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying:
		g.drawGame(screen)
	case StateGameOver:
		g.drawGameOver(screen)
	}
}

// drawMenu 绘制菜单
func (g *Game) drawMenu(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 20, 255})

	menuText := "大众ID.4 CAN总线驾驶游戏\n\n\nW/S - 油门/刹车\nA/D - 切换车道\n躲避障碍物\n\n\n按空格开始"
	ebitenutil.DebugPrintAt(screen, menuText, ScreenWidth/2-100, ScreenHeight/2-80)
}

// drawGame 绘制游戏画面
func (g *Game) drawGame(screen *ebiten.Image) {
	// 背景
	screen.Fill(color.RGBA{50, 150, 50, 255}) // 草地

	// 绘制道路
	g.drawRoad(screen)

	// 绘制障碍物
	for _, obs := range g.obstacles {
		g.drawObstacle(screen, obs)
	}

	// 绘制玩家
	g.drawPlayer(screen)

	// 绘制HUD（抬头显示）
	g.drawHUD(screen)

	// 暂停提示
	if g.isPaused {
		ebitenutil.DebugPrintAt(screen, "游戏暂停 - 按P继续", ScreenWidth/2-100, ScreenHeight/2)
	}
}

// drawRoad 绘制道路
func (g *Game) drawRoad(screen *ebiten.Image) {
	roadLeft := float32(ScreenWidth/2 - roadWidth/2)
	roadRight := float32(ScreenWidth/2 + roadWidth/2)

	// 道路底色
	vector.DrawFilledRect(screen, roadLeft, 0, float32(roadWidth), float32(ScreenHeight),
		color.RGBA{40, 40, 40, 255}, false)

	// 车道分隔线
	dashHeight := float32(40)
	dashGap := float32(60)
	offset := float32(g.roadOffset)

	for lane := 1; lane < 3; lane++ {
		lineX := roadLeft + float32(float64(lane)*laneWidth)
		for y := -offset; y < float32(ScreenHeight); y += dashHeight + dashGap {
			vector.DrawFilledRect(screen, lineX-2, y, 4, dashHeight,
				color.RGBA{255, 255, 255, 255}, false)
		}
	}

	// 道路边缘
	vector.DrawFilledRect(screen, roadLeft-10, 0, 10, float32(ScreenHeight),
		color.RGBA{200, 200, 200, 255}, false)
	vector.DrawFilledRect(screen, roadRight, 0, 10, float32(ScreenHeight),
		color.RGBA{200, 200, 200, 255}, false)
}

// drawPlayer 绘制玩家车辆
func (g *Game) drawPlayer(screen *ebiten.Image) {
	p := g.player

	// 简化：只绘制车身
	vector.DrawFilledRect(screen, float32(p.X), float32(p.Y),
		float32(p.Width), float32(p.Height), p.Color, false)
}

// drawObstacle 绘制障碍物
func (g *Game) drawObstacle(screen *ebiten.Image, obs *Obstacle) {
	// 简化：只绘制主体
	vector.DrawFilledRect(screen, float32(obs.X), float32(obs.Y),
		float32(obs.Width), float32(obs.Height), obs.Color, false)
}

// drawHUD 绘制HUD
func (g *Game) drawHUD(screen *ebiten.Image) {
	// CAN信号数据
	state := g.vehicleECU.State
	gearNames := []string{"P", "R", "N", "D"}
	gearName := "?"
	if state.Gear >= 0 && state.Gear < len(gearNames) {
		gearName = gearNames[state.Gear]
	}

	hudText := fmt.Sprintf(
		"CAN: %.0f km/h | %s档 | 油%.0f%% 刹%.0f%% | 分:%d",
		state.Speed,
		gearName,
		state.AcceleratorPos,
		state.BrakePosition,
		g.score,
	)

	ebitenutil.DebugPrintAt(screen, hudText, 10, 10)
}

// drawGameOver 绘制游戏结束画面
func (g *Game) drawGameOver(screen *ebiten.Image) {
	// 继续显示游戏画面
	g.drawGame(screen)

	// 游戏结束文字
	gameOverText := fmt.Sprintf(
		"\n\n游戏结束！\n\n分数: %d\n时间: %.1fs\n\n[R]重新开始 [Q]退出",
		g.score,
		g.gameTime,
	)

	ebitenutil.DebugPrintAt(screen, gameOverText, ScreenWidth/2-80, ScreenHeight/2-60)
}

// Layout 游戏布局
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
