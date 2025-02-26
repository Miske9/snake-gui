package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	WindowHeight = 800
	WindowWidth  = 800
	ScreenWidth  = 400
	ScreenHeight = 200
	CellSize     = 10
	Cols         = ScreenWidth / CellSize
	Rows         = ScreenHeight / CellSize
	Lives        = 3
	Speed        = 6
)

type Point struct {
	x, y int
}

type Game struct {
	snake       []Point
	food        Point
	dir         Point
	score       int
	level       int
	lives       int
	gameOver    bool
	gamePassed  bool
	frameCount  int
	levelPassed bool
	obstacles   []Point
	MaxLevel    int
	levelTimer  time.Time
}

func (g *Game) Init() {
	g.snake = []Point{{Cols / 2, Rows / 2}}
	g.dir = Point{1, 0}
	g.food = g.placeFood()
	g.level = 1
	g.lives = Lives
	g.score = 0
	g.gameOver = false
	g.levelPassed = false
	g.gamePassed = false
	g.MaxLevel = 4
	g.obstacles = []Point{}
}

func (g *Game) placeFood() Point {
	for {
		x := rand.Intn(Cols)
		y := rand.Intn(Rows)
		if g.isCellOccupied(x, y) {
			continue
		}
		return Point{x, y}
	}
}

func (g *Game) isCellOccupied(x, y int) bool {
	for _, p := range g.snake {
		if p.x == x && p.y == y {
			return true
		}
	}
	for _, obs := range g.obstacles {
		if obs.x == x && obs.y == y {
			return true
		}
	}
	return false
}

func (g *Game) placeObstacles() {
	g.obstacles = []Point{}
	if g.level >= 3 {
		for i := 0; i < 5+g.level; i++ {
			x := rand.Intn(Cols)
			y := rand.Intn(Rows)
			if !g.isCellOccupied(x, y) {
				g.obstacles = append(g.obstacles, Point{x, y})
			}
		}
	}
}

func (g *Game) Update() error {
	if g.gameOver {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			g.Init()
		}
		return nil
	}
	if g.gamePassed {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			g.Init()
		}
		return nil
	}

	// Ako je level završen, čekaj 3 sekunde pre nego što nastavi
	if g.levelPassed {
		if time.Since(g.levelTimer).Seconds() > 3 {
			g.levelPassed = false
			g.food = g.placeFood()
			g.placeObstacles()
		}
		return nil
	}

	// Kontrola kretanja
	if ebiten.IsKeyPressed(ebiten.KeyW) && g.dir.y == 0 {
		g.dir = Point{0, -1}
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) && g.dir.y == 0 {
		g.dir = Point{0, 1}
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) && g.dir.x == 0 {
		g.dir = Point{-1, 0}
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) && g.dir.x == 0 {
		g.dir = Point{1, 0}
	}

	// Usporenje zmije
	g.frameCount++
	if g.frameCount%Speed != 0 {
		return nil
	}

	head := Point{g.snake[0].x + g.dir.x, g.snake[0].y + g.dir.y}

	// Sudar sa zidovima
	if head.x < 0 || head.x >= Cols || head.y < 0 || head.y >= Rows {
		g.lives--
		g.score = 0
		if g.lives == 0 {
			g.gameOver = true
		}
		g.snake = []Point{{Cols / 2, Rows / 2}}
		g.dir = Point{1, 0}
		return nil
	}

	// Sudar sa samim sobom
	for _, p := range g.snake[1:] {
		if p == head {
			g.lives--
			g.score = 0
			if g.lives == 0 {
				g.gameOver = true
			}
			g.snake = []Point{{Cols / 2, Rows / 2}}
			g.dir = Point{1, 0}
			return nil
		}
	}

	// Sudar sa preprekama
	for _, obs := range g.obstacles {
		if obs == head {
			g.lives--
			g.score = 0
			if g.lives == 0 {
				g.gameOver = true
			}
			g.snake = []Point{{Cols / 2, Rows / 2}}
			g.dir = Point{1, 0}
			return nil
		}
	}

	g.snake = append([]Point{head}, g.snake...)

	// Ako zmija pojede hranu
	if head == g.food {
		g.score++
		g.food = g.placeFood()

		if g.score >= g.level*1 {
			g.level++
			g.score = 0
			g.levelPassed = true
			g.levelTimer = time.Now()
			if g.level > g.MaxLevel {
				g.gamePassed = true
				g.levelPassed = false
				g.level = 4
			}
		}
	} else {
		g.snake = g.snake[:len(g.snake)-1]
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Crtanje okvira
	borderColor := color.RGBA{255, 255, 255, 255}
	for x := 0; x < ScreenWidth; x++ {
		screen.Set(x, 0, borderColor)
		screen.Set(x, ScreenHeight-1, borderColor)
	}
	for y := 0; y < ScreenHeight; y++ {
		screen.Set(0, y, borderColor)
		screen.Set(ScreenWidth-1, y, borderColor)
	}

	// Crtanje prepreka
	for _, obs := range g.obstacles {
		obstacleImg := ebiten.NewImage(CellSize, CellSize)
		obstacleImg.Fill(color.RGBA{128, 128, 128, 255})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(obs.x*CellSize), float64(obs.y*CellSize))
		screen.DrawImage(obstacleImg, op)
	}

	// Crtanje zmije
	for _, p := range g.snake {
		snakeImg := ebiten.NewImage(CellSize, CellSize)
		snakeImg.Fill(color.RGBA{0, 255, 0, 255})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(p.x*CellSize), float64(p.y*CellSize))
		screen.DrawImage(snakeImg, op)
	}

	// Crtanje hrane
	foodImg := ebiten.NewImage(CellSize, CellSize)
	foodImg.Fill(color.RGBA{255, 255, 0, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(g.food.x*CellSize), float64(g.food.y*CellSize))
	screen.DrawImage(foodImg, op)

	// Tekst
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Score: %d | Lives: %d | Level: %d", g.score, g.lives, g.level))
	if g.gameOver {
		ebitenutil.DebugPrint(screen, "\nGAME OVER! Press R to Restart")
	}
	if g.gamePassed {
		ebitenutil.DebugPrint(screen, "\nCONGRATULATIONS! You have passed all levels!")
	}
	if g.levelPassed {
		ebitenutil.DebugPrint(screen, "\nLEVEL PASSED! Starting next level...")
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	rand.Seed(time.Now().UnixNano())
	game := &Game{}
	game.Init()
	ebiten.SetWindowSize(WindowWidth, WindowHeight)
	ebiten.SetWindowTitle("Snake Game")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
