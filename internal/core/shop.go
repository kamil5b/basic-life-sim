package core

import (
	buyablefood "basic-life-sim/internal/buyable/food"
	roomitem "basic-life-sim/internal/buyable/room-item"
	"basic-life-sim/internal/model"
	"fmt"
)

func buyItemMenu(char *model.Character) {
	for {
		fmt.Println("\n--- Buy Item ---")
		fmt.Println("1. Appliances")
		fmt.Println("2. Furniture")
		fmt.Println("3. Hygiene")
		fmt.Println("4. Food")
		fmt.Println("0. Back")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 0:
			return
		case 1:
			buyFromCatalog(char, roomitem.Appliances, "Appliances")
		case 2:
			buyFromCatalog(char, roomitem.Furniture, "Furniture")
		case 3:
			buyFromCatalog(char, roomitem.Hygiene, "Hygiene")
		case 4:
			buyFoodMenu(char)
		default:
			fmt.Println("Invalid choice.")
		}
	}
}

func buyFromCatalog(char *model.Character, catalog []model.RoomItem, title string) {
	for {
		fmt.Printf("\n--- %s ---\n", title)
		fmt.Printf("Your money: $%.2f\n\n", char.CurrentStats.Money)

		for i, item := range catalog {
			affordable := ""
			if char.CurrentStats.Money < item.BasePrice {
				affordable = " [can't afford]"
			}
			fmt.Printf("%d. %-22s $%.2f  %dx%d  actions: %v%s\n",
				i+1, item.Name, item.BasePrice,
				item.Width, item.Length,
				item.Actions, affordable,
			)
		}
		fmt.Println("0. Back")

		var choice int
		fmt.Scanln(&choice)
		if choice == 0 {
			return
		}
		if choice < 1 || choice > len(catalog) {
			fmt.Println("Invalid choice.")
			continue
		}

		item := catalog[choice-1]
		if char.CurrentStats.Money < item.BasePrice {
			fmt.Printf("Not enough money. You need $%.2f but have $%.2f.\n", item.BasePrice, char.CurrentStats.Money)
			continue
		}

		placed, err := promptPlacement(char, item)
		if err != nil {
			fmt.Println("Placement cancelled:", err)
			continue
		}

		char.CurrentStats.Money -= item.BasePrice
		char.CurrentHome.RoomItems = append(char.CurrentHome.RoomItems, placed)
		fmt.Printf("Bought and placed %s. Remaining money: $%.2f\n", item.Name, char.CurrentStats.Money)
	}
}

// buyFoodMenu routes to fridge storage or room placement depending on whether a fridge exists.
func buyFoodMenu(char *model.Character) {
	fridges := findFridges(char)
	if len(fridges) == 0 {
		buyFoodToRoom(char)
		return
	}

	// pick fridge
	var fridgeIdx int
	if len(fridges) == 1 {
		fridgeIdx = fridges[0]
		fmt.Printf("Using your %s.\n", char.CurrentHome.RoomItems[fridgeIdx].Item.Name)
	} else {
		fmt.Println("Choose a refrigerator:")
		for i, idx := range fridges {
			placed := char.CurrentHome.RoomItems[idx]
			cap := placed.Item.Storage
			used := usedSlotCount(&char.CurrentHome.RoomItems[idx])
			fmt.Printf("%d. %s at (%d,%d) — %d/%d slots used\n",
				i+1, placed.Item.Name, placed.X, placed.Y, used, cap.TotalSlots())
		}
		fmt.Println("0. Back")
		var choice int
		fmt.Scanln(&choice)
		if choice == 0 {
			return
		}
		if choice < 1 || choice > len(fridges) {
			fmt.Println("Invalid choice.")
			return
		}
		fridgeIdx = fridges[choice-1]
	}

	placed := &char.CurrentHome.RoomItems[fridgeIdx]
	buyFoodIntoFridge(char, placed)
}

func buyFoodIntoFridge(char *model.Character, placed *model.PlacedRoomItem) {
	cap := placed.Item.Storage
	for {
		used := usedSlotCount(placed)
		total := cap.TotalSlots()
		fmt.Printf("\n--- Buy Food -> %s (%d/%d slots used) ---\n", placed.Item.Name, used, total)
		fmt.Printf("Your money: $%.2f\n\n", char.CurrentStats.Money)

		if used >= total {
			fmt.Println("Refrigerator is full. Remove some food first.")
			return
		}

		for i, f := range buyablefood.All {
			affordable := ""
			if char.CurrentStats.Money < f.BasePrice {
				affordable = " [can't afford]"
			}
			fmt.Printf("%d. %-18s $%.2f  %dx%dx%d  [%s]%s\n",
				i+1, f.Name, f.BasePrice,
				f.Width, f.Length, f.Height,
				f.Type, affordable)
		}
		fmt.Println("0. Back")

		var choice int
		fmt.Scanln(&choice)
		if choice == 0 {
			return
		}
		if choice < 1 || choice > len(buyablefood.All) {
			fmt.Println("Invalid choice.")
			continue
		}

		f := buyablefood.All[choice-1]
		if char.CurrentStats.Money < f.BasePrice {
			fmt.Printf("Not enough money. You need $%.2f but have $%.2f.\n", f.BasePrice, char.CurrentStats.Money)
			continue
		}

		slot, ok := promptFridgeSlot(placed, cap, f)
		if !ok {
			continue
		}

		mult := cap.MultiplierAt(slot[0], slot[1], slot[2])
		char.CurrentStats.Money -= f.BasePrice
		placed.Stored = append(placed.Stored, model.StoredFood{
			SlotX:          slot[0],
			SlotY:          slot[1],
			SlotZ:          slot[2],
			PurchaseDate:   char.CurrentDate,
			MultiplierUsed: mult,
			UsesRemaining:  f.UsesTotal,
			Food:           f,
		})
		expiry := model.ExpiryDate(char.CurrentDate, f.BaseExpiryDays, mult)
		ey, em, ed := expiry.Unpack()
		zoneNote := ""
		if mult > cap.ExpiryMultiplier {
			zoneNote = fmt.Sprintf(" [cold zone %.0fx]", mult)
		}
		fmt.Printf("Bought %s → fridge slot (%d,%d,%d)%s, expires %04d-%02d-%02d, uses: %d. Money: $%.2f\n",
			f.Name, slot[0], slot[1], slot[2], zoneNote, ey, em, ed, f.UsesTotal, char.CurrentStats.Money)
	}
}

// buyFoodToRoom lets the player buy food and place it directly on the room floor (no fridge).
func buyFoodToRoom(char *model.Character) {
	fmt.Println("No refrigerator found. Food will be placed in your room at room temperature.")
	for {
		y, m, d := char.CurrentDate.Unpack()
		fmt.Printf("\n--- Buy Food (room storage) --- [Date: %04d-%02d-%02d]\n", y, m, d)
		fmt.Printf("Your money: $%.2f\n\n", char.CurrentStats.Money)

		for i, f := range buyablefood.All {
			affordable := ""
			if char.CurrentStats.Money < f.BasePrice {
				affordable = " [can't afford]"
			}
			expiry := model.ExpiryDate(char.CurrentDate, f.BaseExpiryDays, 1)
			ey, em, ed := expiry.Unpack()
			fmt.Printf("%d. %-18s $%.2f  %dx%dx%d  [%s]  expires %04d-%02d-%02d  uses:%d%s\n",
				i+1, f.Name, f.BasePrice,
				f.Width, f.Length, f.Height,
				f.Type, ey, em, ed, f.UsesTotal, affordable)
		}
		fmt.Println("0. Back")

		var choice int
		fmt.Scanln(&choice)
		if choice == 0 {
			return
		}
		if choice < 1 || choice > len(buyablefood.All) {
			fmt.Println("Invalid choice.")
			continue
		}

		f := buyablefood.All[choice-1]
		if char.CurrentStats.Money < f.BasePrice {
			fmt.Printf("Not enough money. You need $%.2f but have $%.2f.\n", f.BasePrice, char.CurrentStats.Money)
			continue
		}

		fmt.Printf("\n--- Place %s in room (size %dx%dx%d) ---\n", f.Name, f.Width, f.Length, f.Height)
		printRoom(char.CurrentHome)
		fmt.Println("Enter X position:")
		var x int
		fmt.Scanln(&x)
		fmt.Println("Enter Y position:")
		var yPos int
		fmt.Scanln(&yPos)
		fmt.Println("Enter Z position:")
		var z int
		fmt.Scanln(&z)
		if x < 0 || yPos < 0 || z < 0 {
			fmt.Println("Coordinates must be non-negative.")
			continue
		}

		char.CurrentStats.Money -= f.BasePrice
		expiry := model.ExpiryDate(char.CurrentDate, f.BaseExpiryDays, 1)
		ey, em, ed := expiry.Unpack()
		char.CurrentHome.FloorFood = append(char.CurrentHome.FloorFood, model.PlacedFood{
			X:             uint8(x),
			Y:             uint8(yPos),
			Z:             uint8(z),
			PurchaseDate:  char.CurrentDate,
			UsesRemaining: f.UsesTotal,
			Food:          f,
		})
		fmt.Printf("Bought %s → room (%d,%d,%d), expires %04d-%02d-%02d, uses: %d. Money: $%.2f\n",
			f.Name, x, yPos, z, ey, em, ed, f.UsesTotal, char.CurrentStats.Money)
	}
}

// printFridgeGrid renders a 2-D layer-by-layer view of the fridge interior.
// Each layer (z) is shown as a Width×Length grid. Occupied cells show the first letter of the food name.
func printFridgeGrid(placed *model.PlacedRoomItem) {
	cap := placed.Item.Storage
	// build slot map
	slotMap := make(map[[3]uint8]string)
	for _, s := range placed.Stored {
		for dz := uint8(0); dz < s.Food.Height; dz++ {
			for dy := uint8(0); dy < s.Food.Length; dy++ {
				for dx := uint8(0); dx < s.Food.Width; dx++ {
					key := [3]uint8{s.SlotX + dx, s.SlotY + dy, s.SlotZ + dz}
					slotMap[key] = string([]rune(s.Food.Name)[0:1])
				}
			}
		}
	}

	fmt.Printf("\n=== %s interior (W=%d L=%d H=%d) ===\n", placed.Item.Name, cap.Width, cap.Length, cap.Height)
	for z := uint8(0); z < cap.Height; z++ {
		// determine if this layer is a cold zone
		coldNote := ""
		for _, cz := range cap.ColdZones {
			if z >= cz.OriginZ && z < cz.OriginZ+cz.Height {
				coldNote = fmt.Sprintf(" [cold zone %.0fx]", cz.ExpiryMultiplier)
				break
			}
		}
		fmt.Printf(" Layer z=%d%s\n", z, coldNote)
		// column header
		fmt.Print("     ")
		for x := uint8(0); x < cap.Width; x++ {
			fmt.Printf(" x%-2d", x)
		}
		fmt.Println()
		for y := uint8(0); y < cap.Length; y++ {
			fmt.Printf(" y%-2d ", y)
			for x := uint8(0); x < cap.Width; x++ {
				sym, ok := slotMap[[3]uint8{x, y, z}]
				if ok {
					fmt.Printf(" %-3s", sym)
				} else {
					fmt.Print(" .  ")
				}
			}
			fmt.Println()
		}
	}
	fmt.Print(" Legend: ")
	seen := make(map[string]bool)
	for _, s := range placed.Stored {
		if !seen[s.Food.Name] {
			seen[s.Food.Name] = true
			fmt.Printf("%s=%s  ", string([]rune(s.Food.Name)[0:1]), s.Food.Name)
		}
	}
	fmt.Println(". = empty")
	fmt.Println("===========================================")
}

// promptFridgeSlot shows the fridge grid and asks the user to pick a slot for the given food.
// Returns ([3]uint8, true) on success, ([3]uint8{}, false) on cancel or invalid.
func promptFridgeSlot(placed *model.PlacedRoomItem, cap *model.StorageCapacity, f model.Food) ([3]uint8, bool) {
	printFridgeGrid(placed)
	fmt.Printf("Placing: %s (size %dx%dx%d)\n", f.Name, f.Width, f.Length, f.Height)

	// show cold zone info
	if len(cap.ColdZones) > 0 {
		for _, cz := range cap.ColdZones {
			fmt.Printf("  Cold zone: x%d-%d, y%d-%d, z%d-%d → %.0fx expiry\n",
				cz.OriginX, cz.OriginX+cz.Width-1,
				cz.OriginY, cz.OriginY+cz.Length-1,
				cz.OriginZ, cz.OriginZ+cz.Height-1,
				cz.ExpiryMultiplier)
		}
	}

	fmt.Println("Enter X (or -1 to cancel):")
	var x int
	fmt.Scanln(&x)
	if x < 0 {
		return [3]uint8{}, false
	}
	fmt.Println("Enter Y:")
	var y int
	fmt.Scanln(&y)
	fmt.Println("Enter Z:")
	var z int
	fmt.Scanln(&z)

	if x < 0 || y < 0 || z < 0 {
		fmt.Println("Coordinates must be non-negative.")
		return [3]uint8{}, false
	}

	slot := [3]uint8{uint8(x), uint8(y), uint8(z)}
	occupied := occupiedFoodSlots(placed)
	if !canFitFood(occupied, cap, f, slot[0], slot[1], slot[2]) {
		fmt.Printf("Cannot place %s at (%d,%d,%d) — slot occupied or out of bounds.\n", f.Name, x, y, z)
		return [3]uint8{}, false
	}
	return slot, true
}

// findFridges returns indices into char.CurrentHome.RoomItems for all placed items with Storage.
func findFridges(char *model.Character) []int {
	var result []int
	for i, placed := range char.CurrentHome.RoomItems {
		if placed.Item.Storage != nil {
			result = append(result, i)
		}
	}
	return result
}

// occupiedFoodSlots builds a set of all XYZ slots currently occupied inside a placed storage item.
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

// usedSlotCount returns the total number of slots consumed by all stored food.
func usedSlotCount(placed *model.PlacedRoomItem) int {
	return len(occupiedFoodSlots(placed))
}

// canFitFood checks whether a food item fits at anchor (ax, ay, az) inside the storage capacity.
func canFitFood(occupied map[[3]uint8]bool, cap *model.StorageCapacity, f model.Food, ax, ay, az uint8) bool {
	if uint8(ax)+f.Width > cap.Width || uint8(ay)+f.Length > cap.Length || uint8(az)+f.Height > cap.Height {
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

// nextFreeSlot finds the first anchor position where the food fits inside the storage.
// Returns (slot, true) if found, or ([0,0,0], false) if no space.
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

func promptPlacement(char *model.Character, item model.RoomItem) (model.PlacedRoomItem, error) {
	fmt.Printf("\n--- Place %s (size %dx%d height %d) ---\n", item.Name, item.Width, item.Length, item.Height)
	printRoom(char.CurrentHome)

	fmt.Println("Enter X position:")
	var x int
	fmt.Scanln(&x)

	fmt.Println("Enter Y position:")
	var y int
	fmt.Scanln(&y)

	fmt.Printf("Enter Z position (0 = floor, room max height: %d):\n", char.CurrentHome.Type.MaxHeight)
	var z int
	fmt.Scanln(&z)

	fmt.Println("Choose direction:")
	fmt.Println("1. North")
	fmt.Println("2. East")
	fmt.Println("3. South")
	fmt.Println("4. West")
	var dirChoice int
	fmt.Scanln(&dirChoice)

	var dir model.Direction
	switch dirChoice {
	case 1:
		dir = model.North
	case 2:
		dir = model.East
	case 3:
		dir = model.South
	case 4:
		dir = model.West
	default:
		return model.PlacedRoomItem{}, fmt.Errorf("invalid direction")
	}

	if x < 0 || y < 0 || z < 0 {
		return model.PlacedRoomItem{}, fmt.Errorf("coordinates must be non-negative")
	}

	if err := canPlace(char.CurrentHome, item, uint8(x), uint8(y), uint8(z), dir); err != nil {
		return model.PlacedRoomItem{}, err
	}

	return model.PlacedRoomItem{
		X:         uint8(x),
		Y:         uint8(y),
		Z:         uint8(z),
		Direction: dir,
		Item:      item,
	}, nil
}
