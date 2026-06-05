package storage

import "github.com/kamil5b/basic-life-sim/internal/model"

// ── Cupboards (floor-standing cabinets with food storage) ────────────────────

var BasicCupboard = model.RoomItem{
	Name:          "Basic Cupboard",
	Type:          model.RoomItemStorage,
	Category:      model.CategoryLow,
	WillBlockPath: true,
	NeedClearance: true,
	Width:         1, Length: 2, Height: 3,
	BasePrice: 80,
	Storage: &model.StorageCapacity{
		Width: 1, Length: 2, Height: 2,
		ExpiryMultiplier: 1.0,
	},
	Actions:  []string{"open", "store", "take out"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {},
}

var DoubleCupboard = model.RoomItem{
	Name:          "Double Cupboard",
	Type:          model.RoomItemStorage,
	Category:      model.CategoryMedium,
	WillBlockPath: true,
	NeedClearance: true,
	Width:         2, Length: 2, Height: 3,
	BasePrice: 150,
	Storage: &model.StorageCapacity{
		Width: 2, Length: 2, Height: 2,
		ExpiryMultiplier: 1.0,
	},
	Actions:  []string{"open", "store", "take out"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {},
}

var TallPantry = model.RoomItem{
	Name:          "Tall Pantry",
	Type:          model.RoomItemStorage,
	Category:      model.CategoryMediumHigh,
	WillBlockPath: true,
	NeedClearance: true,
	Width:         1, Length: 2, Height: 5,
	BasePrice: 250,
	Storage: &model.StorageCapacity{
		Width: 1, Length: 2, Height: 4,
		ExpiryMultiplier: 1.0,
	},
	Actions:  []string{"open", "store", "take out"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {},
}

// ── Wall Shelves (CanOverhang — place at any Z without floor support) ─────────

var WallShelfSmall = model.RoomItem{
	Name:          "Wall Shelf (S)",
	Type:          model.RoomItemStorage,
	Category:      model.CategoryLow,
	CanOverhang:   true,
	NeedClearance: false,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 25,
	Storage: &model.StorageCapacity{
		Width: 1, Length: 1, Height: 1,
		ExpiryMultiplier: 1.0,
	},
	Actions:  []string{"store", "take out"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {},
}

var WallShelfMedium = model.RoomItem{
	Name:          "Wall Shelf (M)",
	Type:          model.RoomItemStorage,
	Category:      model.CategoryLow,
	CanOverhang:   true,
	NeedClearance: false,
	Width:         2, Length: 1, Height: 1,
	BasePrice: 40,
	Storage: &model.StorageCapacity{
		Width: 2, Length: 1, Height: 1,
		ExpiryMultiplier: 1.0,
	},
	Actions:  []string{"store", "take out"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {},
}

var WallShelfLong = model.RoomItem{
	Name:          "Wall Shelf (L)",
	Type:          model.RoomItemStorage,
	Category:      model.CategoryMediumLow,
	CanOverhang:   true,
	NeedClearance: false,
	Width:         3, Length: 1, Height: 1,
	BasePrice: 60,
	Storage: &model.StorageCapacity{
		Width: 3, Length: 1, Height: 1,
		ExpiryMultiplier: 1.0,
	},
	Actions:  []string{"store", "take out"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {},
}

// ── Racks (free-standing multi-level shelving) ────────────────────────────────

var BasicRack = model.RoomItem{
	Name:          "Basic Rack",
	Type:          model.RoomItemStorage,
	Category:      model.CategoryLow,
	WillBlockPath: true,
	NeedClearance: true,
	Width:         1, Length: 2, Height: 4,
	BasePrice: 60,
	Storage: &model.StorageCapacity{
		Width: 1, Length: 2, Height: 3,
		ExpiryMultiplier: 1.0,
	},
	Actions:  []string{"store", "take out"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {},
}

var WideRack = model.RoomItem{
	Name:          "Wide Rack",
	Type:          model.RoomItemStorage,
	Category:      model.CategoryMedium,
	WillBlockPath: true,
	NeedClearance: true,
	Width:         2, Length: 2, Height: 5,
	BasePrice: 120,
	Storage: &model.StorageCapacity{
		Width: 2, Length: 2, Height: 4,
		ExpiryMultiplier: 1.0,
	},
	Actions:  []string{"store", "take out"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {},
}

// ── Shelf Desk (office desk with overhead shelving) ───────────────────────────

var ShelfDesk = model.RoomItem{
	Name:          "Shelf Desk",
	Type:          model.RoomItemStorage,
	Category:      model.CategoryMedium,
	WillBlockPath: true,
	NeedClearance: true,
	Width:         2, Length: 2, Height: 3,
	BasePrice:    180,
	UtilitySlots: 2, // desk surface can hold small utilities
	Storage: &model.StorageCapacity{
		Width: 1, Length: 2, Height: 1,
		ExpiryMultiplier: 1.0,
	},
	Actions: []string{"work", "store", "take out", "place utility"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		if input == "work" {
			char.Energy.Current = safeSub16(char.Energy.Current, 3)
		}
	},
}

var RackDesk = model.RoomItem{
	Name:          "Rack Desk",
	Type:          model.RoomItemStorage,
	Category:      model.CategoryMediumHigh,
	WillBlockPath: true,
	NeedClearance: true,
	Width:         3, Length: 2, Height: 4,
	BasePrice:    280,
	UtilitySlots: 3,
	Storage: &model.StorageCapacity{
		Width: 2, Length: 2, Height: 1,
		ExpiryMultiplier: 1.0,
	},
	Actions: []string{"work", "store", "take out", "place utility"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		if input == "work" {
			char.Energy.Current = safeSub16(char.Energy.Current, 3)
		}
	},
}

func safeSub16(v, delta uint16) uint16 {
	if delta > v {
		return 0
	}
	return v - delta
}
