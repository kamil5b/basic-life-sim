package core

import (
	"basic-life-sim/internal/model"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// roomPanel draws the room grid and item list. Clicking an item shows its actions.
type roomPanel struct {
	char    *model.Character
	main    *mainScreen
	selItem int    // index into RoomItems, -1 = none
	selAct  int    // index into selected item's actions
	mode    rpMode // what we're currently showing on the right
}

type rpMode int

const (
	rpModeList   rpMode = iota // list of placed items
	rpModeAction               // item action picker
)

func newRoomPanel(char *model.Character, main *mainScreen) *roomPanel {
	return &roomPanel{char: char, main: main, selItem: -1}
}

func (p *roomPanel) update() {
	mx, my := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	char := p.char

	switch p.mode {
	case rpModeList:
		if clicked {
			items := char.CurrentHome.RoomItems
			for i := range items {
				rx, ry, rw, rh := rpItemRowRect(i)
				if isHovered(mx, my, rx, ry, rw, rh) {
					p.selItem = i
					p.selAct = 0
					p.mode = rpModeAction
					return
				}
			}
		}

	case rpModeAction:
		if p.selItem < 0 || p.selItem >= len(char.CurrentHome.RoomItems) {
			p.mode = rpModeList
			return
		}
		placed := &char.CurrentHome.RoomItems[p.selItem]
		actions := placed.Item.Actions

		// back button
		bx, by, bw, bh := rpBackBtnRect()
		if clicked && isHovered(mx, my, bx, by, bw, bh) {
			p.mode = rpModeList
			p.selItem = -1
			return
		}

		for i, act := range actions {
			ax, ay, aw, ah := rpActionRowRect(i)
			if clicked && isHovered(mx, my, ax, ay, aw, ah) {
				p.selAct = i
				p.executeAction(placed, act)
			}
		}
	}
}

func (p *roomPanel) executeAction(placed *model.PlacedRoomItem, action string) {
	char := p.char
	if placed.Item.Storage != nil {
		switch action {
		case "eat":
			if len(placed.Stored) == 0 {
				p.main.setMessage("Fridge is empty.")
				return
			}
			// eat the first non-expired item
			for i, s := range placed.Stored {
				if model.IsExpired(s.PurchaseDate, s.Food.BaseExpiryDays, s.MultiplierUsed, char.CurrentDate) {
					foodExpiredPenalty(char)
					placed.Stored[i].UsesRemaining--
					if placed.Stored[i].UsesRemaining == 0 {
						placed.Stored = append(placed.Stored[:i], placed.Stored[i+1:]...)
					}
					p.main.setMessage(fmt.Sprintf("%s was EXPIRED — stat penalty applied!", s.Food.Name))
					return
				}
				if s.Food.EatAction == nil {
					rawFoodPenalty(char)
					p.main.setMessage(fmt.Sprintf("Eating raw %s was a bad idea.", s.Food.Name))
				} else {
					s.Food.EatAction(&char.CurrentStats, char)
					p.main.setMessage(fmt.Sprintf("Ate %s from fridge. Uses left: %d", s.Food.Name, s.UsesRemaining-1))
				}
				consumeUse(char, "fridge", p.selItem, i)
				return
			}
		case "store food":
			p.main.setMessage("Use the Food tab to move food to the fridge.")
		}
		return
	}

	if placed.Item.CookSurface != nil {
		switch action {
		case "cook":
			if len(placed.OnSurface) == 0 {
				p.main.setMessage("Nothing on the surface. Place food first via Food tab.")
				return
			}
			cooked := 0
			for i := range placed.OnSurface {
				sf := &placed.OnSurface[i]
				if sf.Cooked || sf.Food.CookedResult == nil {
					continue
				}
				result := *sf.Food.CookedResult
				result.UsesTotal = sf.Food.UsesTotal
				sf.Food = result
				sf.Cooked = true
				cooked++
			}
			char.Energy.Current = safeSub16(char.Energy.Current, 5)
			if cooked > 0 {
				p.main.setMessage(fmt.Sprintf("Cooked %d item(s). Energy -5.", cooked))
			} else {
				p.main.setMessage("Nothing to cook (already cooked or not cookable).")
			}
		case "place food":
			p.main.setMessage("Use the Food tab to place food on the surface.")
		case "take out":
			if len(placed.OnSurface) == 0 {
				p.main.setMessage("Surface is empty.")
				return
			}
			sf := placed.OnSurface[0]
			placed.OnSurface = placed.OnSurface[1:]
			char.CurrentHome.FloorFood = append(char.CurrentHome.FloorFood, model.PlacedFood{
				PurchaseDate:  sf.PurchaseDate,
				UsesRemaining: sf.UsesRemaining,
				Food:          sf.Food,
			})
			p.main.setMessage(fmt.Sprintf("Took %s off surface, placed on floor.", sf.Food.Name))
		}
		return
	}

	// generic action
	placed.Item.DoAction(action, &char.CurrentStats, char)
	p.main.setMessage(fmt.Sprintf("Used %s: %s", placed.Item.Name, action))
}

// ── layout helpers ────────────────────────────────────────────────────────────

const (
	rpListX  = panelX + 4
	rpListW  = 260
	rpListY  = panelY + 4
	rpRowH   = float32(26)
	rpGridX  = rpListX + rpListW + 10
	rpGridY  = panelY + 4
	rpCellSz = float32(36)
)

func rpItemRowRect(i int) (x, y, w, h float32) {
	return rpListX, rpListY + float32(i)*rpRowH, rpListW, rpRowH - 2
}

func rpActionRowRect(i int) (x, y, w, h float32) {
	return rpListX, rpListY + float32(i)*42, rpListW, 36
}

func rpBackBtnRect() (x, y, w, h float32) {
	return rpListX, panelY + float32(ScreenH) - float32(tabH) - 50, 100, 32
}

func (p *roomPanel) draw(dst *ebiten.Image) {
	mx, my := ebiten.CursorPosition()
	char := p.char

	// draw room grid on the right portion
	p.drawRoomGrid(dst)

	switch p.mode {
	case rpModeList:
		drawText(dst, "Placed Items", float64(rpListX), float64(panelY)+4, fontS, colorMuted)
		if len(char.CurrentHome.RoomItems) == 0 && len(char.CurrentHome.FloorFood) == 0 {
			drawText(dst, "Room is empty. Buy items in the Shop tab.", float64(rpListX), float64(panelY)+28, fontS, colorMuted)
		}
		for i, placed := range char.CurrentHome.RoomItems {
			rx, ry, rw, rh := rpItemRowRect(i)
			ry += 20
			hov := isHovered(mx, my, rx, ry, rw, rh)
			bg := colorPanel
			if hov {
				bg = colorHighlight
			}
			fillRect(dst, rx, ry, rw, rh, bg)
			strokeRect(dst, rx, ry, rw, rh, colorBorder)
			label := fmt.Sprintf("%s  (%d,%d) %s", placed.Item.Name, placed.X, placed.Y, dirName(placed.Direction))
			drawText(dst, label, float64(rx)+6, float64(ry)+6, fontS, colorText)
		}
		// floor food
		if len(char.CurrentHome.FloorFood) > 0 {
			yo := panelY + 20 + float32(len(char.CurrentHome.RoomItems))*rpRowH + 12
			drawText(dst, "Floor Food:", float64(rpListX), float64(yo), fontS, colorMuted)
			yo += 18
			for _, ff := range char.CurrentHome.FloorFood {
				drawText(dst, fmt.Sprintf("  %s (uses:%d)", ff.Food.Name, ff.UsesRemaining), float64(rpListX), float64(yo), fontS, colorFoodFloor)
				yo += 18
			}
		}

	case rpModeAction:
		if p.selItem < 0 || p.selItem >= len(char.CurrentHome.RoomItems) {
			return
		}
		placed := char.CurrentHome.RoomItems[p.selItem]
		drawText(dst, placed.Item.Name, float64(rpListX), float64(panelY)+4, fontM, colorAccent)
		drawText(dst, fmt.Sprintf("at (%d,%d,z=%d) facing %s", placed.X, placed.Y, placed.Z, dirName(placed.Direction)),
			float64(rpListX), float64(panelY)+24, fontS, colorMuted)

		// storage info
		if placed.Item.Storage != nil {
			used := usedSlotCount(&char.CurrentHome.RoomItems[p.selItem])
			total := placed.Item.Storage.TotalSlots()
			drawText(dst, fmt.Sprintf("Storage: %d/%d slots", used, total),
				float64(rpListX), float64(panelY)+44, fontS, colorText)
			// list stored
			iy := float64(panelY) + 64
			for _, s := range placed.Stored {
				exp := model.IsExpired(s.PurchaseDate, s.Food.BaseExpiryDays, s.MultiplierUsed, char.CurrentDate)
				col := colorText
				tag := ""
				if exp {
					col = colorRed
					tag = " [EXPIRED]"
				}
				expiry := model.ExpiryDate(s.PurchaseDate, s.Food.BaseExpiryDays, s.MultiplierUsed)
				ey, em, ed := expiry.Unpack()
				drawText(dst, fmt.Sprintf("  %s  uses:%d  exp:%04d-%02d-%02d%s",
					s.Food.Name, s.UsesRemaining, ey, em, ed, tag),
					float64(rpListX), iy, fontS, col)
				iy += 18
			}
		}

		// surface info
		if placed.Item.CookSurface != nil {
			drawText(dst, fmt.Sprintf("Surface: %d/%d slots", len(placed.OnSurface), placed.Item.CookSurface.Slots),
				float64(rpListX), float64(panelY)+44, fontS, colorText)
			iy := float64(panelY) + 64
			for _, sf := range placed.OnSurface {
				tag := ""
				if sf.Cooked {
					tag = " [cooked]"
				}
				drawText(dst, fmt.Sprintf("  %s%s uses:%d", sf.Food.Name, tag, sf.UsesRemaining),
					float64(rpListX), iy, fontS, colorText)
				iy += 18
			}
		}

		// action buttons
		actY := float64(panelY) + 180
		drawText(dst, "Actions:", float64(rpListX), actY, fontS, colorMuted)
		actY += 20
		for i, act := range placed.Item.Actions {
			ax, ay, aw, ah := rpActionRowRect(i)
			ay = float32(actY) + float32(i)*42
			hov := isHovered(mx, my, ax, ay, aw, ah)
			drawButton(dst, act, ax, ay, aw, ah, fontM, hov, true)
		}

		// back button
		bx, by, bw, bh := rpBackBtnRect()
		drawButton(dst, "← Back", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)
	}
}

func (p *roomPanel) drawRoomGrid(dst *ebiten.Image) {
	home := p.char.CurrentHome
	layout := home.Type.Layout
	if len(layout) == 0 {
		return
	}
	ox := rpGridX
	oy := rpGridY + 20

	rows := len(layout)
	cols := len(layout[0])

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			cx := ox + float32(c)*rpCellSz
			cy := oy + float32(r)*rpCellSz
			var bg color.RGBA
			switch layout[r][c] {
			case model.HomeCellWall:
				bg = colorWall
			case model.HomeCellFloor:
				bg = colorFloor
			case model.HomeCellDoor:
				bg = colorDoor
			}
			fillRect(dst, cx, cy, rpCellSz-1, rpCellSz-1, bg)
		}
	}

	// overlay placed items
	for _, placed := range home.RoomItems {
		for _, cell := range occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction) {
			cx := ox + float32(cell[0])*rpCellSz
			cy := oy + float32(cell[1])*rpCellSz
			fillRect(dst, cx+1, cy+1, rpCellSz-3, rpCellSz-3, colorItem)
		}
		// first letter in top-left of anchor cell
		lx := ox + float32(placed.X)*rpCellSz + 4
		ly := oy + float32(placed.Y)*rpCellSz + 4
		label := string([]rune(placed.Item.Name)[0:1])
		drawText(dst, label, float64(lx), float64(ly), fontS, colorBg)
		// facing arrow in anchor cell
		drawFacingArrow(dst, ox, oy, int(placed.X), int(placed.Y), placed.Direction)
	}

	// overlay floor food
	for _, ff := range home.FloorFood {
		cx := ox + float32(ff.X)*rpCellSz
		cy := oy + float32(ff.Y)*rpCellSz
		fillRect(dst, cx+8, cy+8, rpCellSz-17, rpCellSz-17, colorFoodFloor)
	}

	// title
	drawText(dst, fmt.Sprintf("Room: %s", home.Type.Name), float64(ox), float64(oy)-18, fontS, colorMuted)
}
