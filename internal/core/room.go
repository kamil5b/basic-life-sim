package core

import (
	"basic-life-sim/internal/model"
	"fmt"
)

// itemFootprint returns the (cols, rows) the item occupies given its direction.
// North/South: Width=cols, Length=rows. East/West: rotated.
func itemFootprint(item model.RoomItem, dir model.Direction) (cols, rows uint8) {
	switch dir {
	case model.East, model.West:
		return item.Length, item.Width
	default:
		return item.Width, item.Length
	}
}

// occupiedCells returns all (x,y) cells an item placed at (ox,oy) covers.
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

// zRangeOverlaps returns true if two items' vertical ranges overlap.
// Height 0 is treated as occupying exactly z..z (a flat item still has presence at its Z level).
func zRangeOverlaps(az, aHeight, bz, bHeight uint8) bool {
	aTop := az + aHeight
	bTop := bz + bHeight
	// ranges are [az, aTop] and [bz, bTop] (inclusive)
	return az <= bTop && bz <= aTop
}

// canPlace checks whether an item can be placed at (x,y,z) in the home.
// Rules:
//   - All XY footprint cells must be floor tiles.
//   - Door cells are only allowed when item.Height == 0 (flat items don't block doorways).
//   - Two items collide only when they share an XY cell AND their Z ranges overlap.
//   - The item's top (z + height) must not exceed the room's MaxHeight.
func canPlace(home model.Home, item model.RoomItem, x, y, z uint8, dir model.Direction) error {
	layout := home.Type.Layout
	rows := uint8(len(layout))
	if rows == 0 {
		return fmt.Errorf("empty layout")
	}
	cols := uint8(len(layout[0]))

	// check Z ceiling
	top := z + item.Height
	if top > home.Type.MaxHeight {
		return fmt.Errorf("item height exceeds room max height (%d)", home.Type.MaxHeight)
	}

	for _, cell := range occupiedCells(x, y, item, dir) {
		cx, cy := cell[0], cell[1]
		if cy >= rows || cx >= cols {
			return fmt.Errorf("item does not fit within the room boundaries")
		}
		switch layout[cy][cx] {
		case model.HomeCellFloor:
			// always ok
		case model.HomeCellDoor:
			if item.Height != 0 {
				return fmt.Errorf("position (%d,%d) is a door — only height-0 items may be placed there", cx, cy)
			}
		default:
			return fmt.Errorf("position (%d,%d) is not a floor tile", cx, cy)
		}
	}

	// check 3-D overlap with already placed items
	for _, placed := range home.RoomItems {
		if !zRangeOverlaps(z, item.Height, placed.Z, placed.Item.Height) {
			continue
		}
		for _, existing := range occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction) {
			for _, cell := range occupiedCells(x, y, item, dir) {
				if existing == cell {
					return fmt.Errorf("position (%d,%d) at z=%d-%d is already occupied by %s (z=%d-%d)",
						cell[0], cell[1], z, top, placed.Item.Name, placed.Z, placed.Z+placed.Item.Height)
				}
			}
		}
	}
	return nil
}

// printRoom renders the home layout with placed items overlaid.
// Items are shown as the first letter of their name.
func printRoom(home model.Home) {
	layout := home.Type.Layout
	rows := len(layout)
	if rows == 0 {
		return
	}
	cols := len(layout[0])

	// build overlay grid
	grid := make([][]rune, rows)
	for r := range grid {
		grid[r] = make([]rune, cols)
		for c := range grid[r] {
			switch layout[r][c] {
			case model.HomeCellWall:
				grid[r][c] = '█'
			case model.HomeCellFloor:
				grid[r][c] = ' '
			case model.HomeCellDoor:
				grid[r][c] = 'D'
			default:
				grid[r][c] = '?'
			}
		}
	}

	for _, placed := range home.RoomItems {
		symbol := rune(placed.Item.Name[0])
		for _, cell := range occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction) {
			cx, cy := int(cell[0]), int(cell[1])
			if cy < rows && cx < cols {
				grid[cy][cx] = symbol
			}
		}
	}

	fmt.Println("===========================================")
	fmt.Printf("Room: %s\n", home.Type.Name)
	fmt.Println("===========================================")
	for _, row := range grid {
		for _, cell := range row {
			fmt.Printf("%c", cell)
		}
		fmt.Println()
	}
	fmt.Println("===========================================")
	fmt.Println("Legend:")
	for _, placed := range home.RoomItems {
		fmt.Printf("  %c = %s (at %d,%d z=%d)\n", rune(placed.Item.Name[0]), placed.Item.Name, placed.X, placed.Y, placed.Z)
	}
	fmt.Println("===========================================")
}

// checkRoomMenu shows the room, lets the player pick a placed item, then pick an action.
func checkRoomMenu(char *model.Character) {
	for {
		if len(char.CurrentHome.RoomItems) == 0 {
			fmt.Println("Your room is empty. Buy some items first.")
			return
		}

		printRoom(char.CurrentHome)

		fmt.Println("Choose an item to interact with (0 to go back):")
		for i, placed := range char.CurrentHome.RoomItems {
			fmt.Printf("%d. %s (at %d,%d z=%d facing %s)\n", i+1, placed.Item.Name, placed.X, placed.Y, placed.Z, dirName(placed.Direction))
		}

		var choice int
		fmt.Scanln(&choice)
		if choice == 0 {
			return
		}
		if choice < 1 || choice > len(char.CurrentHome.RoomItems) {
			fmt.Println("Invalid choice.")
			continue
		}

		placed := &char.CurrentHome.RoomItems[choice-1]
		doItemAction(placed, char)
	}
}

// doItemAction prompts the player to pick an action for an item and executes it.
func doItemAction(placed *model.PlacedRoomItem, char *model.Character) {
	actions := availableActions(placed.Item)
	if len(actions) == 0 {
		fmt.Printf("%s has no available actions.\n", placed.Item.Name)
		return
	}

	fmt.Printf("\n--- %s ---\n", placed.Item.Name)
	fmt.Println("Choose an action (0 to go back):")
	for i, a := range actions {
		fmt.Printf("%d. %s\n", i+1, a)
	}

	var choice int
	fmt.Scanln(&choice)
	if choice == 0 {
		return
	}
	if choice < 1 || choice > len(actions) {
		fmt.Println("Invalid choice.")
		return
	}

	action := actions[choice-1]
	placed.Item.DoAction(action, &char.CurrentStats, char)
	fmt.Printf("You %s using %s.\n", action, placed.Item.Name)
}

// availableActions returns the actions declared on the item.
func availableActions(item model.RoomItem) []string {
	return item.Actions
}

func dirName(d model.Direction) string {
	switch d {
	case model.North:
		return "North"
	case model.East:
		return "East"
	case model.South:
		return "South"
	case model.West:
		return "West"
	default:
		return "Unknown"
	}
}
