package appliance

import "basic-life-sim/internal/model"

var MiniRefrigerator = model.RoomItem{
	Name:          "Mini Refrigerator",
	Type:          "Appliance",
	Category:      "Low",
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 2,
	BasePrice: 120,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "eat":
			stat.Food = safeAdd(stat.Food, 15)
		}
	},
}

var StandardRefrigerator = model.RoomItem{
	Name:          "Standard Refrigerator",
	Type:          "Appliance",
	Category:      "Medium",
	WillBlockPath: true,
	Width:         1, Length: 2, Height: 2,
	BasePrice: 350,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "eat":
			stat.Food = safeAdd(stat.Food, 25)
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
