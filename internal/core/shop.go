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
			_, fits := nextFreeSlot(placed, cap, f)
			noFit := ""
			if !fits {
				noFit = " [no space]"
			}
			fmt.Printf("%d. %-18s $%.2f  %dx%dx%d  [%s]%s%s\n",
				i+1, f.Name, f.BasePrice,
				f.Width, f.Length, f.Height,
				f.Type, affordable, noFit)
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

		slot, fits := nextFreeSlot(placed, cap, f)
		if !fits {
			fmt.Printf("%s (%dx%dx%d) does not fit in the remaining space.\n", f.Name, f.Width, f.Length, f.Height)
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
