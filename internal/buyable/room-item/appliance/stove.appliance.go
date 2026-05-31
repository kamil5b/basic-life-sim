package appliance

import "basic-life-sim/internal/model"

var SingleBurnerStove = model.RoomItem{
	Name:          "Single Burner Stove",
	Type:          model.RoomItemAppliance,
	Category:      model.CategoryLow,
	WillBlockPath: true,
	NeedClearance: true,
	Width:         1, Length: 1, Height: 1,
	BasePrice:   40,
	CookSurface: &model.CookCapacity{Slots: 1},
	Actions:     []string{"place food", "cook", "take out"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		// "place food", "cook", "take out" are handled by cookSurfaceAction in core
	},
}

var GasStove = model.RoomItem{
	Name:          "Gas Stove",
	Type:          model.RoomItemAppliance,
	Category:      model.CategoryMedium,
	WillBlockPath: true,
	NeedClearance: true,
	Width:         2, Length: 1, Height: 1,
	BasePrice:   200,
	CookSurface: &model.CookCapacity{Slots: 4},
	Actions:     []string{"place food", "cook", "take out"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		// "place food", "cook", "take out" are handled by cookSurfaceAction in core
	},
}
