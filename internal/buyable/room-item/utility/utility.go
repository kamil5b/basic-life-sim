package utility

import "github.com/kamil5b/basic-life-sim/internal/model"

var KitchenKnife = model.Utility{
	Name:    "Kitchen Knife",
	Ability: model.ProcessChop,
	Width:   1, Length: 1, Height: 1,
	BasePrice: 15,
	Actions:   []string{"process"},
	DoAction:  func(input string, stat *model.Stats, char *model.Character) {},
}

var CuttingBoard = model.Utility{
	Name:  "Cutting Board",
	Width: 1, Length: 2, Height: 1,
	BasePrice:    10,
	UtilitySlots: 1, // can hold a knife on top
	Actions:      []string{"process"},
	DoAction:     func(input string, stat *model.Stats, char *model.Character) {},
}

var FryingPan = model.Utility{
	Name:  "Frying Pan",
	Width: 1, Length: 1, Height: 1,
	BasePrice:   25,
	CookSurface: &model.CookCapacity{Slots: 3},
	Actions:     []string{"place food", "take out"},
	DoAction:    func(input string, stat *model.Stats, char *model.Character) {},
}

var Wok = model.Utility{
	Name:  "Wok",
	Width: 1, Length: 1, Height: 2,
	BasePrice:   35,
	CookSurface: &model.CookCapacity{Slots: 5},
	Actions:     []string{"place food", "take out"},
	DoAction:    func(input string, stat *model.Stats, char *model.Character) {},
}

var Saucepan = model.Utility{
	Name:  "Saucepan",
	Width: 1, Length: 1, Height: 2,
	BasePrice:   20,
	CookSurface: &model.CookCapacity{Slots: 2},
	Actions:     []string{"place food", "take out"},
	DoAction:    func(input string, stat *model.Stats, char *model.Character) {},
}

func init() {
	model.RegisterUtility(KitchenKnife, CuttingBoard, FryingPan, Wok, Saucepan)
}

var All = []model.Utility{
	KitchenKnife,
	CuttingBoard,
	FryingPan,
	Wok,
	Saucepan,
}
