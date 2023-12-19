package main

import (
	"embed"
	"image"
	_ "image/png"
	"math"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// ------------------------------------------------------
// GLOBAL
// ------------------------------------------------------

//go:embed assets/*
var assets embed.FS

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

func mustLoadImages(pattern string) []*ebiten.Image {
	files, err := assets.ReadDir(".")
	if err != nil {
		panic(err)
	}

	var images []*ebiten.Image
	for _, file := range files {
		match, _ := filepath.Match(pattern, file.Name())
		if match {
			f, err := assets.Open(file.Name())
			if err != nil {
				panic(err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				panic(err)
			}

			images = append(images, ebiten.NewImageFromImage(img))
		}
	}
	return images
}

// ------------------------------------------------------
// METEOR
// ------------------------------------------------------
var MeteorSprites = mustLoadImages("assets/meteors/*.png")

type Meteor struct {
	position Vector
	sprite   *ebiten.Image
}

func NewMeteor() *Meteor {
	sprite := MeteorSprites[rand.Intn(len(MeteorSprites))]

	return &Meteor{
		position: Vector{},
		sprite:   sprite,
	}
}

func (m *Meteor) Update() {

}

func (m *Meteor) Draw(screen *ebiten.Image) {

}

// ------------------------------------------------------
// PLAYER
// ------------------------------------------------------

var PlayerSprite = mustLoadImage("assets/player.png")

type Vector struct {
	X float64
	Y float64
}

type Player struct {
	position Vector
	sprite   *ebiten.Image
	rotation float64
}

func NewPlayer() *Player {
	sprite := PlayerSprite

	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	pos := Vector{
		X: ScreenWidth/2 - halfW,
		Y: ScreenHeight/2 - halfH,
	}

	return &Player{
		position: pos,
		sprite:   sprite,
	}
}

func (p *Player) Update() {
	speed := math.Pi / float64(ebiten.TPS())

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.rotation -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.rotation += speed
	}
}

func (p *Player) Draw(screen *ebiten.Image) {
	bounds := p.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(p.rotation)
	op.GeoM.Translate(halfW, halfH)

	op.GeoM.Translate(p.position.X, p.position.Y)

	screen.DrawImage(p.sprite, op)
}

// ------------------------------------------------------
// GAME
// ------------------------------------------------------

const (
	ScreenWidth  = 800
	ScreenHeight = 600
)

type Game struct {
	player           *Player
	meteorSpawnTimer *Timer
	meteors          []*Meteor
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) Update() error {
	g.player.Update()

	g.meteorSpawnTimer.Update()
	if g.meteorSpawnTimer.IsReady() {
		g.meteorSpawnTimer.Reset()

		m := NewMeteor()
		g.meteors = append(g.meteors, m)
	}

	for _, m := range g.meteors {
		m.Update()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.player.Draw(screen)

	for _, m := range g.meteors {
		m.Draw(screen)
	}
}

// ------------------------------------------------------
// TIMER
// ------------------------------------------------------

type Timer struct {
	currentTicks int
	targetTicks  int
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

// ------------------------------------------------------
// MAIN
// ------------------------------------------------------

func main() {
	g := &Game{
		player: NewPlayer(),
	}

	err := ebiten.RunGame(g)
	if err != nil {
		panic(err)
	}
}
