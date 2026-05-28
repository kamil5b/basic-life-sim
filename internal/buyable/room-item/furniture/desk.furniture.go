package furniture

import "basic-life-sim/internal/model"

var BasicDesk = model.RoomItem{
	Name:          "Basic Desk",
	Type:          model.RoomItemFurniture,
	Category:      model.CategoryLow,
	WillBlockPath: true,
	Width:         2, Length: 1, Height: 2,
	BasePrice: 80,
	Actions:   []string{"study"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "study":
			char.Confidence.Current = safeAdd(char.Confidence.Current, 5)
		}
	},
}

var ComputerDesk = model.RoomItem{
	Name:          "Computer Desk",
	Type:          model.RoomItemFurniture,
	Category:      model.CategoryMediumHigh,
	WillBlockPath: true,
	Width:         2, Length: 1, Height: 2,
	BasePrice: 300,
	Actions:   []string{"study", "work"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "study":
			char.Confidence.Current = safeAdd(char.Confidence.Current, 8)
		case "work":
			char.Confidence.Current = safeAdd(char.Confidence.Current, 10)
		}
	},
}
