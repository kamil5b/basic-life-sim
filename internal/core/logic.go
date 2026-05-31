package core

import (
	"basic-life-sim/internal/model"
	"fmt"
)

// ── spatial helpers ──────────────────────────────────────────────────────────

func itemFootprint(item model.RoomItem, dir model.Direction) (cols, rows uint8) {
	switch dir {
	case model.East, model.West:
		return item.Length, item.Width
	default:
		return item.Width, item.Length
	}
}

func occupiedCells(ox, oy uint8, item model.RoomItem, dir model.Direction) [][2]uint8 {
	cols, rows := itemFootprint(item, dir)
	cells := make([][2]uint8, 0, int(cols)*int(rows))
	for r := uint8(0); r < rows; r++ {
		for c := uint8(0); c < cols; c++ {
			cells = append(cells, [2]uint8{ox + c, oy + r})
		}
	}
	return cells
}

func zRangeOverlaps(az, aHeight, bz, bHeight uint8) bool {
	aTop := az + aHeight
	bTop := bz + bHeight
	return az <= bTop && bz <= aTop
}

// canPlaceOnGrid checks only layout bounds and floor-tile validity for all
// possible directions a pending item could face when anchored at (x, y).
// It does NOT check Z or item-vs-item clearance — those are verified later.
// Returns an error if no direction would produce a valid footprint.
func canPlaceOnGrid(home model.Home, item model.RoomItem, x, y uint8) error {
	layout := home.Type.Layout
	rows := uint8(len(layout))
	if rows == 0 {
		return fmt.Errorf("empty layout")
	}
	cols := uint8(len(layout[0]))
	dirs := []model.Direction{model.North, model.East, model.South, model.West}
	for _, dir := range dirs {
		valid := true
		for _, cell := range occupiedCells(x, y, item, dir) {
			cx, cy := cell[0], cell[1]
			if cy >= rows || cx >= cols {
				valid = false
				break
			}
			switch layout[cy][cx] {
			case model.HomeCellFloor:
				// ok
			case model.HomeCellDoor:
				if item.Height > 0 {
					valid = false
				}
			default:
				valid = false
			}
			if !valid {
				break
			}
		}
		if !valid {
			continue
		}
		// Check direct footprint overlap with existing items.
		newCells := occupiedCells(x, y, item, dir)
		newCellSet := make(map[[2]uint8]bool, len(newCells))
		for _, c := range newCells {
			newCellSet[c] = true
		}
		blocked := false
		for _, placed := range home.RoomItems {
			if !zRangeOverlaps(0, item.Height, placed.Z, placed.Item.Height) {
				continue
			}
			// Direct overlap.
			for _, ec := range occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction) {
				if newCellSet[ec] {
					blocked = true
					break
				}
			}
			if blocked {
				break
			}
			// Clearance zone of existing item.
			if !placed.Item.NeedClearance {
				continue
			}
			sdx, sdy := dirStepXY(placed.Direction)
			existingCells := occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction)
			placedSet := make(map[[2]uint8]bool, len(existingCells))
			for _, c := range existingCells {
				placedSet[c] = true
			}
			for _, ec := range existingCells {
				front := [2]uint8{uint8(int(ec[0]) + sdx), uint8(int(ec[1]) + sdy)}
				if placedSet[front] {
					continue
				}
				c1 := [2]uint8{uint8(int(ec[0]) + sdx), uint8(int(ec[1]) + sdy)}
				if newCellSet[c1] {
					blocked = true
					break
				}
			}
			if blocked {
				break
			}
		}
		if !blocked {
			return nil
		}
	}
	return fmt.Errorf("no valid placement at (%d,%d): blocked by walls or nearby items", x, y)
}

func canPlace(home model.Home, item model.RoomItem, x, y, z uint8, dir model.Direction) error {
	layout := home.Type.Layout
	rows := uint8(len(layout))
	if rows == 0 {
		return fmt.Errorf("empty layout")
	}
	cols := uint8(len(layout[0]))
	top := z + item.Height
	if top > home.Type.MaxHeight {
		return fmt.Errorf("item height exceeds room max height (%d)", home.Type.MaxHeight)
	}
	for _, cell := range occupiedCells(x, y, item, dir) {
		cx, cy := cell[0], cell[1]
		if cy >= rows || cx >= cols {
			return fmt.Errorf("item does not fit within room boundaries")
		}
		switch layout[cy][cx] {
		case model.HomeCellFloor:
		case model.HomeCellDoor:
			if item.Height != 0 {
				return fmt.Errorf("position (%d,%d) is a door", cx, cy)
			}
		default:
			return fmt.Errorf("position (%d,%d) is not a floor tile", cx, cy)
		}
	}
	newCells := occupiedCells(x, y, item, dir)
	newCellSet := make(map[[2]uint8]bool, len(newCells))
	for _, c := range newCells {
		newCellSet[c] = true
	}

	for _, placed := range home.RoomItems {
		if !zRangeOverlaps(z, item.Height, placed.Z, placed.Item.Height) {
			continue
		}
		existingCells := occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction)
		// Direct overlap check.
		for _, existing := range existingCells {
			for _, cell := range newCells {
				if existing == cell {
					return fmt.Errorf("position (%d,%d) occupied by %s", cell[0], cell[1], placed.Item.Name)
				}
			}
		}
		// Clearance check: only applies when the already-placed item needs clearance
		// AND the new item also needs clearance.
		if !placed.Item.NeedClearance {
			continue
		}
		sdx, sdy := dirStepXY(placed.Direction)
		placedSet := make(map[[2]uint8]bool, len(existingCells))
		for _, c := range existingCells {
			placedSet[c] = true
		}
		for _, ec := range existingCells {
			// Only front-edge cells of the placed item.
			front := [2]uint8{uint8(int(ec[0]) + sdx), uint8(int(ec[1]) + sdy)}
			if placedSet[front] {
				continue // not a front-edge cell
			}
			// The immediate cell in front is reserved.
			clear1 := [2]uint8{uint8(int(ec[0]) + sdx), uint8(int(ec[1]) + sdy)}
			if newCellSet[clear1] {
				return fmt.Errorf("position (%d,%d) is in the clearance zone in front of %s", clear1[0], clear1[1], placed.Item.Name)
			}
		}
	}
	return nil
}

// dirStepXY returns the (dx, dy) unit step for a cardinal direction.
func dirStepXY(dir model.Direction) (dx, dy int) {
	switch dir {
	case model.North:
		return 0, -1
	case model.South:
		return 0, 1
	case model.East:
		return 1, 0
	case model.West:
		return -1, 0
	}
	return 0, 0
}

// ── fridge slot helpers ──────────────────────────────────────────────────────

func occupiedFoodSlots(placed *model.PlacedRoomItem) map[[3]uint8]bool {
	occupied := make(map[[3]uint8]bool)
	for _, s := range placed.Stored {
		for dz := uint8(0); dz < s.Food.Height; dz++ {
			for dy := uint8(0); dy < s.Food.Length; dy++ {
				for dx := uint8(0); dx < s.Food.Width; dx++ {
					occupied[[3]uint8{s.SlotX + dx, s.SlotY + dy, s.SlotZ + dz}] = true
				}
			}
		}
	}
	return occupied
}

func usedSlotCount(placed *model.PlacedRoomItem) int {
	return len(occupiedFoodSlots(placed))
}

func canFitFood(occupied map[[3]uint8]bool, cap *model.StorageCapacity, f model.Food, ax, ay, az uint8) bool {
	if ax+f.Width > cap.Width || ay+f.Length > cap.Length || az+f.Height > cap.Height {
		return false
	}
	for dz := uint8(0); dz < f.Height; dz++ {
		for dy := uint8(0); dy < f.Length; dy++ {
			for dx := uint8(0); dx < f.Width; dx++ {
				if occupied[[3]uint8{ax + dx, ay + dy, az + dz}] {
					return false
				}
			}
		}
	}
	return true
}

func nextFreeSlot(placed *model.PlacedRoomItem, cap *model.StorageCapacity, f model.Food) ([3]uint8, bool) {
	occupied := occupiedFoodSlots(placed)
	for z := uint8(0); z < cap.Height; z++ {
		for y := uint8(0); y < cap.Length; y++ {
			for x := uint8(0); x < cap.Width; x++ {
				if canFitFood(occupied, cap, f, x, y, z) {
					return [3]uint8{x, y, z}, true
				}
			}
		}
	}
	return [3]uint8{}, false
}

func findFridges(char *model.Character) []int {
	var out []int
	for i, p := range char.CurrentHome.RoomItems {
		if p.Item.Storage != nil {
			out = append(out, i)
		}
	}
	return out
}

// ── food mutation helpers ────────────────────────────────────────────────────

func safeSub16(v, delta uint16) uint16 {
	if delta > v {
		return 0
	}
	return v - delta
}

func safeAdd16(v, delta uint16) uint16 {
	r := v + delta
	if r < v {
		return ^uint16(0)
	}
	return r
}

func foodExpiredPenalty(char *model.Character) {
	const pen = 10
	char.Food.Current = safeSub16(char.Food.Current, pen)
	char.Energy.Current = safeSub16(char.Energy.Current, pen)
	char.Hygiene.Current = safeSub16(char.Hygiene.Current, pen)
	char.Confidence.Current = safeSub16(char.Confidence.Current, pen)
	char.Strength.Current = safeSub16(char.Strength.Current, pen)
}

func rawFoodPenalty(char *model.Character) {
	const pen = 5
	char.Food.Current = safeSub16(char.Food.Current, pen)
	char.Energy.Current = safeSub16(char.Energy.Current, pen)
	char.Hygiene.Current = safeSub16(char.Hygiene.Current, pen)
}

func needPct(need model.NeedStat) uint16 {
	if need.Max == 0 {
		return 0
	}
	return need.Current * 100 / need.Max
}

func dirName(d model.Direction) string {
	switch d {
	case model.North:
		return "N"
	case model.East:
		return "E"
	case model.South:
		return "S"
	case model.West:
		return "W"
	default:
		return "?"
	}
}

// consumeUse decrements uses on a food source and removes it if exhausted.
func consumeUse(char *model.Character, kind string, itemIdx, foodIdx int) {
	switch kind {
	case "floor":
		f := &char.CurrentHome.FloorFood[foodIdx]
		f.UsesRemaining--
		if f.UsesRemaining == 0 {
			char.CurrentHome.FloorFood = append(char.CurrentHome.FloorFood[:foodIdx], char.CurrentHome.FloorFood[foodIdx+1:]...)
		}
	case "fridge":
		s := &char.CurrentHome.RoomItems[itemIdx].Stored[foodIdx]
		s.UsesRemaining--
		if s.UsesRemaining == 0 {
			ri := &char.CurrentHome.RoomItems[itemIdx]
			ri.Stored = append(ri.Stored[:foodIdx], ri.Stored[foodIdx+1:]...)
		}
	case "surface":
		sf := &char.CurrentHome.RoomItems[itemIdx].OnSurface[foodIdx]
		sf.UsesRemaining--
		if sf.UsesRemaining == 0 {
			ri := &char.CurrentHome.RoomItems[itemIdx]
			ri.OnSurface = append(ri.OnSurface[:foodIdx], ri.OnSurface[foodIdx+1:]...)
		}
	}
}

func removeFoodEntirely(char *model.Character, kind string, itemIdx, foodIdx int) {
	switch kind {
	case "floor":
		char.CurrentHome.FloorFood = append(char.CurrentHome.FloorFood[:foodIdx], char.CurrentHome.FloorFood[foodIdx+1:]...)
	case "fridge":
		ri := &char.CurrentHome.RoomItems[itemIdx]
		ri.Stored = append(ri.Stored[:foodIdx], ri.Stored[foodIdx+1:]...)
	case "surface":
		ri := &char.CurrentHome.RoomItems[itemIdx]
		ri.OnSurface = append(ri.OnSurface[:foodIdx], ri.OnSurface[foodIdx+1:]...)
	}
}
