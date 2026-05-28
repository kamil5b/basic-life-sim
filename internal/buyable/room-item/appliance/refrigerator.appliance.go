package appliance

import "basic-life-sim/internal/model"

var MiniRefrigerator = model.RoomItem{
	Name:          "Mini Refrigerator",
	Type:          model.RoomItemAppliance,
	Category:      model.CategoryLow,
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 2,
	BasePrice: 120,
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "eat":
			char.Food.Current = safeAdd(char.Food.Current, 15)
		}
	},
}

var StandardRefrigerator = model.RoomItem{
	Name:          "Standard Refrigerator",
	Type:          model.RoomItemAppliance,
	Category:      model.CategoryMedium,
	WillBlockPath: true,
	Width:         1, Length: 2, Height: 2,
	BasePrice: 350,
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "eat":
			char.Food.Current = safeAdd(char.Food.Current, 25)
		}
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
