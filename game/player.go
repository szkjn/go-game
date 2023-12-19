package game

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	shootCooldown     = time.Millisecond * 100
	rotationPerSecond = math.Pi
	bulletSpawnOffset = 50.0
)

type Player struct {
	game          *Game
	position      Vector
	rotation      float64
	sprite        *ebiten.Image
	shootCooldown *Timer
}

func NewPlayer(game *Game) *Player {
	sprite := PlayerSprite

	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	pos := Vector{
		X: ScreenWidth/2 - halfW,
		Y: ScreenHeight/2 - halfH,
	}

	return &Player{
		game:          game,
		position:      pos,
		rotation:      0,
		sprite:        sprite,
		shootCooldown: NewTimer(shootCooldown),
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

	p.shootCooldown.Update()
	if p.shootCooldown.IsReady() && ebiten.IsKeyPressed(ebiten.KeySpace) {
		p.shootCooldown.Reset()

		bounds := p.sprite.Bounds()
		halfW := float64(bounds.Dx()) / 2
		halfH := float64(bounds.Dy()) / 2

		spawnPos := Vector{
			p.position.X + halfW + math.Sin(p.rotation)*bulletSpawnOffset,
			p.position.Y + halfH + math.Cos(p.rotation)*bulletSpawnOffset,
		}

		bullet := NewBullet(spawnPos, p.rotation)
		p.game.AddBullet(bullet)
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

func (p *Player) Collider() Rect {
	bounds := p.sprite.Bounds()

	return NewRect(
		p.position.X,
		p.position.Y,
		float64(bounds.Dx()),
		float64(bounds.Dy()),
	)
}
