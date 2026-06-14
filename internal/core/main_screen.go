package core

import (
	"fmt"
	"image"
	"time"

	"github.com/kamil5b/basic-life-sim/internal/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/yohamta/furex/v2"
)

type mainMode int

const (
	modeRoom mainMode = iota
	modeShop
)

const hudW = float32(220)

// mainScreen holds the character and delegates to the active panel.
type mainScreen struct {
	char        model.Character
	mode        mainMode
	room        *roomPanel
	shop        *shopPanel
	menuOverlay *menuOverlay
	message     string
	lastUpdate  time.Time
	game        *Game

	// Content frame from furex layout (set during draw, used during update)
	contentFrame image.Rectangle

	// For door popup → shop transition sent from roomPanel
	requestShop bool
}

func newMainScreen(char model.Character) *mainScreen {
	s := &mainScreen{char: char}
	s.room = newRoomPanel(&s.char, s)
	s.shop = newShopPanel(&s.char, s)
	s.menuOverlay = &menuOverlay{main: s}
	return s
}

func (s *mainScreen) setMessage(msg string) { s.message = msg }

// ── frame-relative layout helpers for panels ──────────────────────────────────

func (s *mainScreen) panelX() float32 { return float32(s.contentFrame.Min.X) }
func (s *mainScreen) panelY() float32 { return float32(s.contentFrame.Min.Y) }
func (s *mainScreen) panelW() float32 { return float32(s.contentFrame.Dx()) }
func (s *mainScreen) panelH() float32 { return float32(s.contentFrame.Dy()) }

func (s *mainScreen) updateContent() {
	if s.menuOverlay.open {
		s.menuOverlay.update()
		return
	}
	switch s.mode {
	case modeRoom:
		s.room.update()
	case modeShop:
		s.shop.update()
	}

	// Handle shop request from room (door popup)
	if s.requestShop {
		s.requestShop = false
		s.mode = modeShop
	}
}

// ── BuildView ─────────────────────────────────────────────────────────────────

func (s *mainScreen) BuildView(g *Game) *furex.View {
	g.main = s
	s.game = g
	s.contentFrame = image.Rect(0, 0, 800, 600) // will be overwritten by furex

	root := &furex.View{
		Direction: furex.Row,
	}

	hud := &furex.View{
		Width:   220,
		Handler: &HUDHandler{s: s, g: g},
	}

	content := &furex.View{
		Grow:    1,
		Handler: &ContentHandler{s: s},
	}

	root.AddChild(hud, content)
	return root
}

// ── HUDHandler ────────────────────────────────────────────────────────────────

type HUDHandler struct {
	s *mainScreen
	g *Game
}

func (h *HUDHandler) Draw(screen *ebiten.Image, frame image.Rectangle, v *furex.View) {
	h.drawHUD(screen, frame)
}

func (h *HUDHandler) HandleJustPressedMouseButtonLeft(x, y int) bool {
	// Speed buttons
	speeds := []float64{0.5, 1, 2, 5, 10, 50, 100}
	baseY := 96
	for i, sp := range speeds {
		col := i % 4
		row := i / 4
		bx := 10 + col*54
		by := baseY + row*24
		if isHovered(x, y, float32(bx), float32(by), 50, 20) {
			h.s.char.TimeScale = sp
			return true
		}
	}

	// Menu button
	if isHovered(x, y, 12, float32(h.s.contentFrame.Dy())-36, float32(hudW)-24, 30) {
		h.s.menuOverlay.open = !h.s.menuOverlay.open
		return true
	}
	return false
}

func (h *HUDHandler) HandleJustReleasedMouseButtonLeft(x, y int) {
}

func (h *HUDHandler) drawHUD(dst *ebiten.Image, frame image.Rectangle) {
	mx32, my32 := float32(hudW), float32(frame.Dy())
	fillRect(dst, 0, 0, mx32, my32, colorPanel)
	strokeRect(dst, 0, 0, mx32, my32, colorBorder)

	char := &h.s.char
	y := 16.0
	nameStr := char.Name
	if char.IsMale {
		nameStr += " (M)"
	} else {
		nameStr += " (F)"
	}
	drawText(dst, nameStr, 10, y, fontM, colorAccent)
	y += 22
	drawText(dst, fmt.Sprintf("Age: %d", char.Age), 10, y, fontS, colorMuted)
	y += 18
	cy, cm, cd := char.CurrentDate.Unpack()
	drawText(dst, fmt.Sprintf("Date: %04d-%02d-%02d", cy, cm, cd), 10, y, fontS, colorMuted)
	y += 18
	drawText(dst, fmt.Sprintf("Time: %02d:%02d (%s)", char.Hour, char.Minute, timeOfDay(char)), 10, y, fontS, colorMuted)
	y += 22

	// speed buttons
	speeds := []float64{0.5, 1, 2, 5, 10, 50, 100}
	speedLabels := []string{"0.5x", "1x", "2x", "5x", "10x", "50x", "100x"}
	mx, my := ebiten.CursorPosition()
	for i, sp := range speeds {
		col := i % 4
		row := i / 4
		bx := float32(10 + col*54)
		by := float32(y) + float32(row*24)
		selected := char.TimeScale == sp
		bg := colorPanel
		txt := colorMuted
		if selected {
			bg = colorSelected
			txt = colorText
		}
		fillRect(dst, bx, by, 50, 20, bg)
		strokeRect(dst, bx, by, 50, 20, colorBorder)
		tw, th := text.Measure(speedLabels[i], fontS, 0)
		tx := float64(bx) + float64(50)/2 - tw/2
		ty := float64(by) + float64(20)/2 - th/2
		drawText(dst, speedLabels[i], tx, ty, fontS, txt)
	}
	y += 50

	drawText(dst, fmt.Sprintf("Money: $%.2f", char.CurrentStats.Money), 10, y, fontM, colorGreen)
	y += 28

	// divider
	fillRect(dst, 6, float32(y), hudW-12, 1, colorBorder)
	y += 10

	drawText(dst, "NEEDS", 10, y, fontS, colorMuted)
	y += 18

	bw := hudW - 16
	type needRow struct {
		label string
		stat  model.NeedStat
	}
	needs := []needRow{
		{"Food", char.Food},
		{"Energy", char.Energy},
		{"Hygiene", char.Hygiene},
		{"Conf.", char.Confidence},
		{"Str.", char.Strength},
	}
	for _, n := range needs {
		needBar(dst, n.label, n.stat.Current, n.stat.Max, 8, float32(y), bw, fontS)
		y += 22
	}

	y += 8
	fillRect(dst, 6, float32(y), hudW-12, 1, colorBorder)
	y += 10
	drawText(dst, char.CurrentHome.Type.Name, 10, y, fontS, colorMuted)

	// Menu button at bottom
	menuY := float32(frame.Dy()) - 36
	menuX := float32(12)
	menuW := hudW - 24
	menuH := float32(30)
	hov := isHovered(mx, my, menuX, menuY, menuW, menuH)
	drawButton(dst, "☰ Menu", menuX, menuY, menuW, menuH, fontM, hov, true)
}

// ── ContentHandler ────────────────────────────────────────────────────────────

type ContentHandler struct {
	s *mainScreen
}

func (h *ContentHandler) Draw(screen *ebiten.Image, frame image.Rectangle, view *furex.View) {
	h.s.contentFrame = frame

	// panel background
	fillRect(screen, float32(frame.Min.X), float32(frame.Min.Y), float32(frame.Dx()), float32(frame.Dy()), colorPanel)
	strokeRect(screen, float32(frame.Min.X), float32(frame.Min.Y), float32(frame.Dx()), float32(frame.Dy()), colorBorder)

	switch h.s.mode {
	case modeRoom:
		h.s.room.draw(screen)
	case modeShop:
		h.s.shop.draw(screen)
	}

	// Message bar at bottom
	if h.s.message != "" {
		fillRect(screen, 0, float32(frame.Dy()+frame.Min.Y)-26, float32(frame.Dx()+frame.Min.X), 26, colorHighlight)
		drawText(screen, h.s.message, float64(frame.Min.X)+10, float64(frame.Dy()+frame.Min.Y)-20, fontS, colorYellow)
	}

	// Menu overlay on top
	if h.s.menuOverlay.open {
		h.s.menuOverlay.draw(screen, frame)
	}
}

// ── menuOverlay ───────────────────────────────────────────────────────────────

type mpMode int

const (
	mpModeNormal mpMode = iota
	mpModeSave
	mpModeLoad
)

type menuOverlay struct {
	main    *mainScreen
	open    bool
	mode    mpMode
	selSlot int
	infos   []SaveInfo
}

func (m *menuOverlay) refresh() {
	m.infos = getSaveInfos()
}

func (m *menuOverlay) update() {
	if (m.mode == mpModeSave || m.mode == mpModeLoad) && len(m.infos) == 0 {
		m.refresh()
	}

	mx, my := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if !clicked {
		return
	}
	frame := m.main.contentFrame
	ox := float32(frame.Min.X)
	oy := float32(frame.Min.Y)
	ow := float32(frame.Dx())
	oh := float32(frame.Dy())

	// Centered panel rect
	panelW := float32(500)
	panelH := float32(400)
	px := ox + (ow-panelW)/2
	py := oy + (oh-panelH)/2

	switch m.mode {
	case mpModeNormal:
		if isHovered(mx, my, px+20, py+60, 460, 44) {
			m.mode = mpModeSave
			m.selSlot = -1
			m.refresh()
			return
		}
		if isHovered(mx, my, px+20, py+116, 460, 44) {
			m.mode = mpModeLoad
			m.selSlot = -1
			m.refresh()
			return
		}
		if isHovered(mx, my, px+20, py+172, 460, 44) {
			m.open = false
			m.main.game.SetScreen(newTitleScreen())
			return
		}

	case mpModeSave:
		if isHovered(mx, my, px+20, py+panelH-50, 100, 32) {
			m.mode = mpModeNormal
			return
		}
		for i := 0; i < maxSaveSlots; i++ {
			if isHovered(mx, my, px+20, py+60+float32(i)*54, 440, 48) {
				m.selSlot = i
				return
			}
		}
		if m.selSlot >= 0 {
			name := slotName(m.selSlot)
			exists := saveExists(name)
			if exists {
				if isHovered(mx, my, px+460, py+60+float32(m.selSlot)*54+6, 100, 36) {
					if err := saveGame(name, m.main.char); err != nil {
						m.main.setMessage(fmt.Sprintf("Save failed: %v", err))
					} else {
						m.main.setMessage(fmt.Sprintf("Saved to %s.", name))
					}
					m.refresh()
					return
				}
				if isHovered(mx, my, px+568, py+60+float32(m.selSlot)*54+6, 100, 36) {
					if err := deleteSave(name); err != nil {
						m.main.setMessage(fmt.Sprintf("Delete failed: %v", err))
					} else {
						m.main.setMessage(fmt.Sprintf("Deleted %s.", name))
					}
					m.selSlot = -1
					m.refresh()
					return
				}
			} else {
				if isHovered(mx, my, px+460, py+60+float32(m.selSlot)*54+6, 100, 36) {
					if err := saveGame(name, m.main.char); err != nil {
						m.main.setMessage(fmt.Sprintf("Save failed: %v", err))
					} else {
						m.main.setMessage(fmt.Sprintf("Saved to %s.", name))
					}
					m.refresh()
					return
				}
			}
		}

	case mpModeLoad:
		if isHovered(mx, my, px+20, py+panelH-50, 100, 32) {
			m.mode = mpModeNormal
			return
		}
		for i := 0; i < maxSaveSlots; i++ {
			if isHovered(mx, my, px+20, py+60+float32(i)*54, 440, 48) {
				m.selSlot = i
				return
			}
		}
		if m.selSlot >= 0 {
			name := slotName(m.selSlot)
			exists := saveExists(name)
			if exists {
				if isHovered(mx, my, px+460, py+60+float32(m.selSlot)*54+6, 100, 36) {
					char, err := loadGame(name)
					if err != nil {
						m.main.setMessage(fmt.Sprintf("Load failed: %v", err))
					} else {
						m.main.game.SetScreen(newMainScreen(char))
					}
					return
				}
				if isHovered(mx, my, px+568, py+60+float32(m.selSlot)*54+6, 100, 36) {
					if err := deleteSave(name); err != nil {
						m.main.setMessage(fmt.Sprintf("Delete failed: %v", err))
					} else {
						m.main.setMessage(fmt.Sprintf("Deleted %s.", name))
					}
					m.selSlot = -1
					m.refresh()
					return
				}
			}
		}
	}
}

func (m *menuOverlay) draw(dst *ebiten.Image, frame image.Rectangle) {
	mx, my := ebiten.CursorPosition()
	ox := float32(frame.Min.X)
	oy := float32(frame.Min.Y)
	ow := float32(frame.Dx())
	oh := float32(frame.Dy())

	// Semi-transparent backdrop
	fillRect(dst, ox, oy, ow, oh, colorPanel)

	// Centered panel
	panelW := float32(500)
	panelH := float32(400)
	px := ox + (ow-panelW)/2
	py := oy + (oh-panelH)/2

	fillRect(dst, px, py, panelW, panelH, colorBg)
	strokeRect(dst, px, py, panelW, panelH, colorAccent)

	switch m.mode {
	case mpModeNormal:
		drawText(dst, "Menu", float64(px+20), float64(py+20), fontM, colorAccent)
		drawButton(dst, "Save Game", px+20, py+60, 460, 44, fontM, isHovered(mx, my, px+20, py+60, 460, 44), true)
		drawButton(dst, "Load Game", px+20, py+116, 460, 44, fontM, isHovered(mx, my, px+20, py+116, 460, 44), true)
		drawButton(dst, "Exit to Title", px+20, py+172, 460, 44, fontM, isHovered(mx, my, px+20, py+172, 460, 44), true)

	case mpModeSave, mpModeLoad:
		title := "Save Game"
		if m.mode == mpModeLoad {
			title = "Load Game"
		}
		drawText(dst, title, float64(px+20), float64(py+20), fontM, colorAccent)
		drawButton(dst, "← Back", px+20, py+panelH-50, 100, 32, fontS, isHovered(mx, my, px+20, py+panelH-50, 100, 32), true)

		for i := 0; i < maxSaveSlots; i++ {
			rx := px + 20
			ry := py + 60 + float32(i)*54
			sel := i == m.selSlot
			bg := colorPanel
			if sel {
				bg = colorSelected
			} else if isHovered(mx, my, rx, ry, 440, 48) {
				bg = colorHighlight
			}
			fillRect(dst, rx, ry, 440, 48, bg)
			strokeRect(dst, rx, ry, 440, 48, colorBorder)

			if i < len(m.infos) && m.infos[i].Exists {
				info := m.infos[i]
				drawText(dst, fmt.Sprintf("%s  %s  Age:%d  %s", slotName(i), info.CharName, info.Age, info.Date),
					float64(rx)+8, float64(ry)+4, fontS, colorText)
				drawText(dst, fmt.Sprintf("Saved: %s", info.SavedAt),
					float64(rx)+8, float64(ry)+22, fontS, colorMuted)
			} else {
				drawText(dst, fmt.Sprintf("%s  [empty]", slotName(i)),
					float64(rx)+8, float64(ry)+8, fontS, colorText)
			}

			if sel {
				if i < len(m.infos) && m.infos[i].Exists {
					if m.mode == mpModeSave {
						drawButton(dst, "Overwrite", px+460, ry+6, 100, 36, fontS, isHovered(mx, my, px+460, ry+6, 100, 36), true)
					} else {
						drawButton(dst, "Load", px+460, ry+6, 100, 36, fontS, isHovered(mx, my, px+460, ry+6, 100, 36), true)
					}
					drawButton(dst, "Delete", px+568, ry+6, 100, 36, fontS, isHovered(mx, my, px+568, ry+6, 100, 36), true)
				} else {
					drawButton(dst, "Save", px+460, ry+6, 100, 36, fontS, isHovered(mx, my, px+460, ry+6, 100, 36), true)
				}
			}
		}
	}
}
