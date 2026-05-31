package core

import "github.com/hajimehoshi/ebiten/v2"

// Screen is implemented by every game screen.
type Screen interface {
	Update(g *Game) error
	Draw(dst *ebiten.Image)
}
