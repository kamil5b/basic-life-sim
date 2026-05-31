package core

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type titleScreen struct{}

func newTitleScreen() *titleScreen { return &titleScreen{} }

var titleButtons = []string{"New Game", "Load Game", "Exit"}

func (s *titleScreen) Update(g *Game) error {
	mx, my := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		for i, lbl := range titleButtons {
			x, y, w, h := titleButtonRect(i)
			if isHovered(mx, my, x, y, w, h) {
				switch lbl {
				case "New Game":
					g.SetScreen(newNewGameScreen())
				case "Exit":
					return ebiten.Termination
				}
			}
		}
	}
	return nil
}

func (s *titleScreen) Draw(dst *ebiten.Image) {
	cx := float64(ScreenW) / 2

	title := "BASIC LIFE SIMULATOR"
	tw, _ := text.Measure(title, fontXL, 0)
	drawText(dst, title, cx-tw/2, 180, fontXL, colorAccent)

	sub := "A life simulation game"
	sw, _ := text.Measure(sub, fontM, 0)
	drawText(dst, sub, cx-sw/2, 224, fontM, colorMuted)

	mx, my := ebiten.CursorPosition()
	for i, lbl := range titleButtons {
		x, y, w, h := titleButtonRect(i)
		enabled := lbl != "Load Game"
		drawButton(dst, lbl, x, y, w, h, fontM, isHovered(mx, my, x, y, w, h) && enabled, enabled)
	}
}

func titleButtonRect(i int) (x, y, w, h float32) {
	w, h = 240, 46
	x = float32(ScreenW)/2 - w/2
	y = float32(310 + i*62)
	return
}
