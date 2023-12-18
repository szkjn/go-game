package main

import (
	"embed"
	"image"
	_ "image/png"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/*
var assets embed.FS

var PlayerSprite = mustLoadImage("assets/player.png")

type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	width := PlayerSprite.Bounds().Dx()
	height := PlayerSprite.Bounds().Dy()
	halfW := float64(width / 2)
	halfH := float64(height / 2)

	op := &ebiten.DrawImageOptions{}
	// op.GeoM.Scale(2, 2)
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(45.0 * math.Pi / 180.0)
	op.GeoM.Translate(halfW, halfH)
	screen.DrawImage(PlayerSprite, op)
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
	g := &Game{}

	err := ebiten.RunGame(g)
	if err != nil {
		panic(err)
	}
}
