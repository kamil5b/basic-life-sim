package core

import (
	"fmt"
	"image"
	"image/color"
	"unicode/utf8"

	"github.com/kamil5b/basic-life-sim/internal/constant"
	"github.com/kamil5b/basic-life-sim/internal/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/furex/v2"
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
	runes   []rune
}

func newNewGameScreen() *newGameScreen {
	return &newGameScreen{
		homes: constant.Level1HomeTypes,
	}
}

func (s *newGameScreen) BuildView(g *Game) *furex.View {
	g.updater = s.update
	return s.build(g)
}

func (s *newGameScreen) update(g *Game) error {
	if s.step == ngStepName {
		s.runes = ebiten.AppendInputChars(s.runes[:0])
		for _, r := range s.runes {
			s.name += string(r)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && utf8.RuneCountInString(s.name) > 0 {
			runes := []rune(s.name)
			s.name = string(runes[:len(runes)-1])
		}
	}
	return nil
}

func (s *newGameScreen) rebuild(g *Game) {
	g.rootView = s.build(g)
}

func (s *newGameScreen) build(g *Game) *furex.View {
	root := &furex.View{
		Direction:  furex.Column,
		Justify:    furex.JustifyCenter,
		AlignItems: furex.AlignItemCenter,
	}

	// Header
	root.AddChild(&furex.View{
		Height:  50,
		Handler: &centerLabel{text: "NEW GAME", font: fontL, clr: colorAccent},
	})

	switch s.step {
	case ngStepName:
		s.buildNameStep(root, g)
	case ngStepGender:
		s.buildGenderStep(root, g)
	case ngStepHome:
		s.buildHomeStep(root, g)
	}

	// Error label
	if s.errMsg != "" {
		root.AddChild(&furex.View{
			Height:  40,
			Handler: &centerLabel{text: s.errMsg, font: fontS, clr: colorRed},
		})
	}

	return root
}

func (s *newGameScreen) buildNameStep(root *furex.View, g *Game) {
	root.AddChild(&furex.View{
		Height:  40,
		Handler: &centerLabel{text: "Enter your character's name:", font: fontM, clr: colorText},
	})

	// Text input display (custom draw, not a furex button since it's keyboard-driven)
	root.AddChild(&furex.View{
		Width:   400,
		Height:  46,
		Handler: &ngInputBox{s: s},
	})

	root.AddChild(&furex.View{Height: 16})

	root.AddChild(&furex.View{
		Width:  200,
		Height: 44,
		Handler: &simpleBtn{
			label: "Continue →",
			font:  fontM,
			action: func() {
				defer s.rebuild(g)
				if s.name == "" {
					s.errMsg = "Please enter a name."
				} else {
					s.errMsg = ""
					s.step = ngStepGender
				}
			},
		},
	})
}

func (s *newGameScreen) buildGenderStep(root *furex.View, g *Game) {
	root.AddChild(&furex.View{
		Height:  40,
		Handler: &centerLabel{text: "Choose your gender:", font: fontM, clr: colorText},
	})

	row := &furex.View{
		Direction: furex.Row,
		Height:    60,
	}
	row.AddChild(&furex.View{
		Width:       200,
		Height:      60,
		MarginRight: 16,
		Handler: &simpleBtn{
			label: "♂  Male",
			font:  fontM,
			action: func() {
				defer s.rebuild(g)
				s.isMale = true
				s.errMsg = ""
				s.step = ngStepHome
			},
		},
	})
	row.AddChild(&furex.View{
		Width:  200,
		Height: 60,
		Handler: &simpleBtn{
			label: "♀  Female",
			font:  fontM,
			action: func() {
				defer s.rebuild(g)
				s.isMale = false
				s.errMsg = ""
				s.step = ngStepHome
			},
		},
	})
	root.AddChild(row)
}

func (s *newGameScreen) buildHomeStep(root *furex.View, g *Game) {
	root.AddChild(&furex.View{
		Height:  40,
		Handler: &centerLabel{text: "Choose your starting home:", font: fontM, clr: colorText},
	})

	// Home cards in a row
	cardsRow := &furex.View{
		Direction: furex.Row,
		Height:    380,
	}

	for i, home := range s.homes {
		idx := i
		hm := home
		cardsRow.AddChild(&furex.View{
			Width:       260,
			Height:      380,
			MarginRight: 16,
			Handler: &ngHomeCard{
				s:    s,
				idx:  idx,
				home: hm,
			},
		})
	}
	root.AddChild(cardsRow)

	root.AddChild(&furex.View{Height: 8})

	root.AddChild(&furex.View{
		Width:  240,
		Height: 44,
		Handler: &simpleBtn{
			label: "Start Game →",
			font:  fontM,
			action: func() {
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
			},
		},
	})
}

// ── new game handlers ────────────────────────────────────────────────────────

type ngInputBox struct {
	s *newGameScreen
}

func (b *ngInputBox) Draw(screen *ebiten.Image, frame image.Rectangle, v *furex.View) {
	x, y := float32(frame.Min.X), float32(frame.Min.Y)
	w, h := float32(frame.Dx()), float32(frame.Dy())
	fillRect(screen, x, y, w, h, colorPanel)
	strokeRect(screen, x, y, w, h, colorAccent)
	display := b.s.name + "|"
	drawText(screen, display, float64(x)+10, float64(y)+12, fontM, colorText)
}

type ngHomeCard struct {
	s    *newGameScreen
	idx  int
	home model.HomeType
}

func (c *ngHomeCard) Draw(screen *ebiten.Image, frame image.Rectangle, v *furex.View) {
	mx, my := ebiten.CursorPosition()
	x, y := float32(frame.Min.X), float32(frame.Min.Y)
	w, h := float32(frame.Dx()), float32(frame.Dy())

	selected := c.idx == c.s.homeIdx
	bg := colorPanel
	border := colorBorder
	if selected {
		border = colorAccent
		bg = colorHighlight
	} else if isHovered(mx, my, x, y, w, h) {
		bg = colorHighlight
	}
	fillRect(screen, x, y, w, h, bg)
	strokeRect(screen, x, y, w, h, border)

	iy := float64(y) + 14
	drawText(screen, c.home.Name, float64(x)+14, iy, fontM, colorAccent)
	iy += 24
	drawText(screen, fmt.Sprintf("Upfront: $%.0f   Monthly: $%.0f", c.home.UpfrontCost, c.home.MonthlyCost), float64(x)+14, iy, fontS, colorText)
	iy += 20
	shared := ""
	if c.home.SharedBathroom {
		shared += "Shared bathroom  "
	}
	if c.home.SharedKitchen {
		shared += "Shared kitchen"
	}
	if shared != "" {
		drawText(screen, shared, float64(x)+14, iy, fontS, colorMuted)
	}
	iy += 20
	drawLayoutMini(screen, c.home, float32(x)+14, float32(iy), 14)
}

func (c *ngHomeCard) HandleJustPressedMouseButtonLeft(x, y int) bool {
	c.s.homeIdx = c.idx
	return true
}

func (c *ngHomeCard) HandleJustReleasedMouseButtonLeft(x, y int) {}

func drawLayoutMini(dst *ebiten.Image, home model.HomeType, ox, oy, cellSize float32) {
	for r, row := range home.Layout {
		for c2, cell := range row {
			cx := ox + float32(c2)*cellSize
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
