package core

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/yohamta/furex/v2"
)

type titleScreen struct {
	loadMode bool
	saves    []SaveInfo
	selIdx   int
	loadMsg  string
}

func newTitleScreen() *titleScreen { return &titleScreen{} }

func (s *titleScreen) BuildView(g *Game) *furex.View {
	return s.build(g)
}

func (s *titleScreen) rebuild(g *Game) {
	g.rootView = s.build(g)
	g.rootView.UpdateWithSize(g.winW, g.winH)
}

func (s *titleScreen) build(g *Game) *furex.View {
	root := &furex.View{
		Direction:  furex.Column,
		Justify:    furex.JustifyCenter,
		AlignItems: furex.AlignItemCenter,
	}

	root.AddChild(&furex.View{
		Height:  80,
		Handler: &centerLabel{text: "BASIC LIFE SIMULATOR", font: fontXL, clr: colorAccent},
	})
	root.AddChild(&furex.View{
		Height:  30,
		Handler: &centerLabel{text: "A life simulation game", font: fontM, clr: colorMuted},
	})
	root.AddChild(&furex.View{Height: 50})

	if s.loadMode {
		s.buildLoadMode(root, g)
	} else {
		for _, lbl := range []string{"New Game", "Load Game", "Exit"} {
			l := lbl
			root.AddChild(&furex.View{
				Width:        240,
				Height:       46,
				MarginBottom: 16,
				Handler: &simpleBtn{
					label: l,
					font:  fontM,
					action: func() {
						switch l {
						case "New Game":
							g.SetScreen(newNewGameScreen())
							return
						case "Load Game":
							defer s.rebuild(g)
							s.loadMode = true
							s.selIdx = -1
							s.loadMsg = ""
							names, _ := listSaveFiles()
							s.saves = make([]SaveInfo, 0, len(names))
							for _, n := range names {
								s.saves = append(s.saves, peekSaveInfo(n))
							}
						case "Exit":
							g.shouldExit = true
						}
					},
				},
			})
		}
	}

	return root
}

func (s *titleScreen) buildLoadMode(root *furex.View, g *Game) {
	root.AddChild(&furex.View{
		Height:  50,
		Handler: &centerLabel{text: "LOAD GAME", font: fontL, clr: colorAccent},
	})

	if len(s.saves) == 0 {
		root.AddChild(&furex.View{
			Height:  40,
			Handler: &centerLabel{text: "No save files found.", font: fontM, clr: colorMuted},
		})
	} else {
		for i, info := range s.saves {
			idx := i
			inf := info
			root.AddChild(&furex.View{
				Width:  500,
				Height: 44,
				Handler: &loadSlot{
					s:    s,
					g:    g,
					idx:  idx,
					info: inf,
				},
			})
		}
	}

	if s.loadMsg != "" {
		root.AddChild(&furex.View{
			Height:  30,
			Handler: &centerLabel{text: s.loadMsg, font: fontS, clr: colorRed},
		})
	}

	root.AddChild(&furex.View{
		Width:     240,
		Height:    46,
		MarginTop: 20,
		Handler: &simpleBtn{
			label: "← Back",
			font:  fontM,
			action: func() {
				defer s.rebuild(g)
				s.loadMode = false
			},
		},
	})
}

// ── shared handlers ──────────────────────────────────────────────────────────

type centerLabel struct {
	text string
	font *text.GoTextFace
	clr  color.RGBA
}

func (c *centerLabel) Draw(screen *ebiten.Image, frame image.Rectangle, v *furex.View) {
	tw, _ := text.Measure(c.text, c.font, 0)
	x := float64(frame.Min.X) + float64(frame.Dx())/2 - tw/2
	y := float64(frame.Min.Y) + float64(frame.Dy())/2 + 4
	drawText(screen, c.text, x, y, c.font, c.clr)
}

type simpleBtn struct {
	label  string
	font   *text.GoTextFace
	action func()
}

func (b *simpleBtn) Draw(screen *ebiten.Image, frame image.Rectangle, v *furex.View) {
	mx, my := ebiten.CursorPosition()
	x, y := float32(frame.Min.X), float32(frame.Min.Y)
	w, h := float32(frame.Dx()), float32(frame.Dy())
	drawButton(screen, b.label, x, y, w, h, b.font, isHovered(mx, my, x, y, w, h), true)
}

func (b *simpleBtn) HandleJustPressedMouseButtonLeft(x, y int) bool {
	if b.action != nil {
		b.action()
	}
	return true
}

func (b *simpleBtn) HandleJustReleasedMouseButtonLeft(x, y int) {}

type loadSlot struct {
	s     *titleScreen
	g     *Game
	idx   int
	info  SaveInfo
	frame image.Rectangle
}

func (ls *loadSlot) Draw(screen *ebiten.Image, frame image.Rectangle, v *furex.View) {
	ls.frame = frame
	mx, my := ebiten.CursorPosition()
	x, y := float32(frame.Min.X), float32(frame.Min.Y)
	w, h := float32(frame.Dx()), float32(frame.Dy())

	sel := ls.idx == ls.s.selIdx
	bg := colorPanel
	if sel {
		bg = colorSelected
	} else if isHovered(mx, my, x, y, w, h) {
		bg = colorHighlight
	}
	fillRect(screen, x, y, w, h, bg)
	strokeRect(screen, x, y, w, h, colorBorder)

	if ls.info.Exists {
		drawText(screen, fmt.Sprintf("%s  %s  Age:%d  %s", slotName(ls.idx), ls.info.CharName, ls.info.Age, ls.info.Date),
			float64(x)+10, float64(y)+4, fontS, colorText)
		drawText(screen, fmt.Sprintf("Saved: %s", ls.info.SavedAt),
			float64(x)+10, float64(y)+22, fontS, colorMuted)

		if sel {
			ax := x + w - 220
			ay := y + 4
			drawButton(screen, "Load", ax, ay, 100, 36, fontM, isHovered(mx, my, ax, ay, 100, 36), true)
			drawButton(screen, "Delete", ax+108, ay, 100, 36, fontM, isHovered(mx, my, ax+108, ay, 100, 36), true)
		}
	} else {
		drawText(screen, fmt.Sprintf("%s  [empty]", slotName(ls.idx)),
			float64(x)+10, float64(y)+12, fontS, colorText)
	}
}

func (ls *loadSlot) HandleJustPressedMouseButtonLeft(px, py int) bool {
	defer ls.s.rebuild(ls.g)
	ls.s.selIdx = ls.idx
	x := float32(ls.frame.Min.X)
	y := float32(ls.frame.Min.Y)
	w := float32(ls.frame.Dx())

	if ls.idx >= 0 && ls.idx < len(ls.s.saves) && ls.s.saves[ls.idx].Exists {
		ax := x + w - 220
		ay := y + 4
		if isHovered(px, py, ax, ay, 100, 36) {
			char, err := loadGame(ls.s.saves[ls.idx].Name)
			if err != nil {
				ls.s.loadMsg = fmt.Sprintf("Load failed: %v", err)
			} else {
				ls.g.SetScreen(newMainScreen(char))
			}
			return true
		}
		if isHovered(px, py, ax+108, ay, 100, 36) {
			if err := deleteSave(ls.s.saves[ls.idx].Name); err != nil {
				ls.s.loadMsg = fmt.Sprintf("Delete failed: %v", err)
			} else {
				names, _ := listSaveFiles()
				ls.s.saves = make([]SaveInfo, 0, len(names))
				for _, n := range names {
					ls.s.saves = append(ls.s.saves, peekSaveInfo(n))
				}
				ls.s.selIdx = -1
			}
			return true
		}
	}
	return true
}

func (ls *loadSlot) HandleJustReleasedMouseButtonLeft(x, y int) {}
