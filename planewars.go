package main

import (
    // "os"
    "fmt"
    // "image"
    "image/color"
    "time"
    "math/rand"
    "math"

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
    keyboardErr error
    keyEventChan chan keyboard.Key
)

func main() {
    // 初始化游戏
    initGame()

    // 开始键盘监听
    if keyboardErr = keyboard.Open(); keyboardErr != nil {
        fmt.Println("Could not start keyboard input:", keyboardErr)
        return
    }
    defer keyboard.Close()

    keyEventChan = make(chan keyboard.Key)

    go func() {
        for {
            // 监听键盘事件
            if _, key, err := keyboard.GetKey(); err == nil {
                // 发送按键事件到通道
                fmt.Println("Received key:", key)
                keyEventChan <- key
            } else {
                fmt.Println("Error:", err)
            }
        }
    }()

    // 创建窗口
    ctx := gg.NewContext(screenWidth, screenHeight)

    // 游戏主循环
    fmt.Println("Entering game loop...")
    for !gameOver {
        fmt.Println("Updating game...")
        // 检查键盘事件
        select {
        case key := <-keyEventChan:
            updateGame(key)
        default:
            updateGame(0) // 添加默认分支
            time.Sleep(10 * time.Millisecond) // 休眠一段时间
        }

        // 渲染游戏画面
        drawGame(ctx)

        // 保存画面为 PNG 图像
        saveImage(ctx, "output.png")
        time.Sleep(time.Millisecond * 10)
    }

    // 游戏结束
    fmt.Println("Game Over! Your score:", score)
}

func initGame() {
    player = Player{X: screenWidth / 2, Y: screenHeight - planeSize}
    score = 0
    gameOver = false
}

func updateGame(key keyboard.Key) {
    // 根据按键事件更新游戏状态
    if key == keyboard.KeyArrowLeft && player.X > 0 {
        player.X -= 5
    }
    if key == keyboard.KeyArrowRight && player.X < screenWidth-planeSize {
        player.X += 5
    }
    if key == keyboard.KeySpace {
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


func drawGame(ctx *gg.Context) {
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
    ctx.DrawString(fmt.Sprintf("Score: %d", score), 10, 20)

    // 如果游戏结束，显示提示
    if gameOver {
        ctx.DrawString("Game Over!", screenWidth/2-50, screenHeight/2)
    }
}

func saveImage(ctx *gg.Context, filename string) {
    err := ctx.SavePNG(filename)
    if err != nil {
        fmt.Println("Error saving image:", err)
    }
}

func random(min, max int) int {
    return min + rand.Intn(max-min)
}

func distance(x1, y1, x2, y2 int) float64 {
    dx := float64(x1 - x2)
    dy := float64(y1 - y2)
    return math.Sqrt(dx*dx + dy*dy)
}
