package main

import (
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

// ------------------------------------------------------
// MAIN
// ------------------------------------------------------

func main() {
	g := NewGame()

	err := ebiten.RunGame(g)
	if err != nil {
		panic(err)
	}
}
