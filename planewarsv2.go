package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	winWidth  = 800
	winHeight = 600
)

type gameState struct {
	playerTexture *sdl.Texture
	playerPos     sdl.Rect
	bullets       []sdl.Rect
	enemies       []enemy
	score         int
}

type enemy struct {
	texture *sdl.Texture
	pos     sdl.Rect
	speed   int
}

func initialize() (*sdl.Window, *sdl.Renderer, *ttf.Font, error) {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("初始化SDL失败：%v", err)
	}

	window, err := sdl.CreateWindow(
		"飞机大战",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建窗口失败：%v", err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建渲染器失败：%v", err)
	}

	err = ttf.Init()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("初始化TTF失败：%v", err)
	}

	font, err := ttf.OpenFont("assets/font.ttf", 24)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("加载字体失败：%v", err)
	}

	return window, renderer, font, nil
}

func loadTexture(renderer *sdl.Renderer, imagePath string) (*sdl.Texture, error) {
	img, err := sdl.LoadBMP(imagePath)
	if err != nil {
		return nil, fmt.Errorf("加载纹理失败：%v", err)
	}
	defer img.Free()

	texture, err := renderer.CreateTextureFromSurface(img)
	if err != nil {
		return nil, fmt.Errorf("创建纹理失败：%v", err)
	}

	return texture, nil
}

func createEnemy(renderer *sdl.Renderer) (enemy, error) {
	enemyTexture, err := loadTexture(renderer, "assets/enemy.bmp")
	if err != nil {
		return enemy{}, err
	}

	rand.Seed(time.Now().UnixNano())
	x := rand.Intn(winWidth)
	y := rand.Intn(winHeight/2) - winHeight
	speed := rand.Intn(5) + 1 // 随机速度，1~5

	return enemy{
		texture: enemyTexture,
		pos: sdl.Rect{
			X: int32(x),
			Y: int32(y),
			W: 50,
			H: 50,
		},
		speed: speed,
	}, nil
}

func handleEvents(gameState *gameState) {
    for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
        switch event.(type) {
        case *sdl.QuitEvent:
            os.Exit(0)
        case *sdl.KeyboardEvent:
            keyEvent := event.(*sdl.KeyboardEvent)
            if keyEvent.Type == sdl.KEYDOWN {
                switch keyEvent.Keysym.Scancode {
                case sdl.SCANCODE_LEFT:
                    gameState.playerPos.X -= 50
                case sdl.SCANCODE_RIGHT:
                    gameState.playerPos.X += 50
                case sdl.SCANCODE_SPACE:
                    gameState.bullets = append(gameState.bullets, sdl.Rect{
                        X: gameState.playerPos.X + gameState.playerPos.W/2 - 5,
                        Y: gameState.playerPos.Y,
                        W: 10,
                        H: 20,
                    })
                }
            }
        }
    }
}


func runGame(window *sdl.Window, renderer *sdl.Renderer, font *ttf.Font) error {
	playerTexture, err := loadTexture(renderer, "assets/player.bmp")
	if err != nil {
		return err
	}
	defer playerTexture.Destroy()

	bulletTexture, err := loadTexture(renderer, "assets/bullet.bmp")
	if err != nil {
		return err
	}
	defer bulletTexture.Destroy()

	gameState := gameState{
		playerTexture: playerTexture,
		playerPos: sdl.Rect{
			X: winWidth/2 - 50,
			Y: winHeight - 70,
			W: 70,
			H: 70,
		},
		bullets: make([]sdl.Rect, 0),
		enemies: make([]enemy, 0),
		score:   0,
	}

	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	for {
		
        handleEvents(&gameState) // 处理玩家输入
		// 添加新的敌机
		if len(gameState.enemies) < 5 { // 控制敌机数量
			enemy, err := createEnemy(renderer)
			if err != nil {
				return err
			}
			gameState.enemies = append(gameState.enemies, enemy)
		}

		// 更新游戏状态
		// 更新子弹位置并移除越界的子弹
		updatedBullets := make([]sdl.Rect, 0, len(gameState.bullets))
		for _, bullet := range gameState.bullets {
			bullet.Y -= 10
			if bullet.Y >= 0 {
				updatedBullets = append(updatedBullets, bullet)
			}
		}
		gameState.bullets = updatedBullets

		for i := 0; i < len(gameState.enemies); {
			gameState.enemies[i].pos.Y += int32(gameState.enemies[i].speed)
			if gameState.enemies[i].pos.Y > winHeight {
				gameState.enemies = append(gameState.enemies[:i], gameState.enemies[i+1:]...)
			} else {
				i++
			}
		}

		// 碰撞检测
		for i := 0; i < len(gameState.bullets); i++ {
			hitEnemy := false
			for j := 0; j < len(gameState.enemies); j++ {
				if checkCollision(gameState.bullets[i], gameState.enemies[j].pos) {
					gameState.score++
					gameState.enemies = append(gameState.enemies[:j], gameState.enemies[j+1:]...)
					hitEnemy = true
					break
				}
			}
			if hitEnemy {
				gameState.bullets = append(gameState.bullets[:i], gameState.bullets[i+1:]...)
				i-- // 因为删除了一个子弹，所以要减少 i
			}
		}

		for i := 0; i < len(gameState.enemies); {
			gameState.enemies[i].pos.Y += int32(gameState.enemies[i].speed)
			if gameState.enemies[i].pos.Y > winHeight {
				gameState.enemies = append(gameState.enemies[:i], gameState.enemies[i+1:]...)
			} else {
				i++
			}
		}

		// 设置绘制颜色为白色
		renderer.SetDrawColor(255, 255, 255, 255)
		// 清空渲染器，以白色填充整个窗口
		renderer.Clear()

		for _, bullet := range gameState.bullets {
			renderer.Copy(bulletTexture, nil, &bullet)
		}

		for _, enemy := range gameState.enemies {
			renderer.Copy(enemy.texture, nil, &enemy.pos)
		}

		renderer.Copy(gameState.playerTexture, nil, &gameState.playerPos)

		// 绘制分数
		scoreText := fmt.Sprintf("Score: %d", gameState.score)
		color := sdl.Color{R: 255, G: 86, B: 2, A: 255}
		surface, err := font.RenderUTF8Solid(scoreText, color)
		if err != nil {
			return fmt.Errorf("渲染文本表面失败：%v", err)
		}
		defer surface.Free()
		texture, err := renderer.CreateTextureFromSurface(surface)
		if err != nil {
			return fmt.Errorf("创建纹理失败：%v", err)
		}
		defer texture.Destroy()
		renderer.Copy(texture, nil, &sdl.Rect{X: 10, Y: 10, W: 100, H: 30})

		renderer.Present()
		<-ticker.C
	}
}

func checkCollision(rect1 sdl.Rect, rect2 sdl.Rect) bool {
	return rect1.X < rect2.X+rect2.W &&
		rect1.X+rect1.W > rect2.X &&
		rect1.Y < rect2.Y+rect2.H &&
		rect1.Y+rect1.H > rect2.Y
}

func main() {
	window, renderer, font, err := initialize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化失败：%v\n", err)
		os.Exit(1)
	}
	defer window.Destroy()
	defer renderer.Destroy()

	err = runGame(window, renderer, font)
	if err != nil {
		fmt.Fprintf(os.Stderr, "游戏运行失败：%v\n", err)
		os.Exit(1)
	}
}

