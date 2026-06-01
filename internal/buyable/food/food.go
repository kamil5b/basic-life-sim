package food

import "github.com/kamil5b/basic-life-sim/internal/model"

// --- Cooked results (defined first so raw foods can reference them) ---

var CookedEgg = model.Food{
	Name:      "Cooked Egg",
	Type:      "cooked protein",
	BasePrice: 0,
	Width:     1, Length: 1, Height: 1,
	BaseExpiryDays: 2,
	UsesTotal:      1,
	Temperature:    70,
	CanBeCooked:    false,
	CanBeMixed:     true,
	EatAction: func(stat *model.Stats, char *model.Character) {
		char.Food.Current = safeAdd(char.Food.Current, 12)
		char.Energy.Current = safeAdd(char.Energy.Current, 5)
	},
}

var CookedChickenBreast = model.Food{
	Name:      "Cooked Chicken Breast",
	Type:      "cooked protein",
	BasePrice: 0,
	Width:     1, Length: 2, Height: 1,
	BaseExpiryDays: 4,
	UsesTotal:      2,
	Temperature:    75,
	CanBeCooked:    false,
	CanBeMixed:     false,
	EatAction: func(stat *model.Stats, char *model.Character) {
		char.Food.Current = safeAdd(char.Food.Current, 20)
		char.Strength.Current = safeAdd(char.Strength.Current, 5)
		char.Energy.Current = safeAdd(char.Energy.Current, 8)
	},
}

var CookedRice = model.Food{
	Name:      "Cooked Rice",
	Type:      "cooked carb",
	BasePrice: 0,
	Width:     1, Length: 1, Height: 2,
	BaseExpiryDays: 3,
	UsesTotal:      6,
	Temperature:    80,
	CanBeCooked:    false,
	CanBeMixed:     true,
	EatAction: func(stat *model.Stats, char *model.Character) {
		char.Food.Current = safeAdd(char.Food.Current, 25)
		char.Energy.Current = safeAdd(char.Energy.Current, 10)
	},
}

var CookedBroccoli = model.Food{
	Name:      "Cooked Broccoli",
	Type:      "cooked vegetable",
	BasePrice: 0,
	Width:     1, Length: 1, Height: 2,
	BaseExpiryDays: 3,
	UsesTotal:      2,
	Temperature:    70,
	CanBeCooked:    false,
	CanBeMixed:     true,
	EatAction: func(stat *model.Stats, char *model.Character) {
		char.Food.Current = safeAdd(char.Food.Current, 12)
		char.Confidence.Current = safeAdd(char.Confidence.Current, 3)
		char.Energy.Current = safeAdd(char.Energy.Current, 4)
	},
}

// --- Raw foods ---

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
	CookedResult:   &CookedEgg,
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
	CookedResult:   &CookedChickenBreast,
	// raw chicken is unsafe to eat — EatAction is nil (penalises stats)
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
	CookedResult:   &CookedRice,
	// raw rice is inedible
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
	CookedResult:   &CookedBroccoli,
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
