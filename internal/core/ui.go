package core

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── palette ──────────────────────────────────────────────────────────────────

var (
	colorBg        = color.RGBA{18, 18, 24, 255}
	colorPanel     = color.RGBA{28, 28, 38, 255}
	colorBorder    = color.RGBA{60, 60, 80, 255}
	colorText      = color.RGBA{220, 220, 230, 255}
	colorMuted     = color.RGBA{120, 120, 140, 255}
	colorAccent    = color.RGBA{100, 180, 255, 255}
	colorGreen     = color.RGBA{80, 200, 120, 255}
	colorRed       = color.RGBA{220, 80, 80, 255}
	colorYellow    = color.RGBA{240, 200, 60, 255}
	colorHighlight = color.RGBA{50, 50, 70, 255}
	colorSelected  = color.RGBA{70, 110, 170, 255}
	colorWall      = color.RGBA{60, 60, 80, 255}
	colorFloor     = color.RGBA{35, 35, 48, 255}
	colorDoor      = color.RGBA{160, 120, 60, 255}
	colorItem      = color.RGBA{100, 160, 240, 255}
	colorFoodFloor = color.RGBA{180, 140, 60, 255}
)

// ── drawing primitives ────────────────────────────────────────────────────────

func fillRect(dst *ebiten.Image, x, y, w, h float32, c color.RGBA) {
	vector.DrawFilledRect(dst, x, y, w, h, c, false)
}

func strokeRect(dst *ebiten.Image, x, y, w, h float32, c color.RGBA) {
	vector.StrokeRect(dst, x, y, w, h, 1, c, false)
}

func drawText(dst *ebiten.Image, str string, x, y float64, fnt *text.GoTextFace, clr color.RGBA) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(dst, str, fnt, op)
}

func drawTextWrapped(dst *ebiten.Image, str string, x, y, maxWidth, lineH float64, fnt *text.GoTextFace, clr color.RGBA) float64 {
	words := strings.Fields(str)
	line := ""
	cy := y
	for _, w := range words {
		test := line
		if test != "" {
			test += " "
		}
		test += w
		adv, _ := text.Measure(test, fnt, 0)
		if adv > maxWidth && line != "" {
			drawText(dst, line, x, cy, fnt, clr)
			cy += lineH
			line = w
		} else {
			line = test
		}
	}
	if line != "" {
		drawText(dst, line, x, cy, fnt, clr)
		cy += lineH
	}
	return cy
}

func isHovered(mx, my int, x, y, w, h float32) bool {
	return float32(mx) >= x && float32(mx) < x+w && float32(my) >= y && float32(my) < y+h
}

// button draws a styled button box (caller decides click logic separately).
func drawButton(dst *ebiten.Image, label string, x, y, w, h float32, fnt *text.GoTextFace, hovered, enabled bool) {
	bg := colorPanel
	border := colorBorder
	txtColor := colorText
	if !enabled {
		txtColor = colorMuted
	} else if hovered {
		bg = colorHighlight
		border = colorAccent
	}
	fillRect(dst, x, y, w, h, bg)
	strokeRect(dst, x, y, w, h, border)
	tw, th := text.Measure(label, fnt, 0)
	tx := float64(x) + float64(w)/2 - tw/2
	ty := float64(y) + float64(h)/2 - th/2
	drawText(dst, label, tx, ty, fnt, txtColor)
}

// needBar draws a labelled progress bar.
func needBar(dst *ebiten.Image, label string, cur, max uint16, x, y, w float32, fnt *text.GoTextFace) {
	const h = float32(14)
	pct := float32(0)
	if max > 0 {
		pct = float32(cur) / float32(max)
	}
	barCol := colorGreen
	if pct < 0.3 {
		barCol = colorRed
	} else if pct < 0.6 {
		barCol = colorYellow
	}
	drawText(dst, label, float64(x), float64(y), fnt, colorMuted)
	bx := x + 84
	bw := w - 88
	fillRect(dst, bx, y, bw, h, colorBorder)
	if pct > 0 {
		fillRect(dst, bx, y, bw*pct, h, barCol)
	}
	strokeRect(dst, bx, y, bw, h, colorBorder)
	valStr := fmt.Sprintf("%d/%d", cur, max)
	vw, _ := text.Measure(valStr, fnt, 0)
	drawText(dst, valStr, float64(bx)+float64(bw)/2-vw/2, float64(y), fnt, colorText)
}

// listRow draws a single list row and returns true if hovered.
func listRow(dst *ebiten.Image, label string, idx, selIdx int, x, y, w, rowH float32, fnt *text.GoTextFace, mx, my int) bool {
	hov := isHovered(mx, my, x, y, w, rowH)
	bg := colorPanel
	if idx == selIdx {
		bg = colorSelected
	} else if hov {
		bg = colorHighlight
	}
	fillRect(dst, x, y, w, rowH, bg)
	strokeRect(dst, x, y, w, rowH, colorBorder)
	drawText(dst, label, float64(x)+8, float64(y)+float64(rowH)/2-7, fnt, colorText)
	return hov
}
