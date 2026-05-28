package appliance

import "basic-life-sim/internal/model"

var MiniRefrigerator = model.RoomItem{
	Name:          "Mini Refrigerator",
	Type:          model.RoomItemAppliance,
	Category:      model.CategoryLow,
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 2,
	BasePrice: 120,
	// 2x2x3 = 12 slots; regular cold: 3x base expiry
	Storage: &model.StorageCapacity{Width: 2, Length: 2, Height: 3, ExpiryMultiplier: 3.0},
	Actions: []string{"eat"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		// "eat" is handled by the room's food interaction logic, not here directly
	},
}

var StandardRefrigerator = model.RoomItem{
	Name:          "Standard Refrigerator",
	Type:          model.RoomItemAppliance,
	Category:      model.CategoryMedium,
	WillBlockPath: true,
	Width:         1, Length: 2, Height: 2,
	BasePrice: 350,
	// 3x3x4 = 36 slots; cold storage: 7x base expiry
	Storage: &model.StorageCapacity{Width: 3, Length: 3, Height: 4, ExpiryMultiplier: 7.0},
	Actions: []string{"eat"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		// "eat" is handled by the room's food interaction logic, not here directly
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
