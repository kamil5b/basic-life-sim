package hygiene

import "basic-life-sim/internal/model"

var BasicToilet = model.RoomItem{
	Name:          "Basic Toilet",
	Type:          model.RoomItemHygiene,
	Category:      model.CategoryLow,
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 60,
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "use":
			char.Hygiene.Current = safeAdd(char.Hygiene.Current, 10)
		}
	},
}

var BidetToilet = model.RoomItem{
	Name:          "Bidet Toilet",
	Type:          model.RoomItemHygiene,
	Category:      model.CategoryMediumHigh,
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 400,
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "use":
			char.Hygiene.Current = safeAdd(char.Hygiene.Current, 20)
			char.Confidence.Current = safeAdd(char.Confidence.Current, 5)
		}
	},
}
