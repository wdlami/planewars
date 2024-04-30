package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/eiannone/keyboard"
	"github.com/fogleman/gg"
)

const (
	screenWidth  = 800
	screenHeight = 600
	planeSize    = 50
	enemySize    = 30
	bulletSize   = 5
	bulletSpeed  = 5
)

type Player struct {
	X, Y int
}

type Enemy struct {
	X, Y int
}

type Bullet struct {
	X, Y int
}

var (
	player      Player
	enemies     []Enemy
	bullets     []Bullet
	score       int
	gameOver    bool
	startButton *widget.Button
	infoLabel   *widget.Label
	canvasObj   fyne.CanvasObject
)

func main() {
	myApp := app.New()
	win := myApp.NewWindow("Plane Wars")

	player = Player{X: screenWidth / 2, Y: screenHeight - planeSize}

	canvasObj = canvas.NewRaster(drawGame)
	canvasObj.Resize(fyne.NewSize(screenWidth, screenHeight))

	startButton = widget.NewButton("Start Game", func() {
		startButton.Hide()
		infoLabel.Hide()
		go gameLoop(win)
	})
	startButton.Importance = widget.HighImportance

	infoLabel = widget.NewLabel("Use arrow keys to move and Space to shoot")

	win.SetContent(container.NewVBox(
		canvasObj,
		container.NewHBox(
			startButton,
			infoLabel,
		),
	))

	win.Resize(fyne.NewSize(screenWidth, screenHeight))
	win.ShowAndRun()
}

func gameLoop(win fyne.Window) {
	// 初始化游戏
	initGame()

	// 创建一个通道用于接收游戏结束信号
	gameOverCh := make(chan struct{})

	// 输出游戏循环开始的消息
	fmt.Println("Game loop started...")

	// 启动按键监听的 goroutine
	go func() {
		// 开始键盘监听
		if err := keyboard.Open(); err != nil {
			fmt.Println("Could not start keyboard input:", err)
			return
		}
		defer keyboard.Close()

		for {
			// 监听键盘事件
			_, key, err := keyboard.GetKey()
			if err != nil {
				fmt.Println("Error getting keyboard input:", err)
				continue
			}

			// 输出按键事件
			fmt.Println("Key pressed:", key)

			// 更新游戏状态
			updateGame(key)

			// 如果游戏结束，向通道发送信号
			if gameOver {
				close(gameOverCh)
				return
			}
		}
	}()

	// 游戏循环
	for !gameOver {
		// 更新画布内容
		canvas.Refresh(canvasObj)

		// 调试输出
		fmt.Println("Game loop running...")

		// 等待一段时间
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("Game Over! Your score:", score)
}

func initGame() {
	player = Player{X: screenWidth / 2, Y: screenHeight - planeSize}
	enemies = nil
	bullets = nil
	score = 0
	gameOver = false
}

func updateGame(key keyboard.Key) {
	// 根据按键事件更新游戏状态
	switch key {
	case keyboard.KeyArrowLeft:
		if player.X > 0 {
			player.X -= 5
		}
	case keyboard.KeyArrowRight:
		if player.X < screenWidth-planeSize {
			player.X += 5
		}
	case keyboard.KeySpace:
		bullets = append(bullets, Bullet{X: player.X + planeSize/2, Y: player.Y})
	}

	// 移动敌人
	for i := range enemies {
		enemies[i].Y += 2
	}

	// 移动子弹并检查碰撞
	for i := 0; i < len(bullets); i++ {
		bullets[i].Y -= bulletSpeed
		for j := 0; j < len(enemies); j++ {
			if distance(enemies[j].X, enemies[j].Y, bullets[i].X, bullets[i].Y) < enemySize {
				score++
				enemies = append(enemies[:j], enemies[j+1:]...)
				bullets = append(bullets[:i], bullets[i+1:]...)
			}
		}
	}

	// 生成新的敌人
	if len(enemies) < 5 {
		enemies = append(enemies, Enemy{X: random(0, screenWidth-enemySize), Y: -enemySize})
	}

	// 检查玩家与敌人的碰撞
	for _, enemy := range enemies {
		if distance(player.X, player.Y, enemy.X, enemy.Y) < enemySize {
			gameOver = true
			break
		}
	}
}

func drawGame(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	ctx := gg.NewContextForRGBA(img)

	// 清空画面
	ctx.SetColor(color.Black)
	ctx.Clear()

	// 绘制玩家飞机
	ctx.SetColor(color.White)
	ctx.DrawRectangle(float64(player.X), float64(player.Y), planeSize, planeSize)
	ctx.Fill()

	// 绘制敌人
	ctx.SetColor(color.RGBA{128, 128, 128, 255}) // 使用灰色绘制敌人
	for _, enemy := range enemies {
		ctx.DrawRectangle(float64(enemy.X), float64(enemy.Y), enemySize, enemySize)
		ctx.Fill()
	}

	// 绘制子弹
	ctx.SetColor(color.RGBA{255, 255, 0, 255})
	for _, bullet := range bullets {
		ctx.DrawRectangle(float64(bullet.X), float64(bullet.Y), bulletSize, bulletSize)
		ctx.Fill()
	}

	// 显示得分
	ctx.SetColor(color.White)
	ctx.DrawString(fmt.Sprintf("Score: %d", score), 10, 20)

	// 如果游戏结束，显示提示
	if gameOver {
		ctx.DrawString("Game Over!", screenWidth/2-50, screenHeight/2)
	}

	return img
}

func random(min, max int) int {
	return min + rand.Intn(max-min)
}

func distance(x1, y1, x2, y2 int) float64 {
	dx := float64(x1 - x2)
	dy := float64(y1 - y2)
	return math.Sqrt(dx*dx + dy*dy)
}
