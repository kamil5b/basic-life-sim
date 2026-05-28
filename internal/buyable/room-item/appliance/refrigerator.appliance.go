package appliance

import "basic-life-sim/internal/model"

var MiniRefrigerator = model.RoomItem{
	Name:          "Mini Refrigerator",
	Type:          model.RoomItemAppliance,
	Category:      model.CategoryLow,
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 2,
	BasePrice: 120,
	// 2x2x3 = 12 slots; uniform 3x expiry, no cold zone
	Storage: &model.StorageCapacity{
		Width: 2, Length: 2, Height: 3,
		ExpiryMultiplier: 3.0,
	},
	Actions: []string{"eat"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		// "eat" is handled by eatFromFridge in core
	},
}

var StandardRefrigerator = model.RoomItem{
	Name:          "Standard Refrigerator",
	Type:          model.RoomItemAppliance,
	Category:      model.CategoryMedium,
	WillBlockPath: true,
	Width:         1, Length: 2, Height: 2,
	BasePrice: 350,
	// 3x3x4 = 36 slots; base 3x expiry everywhere.
	// Cold zone: bottom layer (z=0), full 3x3 footprint — 7x expiry (dedicated cold storage drawer).
	Storage: &model.StorageCapacity{
		Width: 3, Length: 3, Height: 4,
		ExpiryMultiplier: 3.0,
		ColdZones: []model.ColdZone{
			{
				OriginX: 0, OriginY: 0, OriginZ: 0,
				Width: 3, Length: 3, Height: 1,
				ExpiryMultiplier: 7.0,
			},
		},
	},
	Actions: []string{"eat"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		// "eat" is handled by eatFromFridge in core
	},
}

func safeAdd(v, delta uint16) uint16 {
	if v+delta < v {
		return ^uint16(0)
	}
	return v + delta
}

func safeSub(v, delta uint16) uint16 {
	if delta > v {
		return 0
	}
	return v - delta
}
