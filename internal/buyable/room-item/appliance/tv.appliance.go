package appliance

import "basic-life-sim/internal/model"

var SmallTV = model.RoomItem{
	Name:          "Small TV",
	Type:          "Appliance",
	Category:      "Low",
	WillBlockPath: true,
	Width:         2, Length: 1, Height: 1,
	BasePrice: 100,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "watch":
			stat.Energy = safeAdd(stat.Energy, 10)
			stat.Confidence = safeSub(stat.Confidence, 5)
		}
	},
}

var SmartTV = model.RoomItem{
	Name:          "Smart TV",
	Type:          "Appliance",
	Category:      "Medium-High",
	WillBlockPath: true,
	Width:         3, Length: 1, Height: 1,
	BasePrice: 500,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "watch":
			stat.Energy = safeAdd(stat.Energy, 15)
		case "stream":
			stat.Confidence = safeAdd(stat.Confidence, 5)
		}
	},
}
