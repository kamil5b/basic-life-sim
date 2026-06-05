package core

import (
	"fmt"

	"github.com/kamil5b/basic-life-sim/internal/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type mpMode int

const (
	mpModeNormal mpMode = iota
	mpModeSave
	mpModeLoad
)

type menuPanel struct {
	char    *model.Character
	main    *mainScreen
	mode    mpMode
	selSlot int
	infos   []SaveInfo
}

func newMenuPanel(char *model.Character, main *mainScreen) *menuPanel {
	return &menuPanel{char: char, main: main, selSlot: -1}
}

func (p *menuPanel) refresh() {
	p.infos = getSaveInfos()
}

func (p *menuPanel) update(g *Game) {
	if (p.mode == mpModeSave || p.mode == mpModeLoad) && len(p.infos) == 0 {
		p.refresh()
	}

	mx, my := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	switch p.mode {
	case mpModeNormal:
		sx, sy, sw, sh := mpSaveBtnRect()
		if clicked && isHovered(mx, my, sx, sy, sw, sh) {
			p.mode = mpModeSave
			p.selSlot = -1
			p.refresh()
			return
		}
		lx, ly, lw, lh := mpLoadBtnRect()
		if clicked && isHovered(mx, my, lx, ly, lw, lh) {
			p.mode = mpModeLoad
			p.selSlot = -1
			p.refresh()
			return
		}
		ex, ey, ew, eh := mpExitBtnRect()
		if clicked && isHovered(mx, my, ex, ey, ew, eh) {
			g.SetScreen(newTitleScreen())
			return
		}

	case mpModeSave:
		bx, by, bw, bh := mpBackBtnRect()
		if clicked && isHovered(mx, my, bx, by, bw, bh) {
			p.mode = mpModeNormal
			return
		}
		for i := 0; i < maxSaveSlots; i++ {
			rx, ry, rw, rh := mpSlotRect(i)
			if clicked && isHovered(mx, my, rx, ry, rw, rh) {
				p.selSlot = i
			}
		}
		if p.selSlot >= 0 {
			name := slotName(p.selSlot)
			exists := saveExists(name)
			if exists {
				ox, oy, ow, oh := mpActionBtnRect(p.selSlot, 0)
				if clicked && isHovered(mx, my, ox, oy, ow, oh) {
					if err := saveGame(name, *p.char); err != nil {
						p.main.setMessage(fmt.Sprintf("Save failed: %v", err))
					} else {
						p.main.setMessage(fmt.Sprintf("Saved to %s.", name))
					}
					p.refresh()
					return
				}
				dx, dy, dw, dh := mpActionBtnRect(p.selSlot, 1)
				if clicked && isHovered(mx, my, dx, dy, dw, dh) {
					if err := deleteSave(name); err != nil {
						p.main.setMessage(fmt.Sprintf("Delete failed: %v", err))
					} else {
						p.main.setMessage(fmt.Sprintf("Deleted %s.", name))
					}
					p.selSlot = -1
					p.refresh()
					return
				}
			} else {
				sx, sy, sw, sh := mpActionBtnRect(p.selSlot, 0)
				if clicked && isHovered(mx, my, sx, sy, sw, sh) {
					if err := saveGame(name, *p.char); err != nil {
						p.main.setMessage(fmt.Sprintf("Save failed: %v", err))
					} else {
						p.main.setMessage(fmt.Sprintf("Saved to %s.", name))
					}
					p.refresh()
					return
				}
			}
		}

	case mpModeLoad:
		bx, by, bw, bh := mpBackBtnRect()
		if clicked && isHovered(mx, my, bx, by, bw, bh) {
			p.mode = mpModeNormal
			return
		}
		for i := 0; i < maxSaveSlots; i++ {
			rx, ry, rw, rh := mpSlotRect(i)
			if clicked && isHovered(mx, my, rx, ry, rw, rh) {
				p.selSlot = i
			}
		}
		if p.selSlot >= 0 {
			name := slotName(p.selSlot)
			exists := saveExists(name)
			if exists {
				lx, ly, lw, lh := mpActionBtnRect(p.selSlot, 0)
				if clicked && isHovered(mx, my, lx, ly, lw, lh) {
					char, err := loadGame(name)
					if err != nil {
						p.main.setMessage(fmt.Sprintf("Load failed: %v", err))
					} else {
						g.SetScreen(newMainScreen(char))
					}
					return
				}
				dx, dy, dw, dh := mpActionBtnRect(p.selSlot, 1)
				if clicked && isHovered(mx, my, dx, dy, dw, dh) {
					if err := deleteSave(name); err != nil {
						p.main.setMessage(fmt.Sprintf("Delete failed: %v", err))
					} else {
						p.main.setMessage(fmt.Sprintf("Deleted %s.", name))
					}
					p.selSlot = -1
					p.refresh()
					return
				}
			}
		}
	}
}

func (p *menuPanel) draw(dst *ebiten.Image) {
	mx, my := ebiten.CursorPosition()

	switch p.mode {
	case mpModeNormal:
		drawText(dst, "Menu", float64(panelX)+8, float64(panelY)+8, fontM, colorAccent)
		sx, sy, sw, sh := mpSaveBtnRect()
		drawButton(dst, "Save Game", sx, sy, sw, sh, fontM, isHovered(mx, my, sx, sy, sw, sh), true)
		lx, ly, lw, lh := mpLoadBtnRect()
		drawButton(dst, "Load Game", lx, ly, lw, lh, fontM, isHovered(mx, my, lx, ly, lw, lh), true)
		ex, ey, ew, eh := mpExitBtnRect()
		drawButton(dst, "Exit to Title", ex, ey, ew, eh, fontM, isHovered(mx, my, ex, ey, ew, eh), true)

	case mpModeSave, mpModeLoad:
		title := "Save Game"
		if p.mode == mpModeLoad {
			title = "Load Game"
		}
		drawText(dst, title, float64(panelX)+8, float64(panelY)+8, fontM, colorAccent)
		bx, by, bw, bh := mpBackBtnRect()
		drawButton(dst, "← Back", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)

		for i := 0; i < maxSaveSlots; i++ {
			rx, ry, rw, rh := mpSlotRect(i)
			sel := i == p.selSlot
			bg := colorPanel
			if sel {
				bg = colorSelected
			} else if isHovered(mx, my, rx, ry, rw, rh) {
				bg = colorHighlight
			}
			fillRect(dst, rx, ry, rw, rh, bg)
			strokeRect(dst, rx, ry, rw, rh, colorBorder)

			if i < len(p.infos) && p.infos[i].Exists {
				info := p.infos[i]
				drawText(dst, fmt.Sprintf("%s  %s  Age:%d  %s", slotName(i), info.CharName, info.Age, info.Date),
					float64(rx)+8, float64(ry)+4, fontS, colorText)
				drawText(dst, fmt.Sprintf("Saved: %s", info.SavedAt),
					float64(rx)+8, float64(ry)+22, fontS, colorMuted)
			} else {
				drawText(dst, fmt.Sprintf("%s  [empty]", slotName(i)),
					float64(rx)+8, float64(ry)+8, fontS, colorText)
			}

			if sel {
				if i < len(p.infos) && p.infos[i].Exists {
					if p.mode == mpModeSave {
						ox, oy, ow, oh := mpActionBtnRect(i, 0)
						drawButton(dst, "Overwrite", ox, oy, ow, oh, fontS, isHovered(mx, my, ox, oy, ow, oh), true)
					} else {
						lx, ly, lw, lh := mpActionBtnRect(i, 0)
						drawButton(dst, "Load", lx, ly, lw, lh, fontS, isHovered(mx, my, lx, ly, lw, lh), true)
					}
					dx, dy, dw, dh := mpActionBtnRect(i, 1)
					drawButton(dst, "Delete", dx, dy, dw, dh, fontS, isHovered(mx, my, dx, dy, dw, dh), true)
				} else {
					sx, sy, sw, sh := mpActionBtnRect(i, 0)
					drawButton(dst, "Save", sx, sy, sw, sh, fontS, isHovered(mx, my, sx, sy, sw, sh), true)
				}
			}
		}
	}
}

// ── layout helpers ─────────────────────────────────────────────────────────────

func mpSaveBtnRect() (x, y, w, h float32) {
	return panelX + 8, panelY + 60, 200, 44
}

func mpLoadBtnRect() (x, y, w, h float32) {
	return panelX + 8, panelY + 116, 200, 44
}

func mpExitBtnRect() (x, y, w, h float32) {
	return panelX + 8, panelY + 172, 200, 44
}

func mpBackBtnRect() (x, y, w, h float32) {
	return panelX + 8, panelY + float32(ScreenH) - float32(tabH) - 50, 100, 32
}

func mpSlotRect(i int) (x, y, w, h float32) {
	return panelX + 8, panelY + 60 + float32(i)*54, 440, 48
}

func mpActionBtnRect(slotIdx, actionIdx int) (x, y, w, h float32) {
	x = panelX + 460
	y = panelY + 60 + float32(slotIdx)*54 + 6
	w = 100
	h = 36
	if actionIdx == 1 {
		x += 108
	}
	return
}
