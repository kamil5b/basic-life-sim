package furniture

import "basic-life-sim/internal/model"

var BasicChair = model.RoomItem{
	Name:          "Basic Chair",
	Type:          model.RoomItemFurniture,
	Category:      model.CategoryLow,
	WillBlockPath: false,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 30,
	Actions:   []string{"sit"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "sit":
			char.Energy.Current = safeAdd(char.Energy.Current, 5)
		}
	},
}

var GamingChair = model.RoomItem{
	Name:          "Gaming Chair",
	Type:          model.RoomItemFurniture,
	Category:      model.CategoryMediumHigh,
	WillBlockPath: false,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 250,
	Actions:   []string{"sit"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "sit":
			char.Energy.Current = safeAdd(char.Energy.Current, 10)
			char.Confidence.Current = safeAdd(char.Confidence.Current, 3)
		}
	},
}
