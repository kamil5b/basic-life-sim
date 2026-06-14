package core

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/furex/v2"
)

// Game is the root Ebitengine game object.
type Game struct {
	current    Screen
	rootView   *furex.View
	main       *mainScreen
	updater    func(g *Game) error
	shouldExit bool
	winW       int
	winH       int
}

func newGame() *Game {
	g := &Game{}
	g.SetScreen(newTitleScreen())
	return g
}

func (g *Game) SetScreen(s Screen) {
	g.current = s
	g.main = nil
	g.updater = nil
	g.rootView = s.BuildView(g)
}

func (g *Game) Update() error {
	if g.shouldExit {
		return ebiten.Termination
	}
	// Time advancement (only when main screen is active)
	if g.main != nil {
		now := time.Now()
		if !g.main.lastUpdate.IsZero() {
			elapsed := now.Sub(g.main.lastUpdate).Seconds()
			gameMinutes := elapsed * g.main.char.TimeScale
			advanceTime(&g.main.char, gameMinutes)
		}
		g.main.lastUpdate = now
	}

	// Escape key toggles menu overlay
	if g.main != nil && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.main.mode == modeShop {
			g.main.mode = modeRoom
		} else {
			g.main.menuOverlay.open = !g.main.menuOverlay.open
		}
	}

	if g.rootView != nil {
		g.rootView.UpdateWithSize(g.winW, g.winH)
	}

	if g.updater != nil {
		g.updater(g)
	}

	// Room/shop panel updates (precision click detection needs frame-relative coords)
	if g.main != nil {
		g.main.updateContent()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colorBg)
	if g.rootView != nil {
		g.rootView.Draw(screen)
	}
}

func (g *Game) Layout(ow, oh int) (int, int) {
	if ow < 800 {
		ow = 800
	}
	if oh < 600 {
		oh = 600
	}
	g.winW = ow
	g.winH = oh
	return ow, oh
}

// Run initialises fonts and starts the Ebitengine window.
func Run() error {
	if err := initFonts(); err != nil {
		return err
	}
	ebiten.SetWindowTitle("Basic Life Simulator")
	ebiten.SetWindowSize(1100, 700)
	ebiten.SetWindowPosition(200, 100)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(newGame())
}
