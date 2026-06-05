package utility

import "github.com/kamil5b/basic-life-sim/internal/model"

var KitchenKnife = model.RoomItem{
	Name:     "Kitchen Knife",
	Type:     model.RoomItemUtility,
	Category: model.CategoryLow,
	Width:    1, Length: 1, Height: 1,
	BasePrice:      15,
	UtilityAbility: model.ProcessChop,
	Actions:        []string{"process"},
	DoAction:       func(input string, stat *model.Stats, char *model.Character) {},
}

var CuttingBoard = model.RoomItem{
	Name:     "Cutting Board",
	Type:     model.RoomItemUtility,
	Category: model.CategoryLow,
	Width:    1, Length: 2, Height: 1,
	BasePrice:    10,
	UtilitySlots: 1, // can hold a knife on top
	Actions:      []string{"process"},
	DoAction:     func(input string, stat *model.Stats, char *model.Character) {},
}

var FryingPan = model.RoomItem{
	Name:     "Frying Pan",
	Type:     model.RoomItemUtility,
	Category: model.CategoryLow,
	Width:    1, Length: 1, Height: 1,
	BasePrice:   25,
	CookSurface: &model.CookCapacity{Slots: 3},
	Actions:     []string{"place food", "take out"},
	DoAction:    func(input string, stat *model.Stats, char *model.Character) {},
}

var Wok = model.RoomItem{
	Name:     "Wok",
	Type:     model.RoomItemUtility,
	Category: model.CategoryMedium,
	Width:    1, Length: 1, Height: 2,
	BasePrice:   35,
	CookSurface: &model.CookCapacity{Slots: 5},
	Actions:     []string{"place food", "take out"},
	DoAction:    func(input string, stat *model.Stats, char *model.Character) {},
}

var Saucepan = model.RoomItem{
	Name:     "Saucepan",
	Type:     model.RoomItemUtility,
	Category: model.CategoryLow,
	Width:    1, Length: 1, Height: 2,
	BasePrice:   20,
	CookSurface: &model.CookCapacity{Slots: 2},
	Actions:     []string{"place food", "take out"},
	DoAction:    func(input string, stat *model.Stats, char *model.Character) {},
}
