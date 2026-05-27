package hygiene

import "basic-life-sim/internal/model"

var BasicShower = model.RoomItem{
	Name:          "Basic Shower",
	Type:          "Hygiene",
	Category:      "Low",
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 2,
	BasePrice: 150,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "shower":
			stat.Hygiene = safeAdd(stat.Hygiene, 30)
			stat.Energy = safeAdd(stat.Energy, 5)
		}
	},
}

var PremiumShower = model.RoomItem{
	Name:          "Premium Shower",
	Type:          "Hygiene",
	Category:      "High",
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 2,
	BasePrice: 600,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "shower":
			stat.Hygiene = safeAdd(stat.Hygiene, 40)
			stat.Energy = safeAdd(stat.Energy, 10)
			stat.Confidence = safeAdd(stat.Confidence, 5)
		}
	},
}
