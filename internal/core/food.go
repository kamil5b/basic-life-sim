package core

import (
	"basic-life-sim/internal/model"
	"fmt"
)

// eatFromFridge handles the "eat" action on a placed fridge.
// Lists stored food with expiry status, lets the player pick one, consumes a use,
// removes the item when uses hit 0, and refuses expired food.
func eatFromFridge(placed *model.PlacedRoomItem, char *model.Character) {
	if len(placed.Stored) == 0 {
		fmt.Println("The refrigerator is empty.")
		return
	}

	cy, cm, cd := char.CurrentDate.Unpack()
	fmt.Printf("\n--- %s contents --- [Date: %04d-%02d-%02d]\n", placed.Item.Name, cy, cm, cd)

	for i, s := range placed.Stored {
		expiry := model.ExpiryDate(s.PurchaseDate, s.Food.BaseExpiryDays, s.MultiplierUsed)
		expired := model.IsExpired(s.PurchaseDate, s.Food.BaseExpiryDays, s.MultiplierUsed, char.CurrentDate)
		daysLeft := model.DaysBetween(char.CurrentDate, expiry)
		ey, em, ed := expiry.Unpack()

		zoneNote := ""
		if s.MultiplierUsed > placed.Item.Storage.ExpiryMultiplier {
			zoneNote = fmt.Sprintf(" [cold zone %.0fx]", s.MultiplierUsed)
		}
		status := fmt.Sprintf("expires %04d-%02d-%02d (%d days left)%s", ey, em, ed, daysLeft, zoneNote)
		if expired {
			status = fmt.Sprintf("EXPIRED on %04d-%02d-%02d", ey, em, ed)
		}
		fmt.Printf("%d. %-18s uses left: %d  slot(%d,%d,%d)  %s\n",
			i+1, s.Food.Name, s.UsesRemaining, s.SlotX, s.SlotY, s.SlotZ, status)
	}
	fmt.Println("0. Back")

	var choice int
	fmt.Scanln(&choice)
	if choice == 0 {
		return
	}
	if choice < 1 || choice > len(placed.Stored) {
		fmt.Println("Invalid choice.")
		return
	}

	idx := choice - 1
	s := &placed.Stored[idx]

	if model.IsExpired(s.PurchaseDate, s.Food.BaseExpiryDays, s.MultiplierUsed, char.CurrentDate) {
		fmt.Printf("%s is expired and cannot be eaten.\n", s.Food.Name)
		return
	}

	s.Food.EatAction(&char.CurrentStats, char)
	s.UsesRemaining--
	fmt.Printf("You eat %s. Uses remaining: %d\n", s.Food.Name, s.UsesRemaining)

	if s.UsesRemaining == 0 {
		placed.Stored = append(placed.Stored[:idx], placed.Stored[idx+1:]...)
		fmt.Printf("%s is fully consumed and removed from the fridge.\n", s.Food.Name)
	}
}

// eatFromFloor handles eating food placed directly in the room (no fridge).
func eatFromFloor(char *model.Character) {
	if len(char.CurrentHome.FloorFood) == 0 {
		fmt.Println("No food on the floor.")
		return
	}

	cy, cm, cd := char.CurrentDate.Unpack()
	fmt.Printf("\n--- Floor food --- [Date: %04d-%02d-%02d]\n", cy, cm, cd)

	for i, f := range char.CurrentHome.FloorFood {
		expiry := model.ExpiryDate(f.PurchaseDate, f.Food.BaseExpiryDays, 1)
		expired := model.IsExpired(f.PurchaseDate, f.Food.BaseExpiryDays, 1, char.CurrentDate)
		daysLeft := model.DaysBetween(char.CurrentDate, expiry)
		ey, em, ed := expiry.Unpack()

		status := fmt.Sprintf("expires %04d-%02d-%02d (%d days left)", ey, em, ed, daysLeft)
		if expired {
			status = fmt.Sprintf("EXPIRED on %04d-%02d-%02d", ey, em, ed)
		}
		fmt.Printf("%d. %-18s uses left: %d  %s\n", i+1, f.Food.Name, f.UsesRemaining, status)
	}
	fmt.Println("0. Back")

	var choice int
	fmt.Scanln(&choice)
	if choice == 0 {
		return
	}
	if choice < 1 || choice > len(char.CurrentHome.FloorFood) {
		fmt.Println("Invalid choice.")
		return
	}

	idx := choice - 1
	f := &char.CurrentHome.FloorFood[idx]

	if model.IsExpired(f.PurchaseDate, f.Food.BaseExpiryDays, 1, char.CurrentDate) {
		fmt.Printf("%s is expired and cannot be eaten.\n", f.Food.Name)
		return
	}

	f.Food.EatAction(&char.CurrentStats, char)
	f.UsesRemaining--
	fmt.Printf("You eat %s. Uses remaining: %d\n", f.Food.Name, f.UsesRemaining)

	if f.UsesRemaining == 0 {
		char.CurrentHome.FloorFood = append(char.CurrentHome.FloorFood[:idx], char.CurrentHome.FloorFood[idx+1:]...)
		fmt.Printf("%s is fully consumed and removed.\n", f.Food.Name)
	}
}
