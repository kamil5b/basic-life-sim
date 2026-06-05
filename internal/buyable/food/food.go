package food

import "github.com/kamil5b/basic-life-sim/internal/model"

// ── Group-level vars ──────────────────────────────────────────────────────────
// Raw produce
var (
	Apple   model.Food
	Banana  model.Food
	Orange  model.Food
	Tomato  model.Food
	Avocado model.Food
)

// Raw vegetables
var (
	Broccoli model.Food
	Carrot   model.Food
	Onion    model.Food
	Garlic   model.Food
	Lettuce  model.Food
)

// Raw carbs
var (
	Potato  model.Food
	Rice    model.Food
	Pasta   model.Food
	Oatmeal model.Food
)

// Raw proteins
var (
	Egg           model.Food
	ChickenBreast model.Food
	BeefMince     model.Food
	SalmonFillet  model.Food
	Tofu          model.Food
)

// Dairy
var (
	Milk   model.Food
	Cheese model.Food
	Butter model.Food
	Yogurt model.Food
)

// Pantry / liquids
var (
	CookingOil model.Food
	SoySauce   model.Food
)

// Seasonings
var (
	Salt   model.Food
	Pepper model.Food
)

// Bread
var Bread model.Food

// ── Cooked variants ───────────────────────────────────────────────────────────

var (
	CookedEgg      model.Food
	CookedChicken  model.Food
	CookedBeef     model.Food
	CookedSalmon   model.Food
	CookedBroccoli model.Food
	CookedCarrot   model.Food
	CookedPotato   model.Food
	CookedRice     model.Food
	CookedPasta    model.Food
)

// ── Mixed / Prepared dishes ───────────────────────────────────────────────────

var (
	EggFriedRice    model.Food
	ChickenRiceBowl model.Food
	Bolognese       model.Food
	SalmonBowl      model.Food
	ChickenBroccoli model.Food
	Salad           model.Food
	Omelette        model.Food
	MashedPotato    model.Food
	Toast           model.Food
	GrilledCheese   model.Food
)

// ── Processed ingredients ─────────────────────────────────────────────────────

var (
	ChoppedOnion  model.Food
	SlicedOnion   model.Food
	MincedGarlic  model.Food
	ChoppedCarrot model.Food
	SlicedCarrot  model.Food
	SlicedTomato  model.Food
	SlicedCheese  model.Food
	GratedCheese  model.Food
	RiceFlour     model.Food
)

func init() {
	// ── Part 1: Raw foods ──────────────────────────────────────────────────

	Apple = model.Food{
		Name: "Apple", Type: "fruit", BasePrice: 1.00, WeightGrams: 180,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 7, UsesTotal: 1, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 6, Energy: 3},
	}

	Banana = model.Food{
		Name: "Banana", Type: "fruit", BasePrice: 0.80, WeightGrams: 120,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 5, UsesTotal: 1, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 8, Energy: 5},
	}

	Orange = model.Food{
		Name: "Orange", Type: "fruit", BasePrice: 1.20, WeightGrams: 150,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 10, UsesTotal: 1, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 5, Energy: 4, Confidence: 1},
	}

	Tomato = model.Food{
		Name: "Tomato", Type: "fruit", BasePrice: 0.60, WeightGrams: 120,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 5, UsesTotal: 1, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 3, Energy: 2},
		ProcessResults: []model.ProcessResult{
			{Ability: model.ProcessSlice, ResultName: "Sliced Tomato"},
		},
	}

	Avocado = model.Food{
		Name: "Avocado", Type: "fruit", BasePrice: 2.00, WeightGrams: 200,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 4, UsesTotal: 1, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 10, Energy: 3, Confidence: 1},
	}

	Broccoli = model.Food{
		Name: "Broccoli", Type: "raw veg", BasePrice: 1.20, WeightGrams: 100,
		Width: 1, Length: 1, Height: 2,
		BaseExpiryDays: 5, UsesTotal: 2, Temperature: 4,
		CanBeCooked: true, CookedResult: "Cooked Broccoli", CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 5, Energy: 2, Confidence: 1},
	}

	Carrot = model.Food{
		Name: "Carrot", Type: "raw veg", BasePrice: 0.70, WeightGrams: 80,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 10, UsesTotal: 1, Temperature: 4,
		CanBeCooked: true, CookedResult: "Cooked Carrot", CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 4, Energy: 2},
		ProcessResults: []model.ProcessResult{
			{Ability: model.ProcessChop, ResultName: "Chopped Carrot"},
			{Ability: model.ProcessSlice, ResultName: "Sliced Carrot"},
		},
	}

	Onion = model.Food{
		Name: "Onion", Type: "raw veg", BasePrice: 0.50, WeightGrams: 150,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 14, UsesTotal: 1, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 3, Energy: 1},
		ProcessResults: []model.ProcessResult{
			{Ability: model.ProcessChop, ResultName: "Chopped Onion"},
			{Ability: model.ProcessSlice, ResultName: "Sliced Onion"},
		},
	}

	Garlic = model.Food{
		Name: "Garlic", Type: "raw veg", BasePrice: 0.30, WeightGrams: 10,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 30, UsesTotal: 3, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 1},
		ProcessResults: []model.ProcessResult{
			{Ability: model.ProcessGrind, ResultName: "Minced Garlic"},
		},
	}

	Lettuce = model.Food{
		Name: "Lettuce", Type: "raw veg", BasePrice: 1.50, WeightGrams: 100,
		Width: 1, Length: 1, Height: 2,
		BaseExpiryDays: 4, UsesTotal: 4, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 2, Energy: 1},
	}

	Potato = model.Food{
		Name: "Potato", Type: "raw carb", BasePrice: 0.80, WeightGrams: 200,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 14, UsesTotal: 1, Temperature: 4,
		CanBeCooked: true, CookedResult: "Cooked Potato", CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 8, Energy: 5},
	}

	Egg = model.Food{
		Name: "Egg", Type: "raw protein", BasePrice: 0.50, WeightGrams: 50,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 21, UsesTotal: 1, Temperature: 4,
		CanBeCooked: true, CookedResult: "Cooked Egg", CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 5},
	}

	ChickenBreast = model.Food{
		Name: "Chicken Breast", Type: "raw protein", BasePrice: 3.00, WeightGrams: 200,
		Width: 1, Length: 2, Height: 1,
		BaseExpiryDays: 3, UsesTotal: 2, Temperature: 4,
		CanBeCooked: true, CookedResult: "Cooked Chicken", CanBeMixed: false,
		Nutrition: model.Nutrition{}, // raw chicken is unsafe; penalty applied via OnEat=nil check
	}

	BeefMince = model.Food{
		Name: "Beef Mince", Type: "raw protein", BasePrice: 4.00, WeightGrams: 250,
		Width: 1, Length: 2, Height: 1,
		BaseExpiryDays: 2, UsesTotal: 3, Temperature: 4,
		CanBeCooked: true, CookedResult: "Cooked Beef", CanBeMixed: false,
		Nutrition: model.Nutrition{}, // raw beef unsafe
	}

	SalmonFillet = model.Food{
		Name: "Salmon Fillet", Type: "raw protein", BasePrice: 5.00, WeightGrams: 180,
		Width: 1, Length: 2, Height: 1,
		BaseExpiryDays: 2, UsesTotal: 2, Temperature: 4,
		CanBeCooked: true, CookedResult: "Cooked Salmon", CanBeMixed: false,
		Nutrition: model.Nutrition{}, // raw salmon unsafe
	}

	Tofu = model.Food{
		Name: "Tofu", Type: "raw protein", BasePrice: 2.00, WeightGrams: 200,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 7, UsesTotal: 3, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 6, Energy: 2, Strength: 1},
	}

	Milk = model.Food{
		Name: "Milk", Type: "liquid", BasePrice: 1.50, WeightGrams: 250,
		Width: 1, Length: 1, Height: 2,
		BaseExpiryDays: 5, UsesTotal: 4, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 4, Strength: 1},
	}

	Cheese = model.Food{
		Name: "Cheese", Type: "dairy", BasePrice: 3.00, WeightGrams: 40,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 14, UsesTotal: 4, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 5, Energy: 2},
		ProcessResults: []model.ProcessResult{
			{Ability: model.ProcessSlice, ResultName: "Sliced Cheese"},
			{Ability: model.ProcessGrind, ResultName: "Grated Cheese"},
		},
	}

	Butter = model.Food{
		Name: "Butter", Type: "dairy", BasePrice: 1.00, WeightGrams: 20,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 21, UsesTotal: 10, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 3, Energy: 4, Hygiene: -1},
	}

	Yogurt = model.Food{
		Name: "Yogurt", Type: "dairy", BasePrice: 1.80, WeightGrams: 150,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 7, UsesTotal: 2, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 5, Energy: 1},
	}

	Rice = model.Food{
		Name: "Rice", Type: "raw carb", BasePrice: 2.00, WeightGrams: 100,
		Width: 1, Length: 1, Height: 2,
		BaseExpiryDays: 365, UsesTotal: 6, Temperature: 20,
		CanBeCooked: true, CookedResult: "Cooked Rice", CanBeMixed: true,
		Nutrition: model.Nutrition{}, // inedible raw
		ProcessResults: []model.ProcessResult{
			{Ability: model.ProcessGrind, ResultName: "Rice Flour"},
		},
	}

	Bread = model.Food{
		Name: "Bread", Type: "carb", BasePrice: 2.50, WeightGrams: 50,
		Width: 1, Length: 2, Height: 1,
		BaseExpiryDays: 5, UsesTotal: 10, Temperature: 20,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 8, Energy: 5},
	}

	Pasta = model.Food{
		Name: "Pasta", Type: "raw carb", BasePrice: 1.50, WeightGrams: 100,
		Width: 1, Length: 1, Height: 2,
		BaseExpiryDays: 365, UsesTotal: 4, Temperature: 20,
		CanBeCooked: true, CookedResult: "Cooked Pasta", CanBeMixed: true,
		Nutrition: model.Nutrition{}, // inedible raw
	}

	Oatmeal = model.Food{
		Name: "Oatmeal", Type: "raw carb", BasePrice: 1.20, WeightGrams: 50,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 180, UsesTotal: 4, Temperature: 20,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 4, Energy: 3},
	}

	CookingOil = model.Food{
		Name: "Cooking Oil", Type: "liquid", BasePrice: 3.00, WeightGrams: 15,
		Width: 1, Length: 1, Height: 2,
		BaseExpiryDays: 180, UsesTotal: 20, Temperature: 20,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 2, Energy: 5, Hygiene: -1},
	}

	SoySauce = model.Food{
		Name: "Soy Sauce", Type: "liquid", BasePrice: 2.00, WeightGrams: 10,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 365, UsesTotal: 15, Temperature: 20,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{},
	}

	Salt = model.Food{
		Name: "Salt", Type: "seasoning", BasePrice: 1.00, WeightGrams: 5,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 9999, UsesTotal: 30, Temperature: 20,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{},
	}

	Pepper = model.Food{
		Name: "Pepper", Type: "seasoning", BasePrice: 1.50, WeightGrams: 2,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 9999, UsesTotal: 20, Temperature: 20,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{},
	}

	// ── Part 2: Cooked variants ────────────────────────────────────────────

	CookedEgg = model.Food{
		Name: "Cooked Egg", Type: "cooked protein", BasePrice: 0, WeightGrams: 45,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 2, UsesTotal: 1, Temperature: 70,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 12, Energy: 5},
		MixRecipes: []model.MixRecipe{
			{Ingredient: model.MixIngredient{FoodName: "Cooked Rice"}, ResultName: "Egg Fried Rice"},
		},
	}

	CookedChicken = model.Food{
		Name: "Cooked Chicken", Type: "cooked protein", BasePrice: 0, WeightGrams: 180,
		Width: 1, Length: 2, Height: 1,
		BaseExpiryDays: 4, UsesTotal: 2, Temperature: 75,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 20, Energy: 8, Strength: 5},
		MixRecipes: []model.MixRecipe{
			{Ingredient: model.MixIngredient{FoodName: "Cooked Rice"}, ResultName: "Chicken Rice Bowl"},
			{Ingredient: model.MixIngredient{FoodName: "Cooked Broccoli"}, ResultName: "Chicken Broccoli"},
		},
	}

	CookedBeef = model.Food{
		Name: "Cooked Beef", Type: "cooked protein", BasePrice: 0, WeightGrams: 220,
		Width: 1, Length: 2, Height: 1,
		BaseExpiryDays: 3, UsesTotal: 3, Temperature: 75,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 22, Energy: 10, Strength: 6},
		MixRecipes: []model.MixRecipe{
			{Ingredient: model.MixIngredient{FoodName: "Cooked Pasta"}, ResultName: "Bolognese"},
		},
	}

	CookedSalmon = model.Food{
		Name: "Cooked Salmon", Type: "cooked protein", BasePrice: 0, WeightGrams: 160,
		Width: 1, Length: 2, Height: 1,
		BaseExpiryDays: 3, UsesTotal: 2, Temperature: 70,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 18, Energy: 7, Confidence: 1, Strength: 3},
		MixRecipes: []model.MixRecipe{
			{Ingredient: model.MixIngredient{FoodName: "Cooked Rice"}, ResultName: "Salmon Bowl"},
		},
	}

	CookedBroccoli = model.Food{
		Name: "Cooked Broccoli", Type: "cooked vegetable", BasePrice: 0, WeightGrams: 90,
		Width: 1, Length: 1, Height: 2,
		BaseExpiryDays: 3, UsesTotal: 2, Temperature: 70,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 12, Energy: 4, Confidence: 3},
	}

	CookedCarrot = model.Food{
		Name: "Cooked Carrot", Type: "cooked vegetable", BasePrice: 0, WeightGrams: 70,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 3, UsesTotal: 1, Temperature: 70,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 8, Energy: 3, Confidence: 2},
	}

	CookedPotato = model.Food{
		Name: "Cooked Potato", Type: "cooked carb", BasePrice: 0, WeightGrams: 180,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 3, UsesTotal: 1, Temperature: 80,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 14, Energy: 8},
	}

	CookedRice = model.Food{
		Name: "Cooked Rice", Type: "cooked carb", BasePrice: 0, WeightGrams: 300,
		Width: 1, Length: 1, Height: 2,
		BaseExpiryDays: 3, UsesTotal: 6, Temperature: 80,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 25, Energy: 10},
		MixRecipes: []model.MixRecipe{
			{Ingredient: model.MixIngredient{FoodName: "Cooked Egg"}, ResultName: "Egg Fried Rice"},
			{Ingredient: model.MixIngredient{FoodName: "Cooked Chicken"}, ResultName: "Chicken Rice Bowl"},
			{Ingredient: model.MixIngredient{FoodName: "Cooked Salmon"}, ResultName: "Salmon Bowl"},
		},
	}

	CookedPasta = model.Food{
		Name: "Cooked Pasta", Type: "cooked carb", BasePrice: 0, WeightGrams: 250,
		Width: 1, Length: 1, Height: 2,
		BaseExpiryDays: 3, UsesTotal: 4, Temperature: 80,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 20, Energy: 9},
		MixRecipes: []model.MixRecipe{
			{Ingredient: model.MixIngredient{FoodName: "Cooked Beef"}, ResultName: "Bolognese"},
		},
	}

	// ── Part 3: Mixed / Prepared dishes ────────────────────────────────────

	EggFriedRice = model.Food{
		Name: "Egg Fried Rice", Type: "prepared food", BasePrice: 0, WeightGrams: 350,
		Width: 1, Length: 1, Height: 2,
		BaseExpiryDays: 2, UsesTotal: 4, Temperature: 70,
		CanBeCooked: false, CanBeMixed: false,
		Nutrition: model.Nutrition{Food: 35, Energy: 15},
	}

	ChickenRiceBowl = model.Food{
		Name: "Chicken Rice Bowl", Type: "prepared food", BasePrice: 0, WeightGrams: 480,
		Width: 1, Length: 1, Height: 2,
		BaseExpiryDays: 2, UsesTotal: 4, Temperature: 75,
		CanBeCooked: false, CanBeMixed: false,
		Nutrition: model.Nutrition{Food: 42, Energy: 18, Strength: 5},
	}

	Bolognese = model.Food{
		Name: "Bolognese", Type: "prepared food", BasePrice: 0, WeightGrams: 470,
		Width: 1, Length: 1, Height: 2,
		BaseExpiryDays: 2, UsesTotal: 4, Temperature: 75,
		CanBeCooked: false, CanBeMixed: false,
		Nutrition: model.Nutrition{Food: 40, Energy: 19, Strength: 6},
	}

	SalmonBowl = model.Food{
		Name: "Salmon Bowl", Type: "prepared food", BasePrice: 0, WeightGrams: 460,
		Width: 1, Length: 1, Height: 2,
		BaseExpiryDays: 2, UsesTotal: 4, Temperature: 70,
		CanBeCooked: false, CanBeMixed: false,
		Nutrition: model.Nutrition{Food: 40, Energy: 17, Confidence: 1, Strength: 3},
	}

	ChickenBroccoli = model.Food{
		Name: "Chicken Broccoli", Type: "prepared food", BasePrice: 0, WeightGrams: 270,
		Width: 1, Length: 2, Height: 1,
		BaseExpiryDays: 2, UsesTotal: 3, Temperature: 70,
		CanBeCooked: false, CanBeMixed: false,
		Nutrition: model.Nutrition{Food: 30, Energy: 12, Confidence: 3, Strength: 5},
	}

	Salad = model.Food{
		Name: "Salad", Type: "prepared food", BasePrice: 0, WeightGrams: 320,
		Width: 1, Length: 1, Height: 2,
		BaseExpiryDays: 2, UsesTotal: 3, Temperature: 4,
		CanBeCooked: false, CanBeMixed: false,
		Nutrition: model.Nutrition{Food: 15, Energy: 5, Confidence: 5},
	}

	Omelette = model.Food{
		Name: "Omelette", Type: "prepared food", BasePrice: 0, WeightGrams: 100,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 2, UsesTotal: 2, Temperature: 70,
		CanBeCooked: false, CanBeMixed: false,
		Nutrition: model.Nutrition{Food: 20, Energy: 10, Confidence: 2, Strength: 1},
	}

	MashedPotato = model.Food{
		Name: "Mashed Potato", Type: "prepared food", BasePrice: 0, WeightGrams: 250,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 2, UsesTotal: 2, Temperature: 70,
		CanBeCooked: false, CanBeMixed: false,
		Nutrition: model.Nutrition{Food: 25, Energy: 15, Confidence: 1, Strength: 1},
	}

	Toast = model.Food{
		Name: "Toast", Type: "prepared food", BasePrice: 0, WeightGrams: 70,
		Width: 1, Length: 2, Height: 1,
		BaseExpiryDays: 1, UsesTotal: 2, Temperature: 50,
		CanBeCooked: false, CanBeMixed: false,
		Nutrition: model.Nutrition{Food: 14, Energy: 10, Hygiene: -1},
	}

	GrilledCheese = model.Food{
		Name: "Grilled Cheese", Type: "prepared food", BasePrice: 0, WeightGrams: 110,
		Width: 1, Length: 2, Height: 1,
		BaseExpiryDays: 1, UsesTotal: 2, Temperature: 60,
		CanBeCooked: false, CanBeMixed: false,
		Nutrition: model.Nutrition{Food: 18, Energy: 9, Hygiene: -1},
	}

	// ── Part 4: Processed ingredients ───────────────────────────────────────

	ChoppedOnion = model.Food{
		Name: "Chopped Onion", Type: "raw veg", BasePrice: 0, WeightGrams: 140,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 3, UsesTotal: 2, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 3, Energy: 1},
	}

	SlicedOnion = model.Food{
		Name: "Sliced Onion", Type: "raw veg", BasePrice: 0, WeightGrams: 140,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 3, UsesTotal: 2, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 3, Energy: 1},
	}

	MincedGarlic = model.Food{
		Name: "Minced Garlic", Type: "raw veg", BasePrice: 0, WeightGrams: 8,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 3, UsesTotal: 4, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 1},
	}

	ChoppedCarrot = model.Food{
		Name: "Chopped Carrot", Type: "raw veg", BasePrice: 0, WeightGrams: 70,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 3, UsesTotal: 2, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 4, Energy: 2},
	}

	SlicedCarrot = model.Food{
		Name: "Sliced Carrot", Type: "raw veg", BasePrice: 0, WeightGrams: 70,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 3, UsesTotal: 2, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 4, Energy: 2},
	}

	SlicedTomato = model.Food{
		Name: "Sliced Tomato", Type: "raw veg", BasePrice: 0, WeightGrams: 110,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 2, UsesTotal: 2, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 3, Energy: 2},
	}

	SlicedCheese = model.Food{
		Name: "Sliced Cheese", Type: "dairy", BasePrice: 0, WeightGrams: 35,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 10, UsesTotal: 5, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 5, Energy: 2},
	}

	GratedCheese = model.Food{
		Name: "Grated Cheese", Type: "dairy", BasePrice: 0, WeightGrams: 35,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 5, UsesTotal: 5, Temperature: 4,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{Food: 5, Energy: 2},
	}

	RiceFlour = model.Food{
		Name: "Rice Flour", Type: "raw carb", BasePrice: 0, WeightGrams: 95,
		Width: 1, Length: 1, Height: 1,
		BaseExpiryDays: 180, UsesTotal: 6, Temperature: 20,
		CanBeCooked: false, CanBeMixed: true,
		Nutrition: model.Nutrition{},
	}

	// ── Register all ────────────────────────────────────────────────────────

	model.RegisterFood(
		// Raw produce
		Apple, Banana, Orange, Tomato, Avocado,
		// Raw vegetables
		Broccoli, Carrot, Onion, Garlic, Lettuce,
		// Raw carbs
		Potato, Rice, Pasta, Oatmeal,
		// Raw proteins
		Egg, ChickenBreast, BeefMince, SalmonFillet, Tofu,
		// Dairy
		Milk, Cheese, Butter, Yogurt,
		// Pantry
		CookingOil, SoySauce,
		// Seasonings
		Salt, Pepper,
		// Bread
		Bread,
		// Cooked variants
		CookedEgg, CookedChicken, CookedBeef, CookedSalmon,
		CookedBroccoli, CookedCarrot, CookedPotato, CookedRice, CookedPasta,
		// Mixed/Prepared dishes
		EggFriedRice, ChickenRiceBowl, Bolognese, SalmonBowl, ChickenBroccoli,
		Salad, Omelette, MashedPotato, Toast, GrilledCheese,
		// Processed ingredients
		ChoppedOnion, SlicedOnion, MincedGarlic, ChoppedCarrot, SlicedCarrot,
		SlicedTomato, SlicedCheese, GratedCheese, RiceFlour,
	)
}

// All is the slice for shop display (only purchasable raw foods).
var All = []model.Food{
	Apple, Banana, Orange, Tomato, Avocado,
	Broccoli, Carrot, Onion, Garlic, Lettuce,
	Potato,
	Egg, ChickenBreast, BeefMince, SalmonFillet, Tofu,
	Milk, Cheese, Butter, Yogurt,
	Rice, Bread, Pasta, Oatmeal,
	CookingOil, SoySauce,
	Salt, Pepper,
}
