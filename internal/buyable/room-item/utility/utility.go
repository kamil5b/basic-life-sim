package utility

import "github.com/kamil5b/basic-life-sim/internal/model"

var KitchenKnife = model.Utility{
	BaseItem: model.BaseItem{Name: "Kitchen Knife", BasePrice: 15, Width: 1, Length: 1, Height: 1},
	Ability:  model.ProcessChop,
	Actions:  []string{"process"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {},
}

var CuttingBoard = model.Utility{
	BaseItem:     model.BaseItem{Name: "Cutting Board", BasePrice: 10, Width: 1, Length: 2, Height: 1},
	UtilitySlots: 1,
	Actions:      []string{"process"},
	DoAction:     func(input string, stat *model.Stats, char *model.Character) {},
}

var FryingPan = model.Utility{
	BaseItem:    model.BaseItem{Name: "Frying Pan", BasePrice: 25, Width: 1, Length: 1, Height: 1},
	CookSurface: &model.CookCapacity{Slots: 3},
	Actions:     []string{"place food", "take out"},
	DoAction:    func(input string, stat *model.Stats, char *model.Character) {},
}

var Wok = model.Utility{
	BaseItem:    model.BaseItem{Name: "Wok", BasePrice: 35, Width: 1, Length: 1, Height: 2},
	CookSurface: &model.CookCapacity{Slots: 5},
	Actions:     []string{"place food", "take out"},
	DoAction:    func(input string, stat *model.Stats, char *model.Character) {},
}

var Saucepan = model.Utility{
	BaseItem:    model.BaseItem{Name: "Saucepan", BasePrice: 20, Width: 1, Length: 1, Height: 2},
	CookSurface: &model.CookCapacity{Slots: 2},
	Actions:     []string{"place food", "take out"},
	DoAction:    func(input string, stat *model.Stats, char *model.Character) {},
}

func init() {
	model.RegisterUtility(KitchenKnife, CuttingBoard, FryingPan, Wok, Saucepan)

	All = []model.Utility{
		KitchenKnife,
		CuttingBoard,
		FryingPan,
		Wok,
		Saucepan,
	}
}

var All []model.Utility
