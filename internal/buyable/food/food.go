package food

import "basic-life-sim/internal/model"

var Egg = model.Food{
	Name:      "Egg",
	Type:      "raw protein",
	BasePrice: 0.50,
	Width:     1, Length: 1, Height: 1,
	BaseExpiryDays: 21,
	UsesTotal:      1,
	Temperature:    4,
	CanBeCooked:    true,
	CanBeMixed:     true,
	EatAction: func(stat *model.Stats, char *model.Character) {
		char.Food.Current = safeAdd(char.Food.Current, 5)
	},
}

var ChickenBreast = model.Food{
	Name:      "Chicken Breast",
	Type:      "raw protein",
	BasePrice: 3.00,
	Width:     1, Length: 2, Height: 1,
	BaseExpiryDays: 3,
	UsesTotal:      2,
	Temperature:    4,
	CanBeCooked:    true,
	CanBeMixed:     false,
	EatAction: func(stat *model.Stats, char *model.Character) {
		char.Food.Current = safeAdd(char.Food.Current, 8)
		char.Strength.Current = safeAdd(char.Strength.Current, 2)
	},
}

var Apple = model.Food{
	Name:      "Apple",
	Type:      "fruit",
	BasePrice: 1.00,
	Width:     1, Length: 1, Height: 1,
	BaseExpiryDays: 7,
	UsesTotal:      1,
	Temperature:    4,
	CanBeCooked:    false,
	CanBeMixed:     true,
	EatAction: func(stat *model.Stats, char *model.Character) {
		char.Food.Current = safeAdd(char.Food.Current, 6)
		char.Energy.Current = safeAdd(char.Energy.Current, 3)
	},
}

var Milk = model.Food{
	Name:      "Milk",
	Type:      "liquid",
	BasePrice: 1.50,
	Width:     1, Length: 1, Height: 2,
	BaseExpiryDays: 5,
	UsesTotal:      4,
	Temperature:    4,
	CanBeCooked:    false,
	CanBeMixed:     true,
	EatAction: func(stat *model.Stats, char *model.Character) {
		char.Food.Current = safeAdd(char.Food.Current, 4)
		char.Strength.Current = safeAdd(char.Strength.Current, 1)
	},
}

var Rice = model.Food{
	Name:      "Rice",
	Type:      "raw carb",
	BasePrice: 2.00,
	Width:     1, Length: 1, Height: 2,
	BaseExpiryDays: 365,
	UsesTotal:      6,
	Temperature:    20,
	CanBeCooked:    true,
	CanBeMixed:     true,
	EatAction: func(stat *model.Stats, char *model.Character) {
		char.Food.Current = safeAdd(char.Food.Current, 10)
	},
}

var Broccoli = model.Food{
	Name:      "Broccoli",
	Type:      "raw vegetable",
	BasePrice: 1.20,
	Width:     1, Length: 1, Height: 2,
	BaseExpiryDays: 5,
	UsesTotal:      2,
	Temperature:    4,
	CanBeCooked:    true,
	CanBeMixed:     true,
	EatAction: func(stat *model.Stats, char *model.Character) {
		char.Food.Current = safeAdd(char.Food.Current, 5)
		char.Confidence.Current = safeAdd(char.Confidence.Current, 1)
	},
}

var All = []model.Food{
	Egg,
	ChickenBreast,
	Apple,
	Milk,
	Rice,
	Broccoli,
}

func safeAdd(v, delta uint16) uint16 {
	if v+delta < v {
		return ^uint16(0)
	}
	return v + delta
}
