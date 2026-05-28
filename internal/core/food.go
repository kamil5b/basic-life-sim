package core

import (
	"basic-life-sim/internal/model"
	"fmt"
)

// ─── helpers ────────────────────────────────────────────────────────────────

func foodExpiredPenalty(char *model.Character) {
	const pen = 10
	char.Food.Current = safeSub16(char.Food.Current, pen)
	char.Energy.Current = safeSub16(char.Energy.Current, pen)
	char.Hygiene.Current = safeSub16(char.Hygiene.Current, pen)
	char.Confidence.Current = safeSub16(char.Confidence.Current, pen)
	char.Strength.Current = safeSub16(char.Strength.Current, pen)
	fmt.Println("You feel terrible. All stats decreased.")
}

func rawFoodPenalty(char *model.Character, name string) {
	const pen = 5
	char.Food.Current = safeSub16(char.Food.Current, pen)
	char.Energy.Current = safeSub16(char.Energy.Current, pen)
	char.Hygiene.Current = safeSub16(char.Hygiene.Current, pen)
	fmt.Printf("Eating raw %s was a bad idea. Food, Energy and Hygiene decreased.\n", name)
}

func safeSub16(v, delta uint16) uint16 {
	if delta > v {
		return 0
	}
	return v - delta
}

// collectAllFood returns a flat list of all food the character currently has access to:
// floor food, fridge stored food, and stove surface food.
// Returns (label, purchaseDate, multiplierUsed, usesRemaining, isExpired, sourceTag)
// sourceTag is used by takeOutFood / eatFood to know where to look.
type foodSource struct {
	label         string
	purchaseDate  model.CompactDate
	multiplier    float32
	usesRemaining uint8
	cooked        bool // only relevant for surface food
	food          model.Food
	// location
	kind    string // "floor", "fridge", "surface"
	itemIdx int    // index in RoomItems (for fridge/surface)
	foodIdx int    // index within Stored or OnSurface or FloorFood
}

func gatherFoodSources(char *model.Character) []foodSource {
	var sources []foodSource

	// floor food
	for i, f := range char.CurrentHome.FloorFood {
		sources = append(sources, foodSource{
			label: f.Food.Name, purchaseDate: f.PurchaseDate,
			multiplier: 1, usesRemaining: f.UsesRemaining,
			food: f.Food, kind: "floor", foodIdx: i,
		})
	}
	// fridge stored
	for ri, placed := range char.CurrentHome.RoomItems {
		if placed.Item.Storage == nil {
			continue
		}
		for fi, s := range placed.Stored {
			sources = append(sources, foodSource{
				label: s.Food.Name, purchaseDate: s.PurchaseDate,
				multiplier: s.MultiplierUsed, usesRemaining: s.UsesRemaining,
				food: s.Food, kind: "fridge", itemIdx: ri, foodIdx: fi,
			})
		}
	}
	// cooking surface
	for ri, placed := range char.CurrentHome.RoomItems {
		if placed.Item.CookSurface == nil {
			continue
		}
		for fi, s := range placed.OnSurface {
			sources = append(sources, foodSource{
				label: s.Food.Name, purchaseDate: s.PurchaseDate,
				multiplier: 1, usesRemaining: s.UsesRemaining,
				cooked: s.Cooked,
				food:   s.Food, kind: "surface", itemIdx: ri, foodIdx: fi,
			})
		}
	}
	return sources
}

// printFoodSources prints a numbered list of food sources with expiry status.
func printFoodSources(sources []foodSource, currentDate model.CompactDate) {
	for i, s := range sources {
		expiry := model.ExpiryDate(s.purchaseDate, s.food.BaseExpiryDays, s.multiplier)
		expired := model.IsExpired(s.purchaseDate, s.food.BaseExpiryDays, s.multiplier, currentDate)
		daysLeft := model.DaysBetween(currentDate, expiry)
		ey, em, ed := expiry.Unpack()

		loc := s.kind
		if s.kind == "surface" && s.cooked {
			loc = "surface (cooked)"
		}

		status := fmt.Sprintf("expires %04d-%02d-%02d (%d days)", ey, em, ed, daysLeft)
		if expired {
			status = fmt.Sprintf("EXPIRED %04d-%02d-%02d", ey, em, ed)
		}
		edible := ""
		if s.food.EatAction == nil {
			edible = " [inedible raw]"
		}
		fmt.Printf("%d. %-22s uses:%d  [%s]%s  %s\n",
			i+1, s.label, s.usesRemaining, loc, edible, status)
	}
}

// removeFoodSource removes the food at index from its source, decrementing uses first.
// If uses reach 0, the item is deleted from its container.
func consumeUse(char *model.Character, s foodSource) {
	switch s.kind {
	case "floor":
		f := &char.CurrentHome.FloorFood[s.foodIdx]
		f.UsesRemaining--
		if f.UsesRemaining == 0 {
			char.CurrentHome.FloorFood = append(char.CurrentHome.FloorFood[:s.foodIdx], char.CurrentHome.FloorFood[s.foodIdx+1:]...)
		}
	case "fridge":
		stored := &char.CurrentHome.RoomItems[s.itemIdx].Stored[s.foodIdx]
		stored.UsesRemaining--
		if stored.UsesRemaining == 0 {
			ri := &char.CurrentHome.RoomItems[s.itemIdx]
			ri.Stored = append(ri.Stored[:s.foodIdx], ri.Stored[s.foodIdx+1:]...)
		}
	case "surface":
		sf := &char.CurrentHome.RoomItems[s.itemIdx].OnSurface[s.foodIdx]
		sf.UsesRemaining--
		if sf.UsesRemaining == 0 {
			ri := &char.CurrentHome.RoomItems[s.itemIdx]
			ri.OnSurface = append(ri.OnSurface[:s.foodIdx], ri.OnSurface[s.foodIdx+1:]...)
		}
	}
}

// removeFoodSourceEntirely removes a food item from its source without consuming — used for "take out".
func removeFoodSourceEntirely(char *model.Character, s foodSource) {
	switch s.kind {
	case "floor":
		char.CurrentHome.FloorFood = append(char.CurrentHome.FloorFood[:s.foodIdx], char.CurrentHome.FloorFood[s.foodIdx+1:]...)
	case "fridge":
		ri := &char.CurrentHome.RoomItems[s.itemIdx]
		ri.Stored = append(ri.Stored[:s.foodIdx], ri.Stored[s.foodIdx+1:]...)
	case "surface":
		ri := &char.CurrentHome.RoomItems[s.itemIdx]
		ri.OnSurface = append(ri.OnSurface[:s.foodIdx], ri.OnSurface[s.foodIdx+1:]...)
	}
}

// storeFoodInFridge lets the player pick floor/surface food and manually place it into a fridge slot.
func storeFoodInFridge(placed *model.PlacedRoomItem, char *model.Character) {
	cap := placed.Item.Storage

	// gather food not already inside a fridge
	var sources []foodSource
	for _, s := range gatherFoodSources(char) {
		if s.kind != "fridge" {
			sources = append(sources, s)
		}
	}
	if len(sources) == 0 {
		fmt.Println("No food outside the fridge to store.")
		return
	}

	used := usedSlotCount(placed)
	total := cap.TotalSlots()
	if used >= total {
		fmt.Println("Refrigerator is full.")
		return
	}

	cy, cm, cd := char.CurrentDate.Unpack()
	fmt.Printf("\n--- Store food in %s (%d/%d slots used) --- [%04d-%02d-%02d]\n",
		placed.Item.Name, used, total, cy, cm, cd)
	printFoodSources(sources, char.CurrentDate)
	fmt.Println("0. Cancel")

	var choice int
	fmt.Scanln(&choice)
	if choice == 0 {
		return
	}
	if choice < 1 || choice > len(sources) {
		fmt.Println("Invalid choice.")
		return
	}

	src := sources[choice-1]
	slot, ok := promptFridgeSlot(placed, cap, src.food)
	if !ok {
		return
	}

	mult := cap.MultiplierAt(slot[0], slot[1], slot[2])
	removeFoodSourceEntirely(char, src)
	placed.Stored = append(placed.Stored, model.StoredFood{
		SlotX:          slot[0],
		SlotY:          slot[1],
		SlotZ:          slot[2],
		PurchaseDate:   src.purchaseDate,
		MultiplierUsed: mult,
		UsesRemaining:  src.usesRemaining,
		Food:           src.food,
	})
	expiry := model.ExpiryDate(src.purchaseDate, src.food.BaseExpiryDays, mult)
	ey, em, ed := expiry.Unpack()
	zoneNote := ""
	if mult > cap.ExpiryMultiplier {
		zoneNote = fmt.Sprintf(" [cold zone %.0fx]", mult)
	}
	fmt.Printf("Stored %s in slot (%d,%d,%d)%s, expires %04d-%02d-%02d.\n",
		src.food.Name, slot[0], slot[1], slot[2], zoneNote, ey, em, ed)
}

// ─── fridge eat (from checkRoom → doItemAction) ─────────────────────────────

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
		edible := ""
		if s.Food.EatAction == nil {
			edible = " [inedible raw]"
		}
		fmt.Printf("%d. %-18s uses left: %d  slot(%d,%d,%d)%s  %s\n",
			i+1, s.Food.Name, s.UsesRemaining, s.SlotX, s.SlotY, s.SlotZ, edible, status)
	}
	fmt.Println("\nActions: 1. Eat  2. Take out  0. Back")
	var action int
	fmt.Scanln(&action)
	if action == 0 {
		return
	}

	fmt.Println("Choose food (0 to cancel):")
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

	switch action {
	case 1: // eat
		if model.IsExpired(s.PurchaseDate, s.Food.BaseExpiryDays, s.MultiplierUsed, char.CurrentDate) {
			fmt.Printf("%s is expired!\n", s.Food.Name)
			foodExpiredPenalty(char)
			s.UsesRemaining--
			if s.UsesRemaining == 0 {
				placed.Stored = append(placed.Stored[:idx], placed.Stored[idx+1:]...)
			}
			return
		}
		if s.Food.EatAction == nil {
			rawFoodPenalty(char, s.Food.Name)
		} else {
			s.Food.EatAction(&char.CurrentStats, char)
			fmt.Printf("You eat %s. Uses remaining: %d\n", s.Food.Name, s.UsesRemaining-1)
		}
		s.UsesRemaining--
		if s.UsesRemaining == 0 {
			placed.Stored = append(placed.Stored[:idx], placed.Stored[idx+1:]...)
			fmt.Printf("%s fully consumed.\n", s.Food.Name)
		}
	case 2: // take out → place on floor
		sf := model.PlacedFood{
			PurchaseDate:  s.PurchaseDate,
			UsesRemaining: s.UsesRemaining,
			Food:          s.Food,
		}
		placed.Stored = append(placed.Stored[:idx], placed.Stored[idx+1:]...)
		promptPlaceOnFloor(char, sf)
	}
}

// ─── floor food eat ──────────────────────────────────────────────────────────

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
		edible := ""
		if f.Food.EatAction == nil {
			edible = " [inedible raw]"
		}
		fmt.Printf("%d. %-18s uses left: %d%s  %s\n", i+1, f.Food.Name, f.UsesRemaining, edible, status)
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
		fmt.Printf("%s is expired!\n", f.Food.Name)
		foodExpiredPenalty(char)
		f.UsesRemaining--
		if f.UsesRemaining == 0 {
			char.CurrentHome.FloorFood = append(char.CurrentHome.FloorFood[:idx], char.CurrentHome.FloorFood[idx+1:]...)
		}
		return
	}

	if f.Food.EatAction == nil {
		rawFoodPenalty(char, f.Food.Name)
	} else {
		f.Food.EatAction(&char.CurrentStats, char)
		fmt.Printf("You eat %s. Uses remaining: %d\n", f.Food.Name, f.UsesRemaining-1)
	}
	f.UsesRemaining--
	if f.UsesRemaining == 0 {
		char.CurrentHome.FloorFood = append(char.CurrentHome.FloorFood[:idx], char.CurrentHome.FloorFood[idx+1:]...)
		fmt.Printf("%s fully consumed.\n", f.Food.Name)
	}
}

// ─── cooking surface actions ─────────────────────────────────────────────────

// cookSurfaceAction handles "place food", "cook", and "take out" on a stove/appliance.
func cookSurfaceAction(placed *model.PlacedRoomItem, char *model.Character, action string) {
	switch action {
	case "place food":
		placeFoodOnSurface(placed, char)
	case "cook":
		cookOnSurface(placed, char)
	case "take out":
		takeOutFromSurface(placed, char)
	}
}

// placeFoodOnSurface picks food from anywhere accessible and puts it on the cooking surface.
func placeFoodOnSurface(placed *model.PlacedRoomItem, char *model.Character) {
	cap := int(placed.Item.CookSurface.Slots)
	current := len(placed.OnSurface)
	if current >= cap {
		fmt.Printf("%s surface is full (%d/%d slots).\n", placed.Item.Name, current, cap)
		return
	}

	sources := gatherFoodSources(char)
	// filter to only cookable food not already on a surface
	var cookable []foodSource
	for _, s := range sources {
		if s.kind == "surface" {
			continue // already on a surface
		}
		if s.food.CanBeCooked || s.food.CookedResult != nil {
			cookable = append(cookable, s)
		}
	}
	if len(cookable) == 0 {
		fmt.Println("No cookable food available.")
		return
	}

	cy, cm, cd := char.CurrentDate.Unpack()
	fmt.Printf("\n--- Place food on %s (%d/%d slots) --- [%04d-%02d-%02d]\n",
		placed.Item.Name, current, cap, cy, cm, cd)
	printFoodSources(cookable, char.CurrentDate)
	fmt.Println("0. Cancel")

	var choice int
	fmt.Scanln(&choice)
	if choice == 0 {
		return
	}
	if choice < 1 || choice > len(cookable) {
		fmt.Println("Invalid choice.")
		return
	}

	src := cookable[choice-1]
	sf := model.SurfaceFood{
		PurchaseDate:  src.purchaseDate,
		UsesRemaining: src.usesRemaining,
		Food:          src.food,
	}
	removeFoodSourceEntirely(char, src)
	placed.OnSurface = append(placed.OnSurface, sf)
	fmt.Printf("%s placed on %s.\n", src.food.Name, placed.Item.Name)
}

// cookOnSurface transforms all cookable surface food into their cooked results.
func cookOnSurface(placed *model.PlacedRoomItem, char *model.Character) {
	if len(placed.OnSurface) == 0 {
		fmt.Printf("%s has nothing on it to cook.\n", placed.Item.Name)
		return
	}

	cooked := 0
	for i := range placed.OnSurface {
		sf := &placed.OnSurface[i]
		if sf.Cooked {
			fmt.Printf("%s is already cooked.\n", sf.Food.Name)
			continue
		}
		if sf.Food.CookedResult == nil {
			fmt.Printf("%s cannot be cooked.\n", sf.Food.Name)
			continue
		}
		result := *sf.Food.CookedResult
		result.UsesTotal = sf.Food.UsesTotal // preserve uses
		sf.Food = result
		sf.Cooked = true
		cooked++
		fmt.Printf("%s → %s\n", sf.Food.Name, result.Name)
	}
	// Energy cost for cooking
	char.Energy.Current = safeSub16(char.Energy.Current, 5)
	if cooked > 0 {
		fmt.Printf("Cooked %d item(s). Energy -5.\n", cooked)
	}
}

// takeOutFromSurface moves a food item off the cooking surface and places it on the floor.
func takeOutFromSurface(placed *model.PlacedRoomItem, char *model.Character) {
	if len(placed.OnSurface) == 0 {
		fmt.Printf("%s surface is empty.\n", placed.Item.Name)
		return
	}

	fmt.Printf("\n--- Take out from %s ---\n", placed.Item.Name)
	for i, sf := range placed.OnSurface {
		cookedLabel := ""
		if sf.Cooked {
			cookedLabel = " [cooked]"
		}
		fmt.Printf("%d. %s%s  uses:%d\n", i+1, sf.Food.Name, cookedLabel, sf.UsesRemaining)
	}
	fmt.Println("0. Cancel")

	var choice int
	fmt.Scanln(&choice)
	if choice == 0 {
		return
	}
	if choice < 1 || choice > len(placed.OnSurface) {
		fmt.Println("Invalid choice.")
		return
	}
	idx := choice - 1
	sf := placed.OnSurface[idx]
	placed.OnSurface = append(placed.OnSurface[:idx], placed.OnSurface[idx+1:]...)

	pf := model.PlacedFood{
		PurchaseDate:  sf.PurchaseDate,
		UsesRemaining: sf.UsesRemaining,
		Food:          sf.Food,
	}
	promptPlaceOnFloor(char, pf)
}

// promptPlaceOnFloor asks the player for X/Y/Z and places food on the floor.
func promptPlaceOnFloor(char *model.Character, pf model.PlacedFood) {
	printRoom(char.CurrentHome)
	fmt.Printf("Place %s on floor. Enter X:\n", pf.Food.Name)
	var x int
	fmt.Scanln(&x)
	fmt.Println("Enter Y:")
	var y int
	fmt.Scanln(&y)
	fmt.Println("Enter Z:")
	var z int
	fmt.Scanln(&z)
	if x < 0 || y < 0 || z < 0 {
		x, y, z = 0, 0, 0
	}
	pf.X = uint8(x)
	pf.Y = uint8(y)
	pf.Z = uint8(z)
	char.CurrentHome.FloorFood = append(char.CurrentHome.FloorFood, pf)
	fmt.Printf("%s placed on floor at (%d,%d,%d).\n", pf.Food.Name, x, y, z)
}
