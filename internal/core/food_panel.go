package core

import (
	"fmt"

	"github.com/kamil5b/basic-life-sim/internal/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// foodPanel lets the player see all food sources and eat, move-to-fridge, or cook them.
type foodPanel struct {
	char    *model.Character
	main    *mainScreen
	selIdx  int // index into gathered sources
	sources []foodSource
}

// foodSource is a flat view of any food the player has.
type foodSource struct {
	label         string
	purchaseDate  model.CompactDate
	multiplier    float32
	usesRemaining uint8
	cooked        bool
	food          model.Food
	kind          string // "floor", "fridge", "surface"
	itemIdx       int
	foodIdx       int
}

func newFoodPanel(char *model.Character, main *mainScreen) *foodPanel {
	return &foodPanel{char: char, main: main, selIdx: -1}
}

func (p *foodPanel) gatherSources() []foodSource {
	char := p.char
	var out []foodSource
	for i, f := range char.CurrentHome.FloorFood {
		out = append(out, foodSource{
			label: f.Food.Name, purchaseDate: f.PurchaseDate,
			multiplier: 1, usesRemaining: f.UsesRemaining,
			food: f.Food, kind: "floor", foodIdx: i,
		})
	}
	for ri, placed := range char.CurrentHome.RoomItems {
		if placed.Item.Storage != nil {
			for fi, s := range placed.Stored {
				out = append(out, foodSource{
					label: s.Food.Name, purchaseDate: s.PurchaseDate,
					multiplier: s.MultiplierUsed, usesRemaining: s.UsesRemaining,
					food: s.Food, kind: "fridge", itemIdx: ri, foodIdx: fi,
				})
			}
		}
		if placed.Item.CookSurface != nil {
			for fi, s := range placed.OnSurface {
				out = append(out, foodSource{
					label: s.Food.Name, purchaseDate: s.PurchaseDate,
					multiplier: 1, usesRemaining: s.UsesRemaining,
					cooked: s.Cooked,
					food:   s.Food, kind: "surface", itemIdx: ri, foodIdx: fi,
				})
			}
		}
	}
	return out
}

func (p *foodPanel) update() {
	p.sources = p.gatherSources()
	mx, my := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if !clicked {
		return
	}

	for i := range p.sources {
		rx, ry, rw, rh := fpRowRect(i)
		if isHovered(mx, my, rx, ry, rw, rh) {
			p.selIdx = i
		}
	}

	if p.selIdx < 0 || p.selIdx >= len(p.sources) {
		return
	}
	src := p.sources[p.selIdx]
	char := p.char

	// Eat button
	ex, ey, ew, eh := fpActionBtnRect(0)
	if isHovered(mx, my, ex, ey, ew, eh) {
		if model.IsExpired(src.purchaseDate, src.food.BaseExpiryDays, src.multiplier, char.CurrentDate) {
			foodExpiredPenalty(char)
			consumeUse(char, src.kind, src.itemIdx, src.foodIdx)
			p.main.setMessage(fmt.Sprintf("%s was EXPIRED — all stats -10!", src.food.Name))
			p.selIdx = -1
			return
		}
		if src.food.EatAction == nil {
			rawFoodPenalty(char)
			p.main.setMessage(fmt.Sprintf("Eating raw %s penalised Food/Energy/Hygiene -5.", src.food.Name))
		} else {
			src.food.EatAction(&char.CurrentStats, char)
			p.main.setMessage(fmt.Sprintf("Ate %s. Uses left: %d", src.food.Name, src.usesRemaining-1))
		}
		consumeUse(char, src.kind, src.itemIdx, src.foodIdx)
		p.selIdx = -1
		return
	}

	// Move to fridge button
	mx2, my2, mw, mh := fpActionBtnRect(1)
	if isHovered(mx, my, mx2, my2, mw, mh) && src.kind != "fridge" {
		fridges := findFridges(char)
		if len(fridges) == 0 {
			p.main.setMessage("No fridge in room. Buy one from the Shop.")
			return
		}
		placed := &char.CurrentHome.RoomItems[fridges[0]]
		cap := placed.Item.Storage
		slot, ok := nextFreeSlot(placed, cap, src.food)
		if !ok {
			p.main.setMessage("Fridge is full!")
			return
		}
		mult := cap.MultiplierAt(slot[0], slot[1], slot[2])
		removeFoodEntirely(char, src.kind, src.itemIdx, src.foodIdx)
		placed.Stored = append(placed.Stored, model.StoredFood{
			SlotX: slot[0], SlotY: slot[1], SlotZ: slot[2],
			PurchaseDate:   src.purchaseDate,
			MultiplierUsed: mult,
			UsesRemaining:  src.usesRemaining,
			Food:           src.food,
		})
		expiry := model.ExpiryDate(src.purchaseDate, src.food.BaseExpiryDays, mult)
		ey2, em, ed := expiry.Unpack()
		p.main.setMessage(fmt.Sprintf("Moved %s to fridge slot (%d,%d,%d), exp %04d-%02d-%02d",
			src.food.Name, slot[0], slot[1], slot[2], ey2, em, ed))
		p.selIdx = -1
		return
	}

	// Place on stove button
	px, py, pw, ph := fpActionBtnRect(2)
	if isHovered(mx, my, px, py, pw, ph) && src.food.CanBeCooked && src.kind != "surface" {
		stoves := findStoves(char)
		if len(stoves) == 0 {
			p.main.setMessage("No stove in room. Buy one from the Shop.")
			return
		}
		placed := &char.CurrentHome.RoomItems[stoves[0]]
		if len(placed.OnSurface) >= int(placed.Item.CookSurface.Slots) {
			p.main.setMessage("Stove surface is full.")
			return
		}
		removeFoodEntirely(char, src.kind, src.itemIdx, src.foodIdx)
		placed.OnSurface = append(placed.OnSurface, model.SurfaceFood{
			PurchaseDate:  src.purchaseDate,
			UsesRemaining: src.usesRemaining,
			Food:          src.food,
		})
		p.main.setMessage(fmt.Sprintf("Placed %s on stove. Go to Room tab → stove → cook.", src.food.Name))
		p.selIdx = -1
		return
	}
}

func findStoves(char *model.Character) []int {
	var out []int
	for i, p := range char.CurrentHome.RoomItems {
		if p.Item.CookSurface != nil {
			out = append(out, i)
		}
	}
	return out
}

const (
	fpListX  = panelX + 8
	fpListW  = panelW - 16
	fpListY  = panelY + 8
	fpRowH   = float32(28)
	fpBtnY   = float32(ScreenH) - 70
	fpBtnW   = float32(160)
	fpBtnH   = float32(36)
	fpBtnGap = float32(12)
)

func fpRowRect(i int) (x, y, w, h float32) {
	return fpListX, fpListY + float32(i)*fpRowH, fpListW, fpRowH - 2
}

func fpActionBtnRect(i int) (x, y, w, h float32) {
	x = fpListX + float32(i)*(fpBtnW+fpBtnGap)
	y = fpBtnY
	w = fpBtnW
	h = fpBtnH
	return
}

func (p *foodPanel) draw(dst *ebiten.Image) {
	p.sources = p.gatherSources()
	mx, my := ebiten.CursorPosition()
	char := p.char

	if len(p.sources) == 0 {
		drawText(dst, "No food available. Buy food from the Shop tab.", float64(fpListX), float64(fpListY)+20, fontM, colorMuted)
		return
	}

	drawText(dst, "All Food  (click to select, then use action buttons)", float64(fpListX), float64(panelY)+2, fontS, colorMuted)

	for i, src := range p.sources {
		rx, ry, rw, rh := fpRowRect(i)
		ry += 20
		sel := i == p.selIdx
		bg := colorPanel
		if sel {
			bg = colorSelected
		} else if isHovered(mx, my, rx, ry, rw, rh) {
			bg = colorHighlight
		}
		fillRect(dst, rx, ry, rw, rh, bg)
		strokeRect(dst, rx, ry, rw, rh, colorBorder)

		exp := model.IsExpired(src.purchaseDate, src.food.BaseExpiryDays, src.multiplier, char.CurrentDate)
		expiry := model.ExpiryDate(src.purchaseDate, src.food.BaseExpiryDays, src.multiplier)
		ey, em, ed := expiry.Unpack()
		daysLeft := model.DaysBetween(char.CurrentDate, expiry)

		tag := fmt.Sprintf("[%s]", src.kind)
		if src.kind == "surface" && src.cooked {
			tag = "[surface-cooked]"
		}
		inedible := ""
		if src.food.EatAction == nil {
			inedible = " [raw/inedible]"
		}
		expTag := fmt.Sprintf("exp %04d-%02d-%02d (%dd left)", ey, em, ed, daysLeft)
		if exp {
			expTag = fmt.Sprintf("EXPIRED %04d-%02d-%02d", ey, em, ed)
		}
		label := fmt.Sprintf("%-22s uses:%-3d %-18s%s  %s",
			src.label, src.usesRemaining, tag, inedible, expTag)

		tc := colorText
		if exp {
			tc = colorRed
		} else if src.food.EatAction == nil {
			tc = colorYellow
		}
		drawText(dst, label, float64(rx)+8, float64(ry)+7, fontS, tc)
	}

	// action buttons
	if p.selIdx >= 0 && p.selIdx < len(p.sources) {
		src := p.sources[p.selIdx]

		eatLabel := "Eat"
		ex, ey2, ew, eh := fpActionBtnRect(0)
		drawButton(dst, eatLabel, ex, ey2, ew, eh, fontM, isHovered(mx, my, ex, ey2, ew, eh), true)

		fridgeEnabled := src.kind != "fridge" && len(findFridges(char)) > 0
		mx2, my2, mw, mh := fpActionBtnRect(1)
		drawButton(dst, "→ Fridge", mx2, my2, mw, mh, fontM, isHovered(mx, my, mx2, my2, mw, mh) && fridgeEnabled, fridgeEnabled)

		stoveEnabled := src.food.CanBeCooked && src.kind != "surface" && len(findStoves(char)) > 0
		px, py, pw, ph := fpActionBtnRect(2)
		drawButton(dst, "→ Stove", px, py, pw, ph, fontM, isHovered(mx, my, px, py, pw, ph) && stoveEnabled, stoveEnabled)
	}
}
