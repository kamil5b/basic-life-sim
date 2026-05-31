package core

import (
	"basic-life-sim/internal/model"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// roomPanel draws the room grid and item list.
type roomPanel struct {
	char *model.Character
	main *mainScreen

	// list / hover state
	hoverCell [2]int // grid cell the mouse is over; [-1,-1] = none
	selCell   [2]int // clicked cell for filtering; [-1,-1] = no filter active

	// selected item (in filtered or action mode)
	selItem int // index into RoomItems, -1 = none
	selAct  int

	mode   rpMode
	wizard *placementWizard
}

type rpMode int

const (
	rpModeList     rpMode = iota // full list, hover highlights on grid
	rpModeFiltered               // cell clicked — show only items stacked there
	rpModeAction                 // action picker for a selected item
	rpModeMoveGrid               // picking new X,Y for an item
	rpModeMoveZ                  // picking new Z
	rpModeMoveDir                // picking new direction
)

func newRoomPanel(char *model.Character, main *mainScreen) *roomPanel {
	return &roomPanel{
		char:      char,
		main:      main,
		selItem:   -1,
		hoverCell: [2]int{-1, -1},
		selCell:   [2]int{-1, -1},
	}
}

// itemsAtCell returns the indices (into RoomItems) of items that occupy the grid cell (gx,gy)
// at any Z level.
func (p *roomPanel) itemsAtCell(gx, gy int) []int {
	var out []int
	for i, placed := range p.char.CurrentHome.RoomItems {
		for _, cell := range occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction) {
			if int(cell[0]) == gx && int(cell[1]) == gy {
				out = append(out, i)
				break
			}
		}
	}
	return out
}

func (p *roomPanel) gridCellAt(px, py int) (gx, gy int) {
	ox := int(rpGridX)
	oy := int(rpGridY) + 20
	cs := int(rpCellSz)
	layout := p.char.CurrentHome.Type.Layout
	rows := len(layout)
	if rows == 0 {
		return -1, -1
	}
	cols := len(layout[0])
	gx = (px - ox) / cs
	gy = (py - oy) / cs
	if gx < 0 || gy < 0 || gx >= cols || gy >= rows {
		return -1, -1
	}
	return gx, gy
}

func (p *roomPanel) update() {
	mx, my := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	switch p.mode {
	case rpModeList:
		// Always track hover cell
		gx, gy := p.gridCellAt(mx, my)
		p.hoverCell = [2]int{gx, gy}

		if clicked {
			if gx >= 0 {
				stacked := p.itemsAtCell(gx, gy)
				if len(stacked) > 0 {
					p.selCell = [2]int{gx, gy}
					p.selItem = -1
					p.mode = rpModeFiltered
					return
				}
			}
		}

	case rpModeFiltered:
		// Back if clicking outside the list area or on back button
		bx, by, bw, bh := rpBackBtnRect()
		if clicked && isHovered(mx, my, bx, by, bw, bh) {
			p.mode = rpModeList
			p.selCell = [2]int{-1, -1}
			p.selItem = -1
			return
		}

		stacked := p.itemsAtCell(p.selCell[0], p.selCell[1])
		for li, itemIdx := range stacked {
			rx, ry, rw, rh := rpItemRowRect(li)
			ry += 20
			if clicked && isHovered(mx, my, rx, ry, rw, rh) {
				p.selItem = itemIdx
				p.selAct = 0
				p.mode = rpModeAction
				return
			}
		}

	case rpModeAction:
		if p.selItem < 0 || p.selItem >= len(p.char.CurrentHome.RoomItems) {
			p.mode = rpModeFiltered
			return
		}
		placed := &p.char.CurrentHome.RoomItems[p.selItem]
		actions := placed.Item.Actions

		// back button
		bx, by, bw, bh := rpBackBtnRect()
		if clicked && isHovered(mx, my, bx, by, bw, bh) {
			p.mode = rpModeFiltered
			return
		}

		// move button
		mbx, mby, mbw, mbh := rpMoveBtnRect()
		if clicked && isHovered(mx, my, mbx, mby, mbw, mbh) {
			p.wizard = newPlacementWizard(
				p.char, &placed.Item, p.selItem,
				rpGridX, rpGridY+20,
				"← Cancel",
				func(x, y, z uint8, dir model.Direction) error {
					return p.finalizeMove(x, y, z, dir)
				},
				func() { p.wizard = nil; p.mode = rpModeAction },
			)
			p.mode = rpModeMoveGrid
			return
		}

		for i, act := range actions {
			ax, ay, aw, ah := rpActionRowRect(i)
			if clicked && isHovered(mx, my, ax, ay, aw, ah) {
				p.selAct = i
				p.executeAction(placed, act)
			}
		}

	case rpModeMoveGrid, rpModeMoveZ, rpModeMoveDir:
		if p.wizard != nil {
			p.wizard.update()
			if p.wizard != nil {
				switch p.wizard.step() {
				case pwStepZ:
					p.mode = rpModeMoveZ
				case pwStepDir:
					p.mode = rpModeMoveDir
				}
			}
		}
	}
}

func (p *roomPanel) finalizeMove(x, y, z uint8, dir model.Direction) error {
	if p.selItem < 0 || p.selItem >= len(p.char.CurrentHome.RoomItems) {
		return nil
	}
	home := &p.char.CurrentHome
	placed := &home.RoomItems[p.selItem]
	item := placed.Item

	// Temporarily remove self for clean canPlace check
	saved := home.RoomItems[p.selItem]
	home.RoomItems = append(home.RoomItems[:p.selItem], home.RoomItems[p.selItem+1:]...)

	finalZ := z
	if !item.CanOverhang {
		finalZ = 0
	}

	if err := canPlace(*home, item, x, y, finalZ, dir); err != nil {
		// Restore
		home.RoomItems = append(home.RoomItems[:p.selItem], append([]model.PlacedRoomItem{saved}, home.RoomItems[p.selItem:]...)...)
		return err
	}

	saved.X = x
	saved.Y = y
	saved.Z = finalZ
	saved.Direction = dir

	home.RoomItems = append(home.RoomItems[:p.selItem], append([]model.PlacedRoomItem{saved}, home.RoomItems[p.selItem:]...)...)

	// Drop any non-overhang items that no longer have support (their Z > dropZ)
	p.applyGravity()

	p.main.setMessage(fmt.Sprintf("Moved %s to (%d,%d,z=%d) facing %s", item.Name, x, y, finalZ, dirName(dir)))
	p.selCell = [2]int{int(x), int(y)}
	p.wizard = nil
	p.mode = rpModeFiltered
	return nil
}

// applyGravity scans all non-overhang items and drops them to their lowest valid Z.
func (p *roomPanel) applyGravity() {
	home := &p.char.CurrentHome
	for i := range home.RoomItems {
		pl := &home.RoomItems[i]
		if pl.Item.CanOverhang || pl.Z == 0 {
			continue
		}
		// Compute drop Z: find lowest Z where item fits without collision (excluding self).
		removed := home.RoomItems[i]
		home.RoomItems = append(home.RoomItems[:i], home.RoomItems[i+1:]...)
		bestZ := uint8(0)
		for z := uint8(0); z+removed.Item.Height <= home.Type.MaxHeight; z++ {
			if canPlace(*home, removed.Item, removed.X, removed.Y, z, removed.Direction) == nil {
				bestZ = z
				break
			}
		}
		removed.Z = bestZ
		home.RoomItems = append(home.RoomItems[:i], append([]model.PlacedRoomItem{removed}, home.RoomItems[i:]...)...)
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

func rpMoveBtnRect() (x, y, w, h float32) {
	return rpListX + 108, panelY + float32(ScreenH) - float32(tabH) - 50, 120, 32
}

// ── draw ──────────────────────────────────────────────────────────────────────

func (p *roomPanel) draw(dst *ebiten.Image) {
	mx, my := ebiten.CursorPosition()
	char := p.char

	p.drawRoomGrid(dst)

	switch p.mode {
	case rpModeList:
		drawText(dst, "Placed Items", float64(rpListX), float64(panelY)+4, fontS, colorMuted)
		if len(char.CurrentHome.RoomItems) == 0 && len(char.CurrentHome.FloorFood) == 0 {
			drawText(dst, "Room is empty. Buy items in the Shop tab.", float64(rpListX), float64(panelY)+28, fontS, colorMuted)
		}
		// Which items should be highlighted (hovered cell)
		hoveredIndices := make(map[int]bool)
		if p.hoverCell[0] >= 0 {
			for _, idx := range p.itemsAtCell(p.hoverCell[0], p.hoverCell[1]) {
				hoveredIndices[idx] = true
			}
		}
		for i, placed := range char.CurrentHome.RoomItems {
			rx, ry, rw, rh := rpItemRowRect(i)
			ry += 20
			highlighted := hoveredIndices[i]
			rowHov := isHovered(mx, my, rx, ry, rw, rh)
			bg := colorPanel
			if highlighted {
				bg = color.RGBA{60, 110, 60, 255} // green tint for grid-hover highlight
			} else if rowHov {
				bg = colorHighlight
			}
			fillRect(dst, rx, ry, rw, rh, bg)
			strokeRect(dst, rx, ry, rw, rh, colorBorder)
			label := fmt.Sprintf("%s  (%d,%d) z=%d %s", placed.Item.Name, placed.X, placed.Y, placed.Z, dirName(placed.Direction))
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
		if p.hoverCell[0] >= 0 && len(hoveredIndices) > 0 {
			drawText(dst, "Click cell to select", float64(rpListX), float64(panelY)+10, fontS, colorAccent)
		}

	case rpModeFiltered:
		stacked := p.itemsAtCell(p.selCell[0], p.selCell[1])
		drawText(dst, fmt.Sprintf("Cell (%d,%d) — %d item(s)", p.selCell[0], p.selCell[1], len(stacked)),
			float64(rpListX), float64(panelY)+4, fontS, colorAccent)
		for li, itemIdx := range stacked {
			placed := char.CurrentHome.RoomItems[itemIdx]
			rx, ry, rw, rh := rpItemRowRect(li)
			ry += 20
			hov := isHovered(mx, my, rx, ry, rw, rh)
			bg := colorPanel
			if hov {
				bg = colorHighlight
			}
			fillRect(dst, rx, ry, rw, rh, bg)
			strokeRect(dst, rx, ry, rw, rh, colorBorder)
			label := fmt.Sprintf("%s  z=%d %s", placed.Item.Name, placed.Z, dirName(placed.Direction))
			drawText(dst, label, float64(rx)+6, float64(ry)+6, fontS, colorText)
		}
		bx, by, bw, bh := rpBackBtnRect()
		drawButton(dst, "← Back", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)

	case rpModeAction:
		if p.selItem < 0 || p.selItem >= len(char.CurrentHome.RoomItems) {
			return
		}
		placed := char.CurrentHome.RoomItems[p.selItem]
		drawText(dst, placed.Item.Name, float64(rpListX), float64(panelY)+4, fontM, colorAccent)
		drawText(dst, fmt.Sprintf("at (%d,%d,z=%d) facing %s", placed.X, placed.Y, placed.Z, dirName(placed.Direction)),
			float64(rpListX), float64(panelY)+24, fontS, colorMuted)

		if placed.Item.Storage != nil {
			used := usedSlotCount(&char.CurrentHome.RoomItems[p.selItem])
			total := placed.Item.Storage.TotalSlots()
			drawText(dst, fmt.Sprintf("Storage: %d/%d slots", used, total),
				float64(rpListX), float64(panelY)+44, fontS, colorText)
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

		actY := float64(panelY) + 180
		drawText(dst, "Actions:", float64(rpListX), actY, fontS, colorMuted)
		actY += 20
		for i, act := range placed.Item.Actions {
			ax, ay, aw, ah := rpActionRowRect(i)
			ay = float32(actY) + float32(i)*42
			hov := isHovered(mx, my, ax, ay, aw, ah)
			drawButton(dst, act, ax, ay, aw, ah, fontM, hov, true)
		}

		bx, by, bw, bh := rpBackBtnRect()
		drawButton(dst, "← Back", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)
		mbx, mby, mbw, mbh := rpMoveBtnRect()
		drawButton(dst, "✦ Move", mbx, mby, mbw, mbh, fontS, isHovered(mx, my, mbx, mby, mbw, mbh), true)

	case rpModeMoveGrid:
		drawText(dst, "Step 1: Click cell to move item", float64(rpListX), float64(panelY)+8, fontS, colorMuted)
		if p.selItem >= 0 && p.selItem < len(char.CurrentHome.RoomItems) {
			placed := char.CurrentHome.RoomItems[p.selItem]
			drawText(dst, placed.Item.Name, float64(rpListX), float64(panelY)+28, fontM, colorAccent)
		}
		if p.wizard != nil && p.wizard.err != "" {
			drawTextWrapped(dst, p.wizard.err, float64(rpListX), float64(panelY)+54, float64(rpListW), 18, fontS, colorRed)
		}
		bx, by, bw, bh := rpBackBtnRect()
		drawButton(dst, "← Cancel", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)

	case rpModeMoveZ, rpModeMoveDir:
		if p.wizard != nil {
			p.wizard.drawLeftPanel(dst, mx, my)
		}
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

	// Which cells to highlight based on mode
	highlightCells := make(map[[2]int]color.RGBA)

	switch p.mode {
	case rpModeList:
		// Highlight cells belonging to hovered items in the list
		mx, my := ebiten.CursorPosition()
		hoverListIdx := -1
		for i := range home.RoomItems {
			rx, ry, rw, rh := rpItemRowRect(i)
			ry += 20
			if isHovered(mx, my, rx, ry, rw, rh) {
				hoverListIdx = i
				break
			}
		}
		// Highlight cells from grid-hover
		if p.hoverCell[0] >= 0 {
			for _, idx := range p.itemsAtCell(p.hoverCell[0], p.hoverCell[1]) {
				for _, cell := range occupiedCells(home.RoomItems[idx].X, home.RoomItems[idx].Y, home.RoomItems[idx].Item, home.RoomItems[idx].Direction) {
					highlightCells[[2]int{int(cell[0]), int(cell[1])}] = color.RGBA{80, 180, 80, 200}
				}
			}
		}
		// Also highlight cells for list-hover
		if hoverListIdx >= 0 {
			pl := home.RoomItems[hoverListIdx]
			for _, cell := range occupiedCells(pl.X, pl.Y, pl.Item, pl.Direction) {
				highlightCells[[2]int{int(cell[0]), int(cell[1])}] = color.RGBA{80, 180, 80, 200}
			}
		}

	case rpModeFiltered:
		// Highlight the selected cell
		stacked := p.itemsAtCell(p.selCell[0], p.selCell[1])
		for _, idx := range stacked {
			pl := home.RoomItems[idx]
			for _, cell := range occupiedCells(pl.X, pl.Y, pl.Item, pl.Direction) {
				highlightCells[[2]int{int(cell[0]), int(cell[1])}] = color.RGBA{80, 180, 80, 200}
			}
		}

	case rpModeAction:
		if p.selItem >= 0 && p.selItem < len(home.RoomItems) {
			pl := home.RoomItems[p.selItem]
			for _, cell := range occupiedCells(pl.X, pl.Y, pl.Item, pl.Direction) {
				highlightCells[[2]int{int(cell[0]), int(cell[1])}] = colorAccent
			}
		}

	}

	// Determine if we are in a move mode and which item is being moved
	inMoveMode := p.mode == rpModeMoveGrid || p.mode == rpModeMoveZ || p.mode == rpModeMoveDir

	// Draw items — dim the item being moved at its old position
	for i, placed := range home.RoomItems {
		isDimmed := inMoveMode && i == p.selItem
		for _, cell := range occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction) {
			cx := ox + float32(cell[0])*rpCellSz
			cy := oy + float32(cell[1])*rpCellSz
			col := colorItem
			if hc, ok := highlightCells[[2]int{int(cell[0]), int(cell[1])}]; ok {
				col = hc
			}
			if isDimmed {
				col = color.RGBA{col.R / 3, col.G / 3, col.B / 3, 180}
			}
			fillRect(dst, cx+1, cy+1, rpCellSz-3, rpCellSz-3, col)
		}
		lx := ox + float32(placed.X)*rpCellSz + 4
		ly := oy + float32(placed.Y)*rpCellSz + 4
		label := string([]rune(placed.Item.Name)[0:1])
		if isDimmed {
			drawText(dst, label, float64(lx), float64(ly), fontS, colorMuted)
		} else {
			drawText(dst, label, float64(lx), float64(ly), fontS, colorBg)
			drawFacingArrow(dst, ox, oy, int(placed.X), int(placed.Y), placed.Direction)
		}
	}

	// Draw move ghost overlay on top of everything
	if inMoveMode && p.wizard != nil {
		p.wizard.drawGhost(dst, ox, oy)
	}

	// Floor food
	for _, ff := range home.FloorFood {
		cx := ox + float32(ff.X)*rpCellSz
		cy := oy + float32(ff.Y)*rpCellSz
		fillRect(dst, cx+8, cy+8, rpCellSz-17, rpCellSz-17, colorFoodFloor)
	}

	// Hover cursor outline in list/moveGrid mode
	if p.mode == rpModeList || p.mode == rpModeMoveGrid {
		mx, my := ebiten.CursorPosition()
		hx, hy := p.gridCellAt(mx, my)
		if hx >= 0 {
			cx := ox + float32(hx)*rpCellSz
			cy := oy + float32(hy)*rpCellSz
			strokeRect(dst, cx, cy, rpCellSz-1, rpCellSz-1, colorAccent)
		}
	}

	// Selected cell outline in filtered/action mode
	if (p.mode == rpModeFiltered || p.mode == rpModeAction) && p.selCell[0] >= 0 {
		cx := ox + float32(p.selCell[0])*rpCellSz
		cy := oy + float32(p.selCell[1])*rpCellSz
		strokeRect(dst, cx, cy, rpCellSz-1, rpCellSz-1, colorYellow)
	}

	drawText(dst, fmt.Sprintf("Room: %s", home.Type.Name), float64(ox), float64(oy)-18, fontS, colorMuted)
}
