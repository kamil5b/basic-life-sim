package furniture

import "basic-life-sim/internal/model"

var SingleBed = model.RoomItem{
	Name:          "Single Bed",
	Type:          "Furniture",
	Category:      "Low",
	WillBlockPath: true,
	Width:         1, Length: 2, Height: 1,
	BasePrice: 150,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "sleep":
			stat.Energy = safeAdd(stat.Energy, 40)
		}
	},
}

var QueenBed = model.RoomItem{
	Name:          "Queen Bed",
	Type:          "Furniture",
	Category:      "Medium",
	WillBlockPath: true,
	Width:         2, Length: 2, Height: 1,
	BasePrice: 350,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "sleep":
			stat.Energy = safeAdd(stat.Energy, 50)
		}
	},
}

var KingBed = model.RoomItem{
	Name:          "King Bed",
	Type:          "Furniture",
	Category:      "High",
	WillBlockPath: true,
	Width:         2, Length: 2, Height: 1,
	BasePrice: 700,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "sleep":
			stat.Energy = safeAdd(stat.Energy, 60)
			stat.Confidence = safeAdd(stat.Confidence, 5)
		}
	},
}
