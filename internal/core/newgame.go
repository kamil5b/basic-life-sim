package core

import (
	"fmt"
	"image/color"
	"unicode/utf8"

	"github.com/kamil5b/basic-life-sim/internal/constant"
	"github.com/kamil5b/basic-life-sim/internal/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type ngStep int

const (
	ngStepName ngStep = iota
	ngStepGender
	ngStepHome
)

type newGameScreen struct {
	step    ngStep
	name    string
	isMale  bool
	homeIdx int
	errMsg  string
	homes   []model.HomeType
	// text input
	runes []rune
}

func newNewGameScreen() *newGameScreen {
	return &newGameScreen{
		homes: constant.Level1HomeTypes,
	}
}

func (s *newGameScreen) Update(g *Game) error {
	mx, my := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	switch s.step {
	case ngStepName:
		// collect typed characters
		s.runes = ebiten.AppendInputChars(s.runes[:0])
		for _, r := range s.runes {
			s.name += string(r)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && utf8.RuneCountInString(s.name) > 0 {
			runes := []rune(s.name)
			s.name = string(runes[:len(runes)-1])
		}
		if clicked && isHovered(mx, my, ngBtnX(), ngBtnY(0), ngBtnW, ngBtnH) {
			if s.name == "" {
				s.errMsg = "Please enter a name."
			} else {
				s.errMsg = ""
				s.step = ngStepGender
			}
		}

	case ngStepGender:
		if clicked {
			if isHovered(mx, my, ngBtnX()-130, ngBtnY(0), 240, ngBtnH) {
				s.isMale = true
				s.errMsg = ""
				s.step = ngStepHome
			} else if isHovered(mx, my, ngBtnX()+130, ngBtnY(0), 240, ngBtnH) {
				s.isMale = false
				s.errMsg = ""
				s.step = ngStepHome
			}
		}

	case ngStepHome:
		for i := range s.homes {
			hx, hy, hw, hh := ngHomeCardRect(i)
			if clicked && isHovered(mx, my, hx, hy, hw, hh) {
				s.homeIdx = i
			}
		}
		if clicked && isHovered(mx, my, ngBtnX(), ngBtnY(0), ngBtnW, ngBtnH) {
			// start game
			home := s.homes[s.homeIdx]
			char := model.Character{
				Name:        s.name,
				IsMale:      s.isMale,
				Age:         18,
				CurrentDate: model.NewCompactDate(2025, 1, 1),
				Hour:        8,
				Minute:      0,
				TimeScale:   1.0,
				Food:        model.NeedStat{Current: 50, Max: 50},
				Energy:      model.NeedStat{Current: 50, Max: 50},
				Hygiene:     model.NeedStat{Current: 50, Max: 50},
				Confidence:  model.NeedStat{Current: 50, Max: 50},
				Strength:    model.NeedStat{Current: 50, Max: 50},
				CurrentStats: model.Stats{
					Money: 3000 - home.UpfrontCost,
				},
				CurrentHome: model.Home{Type: home},
			}
			g.SetScreen(newMainScreen(char))
		}
	}
	return nil
}

const (
	ngBtnW = 200
	ngBtnH = 44
)

func ngBtnX() float32 { return float32(ScreenW)/2 - ngBtnW/2 }
func ngBtnY(row int) float32 {
	return float32(ScreenH) - 100 + float32(row)*56
}

func (s *newGameScreen) Draw(dst *ebiten.Image) {
	cx := float64(ScreenW) / 2

	// title
	drawText(dst, "NEW GAME", cx-50, 40, fontL, colorAccent)

	mx, my := ebiten.CursorPosition()

	switch s.step {
	case ngStepName:
		drawText(dst, "Enter your character's name:", cx-160, 160, fontM, colorText)
		// input box
		bx, by, bw, bh := float32(cx)-200, float32(220), float32(400), float32(46)
		fillRect(dst, bx, by, bw, bh, colorPanel)
		strokeRect(dst, bx, by, bw, bh, colorAccent)
		display := s.name + "|"
		drawText(dst, display, float64(bx)+10, float64(by)+12, fontM, colorText)
		drawButton(dst, "Continue →", ngBtnX(), ngBtnY(0), ngBtnW, ngBtnH, fontM,
			isHovered(mx, my, ngBtnX(), ngBtnY(0), ngBtnW, ngBtnH), true)

	case ngStepGender:
		drawText(dst, "Choose your gender:", cx-110, 160, fontM, colorText)
		mx32, my32 := float32(mx), float32(my)
		_ = mx32
		_ = my32

		maleX := float32(cx) - 270
		femX := float32(cx) + 50
		bw, bh := float32(200), float32(60)

		mHov := isHovered(mx, my, maleX, ngBtnY(0)-10, bw, bh)
		fHov := isHovered(mx, my, femX, ngBtnY(0)-10, bw, bh)
		drawButton(dst, "♂  Male", maleX, ngBtnY(0)-10, bw, bh, fontM, mHov, true)
		drawButton(dst, "♀  Female", femX, ngBtnY(0)-10, bw, bh, fontM, fHov, true)

	case ngStepHome:
		drawText(dst, "Choose your starting home:", cx-160, 100, fontM, colorText)
		for i, home := range s.homes {
			hx, hy, hw, hh := ngHomeCardRect(i)
			selected := i == s.homeIdx
			bg := colorPanel
			border := colorBorder
			if selected {
				border = colorAccent
				bg = colorHighlight
			} else if isHovered(mx, my, hx, hy, hw, hh) {
				bg = colorHighlight
			}
			fillRect(dst, hx, hy, hw, hh, bg)
			strokeRect(dst, hx, hy, hw, hh, border)

			iy := float64(hy) + 14
			drawText(dst, home.Name, float64(hx)+14, iy, fontM, colorAccent)
			iy += 24
			drawText(dst, fmt.Sprintf("Upfront: $%.0f   Monthly: $%.0f", home.UpfrontCost, home.MonthlyCost), float64(hx)+14, iy, fontS, colorText)
			iy += 20
			shared := ""
			if home.SharedBathroom {
				shared += "Shared bathroom  "
			}
			if home.SharedKitchen {
				shared += "Shared kitchen"
			}
			if shared != "" {
				drawText(dst, shared, float64(hx)+14, iy, fontS, colorMuted)
			}
			iy += 20
			// draw small layout grid
			drawLayoutMini(dst, home, float32(hx)+14, float32(iy), 14)
		}
		drawButton(dst, "Start Game →", ngBtnX(), ngBtnY(0), ngBtnW+40, ngBtnH, fontM,
			isHovered(mx, my, ngBtnX(), ngBtnY(0), ngBtnW+40, ngBtnH), true)
	}

	if s.errMsg != "" {
		ew, _ := text.Measure(s.errMsg, fontS, 0)
		drawText(dst, s.errMsg, cx-ew/2, float64(ngBtnY(1))+6, fontS, colorRed)
	}
}

func ngHomeCardRect(i int) (x, y, w, h float32) {
	w = float32(ScreenW)/float32(len(constant.Level1HomeTypes)) - 40
	h = 360
	x = 40 + float32(i)*(w+20)
	y = 140
	return
}

func drawLayoutMini(dst *ebiten.Image, home model.HomeType, ox, oy, cellSize float32) {
	for r, row := range home.Layout {
		for c, cell := range row {
			cx := ox + float32(c)*cellSize
			cy := oy + float32(r)*cellSize
			var col color.RGBA
			switch cell {
			case model.HomeCellWall:
				col = colorWall
			case model.HomeCellFloor:
				col = colorFloor
			case model.HomeCellDoor:
				col = colorDoor
			default:
				continue
			}
			fillRect(dst, cx, cy, cellSize-1, cellSize-1, col)
		}
	}
}
