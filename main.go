package main

import (
	"embed"
	"image"
	_ "image/png"
	"io/fs"
	"math"
	"math/rand"
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

func mustLoadImages(path string) []*ebiten.Image {
	matches, err := fs.Glob(assets, path)
	if err != nil {
		panic(err)
	}

	images := make([]*ebiten.Image, len(matches))
	for i, match := range matches {
		images[i] = mustLoadImage(match)
	}

	return images
}

// ------------------------------------------------------
// VECTOR
// ------------------------------------------------------

type Vector struct {
	X float64
	Y float64
}

func (v Vector) Normalize() Vector {
	magnitude := math.Sqrt(v.X*v.X + v.Y*v.Y)
	return Vector{v.X / magnitude, v.Y / magnitude}
}

// ------------------------------------------------------
// BULLET
// ------------------------------------------------------
const (
	bulletSpeedPerSecond = 350.0
)

var LaserSprite = mustLoadImage("assets/laser.png")

type Bullet struct {
	position Vector
	rotation float64
	sprite   *ebiten.Image
}

func NewBullet(pos Vector, rotation float64) *Bullet {
	sprite := LaserSprite

	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	pos.X -= halfW
	pos.Y -= halfH
	b := &Bullet{
		position: pos,
		rotation: rotation,
		sprite:   sprite,
	}

	return b
}

func (b *Bullet) Update() {
	speed := bulletSpeedPerSecond / float64(ebiten.TPS())

	b.position.X += math.Sin(b.rotation) * speed
	b.position.Y += math.Cos(b.rotation) * -speed
}

func (b *Bullet) Draw(screen *ebiten.Image) {
	bounds := b.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(b.rotation)
	op.GeoM.Translate(halfW, halfH)

	op.GeoM.Translate(b.position.X, b.position.Y)

	screen.DrawImage(b.sprite, op)
}

func (b *Bullet) Collider() Rect {
	bounds := b.sprite.Bounds()

	return NewRect(
		b.position.X,
		b.position.Y,
		float64(bounds.Dx()),
		float64(bounds.Dy()),
	)
}

// ------------------------------------------------------
// METEOR
// ------------------------------------------------------
const (
	rotationSpeedMin = -0.02
	rotationSpeedMax = 0.02
)

var MeteorSprites = mustLoadImages("assets/meteors/*.png")

type Meteor struct {
	position      Vector
	movement      Vector
	rotation      float64
	rotationSpeed float64
	sprite        *ebiten.Image
}

func NewMeteor() *Meteor {
	target := Vector{
		X: ScreenWidth / 2,
		Y: ScreenHeight / 2,
	}

	r := ScreenWidth / 2.0
	angle := rand.Float64() * 2 * math.Pi

	pos := Vector{
		X: target.X + math.Cos(angle)*r,
		Y: target.Y + math.Sin(angle)*r,
	}

	velocity := 0.25 + rand.Float64()*1.5
	rotationSpeed := rotationSpeedMin + rand.Float64()*(rotationSpeedMax-rotationSpeedMin)

	direction := Vector{
		X: target.X - pos.X,
		Y: target.Y - pos.Y,
	}

	normalizedDirection := direction.Normalize()

	movement := Vector{
		X: normalizedDirection.X * velocity,
		Y: normalizedDirection.Y * velocity,
	}

	if len(MeteorSprites) == 0 {
		panic("no meteor sprites loaded")
	}
	sprite := MeteorSprites[rand.Intn(len(MeteorSprites))]

	return &Meteor{
		position: pos,
		movement: movement,
		rotation: rotationSpeed,
		sprite:   sprite,
	}
}

func (m *Meteor) Update() {
	m.position.X += m.movement.X
	m.position.Y += m.movement.Y
	m.rotation += m.rotationSpeed
}

func (m *Meteor) Draw(screen *ebiten.Image) {
	bounds := m.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	// op.GeoM.Rotate(m.rotation)
	op.GeoM.Translate(halfW, halfH)

	op.GeoM.Translate(m.position.X, m.position.Y)

	screen.DrawImage(m.sprite, op)
}

func (m *Meteor) Collider() Rect {
	bounds := m.sprite.Bounds()

	return NewRect(
		m.position.X,
		m.position.Y,
		float64(bounds.Dx()),
		float64(bounds.Dy()),
	)
}

// ------------------------------------------------------
// PLAYER
// ------------------------------------------------------

var PlayerSprite = mustLoadImage("assets/player.png")

const (
	shootCooldown     = time.Millisecond * 500
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

// ------------------------------------------------------
// RECT
// ------------------------------------------------------
type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

func NewRect(x, y, width, height float64) Rect {
	return Rect{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

func (r Rect) MaxX() float64 {
	return r.X + r.Width
}

func (r Rect) MaxY() float64 {
	return r.Y + r.Height
}

func (r Rect) Intersects(other Rect) bool {
	return r.X <= other.MaxX() && other.X <= r.MaxX() && r.Y <= other.MaxY() && other.Y <= r.MaxY()
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
	bullets          []*Bullet
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

	for _, b := range g.bullets {
		b.Update()
	}

	for i, m := range g.meteors {
		for j, b := range g.bullets {
			if m.Collider().Intersects(b.Collider()) {
				g.meteors = append(g.meteors[:i], g.meteors[i+1:]...)
				g.bullets = append(g.bullets[:j], g.bullets[j+1:]...)
			}
		}
	}

	for _, m := range g.meteors {
		if m.Collider().Intersects(g.player.Collider()) {
			g.Reset()
			break
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.player.Draw(screen)

	for _, m := range g.meteors {
		m.Draw(screen)
	}

	for _, b := range g.bullets {
		b.Draw(screen)
	}
}

func (g *Game) AddBullet(b *Bullet) {
	g.bullets = append(g.bullets, b)
}
func (g *Game) Reset() {
	g.player = NewPlayer(g)
	g.meteors = nil
	g.bullets = nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
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
		meteorSpawnTimer: NewTimer(20),
	}

	g.player = NewPlayer(g)

	err := ebiten.RunGame(g)
	if err != nil {
		panic(err)
	}
}
