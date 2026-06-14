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
	selFloorItem int // index into FloorItems, -1 = none
	selAct       int

	mode   rpMode
	wizard *placementWizard

	// storage organizer state
	storageWiz *storageWizard

	// floor food move state
	floorMoveMode           bool // true = waiting for user to click target cell
	floorFoodFridgeTarget   int  // RoomItems index of fridge being targeted (-1 = none)
	floorFoodFridgeTargetXY [2]int
	utilTarget              int // FloorItems index of utility being targeted for interaction

	// door popup
	doorPopupOpen bool
	doorCell      [2]int
}

type rpMode int

const (
	rpModeList                  rpMode = iota // full list, hover highlights on grid
	rpModeFiltered                            // cell clicked — show only items stacked there
	rpModeAction                              // action picker for a selected item
	rpModeMoveGrid                            // picking new X,Y for an item
	rpModeMoveZ                               // picking new Z
	rpModeMoveDir                             // picking new direction
	rpModeFridgeGrid                          // organising food inside a storage container (unused — handled by storageWiz)
	rpModeFloorItemAction                     // action picker for a selected floor item (food or utility)
	rpModeFloorFoodFridgeChoice               // "In World" vs "In Fridge" when dropping food onto a fridge cell
	rpModeUtilityChoice                       // use utility on food vs place separately
	rpModeTimeSkip                            // choose how many hours to skip for an activity
	rpModeDoorPopup                           // door popup (Go to Shop)
)

func newRoomPanel(char *model.Character, main *mainScreen) *roomPanel {
	return &roomPanel{
		char:                  char,
		main:                  main,
		selItem:               -1,
		selFloorItem:          -1,
		hoverCell:             [2]int{-1, -1},
		selCell:               [2]int{-1, -1},
		floorFoodFridgeTarget: -1,
		doorCell:              [2]int{-1, -1},
	}
}

// doorTypeAt returns the door type at grid cell (gx, gy), or -1 if not a door.
func (p *roomPanel) doorTypeAt(gx, gy int) int {
	for _, d := range p.char.CurrentHome.Type.Doors {
		if int(d.X) == gx && int(d.Y) == gy {
			return int(d.Type)
		}
	}
	return -1
}

// isDoorCell checks if the grid cell is a door in the layout.
func (p *roomPanel) isDoorCell(gx, gy int) bool {
	lt := p.char.CurrentHome.Type.Layout
	if gy < 0 || gy >= len(lt) || gx < 0 || gx >= len(lt[0]) {
		return false
	}
	return lt[gy][gx] == model.HomeCellDoor
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

// floorItemsAtCell returns the indices (into FloorItems) of items at grid cell (gx,gy).
func (p *roomPanel) floorItemsAtCell(gx, gy int) []int {
	var out []int
	for i, fi := range p.char.CurrentHome.FloorItems {
		if int(fi.X) == gx && int(fi.Y) == gy {
			out = append(out, i)
		}
	}
	return out
}

// floorFoodAtCell returns the indices (into FloorItems) of food at grid cell (gx,gy).
func (p *roomPanel) floorFoodAtCell(gx, gy int) []int {
	var out []int
	for _, idx := range p.floorItemsAtCell(gx, gy) {
		if p.char.CurrentHome.FloorItems[idx].Item.Kind == model.FloorKindFood {
			out = append(out, idx)
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
	ox := int(p.main.panelX() + 274)
	oy := int(p.main.panelY()+4) + 20
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
	if p.storageWiz != nil {
		if p.storageWiz.update() {
			p.storageWiz = nil
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
				// Door click: check door type
				if p.isDoorCell(gx, gy) {
					dt := p.doorTypeAt(gx, gy)
					if dt == int(model.DoorRoomExit) {
						p.doorPopupOpen = true
						p.doorCell = [2]int{gx, gy}
						p.mode = rpModeDoorPopup
						return
					}
					// DoorInternal: reserved, no action
					return
				}
				stacked := p.itemsAtCell(gx, gy)
				floorItems := p.floorItemsAtCell(gx, gy)
				if len(stacked) > 0 || len(floorItems) > 0 {
					p.selCell = [2]int{gx, gy}
					p.selItem = -1
					p.selFloorItem = -1
					p.mode = rpModeFiltered
					return
				}
			}
		}

	case rpModeDoorPopup:
		// Click anywhere outside door popup closes it
		if clicked {
			// "Go to Shop" button
			ox := p.main.panelX() + 274
			oy := p.main.panelY() + 4 + 20
			cx := ox + float32(p.doorCell[0])*rpCellSz
			cy := oy + float32(p.doorCell[1])*rpCellSz
			bx, by, bw, bh := cx-40, cy+rpCellSz+4, float32(140), float32(32)
			if isHovered(mx, my, bx, by, bw, bh) {
				p.doorPopupOpen = false
				p.main.mode = modeShop
				p.mode = rpModeList
				return
			}
			// Click anywhere else closes popup
			p.doorPopupOpen = false
			p.mode = rpModeList
		}

	case rpModeFiltered:
		// Back if clicking outside the list area or on back button
		bx, by, bw, bh := p.main.panelX()+4, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
		if clicked && isHovered(mx, my, bx, by, bw, bh) {
			p.mode = rpModeList
			p.selCell = [2]int{-1, -1}
			p.selItem = -1
			p.selFloorItem = -1
			return
		}

		stacked := p.itemsAtCell(p.selCell[0], p.selCell[1])
		for li, itemIdx := range stacked {
			rx, ry, rw, rh := p.main.panelX()+4, p.main.panelY()+4+float32(li)*rpRowH, float32(260), rpRowH-2
			ry += 20
			if clicked && isHovered(mx, my, rx, ry, rw, rh) {
				p.selItem = itemIdx
				p.selAct = 0
				p.mode = rpModeAction
				return
			}
		}
		// Floor item rows appear below room-item rows
		floorItems := p.floorItemsAtCell(p.selCell[0], p.selCell[1])
		baseRow := len(stacked)
		for li, fiIdx := range floorItems {
			rx, ry, rw, rh := p.main.panelX()+4, p.main.panelY()+4+float32(baseRow+li)*rpRowH, float32(260), rpRowH-2
			ry += 20
			if clicked && isHovered(mx, my, rx, ry, rw, rh) {
				p.selFloorItem = fiIdx
				p.floorMoveMode = false
				p.mode = rpModeFloorItemAction
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
		bx, by, bw, bh := p.main.panelX()+4, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
		if clicked && isHovered(mx, my, bx, by, bw, bh) {
			p.mode = rpModeFiltered
			return
		}

		// move button
		mbx, mby, mbw, mbh := p.main.panelX()+4+108, p.main.panelY()+4+rpRowH*20-50, float32(120), float32(32)
		if clicked && isHovered(mx, my, mbx, mby, mbw, mbh) {
			p.wizard = newPlacementWizard(
				p.char, &placed.Item, p.selItem,
				p.main.panelX()+274, p.main.panelY()+4+20,
				p.main.panelX()+4, p.main.panelY()+4, p.main.panelW(), p.main.panelH(),
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
		sbx, sby, sbw, sbh := p.main.panelX()+4+236, p.main.panelY()+4+rpRowH*20-50, float32(80), float32(32)
		if clicked && isHovered(mx, my, sbx, sby, sbw, sbh) {
			p.executeSell(placed)
			return
		}
		// trash button
		tbx, tby, tbw, tbh := p.main.panelX()+4+324, p.main.panelY()+4+rpRowH*20-50, float32(80), float32(32)
		if clicked && isHovered(mx, my, tbx, tby, tbw, tbh) {
			p.executeTrash(placed)
			return
		}
		// organize fridge button (only for storage items)
		if placed.Item.Storage != nil {
			obx, oby, obw, obh := p.main.panelX()+4+412, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
			if clicked && isHovered(mx, my, obx, oby, obw, obh) {
				p.storageWiz = newStorageWizard(
					p.char, p.selItem,
					p.main.panelX()+274, p.main.panelY()+4+20, p.main.panelX()+4,
					p.main.panelY()+4-4, p.main.panelH(),
					nil, nil, 0, false,
					p.main.setMessage,
					func() { p.storageWiz = nil },
					func() { p.storageWiz = nil },
				)
				return
			}
		}

		for i, act := range actions {
			ax, ay, aw, ah := p.main.panelX()+4, p.main.panelY()+4+float32(i)*42, float32(260), float32(36)
			if clicked && isHovered(mx, my, ax, ay, aw, ah) {
				p.selAct = i
				if canSkipTime(act) {
					p.mode = rpModeTimeSkip
					return
				}
				p.executeAction(placed, act)
			}
		}

	case rpModeFloorItemAction:
		if p.selFloorItem < 0 || p.selFloorItem >= len(p.char.CurrentHome.FloorItems) {
			p.mode = rpModeFiltered
			return
		}
		fi := &p.char.CurrentHome.FloorItems[p.selFloorItem]
		isFood := fi.Item.Kind == model.FloorKindFood

		// Back
		bx, by, bw, bh := p.main.panelX()+4, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
		if clicked && isHovered(mx, my, bx, by, bw, bh) {
			p.floorMoveMode = false
			p.mode = rpModeFiltered
			return
		}

		if isFood {
			// Eat
			eax, eay, eaw, eah := p.main.panelX()+4+108, p.main.panelY()+4+rpRowH*20-50, float32(80), float32(32)
			if clicked && isHovered(mx, my, eax, eay, eaw, eah) && !p.floorMoveMode {
				char := p.char
				if model.IsExpired(fi.Item.PurchaseDate, fi.Item.Food.BaseExpiryDays, 1, char.CurrentDate) {
					foodExpiredPenalty(char)
					p.main.setMessage(fmt.Sprintf("%s was EXPIRED — all stats -10!", fi.Item.Food.Name))
				} else if isInedible(fi.Item.Food) {
					rawFoodPenalty(char)
					p.main.setMessage(fmt.Sprintf("Eating raw %s penalised stats -5.", fi.Item.Food.Name))
				} else {
					applyNutrition(char, fi.Item.Food)
					p.main.setMessage(fmt.Sprintf("Ate %s.", fi.Item.Food.Name))
				}
				consumeUse(char, "floor", 0, p.selFloorItem)
				p.selFloorItem = -1
				p.mode = rpModeFiltered
				return
			}
		} else {
			// Use — check for food at same cell to process
			ex, ey, ew, eh := p.main.panelX()+4+108, p.main.panelY()+4+rpRowH*20-50, float32(80), float32(32)
			if clicked && isHovered(mx, my, ex, ey, ew, eh) && !p.floorMoveMode {
				processed := false
				for oi, o := range p.char.CurrentHome.FloorItems {
					if o.Item.Kind == model.FloorKindFood && o.X == fi.X && o.Y == fi.Y {
						tryUseUtilityOnFood(p.char, oi, p.selFloorItem)
						processed = true
						break
					}
				}
				if !processed {
					fi.Item.Utility.DoAction("use", &p.char.CurrentStats, p.char)
					p.main.setMessage(fmt.Sprintf("Used %s.", fi.Item.Utility.Name))
				}
			}
		}

		// Move: toggle waiting-for-click mode
		mmx, mmy, mmw, mmh := p.main.panelX()+4+196, p.main.panelY()+4+rpRowH*20-50, float32(90), float32(32)
		if clicked && isHovered(mx, my, mmx, mmy, mmw, mmh) {
			p.floorMoveMode = !p.floorMoveMode
			if p.floorMoveMode {
				p.main.setMessage("Click a floor cell to move the item there.")
			}
			return
		}

		// Trash
		tx, ty, tw, th := p.main.panelX()+4+294, p.main.panelY()+4+rpRowH*20-50, float32(80), float32(32)
		if clicked && isHovered(mx, my, tx, ty, tw, th) && !p.floorMoveMode {
			name := fi.Item.Food.Name
			if !isFood {
				name = fi.Item.Utility.Name
			}
			p.char.CurrentHome.FloorItems = append(
				p.char.CurrentHome.FloorItems[:p.selFloorItem],
				p.char.CurrentHome.FloorItems[p.selFloorItem+1:]...)
			p.main.setMessage(fmt.Sprintf("Trashed %s.", name))
			p.selFloorItem = -1
			p.mode = rpModeFiltered
			return
		}

		// Grid click while in move mode
		if p.floorMoveMode && clicked {
			gx, gy := p.gridCellAt(mx, my)
			if gx >= 0 {
				fi.X = uint8(gx)
				fi.Y = uint8(gy)
				p.selCell = [2]int{gx, gy}
				name := fi.Item.Food.Name
				if !isFood {
					name = fi.Item.Utility.Name
				}
				// If food moved onto a utility, offer choice
				if isFood {
					for oi, o := range p.char.CurrentHome.FloorItems {
						if o.Item.Kind == model.FloorKindUtility && int(o.X) == gx && int(o.Y) == gy && o.Item.Utility.Ability != "" {
							fi.X = uint8(gx)
							fi.Y = uint8(gy)
							p.utilTarget = oi
							p.floorMoveMode = false
							p.selCell = [2]int{gx, gy}
							p.mode = rpModeUtilityChoice
							return
						}
					}
				}
				p.main.setMessage(fmt.Sprintf("Moved %s to (%d,%d).", name, gx, gy))
			}
		}

	case rpModeUtilityChoice:
		if p.selFloorItem < 0 || p.utilTarget < 0 {
			p.mode = rpModeFiltered
			return
		}
		bx, by, bw, bh := p.main.panelX()+4, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
		if clicked && isHovered(mx, my, bx, by, bw, bh) {
			p.utilTarget = -1
			p.mode = rpModeFiltered
			return
		}
		// Use utility
		mx2, my2, mw, mh := p.main.panelX()+4+108, p.main.panelY()+4+rpRowH*20-50, float32(80), float32(32)
		if clicked && isHovered(mx, my, mx2, my2, mw, mh) {
			tryUseUtilityOnFood(p.char, p.selFloorItem, p.utilTarget)
			p.utilTarget = -1
			p.mode = rpModeFiltered
			return
		}
		// Place separately (just stay, already moved)
		mmx, mmy, mmw, mmh := p.main.panelX()+4+196, p.main.panelY()+4+rpRowH*20-50, float32(90), float32(32)
		if clicked && isHovered(mx, my, mmx, mmy, mmw, mmh) {
			p.utilTarget = -1
			p.mode = rpModeFiltered
			return
		}

	case rpModeFloorFoodFridgeChoice:
		if p.selFloorItem < 0 || p.floorFoodFridgeTarget < 0 ||
			p.selFloorItem >= len(p.char.CurrentHome.FloorItems) ||
			p.floorFoodFridgeTarget >= len(p.char.CurrentHome.RoomItems) {
			p.mode = rpModeFloorItemAction
			return
		}
		ff := &p.char.CurrentHome.FloorItems[p.selFloorItem]
		fridgePlaced := p.char.CurrentHome.RoomItems[p.floorFoodFridgeTarget]

		// Back
		bx, by, bw, bh := p.main.panelX()+4, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
		if clicked && isHovered(mx, my, bx, by, bw, bh) {
			p.floorFoodFridgeTarget = -1
			p.mode = rpModeFloorItemAction
			return
		}

		// In World: place on top of the fridge
		iwx, iwy, iww, iwh := p.main.panelX()+4, p.main.panelY()+4+120, float32(200), float32(48)
		if clicked && isHovered(mx, my, iwx, iwy, iww, iwh) {
			topZ := fridgePlaced.Z + fridgePlaced.Item.Height
			ff.X = uint8(p.floorFoodFridgeTargetXY[0])
			ff.Y = uint8(p.floorFoodFridgeTargetXY[1])
			ff.Z = topZ
			p.selCell = p.floorFoodFridgeTargetXY
			p.main.setMessage(fmt.Sprintf("Placed %s on top of %s (z=%d).", ff.Item.Food.Name, fridgePlaced.Item.Name, topZ))
			p.floorFoodFridgeTarget = -1
			p.selFloorItem = -1
			p.mode = rpModeFiltered
			return
		}

		// In Fridge: open fridge wizard
		ifx, ify, ifw, ifh := p.main.panelX()+4, p.main.panelY()+4+180, float32(200), float32(48)
		if clicked && isHovered(mx, my, ifx, ify, ifw, ifh) {
			storage := fridgePlaced.Item.Storage
			if storage.TotalSlots() <= usedSlotCount(&p.char.CurrentHome.RoomItems[p.floorFoodFridgeTarget]) {
				p.main.setMessage("Fridge is full!")
				return
			}
			ffCopy := ff.Item.Food
			ffIdx := p.selFloorItem
			fridgeIdx := p.floorFoodFridgeTarget
			p.storageWiz = newStorageWizard(
				p.char, fridgeIdx,
				p.main.panelX()+274, p.main.panelY()+4+20, p.main.panelX()+4,
				p.main.panelY()+4-4, p.main.panelH(),
				&ffCopy, nil, model.FloorKindFood, false,
				p.main.setMessage,
				func() { p.storageWiz = nil },
				func() {
					// Placed in storage — remove from floor
					p.char.CurrentHome.FloorItems = append(
						p.char.CurrentHome.FloorItems[:ffIdx],
						p.char.CurrentHome.FloorItems[ffIdx+1:]...)
					p.storageWiz = nil
					p.floorFoodFridgeTarget = -1
					p.selFloorItem = -1
					p.mode = rpModeFiltered
				},
			)
			p.floorFoodFridgeTarget = -1
			return
		}

	case rpModeTimeSkip:
		if p.selItem < 0 || p.selItem >= len(p.char.CurrentHome.RoomItems) {
			p.mode = rpModeAction
			return
		}
		placed := &p.char.CurrentHome.RoomItems[p.selItem]
		if p.selAct < 0 || p.selAct >= len(placed.Item.Actions) {
			p.mode = rpModeAction
			return
		}
		action := placed.Item.Actions[p.selAct]

		bx, by, bw, bh := p.main.panelX()+4, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
		if clicked && isHovered(mx, my, bx, by, bw, bh) {
			p.mode = rpModeAction
			return
		}

		hours := []int{1, 2, 4, 8}
		for i, h := range hours {
			ax, ay, aw, ah := p.main.panelX()+4, p.main.panelY()+4+float32(i)*42, float32(260), float32(36)
			ay += 40
			if clicked && isHovered(mx, my, ax, ay, aw, ah) {
				advanceTime(p.char, float64(h)*60)
				p.executeAction(placed, action)
				p.main.setMessage(fmt.Sprintf("%s for %d hour(s). Time advanced.", action, h))
				p.mode = rpModeAction
				return
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
				if model.IsExpired(s.Item.PurchaseDate, s.Item.Food.BaseExpiryDays, s.Item.MultiplierUsed, char.CurrentDate) {
					foodExpiredPenalty(char)
					placed.Stored[i].Item.UsesRemaining--
					if placed.Stored[i].Item.UsesRemaining == 0 {
						placed.Stored = append(placed.Stored[:i], placed.Stored[i+1:]...)
					}
					p.main.setMessage(fmt.Sprintf("%s was EXPIRED — stat penalty applied!", s.Item.Food.Name))
					return
				}
				inedible := s.Item.Food.Nutrition.Food == 0 && s.Item.Food.Nutrition.Energy == 0 &&
					s.Item.Food.Nutrition.Hygiene == 0 && s.Item.Food.Nutrition.Confidence == 0 &&
					s.Item.Food.Nutrition.Strength == 0 && s.Item.Food.OnEat == nil
				if inedible {
					rawFoodPenalty(char)
					p.main.setMessage(fmt.Sprintf("Eating raw %s was a bad idea.", s.Item.Food.Name))
				} else {
					applyNutrition(char, s.Item.Food)
					p.main.setMessage(fmt.Sprintf("Ate %s from fridge. Uses left: %d", s.Item.Food.Name, s.Item.UsesRemaining-1))
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
			char.CurrentHome.FloorItems = append(char.CurrentHome.FloorItems, model.FloorItem{
				Item: model.InventoryItem{
					Kind:          model.FloorKindFood,
					PurchaseDate:  sf.PurchaseDate,
					UsesRemaining: sf.UsesRemaining,
					Food:          sf.Food,
				},
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

const rpRowH = float32(26)
const rpCellSz = float32(36)

// ── draw ──────────────────────────────────────────────────────────────────────

func (p *roomPanel) draw(dst *ebiten.Image) {
	if p.storageWiz != nil {
		p.storageWiz.draw(dst)
		return
	}

	mx, my := ebiten.CursorPosition()
	char := p.char

	p.drawRoomGrid(dst)

	switch p.mode {
	case rpModeList:
		drawText(dst, "Placed Items", float64(p.main.panelX()+4), float64(p.main.panelY()+4)-4, fontS, colorMuted)
		if len(char.CurrentHome.RoomItems) == 0 && len(char.CurrentHome.FloorItems) == 0 {
			drawText(dst, "Room is empty. Buy items in the Shop tab.", float64(p.main.panelX()+4), float64(p.main.panelY()+4)+20, fontS, colorMuted)
		}
		// Which items should be highlighted (hovered cell)
		hoveredIndices := make(map[int]bool)
		if p.hoverCell[0] >= 0 {
			for _, idx := range p.itemsAtCell(p.hoverCell[0], p.hoverCell[1]) {
				hoveredIndices[idx] = true
			}
		}
		for i, placed := range char.CurrentHome.RoomItems {
			rx, ry, rw, rh := p.main.panelX()+4, p.main.panelY()+4+float32(i)*rpRowH, float32(260), rpRowH-2
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
		// floor items summary
		if len(char.CurrentHome.FloorItems) > 0 {
			yo := p.main.panelY() + 4 + 16 + float32(len(char.CurrentHome.RoomItems))*rpRowH + 12
			drawText(dst, "Floor Items:", float64(p.main.panelX()+4), float64(yo), fontS, colorMuted)
			yo += 18
			for _, fi := range char.CurrentHome.FloorItems {
				if fi.Item.Kind == model.FloorKindFood {
					drawText(dst, fmt.Sprintf("  %s (uses:%d)", fi.Item.Food.Name, fi.Item.UsesRemaining), float64(p.main.panelX()+4), float64(yo), fontS, colorFoodFloor)
				} else {
					drawText(dst, fmt.Sprintf("  %s", fi.Item.Utility.Name), float64(p.main.panelX()+4), float64(yo), fontS, colorUtilFloor)
				}
				yo += 18
			}
		}
		if p.hoverCell[0] >= 0 && len(hoveredIndices) > 0 {
			drawText(dst, "Click cell to select", float64(p.main.panelX()+4), float64(p.main.panelY()+4)+6, fontS, colorAccent)
		}

	case rpModeFiltered:
		stacked := p.itemsAtCell(p.selCell[0], p.selCell[1])
		floorItems := p.floorItemsAtCell(p.selCell[0], p.selCell[1])
		total := len(stacked) + len(floorItems)
		drawText(dst, fmt.Sprintf("Cell (%d,%d) — %d item(s)", p.selCell[0], p.selCell[1], total),
			float64(p.main.panelX()+4), float64(p.main.panelY()+4), fontS, colorAccent)
		for li, itemIdx := range stacked {
			placed := char.CurrentHome.RoomItems[itemIdx]
			rx, ry, rw, rh := p.main.panelX()+4, p.main.panelY()+4+float32(li)*rpRowH, float32(260), rpRowH-2
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
		// Floor item rows (food & utilities)
		baseRow := len(stacked)
		for li, fiIdx := range floorItems {
			fi := char.CurrentHome.FloorItems[fiIdx]
			rx, ry, rw, rh := p.main.panelX()+4, p.main.panelY()+4+float32(baseRow+li)*rpRowH, float32(260), rpRowH-2
			ry += 20
			hov := isHovered(mx, my, rx, ry, rw, rh)
			bg := colorPanel
			if hov {
				bg = colorHighlight
			}
			fillRect(dst, rx, ry, rw, rh, bg)
			strokeRect(dst, rx, ry, rw, rh, colorBorder)
			var label string
			var tc color.RGBA
			if fi.Item.Kind == model.FloorKindFood {
				exp := model.IsExpired(fi.Item.PurchaseDate, fi.Item.Food.BaseExpiryDays, 1, char.CurrentDate)
				tc = colorFoodFloor
				if exp {
					tc = colorRed
				}
				label = fmt.Sprintf("🍞 %s  uses:%d  z=%d", fi.Item.Food.Name, fi.Item.UsesRemaining, fi.Z)
				if exp {
					label += " [EXPIRED]"
				}
			} else {
				tc = colorUtilFloor
				label = fmt.Sprintf("🔧 %s  z=%d", fi.Item.Utility.Name, fi.Z)
			}
			drawText(dst, label, float64(rx)+6, float64(ry)+6, fontS, tc)
		}
		bx, by, bw, bh := p.main.panelX()+4, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
		drawButton(dst, "← Back", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)

	case rpModeAction:
		if p.selItem < 0 || p.selItem >= len(char.CurrentHome.RoomItems) {
			return
		}
		placed := char.CurrentHome.RoomItems[p.selItem]
		drawText(dst, placed.Item.Name, float64(p.main.panelX()+4), float64(p.main.panelY()+4), fontM, colorAccent)
		drawText(dst, fmt.Sprintf("at (%d,%d,z=%d) facing %s", placed.X, placed.Y, placed.Z, dirName(placed.Direction)),
			float64(p.main.panelX()+4), float64(p.main.panelY()+4)+20, fontS, colorMuted)

		if placed.Item.Storage != nil {
			used := usedSlotCount(&char.CurrentHome.RoomItems[p.selItem])
			total := placed.Item.Storage.TotalSlots()
			drawText(dst, fmt.Sprintf("Storage: %d/%d slots", used, total),
				float64(p.main.panelX()+4), float64(p.main.panelY()+4)+40, fontS, colorText)
			iy := float64(p.main.panelY()+4) + 60
			for _, s := range placed.Stored {
				exp := model.IsExpired(s.Item.PurchaseDate, s.Item.Food.BaseExpiryDays, s.Item.MultiplierUsed, char.CurrentDate)
				col := colorText
				tag := ""
				if exp {
					col = colorRed
					tag = " [EXPIRED]"
				}
				expiry := model.ExpiryDate(s.Item.PurchaseDate, s.Item.Food.BaseExpiryDays, s.Item.MultiplierUsed)
				ey, em, ed := expiry.Unpack()
				drawText(dst, fmt.Sprintf("  %s  uses:%d  exp:%04d-%02d-%02d%s",
					s.Item.Food.Name, s.Item.UsesRemaining, ey, em, ed, tag),
					float64(p.main.panelX()+4), iy, fontS, col)
				iy += 18
			}
		}

		if placed.Item.CookSurface != nil {
			drawText(dst, fmt.Sprintf("Surface: %d/%d slots", len(placed.OnSurface), placed.Item.CookSurface.Slots),
				float64(p.main.panelX()+4), float64(p.main.panelY()+4)+40, fontS, colorText)
			iy := float64(p.main.panelY()+4) + 60
			for _, sf := range placed.OnSurface {
				tag := ""
				if sf.Cooked {
					tag = " [cooked]"
				}
				drawText(dst, fmt.Sprintf("  %s%s uses:%d", sf.Food.Name, tag, sf.UsesRemaining),
					float64(p.main.panelX()+4), iy, fontS, colorText)
				iy += 18
			}
		}

		actY := float64(p.main.panelY()+4) + 176
		drawText(dst, "Actions:", float64(p.main.panelX()+4), actY, fontS, colorMuted)
		actY += 20
		for i, act := range placed.Item.Actions {
			ax, ay, aw, ah := p.main.panelX()+4, p.main.panelY()+4+float32(i)*42, float32(260), float32(36)
			ay = float32(actY) + float32(i)*42
			hov := isHovered(mx, my, ax, ay, aw, ah)
			drawButton(dst, act, ax, ay, aw, ah, fontM, hov, true)
		}

		bx, by, bw, bh := p.main.panelX()+4, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
		drawButton(dst, "← Back", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)
		mbx, mby, mbw, mbh := p.main.panelX()+4+108, p.main.panelY()+4+rpRowH*20-50, float32(120), float32(32)
		drawButton(dst, "✦ Move", mbx, mby, mbw, mbh, fontS, isHovered(mx, my, mbx, mby, mbw, mbh), true)
		sbx, sby, sbw, sbh := p.main.panelX()+4+236, p.main.panelY()+4+rpRowH*20-50, float32(80), float32(32)
		drawButton(dst, "Sell", sbx, sby, sbw, sbh, fontS, isHovered(mx, my, sbx, sby, sbw, sbh), true)
		tbx, tby, tbw, tbh := p.main.panelX()+4+324, p.main.panelY()+4+rpRowH*20-50, float32(80), float32(32)
		drawButton(dst, "Trash", tbx, tby, tbw, tbh, fontS, isHovered(mx, my, tbx, tby, tbw, tbh), true)
		if placed.Item.Storage != nil {
			obx, oby, obw, obh := p.main.panelX()+4+412, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
			drawButton(dst, "📦 Organize", obx, oby, obw, obh, fontS, isHovered(mx, my, obx, oby, obw, obh), true)
		}

	case rpModeUtilityChoice:
		if p.selFloorItem < 0 || p.utilTarget < 0 ||
			p.selFloorItem >= len(char.CurrentHome.FloorItems) ||
			p.utilTarget >= len(char.CurrentHome.FloorItems) {
			return
		}
		fi := char.CurrentHome.FloorItems[p.selFloorItem]
		util := char.CurrentHome.FloorItems[p.utilTarget]
		drawText(dst, "Use Utility?", float64(p.main.panelX()+4), float64(p.main.panelY()+4)+12, fontM, colorAccent)
		drawText(dst, fmt.Sprintf("%s + %s", fi.Item.Food.Name, util.Item.Utility.Name),
			float64(p.main.panelX()+4), float64(p.main.panelY()+4)+40, fontS, colorText)

		ex, ey, ew, eh := p.main.panelX()+4+108, p.main.panelY()+4+rpRowH*20-50, float32(80), float32(32)
		drawButton(dst, fmt.Sprintf("Use %s", util.Item.Utility.Name), ex, ey, ew, eh, fontS, isHovered(mx, my, ex, ey, ew, eh), true)
		mmx, mmy, mmw, mmh := p.main.panelX()+4+196, p.main.panelY()+4+rpRowH*20-50, float32(90), float32(32)
		drawButton(dst, "Place Separately", mmx, mmy, mmw, mmh, fontS, isHovered(mx, my, mmx, mmy, mmw, mmh), true)
		bx, by, bw, bh := p.main.panelX()+4, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
		drawButton(dst, "← Back", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)

	case rpModeFloorItemAction:
		if p.selFloorItem < 0 || p.selFloorItem >= len(char.CurrentHome.FloorItems) {
			return
		}
		fi := char.CurrentHome.FloorItems[p.selFloorItem]
		isFood := fi.Item.Kind == model.FloorKindFood
		headCol := colorUtilFloor
		name := fi.Item.Utility.Name
		if isFood {
			headCol = colorFoodFloor
			name = fi.Item.Food.Name
			exp := model.IsExpired(fi.Item.PurchaseDate, fi.Item.Food.BaseExpiryDays, 1, char.CurrentDate)
			if exp {
				headCol = colorRed
			}
		}
		drawText(dst, name, float64(p.main.panelX()+4), float64(p.main.panelY()+4), fontM, headCol)
		drawText(dst, fmt.Sprintf("at (%d,%d,z=%d)", fi.X, fi.Y, fi.Z),
			float64(p.main.panelX()+4), float64(p.main.panelY()+4)+18, fontS, colorMuted)
		if isFood {
			exp := model.IsExpired(fi.Item.PurchaseDate, fi.Item.Food.BaseExpiryDays, 1, char.CurrentDate)
			expiry := model.ExpiryDate(fi.Item.PurchaseDate, fi.Item.Food.BaseExpiryDays, 1)
			ey, em, ed := expiry.Unpack()
			expTag := fmt.Sprintf("Expires: %04d-%02d-%02d", ey, em, ed)
			if exp {
				expTag = fmt.Sprintf("EXPIRED %04d-%02d-%02d", ey, em, ed)
			}
			drawText(dst, expTag, float64(p.main.panelX()+4), float64(p.main.panelY()+4)+40, fontS, headCol)
			if isInedible(fi.Item.Food) {
				drawText(dst, "[raw / inedible — eating will penalise stats]", float64(p.main.panelX()+4), float64(p.main.panelY()+4)+58, fontS, colorYellow)
			}
		}
		if p.floorMoveMode {
			drawText(dst, "→ Click a floor cell to move item there", float64(p.main.panelX()+4), float64(p.main.panelY()+4)+78, fontS, colorYellow)
		}

		bx, by, bw, bh := p.main.panelX()+4, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
		drawButton(dst, "← Back", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)
		ex, ey, ew, eh := p.main.panelX()+4+108, p.main.panelY()+4+rpRowH*20-50, float32(80), float32(32)
		actionLabel := "Eat"
		if !isFood {
			actionLabel = "Use"
		}
		drawButton(dst, actionLabel, ex, ey, ew, eh, fontS, isHovered(mx, my, ex, ey, ew, eh) && !p.floorMoveMode, !p.floorMoveMode)
		mmx, mmy, mmw, mmh := p.main.panelX()+4+196, p.main.panelY()+4+rpRowH*20-50, float32(90), float32(32)
		moveLabel := "✦ Move"
		if p.floorMoveMode {
			moveLabel = "✦ Cancel"
		}
		drawButton(dst, moveLabel, mmx, mmy, mmw, mmh, fontS, isHovered(mx, my, mmx, mmy, mmw, mmh), true)
		tx, ty, tw, th := p.main.panelX()+4+294, p.main.panelY()+4+rpRowH*20-50, float32(80), float32(32)
		drawButton(dst, "Trash", tx, ty, tw, th, fontS, isHovered(mx, my, tx, ty, tw, th) && !p.floorMoveMode, !p.floorMoveMode)

	case rpModeFloorFoodFridgeChoice:
		if p.selFloorItem < 0 || p.floorFoodFridgeTarget < 0 ||
			p.selFloorItem >= len(char.CurrentHome.FloorItems) ||
			p.floorFoodFridgeTarget >= len(char.CurrentHome.RoomItems) {
			return
		}
		ff2 := char.CurrentHome.FloorItems[p.selFloorItem]
		fridgePlaced := char.CurrentHome.RoomItems[p.floorFoodFridgeTarget]
		topZ := fridgePlaced.Z + fridgePlaced.Item.Height

		drawText(dst, "Where to put "+ff2.Item.Food.Name+"?", float64(p.main.panelX()+4), float64(p.main.panelY()+4)+8, fontM, colorAccent)
		drawText(dst, fmt.Sprintf("Dropping onto: %s at (%d,%d)",
			fridgePlaced.Item.Name, p.floorFoodFridgeTargetXY[0], p.floorFoodFridgeTargetXY[1]),
			float64(p.main.panelX()+4), float64(p.main.panelY()+4)+32, fontS, colorMuted)
		drawText(dst, fmt.Sprintf("In World: places on top of fridge at z=%d", topZ),
			float64(p.main.panelX()+4), float64(p.main.panelY()+4)+96, fontS, colorMuted)
		drawText(dst, fmt.Sprintf("In Fridge: store inside (%d/%d slots used)",
			usedSlotCount(&char.CurrentHome.RoomItems[p.floorFoodFridgeTarget]),
			fridgePlaced.Item.Storage.TotalSlots()),
			float64(p.main.panelX()+4), float64(p.main.panelY()+4)+158, fontS, colorMuted)

		iwx, iwy, iww, iwh := p.main.panelX()+4, p.main.panelY()+4+120, float32(200), float32(48)
		drawButton(dst, "📦 In World (on top)", iwx, iwy, iww, iwh, fontM, isHovered(mx, my, iwx, iwy, iww, iwh), true)
		ifx, ify, ifw, ifh := p.main.panelX()+4, p.main.panelY()+4+180, float32(200), float32(48)
		fridgeFull := fridgePlaced.Item.Storage.TotalSlots() <= usedSlotCount(&char.CurrentHome.RoomItems[p.floorFoodFridgeTarget])
		drawButton(dst, "❄ In Fridge", ifx, ify, ifw, ifh, fontM, isHovered(mx, my, ifx, ify, ifw, ifh) && !fridgeFull, !fridgeFull)
		if fridgeFull {
			drawText(dst, "(fridge is full)", float64(ifx)+float64(ifw)+8, float64(ify)+14, fontS, colorRed)
		}

		bx, by, bw, bh := p.main.panelX()+4, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
		drawButton(dst, "← Back", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)

	case rpModeMoveGrid:
		drawText(dst, "Step 1: Click cell to move item", float64(p.main.panelX()+4), float64(p.main.panelY()+4)+4, fontS, colorMuted)
		if p.selItem >= 0 && p.selItem < len(char.CurrentHome.RoomItems) {
			placed := char.CurrentHome.RoomItems[p.selItem]
			drawText(dst, placed.Item.Name, float64(p.main.panelX()+4), float64(p.main.panelY()+4)+24, fontM, colorAccent)
		}
		if p.wizard != nil && p.wizard.err != "" {
			drawTextWrapped(dst, p.wizard.err, float64(p.main.panelX()+4), float64(p.main.panelY()+4)+50, float64(float32(260)), 18, fontS, colorRed)
		}
		bx, by, bw, bh := p.main.panelX()+4, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
		drawButton(dst, "← Cancel", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)

	case rpModeTimeSkip:
		if p.selItem < 0 || p.selItem >= len(char.CurrentHome.RoomItems) {
			return
		}
		placed := char.CurrentHome.RoomItems[p.selItem]
		drawText(dst, placed.Item.Name, float64(p.main.panelX()+4), float64(p.main.panelY()+4), fontM, colorAccent)
		if p.selAct >= 0 && p.selAct < len(placed.Item.Actions) {
			drawText(dst, fmt.Sprintf("Choose duration for: %s", placed.Item.Actions[p.selAct]),
				float64(p.main.panelX()+4), float64(p.main.panelY()+4)+24, fontS, colorText)
		}
		hours := []int{1, 2, 4, 8}
		for i, h := range hours {
			ax, ay, aw, ah := p.main.panelX()+4, p.main.panelY()+4+float32(i)*42, float32(260), float32(36)
			ay += 40
			hov := isHovered(mx, my, ax, ay, aw, ah)
			drawButton(dst, fmt.Sprintf("%d hour(s)", h), ax, ay, aw, ah, fontM, hov, true)
		}
		bx, by, bw, bh := p.main.panelX()+4, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
		drawButton(dst, "← Cancel", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)

	case rpModeMoveZ, rpModeMoveDir:
		if p.wizard != nil {
			p.wizard.drawLeftPanel(dst, mx, my)
		}
		bx, by, bw, bh := p.main.panelX()+4, p.main.panelY()+4+rpRowH*20-50, float32(100), float32(32)
		drawButton(dst, "← Back", bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)

	case rpModeDoorPopup:
		// Popup drawn in drawRoomGrid overlay
	}
}

func (p *roomPanel) drawRoomGrid(dst *ebiten.Image) {
	home := p.char.CurrentHome
	layout := home.Type.Layout
	if len(layout) == 0 {
		return
	}
	ox := p.main.panelX() + 274
	oy := p.main.panelY() + 4 + 20

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
			rx, ry, rw, rh := p.main.panelX()+4, p.main.panelY()+4+float32(i)*rpRowH, float32(260), rpRowH-2
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

	// Floor food & utilities
	for _, fi := range home.FloorItems {
		cx := ox + float32(fi.X)*rpCellSz
		cy := oy + float32(fi.Y)*rpCellSz
		if fi.Item.Kind == model.FloorKindFood {
			fillRect(dst, cx+8, cy+8, rpCellSz-17, rpCellSz-17, colorFoodFloor)
		} else {
			fillRect(dst, cx+4, cy+4, rpCellSz-9, rpCellSz-9, colorUtilFloor)
		}
	}

	// Hover cursor outline in list/moveGrid mode
	if p.mode == rpModeList || p.mode == rpModeMoveGrid {
		mx, my := ebiten.CursorPosition()
		hx, hy := p.gridCellAt(mx, my)
		if hx >= 0 {
			cx := ox + float32(hx)*rpCellSz
			cy := oy + float32(hy)*rpCellSz
			borderColor := colorAccent
			if p.isDoorCell(hx, hy) {
				borderColor = colorYellow
			}
			strokeRect(dst, cx, cy, rpCellSz-1, rpCellSz-1, borderColor)
		}
	}

	// Selected cell outline in filtered/action mode
	if (p.mode == rpModeFiltered || p.mode == rpModeAction) && p.selCell[0] >= 0 {
		cx := ox + float32(p.selCell[0])*rpCellSz
		cy := oy + float32(p.selCell[1])*rpCellSz
		strokeRect(dst, cx, cy, rpCellSz-1, rpCellSz-1, colorYellow)
	}

	drawText(dst, fmt.Sprintf("Room: %s", home.Type.Name), float64(ox), float64(oy)-18, fontS, colorMuted)

	// Door popup overlay
	if p.doorPopupOpen {
		// Semi-transparent backdrop on grid area only
		// Draw popup box near the door cell
		mx, my := ebiten.CursorPosition()
		ox := p.main.panelX() + 274
		oy := p.main.panelY() + 4 + 20
		cx := ox + float32(p.doorCell[0])*rpCellSz
		cy := oy + float32(p.doorCell[1])*rpCellSz
		bx, by, bw, bh := cx-40, cy+rpCellSz+4, float32(140), float32(32)
		// Popup background
		popX := bx - 30
		popY := by - 38
		popW := bw + 60
		popH := bh + 46
		fillRect(dst, popX, popY, popW, popH, colorBg)
		strokeRect(dst, popX, popY, popW, popH, colorAccent)
		drawText(dst, "🚪 Door", float64(popX)+10, float64(popY)+10, fontS, colorMuted)
		drawButton(dst, "Go to Shop", bx, by, bw, bh, fontM, isHovered(mx, my, bx, by, bw, bh), true)
	}
}
