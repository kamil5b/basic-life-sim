package furniture

import "basic-life-sim/internal/model"

var BasicDesk = model.RoomItem{
	Name:          "Basic Desk",
	Type:          "Furniture",
	Category:      "Low",
	WillBlockPath: true,
	Width:         2, Length: 1, Height: 2,
	BasePrice: 80,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "study":
			stat.Confidence = safeAdd(stat.Confidence, 5)
		}
	},
}

var ComputerDesk = model.RoomItem{
	Name:          "Computer Desk",
	Type:          "Furniture",
	Category:      "Medium-High",
	WillBlockPath: true,
	Width:         2, Length: 1, Height: 2,
	BasePrice: 300,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "study":
			stat.Confidence = safeAdd(stat.Confidence, 8)
		case "work":
			stat.Confidence = safeAdd(stat.Confidence, 10)
		}
	},
}
