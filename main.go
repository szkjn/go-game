package main

import (
	"embed"
	"image"
	_ "image/png"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/*
var assets embed.FS

var PlayerSprite = mustLoadImage("assets/player.png")

type Vector struct {
	X float64
	Y float64
}

type Game struct {
	player *Player
}

type Timer struct {
	currentTicks int
	targetTicks  int
}

type Player struct {
	position Vector
	sprite   *ebiten.Image
}

func NewPlayer() *Player {
	return &Player{
		position: Vector{X: 100, Y: 100},
		sprite:   PlayerSprite,
	}
}

func (p *Player) Update() {

}

func (p *Player) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.position.X, p.position.Y)
	screen.DrawImage(p.sprite, op)
}

func NewTimer(d time.Duration) *Timer {
	return &Timer{
		currentTicks: 0,
		targetTicks:  int(d.Milliseconds()) * ebiten.TPS() / 1000,
	}
}

func (t *Timer) Update() {
	if t.currentTicks < t.targetTicks {
		t.currentTicks++
	}
}

func (t *Timer) IsReady() bool {
	return t.currentTicks >= t.targetTicks
}

func (t *Timer) Reset() {
	t.currentTicks = 0
}

func (g *Game) Update() error {
	g.player.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.player.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func mustLoadImage(name string) *ebiten.Image {
	f, err := assets.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}

func main() {
	g := &Game{
		playerPosition: Vector{X: 100, Y: 100},
		attackTimer:    NewTimer(5 * time.Second),
	}

	err := ebiten.RunGame(g)
	if err != nil {
		panic(err)
	}
}
