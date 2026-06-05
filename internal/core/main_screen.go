package core

import (
	"fmt"
	"time"

	"github.com/kamil5b/basic-life-sim/internal/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type mainTab int

const (
	tabRoom mainTab = iota
	tabShop
	tabFood
	tabMenu
)

var tabLabels = []string{"Room", "Shop", "Food", "Menu"}

// mainScreen holds the character and delegates to the active panel.
type mainScreen struct {
	char       model.Character
	tab        mainTab
	room       *roomPanel
	shop       *shopPanel
	food       *foodPanel
	menu       *menuPanel
	message    string // transient feedback message
	lastUpdate time.Time
}

func newMainScreen(char model.Character) *mainScreen {
	s := &mainScreen{char: char}
	s.room = newRoomPanel(&s.char, s)
	s.shop = newShopPanel(&s.char, s)
	s.food = newFoodPanel(&s.char, s)
	s.menu = newMenuPanel(&s.char, s)
	return s
}

func (s *mainScreen) setMessage(msg string) { s.message = msg }

const (
	hudW   = 220
	tabH   = 36
	panelX = float32(hudW + 1)
	panelW = float32(ScreenW) - hudW - 1
	panelY = float32(tabH + 1)
	panelH = float32(ScreenH) - float32(tabH) - 1
)

func (s *mainScreen) Update(g *Game) error {
	mx, my := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if clicked {
		// tab clicks
		for i := range tabLabels {
			tx, ty, tw, th := tabRect(i)
			if isHovered(mx, my, tx, ty, tw, th) {
				s.tab = mainTab(i)
				s.message = ""
			}
		}
	}

	// time advancement
	now := time.Now()
	if !s.lastUpdate.IsZero() {
		elapsed := now.Sub(s.lastUpdate).Seconds()
		gameMinutes := elapsed * s.char.TimeScale
		advanceTime(&s.char, gameMinutes)
	}
	s.lastUpdate = now

	// speed buttons in HUD area
	speeds := []float64{0.5, 1, 2, 5, 10, 50, 100}
	for i, sp := range speeds {
		col := i % 4
		row := i / 4
		bx := float32(10 + col*54)
		by := float32(96 + row*24)
		if clicked && isHovered(mx, my, bx, by, 50, 20) {
			s.char.TimeScale = sp
		}
	}

	switch s.tab {
	case tabRoom:
		s.room.update()
	case tabShop:
		s.shop.update()
	case tabFood:
		s.food.update()
	case tabMenu:
		s.menu.update(g)
	}
	return nil
}

func (s *mainScreen) Draw(dst *ebiten.Image) {
	s.drawHUD(dst)
	s.drawTabs(dst)

	// panel area background
	fillRect(dst, panelX, panelY, panelW, panelH, colorPanel)
	strokeRect(dst, panelX, panelY, panelW, panelH, colorBorder)

	switch s.tab {
	case tabRoom:
		s.room.draw(dst)
	case tabShop:
		s.shop.draw(dst)
	case tabFood:
		s.food.draw(dst)
	case tabMenu:
		s.menu.draw(dst)
	}

	// message bar at the bottom
	if s.message != "" {
		fillRect(dst, 0, float32(ScreenH)-26, float32(ScreenW), 26, colorHighlight)
		drawText(dst, s.message, 10, float64(ScreenH)-20, fontS, colorYellow)
	}
}

func (s *mainScreen) drawHUD(dst *ebiten.Image) {
	fillRect(dst, 0, 0, float32(hudW), float32(ScreenH), colorPanel)
	strokeRect(dst, 0, 0, float32(hudW), float32(ScreenH), colorBorder)

	char := &s.char
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
	fillRect(dst, 6, float32(y), float32(hudW)-12, 1, colorBorder)
	y += 10

	drawText(dst, "NEEDS", 10, y, fontS, colorMuted)
	y += 18

	bw := float32(hudW) - 16
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
	fillRect(dst, 6, float32(y), float32(hudW)-12, 1, colorBorder)
	y += 10
	drawText(dst, char.CurrentHome.Type.Name, 10, y, fontS, colorMuted)
}

func (s *mainScreen) drawTabs(dst *ebiten.Image) {
	mx, my := ebiten.CursorPosition()
	for i, lbl := range tabLabels {
		tx, ty, tw, th := tabRect(i)
		active := mainTab(i) == s.tab
		bg := colorPanel
		border := colorBorder
		textCol := colorMuted
		if active {
			bg = colorSelected
			border = colorAccent
			textCol = colorText
		} else if isHovered(mx, my, tx, ty, tw, th) {
			bg = colorHighlight
		}
		fillRect(dst, tx, ty, tw, th, bg)
		strokeRect(dst, tx, ty, tw, th, border)
		tw2, _ := text.Measure(lbl, fontM, 0)
		drawText(dst, lbl, float64(tx)+float64(tw)/2-tw2/2, float64(ty)+8, fontM, textCol)
	}
}

func tabRect(i int) (x, y, w, h float32) {
	count := len(tabLabels)
	w = (float32(ScreenW) - float32(hudW)) / float32(count)
	h = float32(tabH)
	x = float32(hudW) + float32(i)*w
	y = 0
	return
}
