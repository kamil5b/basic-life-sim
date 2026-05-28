package core

import (
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
