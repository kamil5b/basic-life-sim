package core

import (
	"fmt"
	"image/color"

	"github.com/kamil5b/basic-life-sim/internal/model"

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
	selItem      int // index into RoomItems, -1 = none
	selFloorFood int // index into FloorFood, -1 = none
	selAct       int

	mode   rpMode
	wizard *placementWizard

	// fridge organizer state
	fridgeWiz *fridgeWizard

	// floor food move state
	floorMoveMode           bool // true = waiting for user to click target cell
	floorFoodFridgeTarget   int  // RoomItems index of fridge being targeted (-1 = none)
	floorFoodFridgeTargetXY [2]int
}

type rpMode int

const (
	rpModeList                  rpMode = iota // full list, hover highlights on grid
	rpModeFiltered                            // cell clicked — show only items stacked there
	rpModeAction                              // action picker for a selected item
	rpModeMoveGrid                            // picking new X,Y for an item
	rpModeMoveZ                               // picking new Z
	rpModeMoveDir                             // picking new direction
	rpModeFridgeGrid                          // organising food inside a fridge (unused — handled by fridgeWiz)
	rpModeFloorFoodAction                     // action picker for a selected floor food item
	rpModeFloorFoodFridgeChoice               // "In World" vs "In Fridge" when dropping food onto a fridge cell
)

func newRoomPanel(char *model.Character, main *mainScreen) *roomPanel {
	return &roomPanel{
		char:                  char,
		main:                  main,
		selItem:               -1,
		selFloorFood:          -1,
		hoverCell:             [2]int{-1, -1},
		selCell:               [2]int{-1, -1},
		floorFoodFridgeTarget: -1,
	}
}

// fridgeAtCell returns the RoomItems index of a storage item whose footprint covers (gx,gy), or -1.
func (p *roomPanel) fridgeAtCell(gx, gy int) int {
	for i, placed := range p.char.CurrentHome.RoomItems {
		if placed.Item.Storage == nil {
			continue
		}
		for _, cell := range occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction) {
			if int(cell[0]) == gx && int(cell[1]) == gy {
				return i
			}
		}
	}
	return -1
}

// floorFoodAtCell returns the indices (into FloorFood) of food items at grid cell (gx,gy).
func (p *roomPanel) floorFoodAtCell(gx, gy int) []int {
	var out []int
	for i, ff := range p.char.CurrentHome.FloorFood {
		if int(ff.X) == gx && int(ff.Y) == gy {
			out = append(out, i)
		}
	}
	return out
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
	if p.fridgeWiz != nil {
		if p.fridgeWiz.update() {
			p.fridgeWiz = nil
		}
		return
	}

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
				floorFood := p.floorFoodAtCell(gx, gy)
				if len(stacked) > 0 || len(floorFood) > 0 {
					p.selCell = [2]int{gx, gy}
					p.selItem = -1
					p.selFloorFood = -1
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
			p.selFloorFood = -1
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
		// Floor food rows appear below room-item rows
		floorFood := p.floorFoodAtCell(p.selCell[0], p.selCell[1])
		baseRow := len(stacked)
		for li, ffIdx := range floorFood {
			rx, ry, rw, rh := rpItemRowRect(baseRow + li)
			ry += 20
			if clicked && isHovered(mx, my, rx, ry, rw, rh) {
				p.selFloorFood = ffIdx
				p.floorMoveMode = false
				p.mode = rpModeFloorFoodAction
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

		// sell button
		sbx, sby, sbw, sbh := rpSellBtnRect()
		if clicked && isHovered(mx, my, sbx, sby, sbw, sbh) {
			p.executeSell(placed)
			return
		}
		// trash button
		tbx, tby, tbw, tbh := rpTrashBtnRect()
		if clicked && isHovered(mx, my, tbx, tby, tbw, tbh) {
			p.executeTrash(placed)
			return
		}
		// organize fridge button (only for storage items)
		if placed.Item.Storage != nil {
			obx, oby, obw, obh := rpOrganizeBtnRect()
			if clicked && isHovered(mx, my, obx, oby, obw, obh) {
				p.fridgeWiz = newFridgeWizard(
					p.char, p.selItem,
					rpGridX, rpGridY+20, rpListX,
					nil, false,
					p.main.setMessage,
					func() { p.fridgeWiz = nil },
					func() { p.fridgeWiz = nil },
				)
				return
			}
		}

		for i, act := range actions {
			ax, ay, aw, ah := rpActionRowRect(i)
			if clicked && isHovered(mx, my, ax, ay, aw, ah) {
				p.selAct = i
				p.executeAction(placed, act)
			}
		}

	case rpModeFloorFoodAction:
		if p.selFloorFood < 0 || p.selFloorFood >= len(p.char.CurrentHome.FloorFood) {
			p.mode = rpModeFiltered
			return
		}
		ff := &p.char.CurrentHome.FloorFood[p.selFloorFood]

		// Back
		bx, by, bw, bh := rpBackBtnRect()
		if clicked && isHovered(mx, my, bx, by, bw, bh) {
			p.floorMoveMode = false
			p.mode = rpModeFiltered
			return
		}

		// Eat
		eax, eay, eaw, eah := rpFloorFoodEatBtnRect()
		if clicked && isHovered(mx, my, eax, eay, eaw, eah) && !p.floorMoveMode {
			char := p.char
			if model.IsExpired(ff.PurchaseDate, ff.Food.BaseExpiryDays, 1, char.CurrentDate) {
				foodExpiredPenalty(char)
				p.main.setMessage(fmt.Sprintf("%s was EXPIRED — all stats -10!", ff.Food.Name))
			} else if isInedible(ff.Food) {
				rawFoodPenalty(char)
				p.main.setMessage(fmt.Sprintf("Eating raw %s penalised stats -5.", ff.Food.Name))
			} else {
				applyNutrition(char, ff.Food)
				p.main.setMessage(fmt.Sprintf("Ate %s.", ff.Food.Name))
			}
			consumeUse(char, "floor", 0, p.selFloorFood)
			p.selFloorFood = -1
			p.mode = rpModeFiltered
			return
		}

		// Move: toggle waiting-for-click mode
		mmx, mmy, mmw, mmh := rpFloorFoodMoveBtnRect()
		if clicked && isHovered(mx, my, mmx, mmy, mmw, mmh) {
			p.floorMoveMode = !p.floorMoveMode
			if p.floorMoveMode {
				p.main.setMessage("Click a floor cell to move the food there.")
			}
			return
		}

		// Trash
		tx, ty, tw, th := rpFloorFoodTrashBtnRect()
		if clicked && isHovered(mx, my, tx, ty, tw, th) && !p.floorMoveMode {
			name := ff.Food.Name
			p.char.CurrentHome.FloorFood = append(
				p.char.CurrentHome.FloorFood[:p.selFloorFood],
				p.char.CurrentHome.FloorFood[p.selFloorFood+1:]...)
			p.main.setMessage(fmt.Sprintf("Trashed %s.", name))
			p.selFloorFood = -1
			p.mode = rpModeFiltered
			return
		}

		// Grid click while in move mode
		if p.floorMoveMode && clicked {
			gx, gy := p.gridCellAt(mx, my)
			if gx >= 0 {
				fridgeIdx := p.fridgeAtCell(gx, gy)
				if fridgeIdx >= 0 {
					// Dropped onto a fridge — ask In World or In Fridge
					p.floorFoodFridgeTarget = fridgeIdx
					p.floorFoodFridgeTargetXY = [2]int{gx, gy}
					p.floorMoveMode = false
					p.mode = rpModeFloorFoodFridgeChoice
				} else {
					ff.X = uint8(gx)
					ff.Y = uint8(gy)
					p.selCell = [2]int{gx, gy}
					p.main.setMessage(fmt.Sprintf("Moved %s to (%d,%d).", ff.Food.Name, gx, gy))
					p.floorMoveMode = false
				}
			}
		}

	case rpModeFloorFoodFridgeChoice:
		if p.selFloorFood < 0 || p.floorFoodFridgeTarget < 0 {
			p.mode = rpModeFloorFoodAction
			return
		}
		ff := &p.char.CurrentHome.FloorFood[p.selFloorFood]
		fridgePlaced := p.char.CurrentHome.RoomItems[p.floorFoodFridgeTarget]

		// Back
		bx, by, bw, bh := rpBackBtnRect()
		if clicked && isHovered(mx, my, bx, by, bw, bh) {
			p.floorFoodFridgeTarget = -1
			p.mode = rpModeFloorFoodAction
			return
		}

		// In World: place on top of the fridge
		iwx, iwy, iww, iwh := rpChoiceInWorldBtnRect()
		if clicked && isHovered(mx, my, iwx, iwy, iww, iwh) {
			topZ := fridgePlaced.Z + fridgePlaced.Item.Height
			ff.X = uint8(p.floorFoodFridgeTargetXY[0])
			ff.Y = uint8(p.floorFoodFridgeTargetXY[1])
			ff.Z = topZ
			p.selCell = p.floorFoodFridgeTargetXY
			p.main.setMessage(fmt.Sprintf("Placed %s on top of %s (z=%d).", ff.Food.Name, fridgePlaced.Item.Name, topZ))
			p.floorFoodFridgeTarget = -1
			p.selFloorFood = -1
			p.mode = rpModeFiltered
			return
		}

		// In Fridge: open fridge wizard
		ifx, ify, ifw, ifh := rpChoiceInFridgeBtnRect()
		if clicked && isHovered(mx, my, ifx, ify, ifw, ifh) {
			storage := fridgePlaced.Item.Storage
			if storage.TotalSlots() <= usedSlotCount(&p.char.CurrentHome.RoomItems[p.floorFoodFridgeTarget]) {
				p.main.setMessage("Fridge is full!")
				return
			}
			ffCopy := ff.Food
			ffIdx := p.selFloorFood
			fridgeIdx := p.floorFoodFridgeTarget
			p.fridgeWiz = newFridgeWizard(
				p.char, fridgeIdx,
				rpGridX, rpGridY+20, rpListX,
				&ffCopy, false,
				p.main.setMessage,
				func() {
					// Cancelled — stay in choice
					p.fridgeWiz = nil
				},
				func() {
					// Placed in fridge — remove from floor
					p.char.CurrentHome.FloorFood = append(
						p.char.CurrentHome.FloorFood[:ffIdx],
						p.char.CurrentHome.FloorFood[ffIdx+1:]...)
					p.fridgeWiz = nil
					p.floorFoodFridgeTarget = -1
					p.selFloorFood = -1
					p.mode = rpModeFiltered
				},
			)
			p.floorFoodFridgeTarget = -1
			return
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
	item := home.RoomItems[p.selItem].Item

	// Collect stack: items riding on top of the moved item (same XY footprint overlap, Z immediately above), recursively.
	stack := p.collectStack(p.selItem)

	// Temporarily remove the whole stack for a clean canPlace check.
	// Work highest index first so removals don't shift lower indices.
	saved := make([]model.PlacedRoomItem, len(stack))
	for i, idx := range stack {
		saved[i] = home.RoomItems[idx]
	}
	p.removeIndices(stack)

	finalZ := z
	if !item.CanOverhang {
		finalZ = 0
	}

	if err := canPlace(*home, item, x, y, finalZ, dir); err != nil {
		// Restore everything back
		home.RoomItems = append(home.RoomItems, saved...)
		return err
	}

	// Place the moved item.
	base := saved[0]
	oldX, oldY, oldZ := base.X, base.Y, base.Z
	deltaZ := int(finalZ) - int(oldZ)
	base.X = x
	base.Y = y
	base.Z = finalZ
	base.Direction = dir
	home.RoomItems = append(home.RoomItems, base)

	// Re-place stacked items at the new position, shifted by the same delta.
	for _, s := range saved[1:] {
		s.X = x + (s.X - oldX)
		s.Y = y + (s.Y - oldY)
		newZ := int(s.Z) + deltaZ
		if newZ < 0 {
			newZ = 0
		}
		s.Z = uint8(newZ)
		home.RoomItems = append(home.RoomItems, s)
	}

	// Update selItem to point to the newly appended base item.
	p.selItem = len(home.RoomItems) - len(saved)

	p.main.setMessage(fmt.Sprintf("Moved %s (+%d stacked) to (%d,%d,z=%d) facing %s",
		item.Name, len(saved)-1, x, y, finalZ, dirName(dir)))
	p.selCell = [2]int{int(x), int(y)}
	p.wizard = nil
	p.mode = rpModeFiltered
	return nil
}

// collectStack returns the indices of the moved item (index 0) followed by all
// items recursively stacked on top of it, in bottom-up order.
func (p *roomPanel) collectStack(baseIdx int) []int {
	home := &p.char.CurrentHome
	result := []int{baseIdx}
	visited := map[int]bool{baseIdx: true}
	queue := []int{baseIdx}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		cpl := home.RoomItems[current]
		cTopZ := cpl.Z + cpl.Item.Height
		cCells := occupiedCells(cpl.X, cpl.Y, cpl.Item, cpl.Direction)
		for i, other := range home.RoomItems {
			if visited[i] {
				continue
			}
			// other sits directly on top if its Z == cTopZ and footprints overlap.
			if other.Z != cTopZ {
				continue
			}
			oCells := occupiedCells(other.X, other.Y, other.Item, other.Direction)
			if footprintsOverlap(cCells, oCells) {
				visited[i] = true
				result = append(result, i)
				queue = append(queue, i)
			}
		}
	}
	return result
}

// footprintsOverlap returns true if two cell lists share at least one cell.
func footprintsOverlap(a, b [][2]uint8) bool {
	set := make(map[[2]uint8]bool, len(a))
	for _, c := range a {
		set[c] = true
	}
	for _, c := range b {
		if set[c] {
			return true
		}
	}
	return false
}

// removeIndices removes items at the given indices (must be sorted descending or we sort here).
func (p *roomPanel) removeIndices(indices []int) {
	home := &p.char.CurrentHome
	// Sort descending so removals don't shift subsequent indices.
	for i := len(indices) - 1; i >= 0; i-- {
		for j := i - 1; j >= 0; j-- {
			if indices[j] < indices[i] {
				indices[i], indices[j] = indices[j], indices[i]
			}
		}
	}
	for _, idx := range indices {
		home.RoomItems = append(home.RoomItems[:idx], home.RoomItems[idx+1:]...)
	}
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
				inedible := s.Food.Nutrition.Food == 0 && s.Food.Nutrition.Energy == 0 &&
					s.Food.Nutrition.Hygiene == 0 && s.Food.Nutrition.Confidence == 0 &&
					s.Food.Nutrition.Strength == 0 && s.Food.OnEat == nil
				if inedible {
					rawFoodPenalty(char)
					p.main.setMessage(fmt.Sprintf("Eating raw %s was a bad idea.", s.Food.Name))
				} else {
					applyNutrition(char, s.Food)
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
				if sf.Cooked {
					continue
				}
				result := model.ResolveCookedResult(sf.Food)
				if result == nil {
					continue
				}
				result.UsesTotal = sf.Food.UsesTotal
				sf.Food = *result
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

func (p *roomPanel) executeSell(placed *model.PlacedRoomItem) {
	if p.selItem < 0 || p.selItem >= len(p.char.CurrentHome.RoomItems) {
		return
	}
	refund := placed.Item.BasePrice * 0.5
	p.char.CurrentStats.Money += refund
	p.char.CurrentHome.RoomItems = append(
		p.char.CurrentHome.RoomItems[:p.selItem],
		p.char.CurrentHome.RoomItems[p.selItem+1:]...,
	)
	p.main.setMessage(fmt.Sprintf("Sold %s for $%.2f (50%% refund). Money: $%.2f",
		placed.Item.Name, refund, p.char.CurrentStats.Money))
	p.selItem = -1
	p.selCell = [2]int{-1, -1}
	p.mode = rpModeList
}

func (p *roomPanel) executeTrash(placed *model.PlacedRoomItem) {
	if p.selItem < 0 || p.selItem >= len(p.char.CurrentHome.RoomItems) {
		return
	}
	name := placed.Item.Name
	p.char.CurrentHome.RoomItems = append(
		p.char.CurrentHome.RoomItems[:p.selItem],
		p.char.CurrentHome.RoomItems[p.selItem+1:]...,
	)
	p.main.setMessage(fmt.Sprintf("Trashed %s.", name))
	p.selItem = -1
	p.selCell = [2]int{-1, -1}
	p.mode = rpModeList
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

func rpSellBtnRect() (x, y, w, h float32) {
	return rpListX + 236, panelY + float32(ScreenH) - float32(tabH) - 50, 80, 32
}

func rpTrashBtnRect() (x, y, w, h float32) {
	return rpListX + 324, panelY + float32(ScreenH) - float32(tabH) - 50, 80, 32
}

func rpOrganizeBtnRect() (x, y, w, h float32) {
	return rpListX + 412, panelY + float32(ScreenH) - float32(tabH) - 50, 100, 32
}

func rpFloorFoodEatBtnRect() (x, y, w, h float32) {
	return rpListX + 108, panelY + float32(ScreenH) - float32(tabH) - 50, 80, 32
}

func rpFloorFoodMoveBtnRect() (x, y, w, h float32) {
	return rpListX + 196, panelY + float32(ScreenH) - float32(tabH) - 50, 90, 32
}

func rpFloorFoodTrashBtnRect() (x, y, w, h float32) {
	return rpListX + 294, panelY + float32(ScreenH) - float32(tabH) - 50, 80, 32
}

func rpChoiceInWorldBtnRect() (x, y, w, h float32) {
	return rpListX, panelY + 120, 200, 48
}

func rpChoiceInFridgeBtnRect() (x, y, w, h float32) {
	return rpListX, panelY + 180, 200, 48
}

// ── draw ──────────────────────────────────────────────────────────────────────

func (p *roomPanel) draw(dst *ebiten.Image) {
	if p.fridgeWiz != nil {
		p.fridgeWiz.draw(dst)
		return
	}

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
		floorFood := p.floorFoodAtCell(p.selCell[0], p.selCell[1])
		total := len(stacked) + len(floorFood)
		drawText(dst, fmt.Sprintf("Cell (%d,%d) — %d item(s)", p.selCell[0], p.selCell[1], total),
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
		// Floor food rows
		baseRow := len(stacked)
		for li, ffIdx := range floorFood {
			ff := char.CurrentHome.FloorFood[ffIdx]
			rx, ry, rw, rh := rpItemRowRect(baseRow + li)
			ry += 20
			hov := isHovered(mx, my, rx, ry, rw, rh)
			bg := colorPanel
			if hov {
				bg = colorHighlight
			}
			fillRect(dst, rx, ry, rw, rh, bg)
			strokeRect(dst, rx, ry, rw, rh, colorBorder)
			exp := model.IsExpired(ff.PurchaseDate, ff.Food.BaseExpiryDays, 1, char.CurrentDate)
			tc := colorFoodFloor
			if exp {
				tc = colorRed
			}
			label := fmt.Sprintf("🍞 %s  uses:%d  z=%d", ff.Food.Name, ff.UsesRemaining, ff.Z)
			if exp {
				label += " [EXPIRED]"
			}
			drawText(dst, label, float64(rx)+6, float64(ry)+6, fontS, tc)
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
		sbx, sby, sbw, sbh := rpSellBtnRect()
		drawButton(dst, "Sell", sbx, sby, sbw, sbh, fontS, isHovered(mx, my, sbx, sby, sbw, sbh), true)
		tbx, tby, tbw, tbh := rpTrashBtnRect()
		drawButton(dst, "Trash", tbx, tby, tbw, tbh, fontS, isHovered(mx, my, tbx, tby, tbw, tbh), true)
		if placed.Item.Storage != nil {
			obx, oby, obw, obh := rpOrganizeBtnRect()
			drawButton(dst, "📦 Organize", obx, oby, obw, obh, fontS, isHovered(mx, my, obx, oby, obw, obh), true)
		}

	case rpModeFloorFoodAction:
		if p.selFloorFood < 0 || p.selFloorFood >= len(char.CurrentHome.FloorFood) {
			return
		}
		ff := char.CurrentHome.FloorFood[p.selFloorFood]
		exp := model.IsExpired(ff.PurchaseDate, ff.Food.BaseExpiryDays, 1, char.CurrentDate)
		expiry := model.ExpiryDate(ff.PurchaseDate, ff.Food.BaseExpiryDays, 1)
		ey, em, ed := expiry.Unpack()
		headCol := colorFoodFloor
		if exp {
			headCol = colorRed
		}
		drawText(dst, ff.Food.Name, float64(rpListX), float64(panelY)+4, fontM, headCol)
		drawText(dst, fmt.Sprintf("at (%d,%d,z=%d)  uses:%d", ff.X, ff.Y, ff.Z, ff.UsesRemaining),
			float64(rpListX), float64(panelY)+26, fontS, colorMuted)
		expTag := fmt.Sprintf("Expires: %04d-%02d-%02d", ey, em, ed)
		if exp {
			expTag = fmt.Sprintf("EXPIRED %04d-%02d-%02d", ey, em, ed)
		}
		drawText(dst, expTag, float64(rpListX), float64(panelY)+44, fontS, headCol)
		if isInedible(ff.Food) {
			drawText(dst, "[raw / inedible — eating will penalise stats]", float64(rpListX), float64(panelY)+62, fontS, colorYellow)
		}
		if p.floorMoveMode {
			drawText(dst, "→ Click a floor cell to place food there", float64(rpListX), float64(panelY)+82, fontS, colorYellow)
		}

		bx, by, bw, bh := rpBackBtnRect()
		drawButton(dst, "← Back", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)
		eax, eay, eaw, eah := rpFloorFoodEatBtnRect()
		drawButton(dst, "Eat", eax, eay, eaw, eah, fontS, isHovered(mx, my, eax, eay, eaw, eah) && !p.floorMoveMode, !p.floorMoveMode)
		mmx, mmy, mmw, mmh := rpFloorFoodMoveBtnRect()
		moveLabel := "✦ Move"
		if p.floorMoveMode {
			moveLabel = "✦ Cancel"
		}
		drawButton(dst, moveLabel, mmx, mmy, mmw, mmh, fontS, isHovered(mx, my, mmx, mmy, mmw, mmh), true)
		tx, ty, tw, th := rpFloorFoodTrashBtnRect()
		drawButton(dst, "Trash", tx, ty, tw, th, fontS, isHovered(mx, my, tx, ty, tw, th) && !p.floorMoveMode, !p.floorMoveMode)

	case rpModeFloorFoodFridgeChoice:
		if p.selFloorFood < 0 || p.floorFoodFridgeTarget < 0 ||
			p.selFloorFood >= len(char.CurrentHome.FloorFood) ||
			p.floorFoodFridgeTarget >= len(char.CurrentHome.RoomItems) {
			return
		}
		ff2 := char.CurrentHome.FloorFood[p.selFloorFood]
		fridgePlaced := char.CurrentHome.RoomItems[p.floorFoodFridgeTarget]
		topZ := fridgePlaced.Z + fridgePlaced.Item.Height

		drawText(dst, "Where to put "+ff2.Food.Name+"?", float64(rpListX), float64(panelY)+12, fontM, colorAccent)
		drawText(dst, fmt.Sprintf("Dropping onto: %s at (%d,%d)",
			fridgePlaced.Item.Name, p.floorFoodFridgeTargetXY[0], p.floorFoodFridgeTargetXY[1]),
			float64(rpListX), float64(panelY)+36, fontS, colorMuted)
		drawText(dst, fmt.Sprintf("In World: places on top of fridge at z=%d", topZ),
			float64(rpListX), float64(panelY)+100, fontS, colorMuted)
		drawText(dst, fmt.Sprintf("In Fridge: store inside (%d/%d slots used)",
			usedSlotCount(&char.CurrentHome.RoomItems[p.floorFoodFridgeTarget]),
			fridgePlaced.Item.Storage.TotalSlots()),
			float64(rpListX), float64(panelY)+162, fontS, colorMuted)

		iwx, iwy, iww, iwh := rpChoiceInWorldBtnRect()
		drawButton(dst, "📦 In World (on top)", iwx, iwy, iww, iwh, fontM, isHovered(mx, my, iwx, iwy, iww, iwh), true)
		ifx, ify, ifw, ifh := rpChoiceInFridgeBtnRect()
		fridgeFull := fridgePlaced.Item.Storage.TotalSlots() <= usedSlotCount(&char.CurrentHome.RoomItems[p.floorFoodFridgeTarget])
		drawButton(dst, "❄ In Fridge", ifx, ify, ifw, ifh, fontM, isHovered(mx, my, ifx, ify, ifw, ifh) && !fridgeFull, !fridgeFull)
		if fridgeFull {
			drawText(dst, "(fridge is full)", float64(ifx)+float64(ifw)+8, float64(ify)+14, fontS, colorRed)
		}

		bx, by, bw, bh := rpBackBtnRect()
		drawButton(dst, "← Back", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)

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
