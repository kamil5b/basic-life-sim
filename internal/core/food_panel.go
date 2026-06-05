package core

import (
	"fmt"
	"image/color"

	"github.com/kamil5b/basic-life-sim/internal/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type fpMode int

const (
	fpModeNormal fpMode = iota
	fpModeMove
)

// foodPanel lets the player see all food sources and eat, move-to-fridge, or cook them.
type foodPanel struct {
	char    *model.Character
	main    *mainScreen
	selIdx  int // index into gathered sources
	sources []foodSource
	mode    fpMode
	moveIdx int // index into sources of the item being moved
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
	return &foodPanel{char: char, main: main, selIdx: -1, moveIdx: -1}
}

func (p *foodPanel) gatherSources() []foodSource {
	char := p.char
	var out []foodSource
	for i, fi := range char.CurrentHome.FloorItems {
		if fi.Kind != model.FloorKindFood {
			continue
		}
		out = append(out, foodSource{
			label: fi.Food.Name, purchaseDate: fi.PurchaseDate,
			multiplier: 1, usesRemaining: fi.UsesRemaining,
			food: fi.Food, kind: "floor", foodIdx: i,
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

	// Move mode: picking a target cell on the room grid
	if p.mode == fpModeMove {
		// Cancel
		if clicked {
			cx2, cy2, cw, ch := fpCancelBtnRect()
			if isHovered(mx, my, cx2, cy2, cw, ch) {
				p.mode = fpModeNormal
				p.moveIdx = -1
				return
			}
			// Click on grid cell
			gx, gy := gridCellAtOrigin(mx, my, fpGridOriginX, fpGridOriginY, p.char)
			if gx >= 0 && p.moveIdx >= 0 && p.moveIdx < len(p.sources) {
				src := p.sources[p.moveIdx]
				if src.kind == "floor" {
					// move to target cell
					char := p.char
					char.CurrentHome.FloorItems[src.foodIdx].X = uint8(gx)
					char.CurrentHome.FloorItems[src.foodIdx].Y = uint8(gy)
					p.main.setMessage(fmt.Sprintf("Moved %s to (%d,%d).", src.food.Name, gx, gy))
				}
				p.mode = fpModeNormal
				p.moveIdx = -1
			}
		}
		return
	}

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
		inedible := isInedible(src.food)
		if inedible {
			rawFoodPenalty(char)
			p.main.setMessage(fmt.Sprintf("Eating raw %s penalised Food/Energy/Hygiene -5.", src.food.Name))
		} else {
			applyNutrition(char, src.food)
			p.main.setMessage(fmt.Sprintf("Ate %s. Uses left: %d", src.food.Name, src.usesRemaining-1))
		}
		consumeUse(char, src.kind, src.itemIdx, src.foodIdx)
		p.selIdx = -1
		return
	}

	// → Fridge button
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

	// → Stove button
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

	// → Floor button (take out of fridge)
	fx, fy, fw, fh := fpActionBtnRect(3)
	if isHovered(mx, my, fx, fy, fw, fh) && src.kind == "fridge" {
		daysElapsed := model.DaysBetween(src.purchaseDate, char.CurrentDate)
		fridgeDaysConsumed := float32(daysElapsed) / src.multiplier
		remaining := float32(src.food.BaseExpiryDays) - fridgeDaysConsumed
		if remaining < 1 {
			remaining = 1
		}
		newFood := src.food
		newFood.BaseExpiryDays = uint16(remaining)
		removeFoodEntirely(char, src.kind, src.itemIdx, src.foodIdx)
		char.CurrentHome.FloorItems = append(char.CurrentHome.FloorItems, model.FloorItem{
			X: 0, Y: 0, Z: 0,
			Kind:          model.FloorKindFood,
			PurchaseDate:  char.CurrentDate,
			UsesRemaining: src.usesRemaining,
			Food:          newFood,
		})
		expiry := model.ExpiryDate(char.CurrentDate, newFood.BaseExpiryDays, 1)
		ey2, em, ed := expiry.Unpack()
		p.main.setMessage(fmt.Sprintf("Took %s out of fridge → floor. New exp %04d-%02d-%02d",
			src.food.Name, ey2, em, ed))
		p.selIdx = -1
		return
	}

	// ✦ Move button (floor food)
	mvx, mvy, mvw, mvh := fpActionBtnRect(4)
	if isHovered(mx, my, mvx, mvy, mvw, mvh) && src.kind == "floor" {
		p.moveIdx = p.selIdx
		p.mode = fpModeMove
		p.main.setMessage(fmt.Sprintf("Moving %s — click a floor cell.", src.food.Name))
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
	fpGridOriginX = panelX + 320
	fpGridOriginY = panelY + 40
)

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

func fpCancelBtnRect() (x, y, w, h float32) {
	return fpGridOriginX, fpGridOriginY - 26, 90, 22
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
		if isInedible(src.food) {
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
		} else if isInedible(src.food) {
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

		floorEnabled := src.kind == "fridge"
		fax, fay, faw, fah := fpActionBtnRect(3)
		drawButton(dst, "→ Floor", fax, fay, faw, fah, fontM, isHovered(mx, my, fax, fay, faw, fah) && floorEnabled, floorEnabled)

		moveEnabled := src.kind == "floor"
		mvx, mvy, mvw, mvh := fpActionBtnRect(4)
		drawButton(dst, "✦ Move", mvx, mvy, mvw, mvh, fontM, isHovered(mx, my, mvx, mvy, mvw, mvh) && moveEnabled, moveEnabled)
	}

	// Move mode: draw grid overlay
	if p.mode == fpModeMove {
		home := p.char.CurrentHome
		layout := home.Type.Layout
		if len(layout) > 0 {
			ox := fpGridOriginX
			oy := fpGridOriginY
			rows := len(layout)
			cols := len(layout[0])
			drawText(dst, "Click cell to move food here:", float64(ox), float64(oy)-30, fontS, colorAccent)
			cx2, cy2, cw, ch := fpCancelBtnRect()
			drawButton(dst, "✕ Cancel", cx2, cy2, cw, ch, fontS, isHovered(mx, my, cx2, cy2, cw, ch), true)
			for r := 0; r < rows; r++ {
				for c := 0; c < cols; c++ {
					cellX := ox + float32(c)*rpCellSz
					cellY := oy + float32(r)*rpCellSz
					var bg color.RGBA
					switch layout[r][c] {
					case model.HomeCellWall:
						bg = colorWall
					case model.HomeCellFloor:
						bg = colorFloor
					case model.HomeCellDoor:
						bg = colorDoor
					}
					fillRect(dst, cellX, cellY, rpCellSz-1, rpCellSz-1, bg)
				}
			}
			// existing items overlay
			for _, placed := range home.RoomItems {
				for _, cell := range occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction) {
					cellX := ox + float32(cell[0])*rpCellSz
					cellY := oy + float32(cell[1])*rpCellSz
					fillRect(dst, cellX+1, cellY+1, rpCellSz-3, rpCellSz-3, colorItem)
				}
			}
			// floor food overlay
			for _, fi := range home.FloorItems {
				cellX := ox + float32(fi.X)*rpCellSz
				cellY := oy + float32(fi.Y)*rpCellSz
				if fi.Kind == model.FloorKindFood {
					fillRect(dst, cellX+8, cellY+8, rpCellSz-17, rpCellSz-17, colorFoodFloor)
				} else {
					fillRect(dst, cellX+4, cellY+4, rpCellSz-9, rpCellSz-9, colorUtilFloor)
				}
			}
			// cursor hover highlight
			hx, hy := gridCellAtOrigin(mx, my, ox, oy, p.char)
			if hx >= 0 {
				cellX := ox + float32(hx)*rpCellSz
				cellY := oy + float32(hy)*rpCellSz
				strokeRect(dst, cellX, cellY, rpCellSz-1, rpCellSz-1, colorAccent)
			}
		}
	}
}
