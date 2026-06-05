package core

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type titleScreen struct {
	loadMode bool
	saves    []SaveInfo
	selIdx   int
	loadMsg  string
}

func newTitleScreen() *titleScreen { return &titleScreen{} }

var titleButtons = []string{"New Game", "Load Game", "Exit"}

func (s *titleScreen) Update(g *Game) error {
	mx, my := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	if s.loadMode {
		backX := float32(ScreenW)/2 - 120
		backY := float32(ScreenH) - 100
		if clicked && isHovered(mx, my, backX, backY, 240, 46) {
			s.loadMode = false
			s.loadMsg = ""
			return nil
		}

		for i := range s.saves {
			rx, ry, rw, rh := titleSaveRowRect(i)
			if clicked && isHovered(mx, my, rx, ry, rw, rh) {
				s.selIdx = i
			}
		}

		if s.selIdx >= 0 && s.selIdx < len(s.saves) {
			lx, ly, lw, lh := titleLoadActionRect(s.selIdx, 0)
			if clicked && isHovered(mx, my, lx, ly, lw, lh) {
				char, err := loadGame(s.saves[s.selIdx].Name)
				if err != nil {
					s.loadMsg = fmt.Sprintf("Load failed: %v", err)
				} else {
					g.SetScreen(newMainScreen(char))
				}
				return nil
			}
			dx, dy, dw, dh := titleLoadActionRect(s.selIdx, 1)
			if clicked && isHovered(mx, my, dx, dy, dw, dh) {
				if err := deleteSave(s.saves[s.selIdx].Name); err != nil {
					s.loadMsg = fmt.Sprintf("Delete failed: %v", err)
				} else {
					names, _ := listSaveFiles()
					s.saves = make([]SaveInfo, 0, len(names))
					for _, n := range names {
						s.saves = append(s.saves, peekSaveInfo(n))
					}
					s.selIdx = -1
				}
				return nil
			}
		}
		return nil
	}

	if clicked {
		for i, lbl := range titleButtons {
			x, y, w, h := titleButtonRect(i)
			if isHovered(mx, my, x, y, w, h) {
				switch lbl {
				case "New Game":
					g.SetScreen(newNewGameScreen())
				case "Load Game":
					s.loadMode = true
					s.selIdx = -1
					s.loadMsg = ""
					names, _ := listSaveFiles()
					s.saves = make([]SaveInfo, 0, len(names))
					for _, n := range names {
						s.saves = append(s.saves, peekSaveInfo(n))
					}
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

	if s.loadMode {
		drawText(dst, "LOAD GAME", cx-60, 160, fontL, colorAccent)
		mx, my := ebiten.CursorPosition()

		if len(s.saves) == 0 {
			drawText(dst, "No save files found.", cx-100, 240, fontM, colorMuted)
		} else {
			for i, info := range s.saves {
				rx, ry, rw, rh := titleSaveRowRect(i)
				sel := i == s.selIdx
				bg := colorPanel
				if sel {
					bg = colorSelected
				} else if isHovered(mx, my, rx, ry, rw, rh) {
					bg = colorHighlight
				}
				fillRect(dst, rx, ry, rw, rh, bg)
				strokeRect(dst, rx, ry, rw, rh, colorBorder)
				if info.Exists {
					drawText(dst, fmt.Sprintf("%s  %s  Age:%d  %s", info.Name, info.CharName, info.Age, info.Date),
						float64(rx)+10, float64(ry)+4, fontS, colorText)
					drawText(dst, fmt.Sprintf("Saved: %s", info.SavedAt),
						float64(rx)+10, float64(ry)+22, fontS, colorMuted)
				} else {
					drawText(dst, info.Name, float64(rx)+10, float64(ry)+12, fontM, colorText)
				}
			}

			if s.selIdx >= 0 && s.selIdx < len(s.saves) {
				lx, ly, lw, lh := titleLoadActionRect(s.selIdx, 0)
				drawButton(dst, "Load", lx, ly, lw, lh, fontM, isHovered(mx, my, lx, ly, lw, lh), true)
				dx, dy, dw, dh := titleLoadActionRect(s.selIdx, 1)
				drawButton(dst, "Delete", dx, dy, dw, dh, fontM, isHovered(mx, my, dx, dy, dw, dh), true)
			}
		}

		backX := float32(ScreenW)/2 - 120
		backY := float32(ScreenH) - 100
		drawButton(dst, "← Back", backX, backY, 240, 46, fontM, isHovered(mx, my, backX, backY, 240, 46), true)

		if s.loadMsg != "" {
			mw, _ := text.Measure(s.loadMsg, fontS, 0)
			drawText(dst, s.loadMsg, cx-mw/2, float64(ScreenH)-140, fontS, colorRed)
		}
		return
	}

	title := "BASIC LIFE SIMULATOR"
	tw, _ := text.Measure(title, fontXL, 0)
	drawText(dst, title, cx-tw/2, 180, fontXL, colorAccent)

	sub := "A life simulation game"
	sw, _ := text.Measure(sub, fontM, 0)
	drawText(dst, sub, cx-sw/2, 224, fontM, colorMuted)

	mx, my := ebiten.CursorPosition()
	for i, lbl := range titleButtons {
		x, y, w, h := titleButtonRect(i)
		drawButton(dst, lbl, x, y, w, h, fontM, isHovered(mx, my, x, y, w, h), true)
	}
}

func titleButtonRect(i int) (x, y, w, h float32) {
	w, h = 240, 46
	x = float32(ScreenW)/2 - w/2
	y = float32(310 + i*62)
	return
}

func titleSaveRowRect(i int) (x, y, w, h float32) {
	w = 500
	h = 44
	x = float32(ScreenW)/2 - w/2
	y = float32(230 + i*54)
	return
}

func titleLoadActionRect(row, actionIdx int) (x, y, w, h float32) {
	w = 100
	h = 36
	x = float32(ScreenW)/2 + 260
	y = float32(230 + row*54 + 4)
	if actionIdx == 1 {
		x += 108
	}
	return
}
