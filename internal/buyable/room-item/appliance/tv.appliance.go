package appliance

import "github.com/kamil5b/basic-life-sim/internal/model"

var SmallTV = model.RoomItem{
	Name:          "Small TV",
	Type:          model.RoomItemAppliance,
	Category:      model.CategoryLow,
	WillBlockPath: true,
	NeedClearance: true,
	Width:         2, Length: 1, Height: 1,
	BasePrice: 100,
	Actions:   []string{"watch"},
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
	NeedClearance: true,
	Width:         3, Length: 1, Height: 1,
	BasePrice: 500,
	Actions:   []string{"watch", "stream"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "watch":
			char.Energy.Current = safeAdd(char.Energy.Current, 15)
		case "stream":
			char.Confidence.Current = safeAdd(char.Confidence.Current, 5)
		}
	},
}
