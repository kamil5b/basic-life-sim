package core

import (
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenW = 1100
	ScreenH = 700
)

// Game is the root Ebitengine game object.
type Game struct {
	current Screen
}

func newGame() *Game {
	return &Game{current: newTitleScreen()}
}

func (g *Game) SetScreen(s Screen) { g.current = s }

func (g *Game) Update() error {
	return g.current.Update(g)
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colorBg)
	g.current.Draw(screen)
}

func (g *Game) Layout(_, _ int) (int, int) { return ScreenW, ScreenH }

// Run initialises fonts and starts the Ebitengine window.
func Run() error {
	if err := initFonts(); err != nil {
		return err
	}
	ebiten.SetWindowTitle("Basic Life Simulator")
	ebiten.SetWindowSize(ScreenW, ScreenH)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(newGame())
}
