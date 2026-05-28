package appliance

import "basic-life-sim/internal/model"

var SingleBurnerStove = model.RoomItem{
	Name:          "Single Burner Stove",
	Type:          model.RoomItemAppliance,
	Category:      model.CategoryLow,
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 40,
	Actions:   []string{"cook"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "cook":
			char.Food.Current = safeAdd(char.Food.Current, 30)
			char.Energy.Current = safeSub(char.Energy.Current, 5)
		}
	},
}

var GasStove = model.RoomItem{
	Name:          "Gas Stove",
	Type:          model.RoomItemAppliance,
	Category:      model.CategoryMedium,
	WillBlockPath: true,
	Width:         2, Length: 1, Height: 1,
	BasePrice: 200,
	Actions:   []string{"cook"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "cook":
			char.Food.Current = safeAdd(char.Food.Current, 40)
			char.Energy.Current = safeSub(char.Energy.Current, 5)
		}
	},
}
