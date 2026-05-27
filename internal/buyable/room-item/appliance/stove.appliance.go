package appliance

import "basic-life-sim/internal/model"

var SingleBurnerStove = model.RoomItem{
	Name:          "Single Burner Stove",
	Type:          "Appliance",
	Category:      "Low",
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 40,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "cook":
			stat.Food = safeAdd(stat.Food, 30)
			stat.Energy = safeSub(stat.Energy, 5)
		}
	},
}

var GasStove = model.RoomItem{
	Name:          "Gas Stove",
	Type:          "Appliance",
	Category:      "Medium",
	WillBlockPath: true,
	Width:         2, Length: 1, Height: 1,
	BasePrice: 200,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "cook":
			stat.Food = safeAdd(stat.Food, 40)
			stat.Energy = safeSub(stat.Energy, 5)
		}
	},
}
