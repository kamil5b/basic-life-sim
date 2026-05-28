package appliance

import "basic-life-sim/internal/model"

var SmallTV = model.RoomItem{
	Name:          "Small TV",
	Type:          model.RoomItemAppliance,
	Category:      model.CategoryLow,
	WillBlockPath: true,
	Width:         2, Length: 1, Height: 1,
	BasePrice: 100,
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "watch":
			char.Energy.Current = safeAdd(char.Energy.Current, 10)
			char.Confidence.Current = safeSub(char.Confidence.Current, 5)
		}
	},
}

var SmartTV = model.RoomItem{
	Name:          "Smart TV",
	Type:          model.RoomItemAppliance,
	Category:      model.CategoryMediumHigh,
	WillBlockPath: true,
	Width:         3, Length: 1, Height: 1,
	BasePrice: 500,
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "watch":
			char.Energy.Current = safeAdd(char.Energy.Current, 15)
		case "stream":
			char.Confidence.Current = safeAdd(char.Confidence.Current, 5)
		}
	},
}
