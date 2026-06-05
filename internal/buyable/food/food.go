package food

import "github.com/kamil5b/basic-life-sim/internal/model"

// ── Group-level vars ──────────────────────────────────────────────────────────
var (
	Apple, Banana, Orange, Tomato, Avocado                                model.Food
	Broccoli, Carrot, Onion, Garlic, Lettuce                              model.Food
	Potato, Rice, Pasta, Oatmeal                                          model.Food
	Egg, ChickenBreast, BeefMince, SalmonFillet, Tofu                     model.Food
	Milk, Cheese, Butter, Yogurt                                          model.Food
	CookingOil, SoySauce                                                  model.Food
	Salt, Pepper                                                          model.Food
	Bread                                                                 model.Food
	CookedEgg, CookedChicken, CookedBeef, CookedSalmon                    model.Food
	CookedBroccoli, CookedCarrot, CookedPotato, CookedRice, CookedPasta   model.Food
	EggFriedRice, ChickenRiceBowl, Bolognese, SalmonBowl, ChickenBroccoli model.Food
	Salad, Omelette, MashedPotato, Toast, GrilledCheese                   model.Food
	ChoppedOnion, SlicedOnion, MincedGarlic, ChoppedCarrot, SlicedCarrot  model.Food
	SlicedTomato, SlicedCheese, GratedCheese, RiceFlour                   model.Food
)

func init() {

	Apple = model.Food{BaseItem: model.BaseItem{Name: "Apple", BasePrice: 1.00, Width: 1, Length: 1, Height: 1}, Type: "fruit", WeightGrams: 180, BaseExpiryDays: 7, UsesTotal: 1, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 6, Energy: 3}}
	Banana = model.Food{BaseItem: model.BaseItem{Name: "Banana", BasePrice: 0.80, Width: 1, Length: 1, Height: 1}, Type: "fruit", WeightGrams: 120, BaseExpiryDays: 5, UsesTotal: 1, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 8, Energy: 5}}
	Orange = model.Food{BaseItem: model.BaseItem{Name: "Orange", BasePrice: 1.20, Width: 1, Length: 1, Height: 1}, Type: "fruit", WeightGrams: 150, BaseExpiryDays: 10, UsesTotal: 1, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 5, Energy: 4, Confidence: 1}}
	Tomato = model.Food{BaseItem: model.BaseItem{Name: "Tomato", BasePrice: 0.60, Width: 1, Length: 1, Height: 1}, Type: "fruit", WeightGrams: 120, BaseExpiryDays: 5, UsesTotal: 1, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 3, Energy: 2}, ProcessResults: []model.ProcessResult{{Ability: model.ProcessSlice, ResultName: "Sliced Tomato"}}}
	Avocado = model.Food{BaseItem: model.BaseItem{Name: "Avocado", BasePrice: 2.00, Width: 1, Length: 1, Height: 1}, Type: "fruit", WeightGrams: 200, BaseExpiryDays: 4, UsesTotal: 1, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 10, Energy: 3, Confidence: 1}}

	Broccoli = model.Food{BaseItem: model.BaseItem{Name: "Broccoli", BasePrice: 1.20, Width: 1, Length: 1, Height: 2}, Type: "raw veg", WeightGrams: 100, BaseExpiryDays: 5, UsesTotal: 2, Temperature: 4, CanBeCooked: true, CookedResult: "Cooked Broccoli", CanBeMixed: true, Nutrition: model.Nutrition{Food: 5, Energy: 2, Confidence: 1}}
	Carrot = model.Food{BaseItem: model.BaseItem{Name: "Carrot", BasePrice: 0.70, Width: 1, Length: 1, Height: 1}, Type: "raw veg", WeightGrams: 80, BaseExpiryDays: 10, UsesTotal: 1, Temperature: 4, CanBeCooked: true, CookedResult: "Cooked Carrot", CanBeMixed: true, Nutrition: model.Nutrition{Food: 4, Energy: 2}, ProcessResults: []model.ProcessResult{{Ability: model.ProcessChop, ResultName: "Chopped Carrot"}, {Ability: model.ProcessSlice, ResultName: "Sliced Carrot"}}}
	Onion = model.Food{BaseItem: model.BaseItem{Name: "Onion", BasePrice: 0.50, Width: 1, Length: 1, Height: 1}, Type: "raw veg", WeightGrams: 150, BaseExpiryDays: 14, UsesTotal: 1, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 3, Energy: 1}, ProcessResults: []model.ProcessResult{{Ability: model.ProcessChop, ResultName: "Chopped Onion"}, {Ability: model.ProcessSlice, ResultName: "Sliced Onion"}}}
	Garlic = model.Food{BaseItem: model.BaseItem{Name: "Garlic", BasePrice: 0.30, Width: 1, Length: 1, Height: 1}, Type: "raw veg", WeightGrams: 10, BaseExpiryDays: 30, UsesTotal: 3, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 1}, ProcessResults: []model.ProcessResult{{Ability: model.ProcessGrind, ResultName: "Minced Garlic"}}}
	Lettuce = model.Food{BaseItem: model.BaseItem{Name: "Lettuce", BasePrice: 1.50, Width: 1, Length: 1, Height: 2}, Type: "raw veg", WeightGrams: 100, BaseExpiryDays: 4, UsesTotal: 4, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 2, Energy: 1}}

	Potato = model.Food{BaseItem: model.BaseItem{Name: "Potato", BasePrice: 0.80, Width: 1, Length: 1, Height: 1}, Type: "raw carb", WeightGrams: 200, BaseExpiryDays: 14, UsesTotal: 1, Temperature: 4, CanBeCooked: true, CookedResult: "Cooked Potato", CanBeMixed: true, Nutrition: model.Nutrition{Food: 8, Energy: 5}}
	Rice = model.Food{BaseItem: model.BaseItem{Name: "Rice", BasePrice: 2.00, Width: 1, Length: 1, Height: 2}, Type: "raw carb", WeightGrams: 100, BaseExpiryDays: 365, UsesTotal: 6, Temperature: 20, CanBeCooked: true, CookedResult: "Cooked Rice", CanBeMixed: true, ProcessResults: []model.ProcessResult{{Ability: model.ProcessGrind, ResultName: "Rice Flour"}}}
	Pasta = model.Food{BaseItem: model.BaseItem{Name: "Pasta", BasePrice: 1.50, Width: 1, Length: 1, Height: 2}, Type: "raw carb", WeightGrams: 100, BaseExpiryDays: 365, UsesTotal: 4, Temperature: 20, CanBeCooked: true, CookedResult: "Cooked Pasta", CanBeMixed: true}
	Oatmeal = model.Food{BaseItem: model.BaseItem{Name: "Oatmeal", BasePrice: 1.20, Width: 1, Length: 1, Height: 1}, Type: "raw carb", WeightGrams: 50, BaseExpiryDays: 180, UsesTotal: 4, Temperature: 20, CanBeMixed: true, Nutrition: model.Nutrition{Food: 4, Energy: 3}}

	Egg = model.Food{BaseItem: model.BaseItem{Name: "Egg", BasePrice: 0.50, Width: 1, Length: 1, Height: 1}, Type: "raw protein", WeightGrams: 50, BaseExpiryDays: 21, UsesTotal: 1, Temperature: 4, CanBeCooked: true, CookedResult: "Cooked Egg", CanBeMixed: true, Nutrition: model.Nutrition{Food: 5}}
	ChickenBreast = model.Food{BaseItem: model.BaseItem{Name: "Chicken Breast", BasePrice: 3.00, Width: 1, Length: 2, Height: 1}, Type: "raw protein", WeightGrams: 200, BaseExpiryDays: 3, UsesTotal: 2, Temperature: 4, CanBeCooked: true, CookedResult: "Cooked Chicken", CanBeMixed: false}
	BeefMince = model.Food{BaseItem: model.BaseItem{Name: "Beef Mince", BasePrice: 4.00, Width: 1, Length: 2, Height: 1}, Type: "raw protein", WeightGrams: 250, BaseExpiryDays: 2, UsesTotal: 3, Temperature: 4, CanBeCooked: true, CookedResult: "Cooked Beef", CanBeMixed: false}
	SalmonFillet = model.Food{BaseItem: model.BaseItem{Name: "Salmon Fillet", BasePrice: 5.00, Width: 1, Length: 2, Height: 1}, Type: "raw protein", WeightGrams: 180, BaseExpiryDays: 2, UsesTotal: 2, Temperature: 4, CanBeCooked: true, CookedResult: "Cooked Salmon", CanBeMixed: false}
	Tofu = model.Food{BaseItem: model.BaseItem{Name: "Tofu", BasePrice: 2.00, Width: 1, Length: 1, Height: 1}, Type: "raw protein", WeightGrams: 200, BaseExpiryDays: 7, UsesTotal: 3, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 6, Energy: 2, Strength: 1}}

	Milk = model.Food{BaseItem: model.BaseItem{Name: "Milk", BasePrice: 1.50, Width: 1, Length: 1, Height: 2}, Type: "liquid", WeightGrams: 250, BaseExpiryDays: 5, UsesTotal: 4, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 4, Strength: 1}}
	Cheese = model.Food{BaseItem: model.BaseItem{Name: "Cheese", BasePrice: 3.00, Width: 1, Length: 1, Height: 1}, Type: "dairy", WeightGrams: 40, BaseExpiryDays: 14, UsesTotal: 4, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 5, Energy: 2}, ProcessResults: []model.ProcessResult{{Ability: model.ProcessSlice, ResultName: "Sliced Cheese"}, {Ability: model.ProcessGrind, ResultName: "Grated Cheese"}}}
	Butter = model.Food{BaseItem: model.BaseItem{Name: "Butter", BasePrice: 1.00, Width: 1, Length: 1, Height: 1}, Type: "dairy", WeightGrams: 20, BaseExpiryDays: 21, UsesTotal: 10, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 3, Energy: 4, Hygiene: -1}}
	Yogurt = model.Food{BaseItem: model.BaseItem{Name: "Yogurt", BasePrice: 1.80, Width: 1, Length: 1, Height: 1}, Type: "dairy", WeightGrams: 150, BaseExpiryDays: 7, UsesTotal: 2, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 5, Energy: 1}}

	Bread = model.Food{BaseItem: model.BaseItem{Name: "Bread", BasePrice: 2.50, Width: 1, Length: 2, Height: 1}, Type: "carb", WeightGrams: 50, BaseExpiryDays: 5, UsesTotal: 10, Temperature: 20, CanBeMixed: true, Nutrition: model.Nutrition{Food: 8, Energy: 5}}

	CookingOil = model.Food{BaseItem: model.BaseItem{Name: "Cooking Oil", BasePrice: 3.00, Width: 1, Length: 1, Height: 2}, Type: "liquid", WeightGrams: 15, BaseExpiryDays: 180, UsesTotal: 20, Temperature: 20, CanBeMixed: true, Nutrition: model.Nutrition{Food: 2, Energy: 5, Hygiene: -1}}
	SoySauce = model.Food{BaseItem: model.BaseItem{Name: "Soy Sauce", BasePrice: 2.00, Width: 1, Length: 1, Height: 1}, Type: "liquid", WeightGrams: 10, BaseExpiryDays: 365, UsesTotal: 15, Temperature: 20, CanBeMixed: true}
	Salt = model.Food{BaseItem: model.BaseItem{Name: "Salt", BasePrice: 1.00, Width: 1, Length: 1, Height: 1}, Type: "seasoning", WeightGrams: 5, BaseExpiryDays: 9999, UsesTotal: 30, Temperature: 20, CanBeMixed: true}
	Pepper = model.Food{BaseItem: model.BaseItem{Name: "Pepper", BasePrice: 1.50, Width: 1, Length: 1, Height: 1}, Type: "seasoning", WeightGrams: 2, BaseExpiryDays: 9999, UsesTotal: 20, Temperature: 20, CanBeMixed: true}

	// ── Cooked ──
	CookedEgg = model.Food{BaseItem: model.BaseItem{Name: "Cooked Egg", BasePrice: 0, Width: 1, Length: 1, Height: 1}, Type: "cooked protein", WeightGrams: 45, BaseExpiryDays: 2, UsesTotal: 1, Temperature: 70, CanBeMixed: true, Nutrition: model.Nutrition{Food: 12, Energy: 5}, MixRecipes: []model.MixRecipe{{Ingredient: model.MixIngredient{FoodName: "Cooked Rice"}, ResultName: "Egg Fried Rice"}}}
	CookedChicken = model.Food{BaseItem: model.BaseItem{Name: "Cooked Chicken", BasePrice: 0, Width: 1, Length: 2, Height: 1}, Type: "cooked protein", WeightGrams: 180, BaseExpiryDays: 4, UsesTotal: 2, Temperature: 75, CanBeMixed: true, Nutrition: model.Nutrition{Food: 20, Energy: 8, Strength: 5}, MixRecipes: []model.MixRecipe{{Ingredient: model.MixIngredient{FoodName: "Cooked Rice"}, ResultName: "Chicken Rice Bowl"}, {Ingredient: model.MixIngredient{FoodName: "Cooked Broccoli"}, ResultName: "Chicken Broccoli"}}}
	CookedBeef = model.Food{BaseItem: model.BaseItem{Name: "Cooked Beef", BasePrice: 0, Width: 1, Length: 2, Height: 1}, Type: "cooked protein", WeightGrams: 220, BaseExpiryDays: 3, UsesTotal: 3, Temperature: 75, CanBeMixed: true, Nutrition: model.Nutrition{Food: 22, Energy: 10, Strength: 6}, MixRecipes: []model.MixRecipe{{Ingredient: model.MixIngredient{FoodName: "Cooked Pasta"}, ResultName: "Bolognese"}}}
	CookedSalmon = model.Food{BaseItem: model.BaseItem{Name: "Cooked Salmon", BasePrice: 0, Width: 1, Length: 2, Height: 1}, Type: "cooked protein", WeightGrams: 160, BaseExpiryDays: 3, UsesTotal: 2, Temperature: 70, CanBeMixed: true, Nutrition: model.Nutrition{Food: 18, Energy: 7, Confidence: 1, Strength: 3}, MixRecipes: []model.MixRecipe{{Ingredient: model.MixIngredient{FoodName: "Cooked Rice"}, ResultName: "Salmon Bowl"}}}
	CookedBroccoli = model.Food{BaseItem: model.BaseItem{Name: "Cooked Broccoli", BasePrice: 0, Width: 1, Length: 1, Height: 2}, Type: "cooked vegetable", WeightGrams: 90, BaseExpiryDays: 3, UsesTotal: 2, Temperature: 70, CanBeMixed: true, Nutrition: model.Nutrition{Food: 12, Energy: 4, Confidence: 3}}
	CookedCarrot = model.Food{BaseItem: model.BaseItem{Name: "Cooked Carrot", BasePrice: 0, Width: 1, Length: 1, Height: 1}, Type: "cooked vegetable", WeightGrams: 70, BaseExpiryDays: 3, UsesTotal: 1, Temperature: 70, CanBeMixed: true, Nutrition: model.Nutrition{Food: 8, Energy: 3, Confidence: 2}}
	CookedPotato = model.Food{BaseItem: model.BaseItem{Name: "Cooked Potato", BasePrice: 0, Width: 1, Length: 1, Height: 1}, Type: "cooked carb", WeightGrams: 180, BaseExpiryDays: 3, UsesTotal: 1, Temperature: 80, CanBeMixed: true, Nutrition: model.Nutrition{Food: 14, Energy: 8}}
	CookedRice = model.Food{BaseItem: model.BaseItem{Name: "Cooked Rice", BasePrice: 0, Width: 1, Length: 1, Height: 2}, Type: "cooked carb", WeightGrams: 300, BaseExpiryDays: 3, UsesTotal: 6, Temperature: 80, CanBeMixed: true, Nutrition: model.Nutrition{Food: 25, Energy: 10}, MixRecipes: []model.MixRecipe{{Ingredient: model.MixIngredient{FoodName: "Cooked Egg"}, ResultName: "Egg Fried Rice"}, {Ingredient: model.MixIngredient{FoodName: "Cooked Chicken"}, ResultName: "Chicken Rice Bowl"}, {Ingredient: model.MixIngredient{FoodName: "Cooked Salmon"}, ResultName: "Salmon Bowl"}}}
	CookedPasta = model.Food{BaseItem: model.BaseItem{Name: "Cooked Pasta", BasePrice: 0, Width: 1, Length: 1, Height: 2}, Type: "cooked carb", WeightGrams: 250, BaseExpiryDays: 3, UsesTotal: 4, Temperature: 80, CanBeMixed: true, Nutrition: model.Nutrition{Food: 20, Energy: 9}, MixRecipes: []model.MixRecipe{{Ingredient: model.MixIngredient{FoodName: "Cooked Beef"}, ResultName: "Bolognese"}}}

	// ── Prepared dishes ──
	EggFriedRice = model.Food{BaseItem: model.BaseItem{Name: "Egg Fried Rice", BasePrice: 0, Width: 1, Length: 1, Height: 2}, Type: "prepared food", WeightGrams: 350, BaseExpiryDays: 2, UsesTotal: 4, Temperature: 70, Nutrition: model.Nutrition{Food: 35, Energy: 15}}
	ChickenRiceBowl = model.Food{BaseItem: model.BaseItem{Name: "Chicken Rice Bowl", BasePrice: 0, Width: 1, Length: 1, Height: 2}, Type: "prepared food", WeightGrams: 480, BaseExpiryDays: 2, UsesTotal: 4, Temperature: 75, Nutrition: model.Nutrition{Food: 42, Energy: 18, Strength: 5}}
	Bolognese = model.Food{BaseItem: model.BaseItem{Name: "Bolognese", BasePrice: 0, Width: 1, Length: 1, Height: 2}, Type: "prepared food", WeightGrams: 470, BaseExpiryDays: 2, UsesTotal: 4, Temperature: 75, Nutrition: model.Nutrition{Food: 40, Energy: 19, Strength: 6}}
	SalmonBowl = model.Food{BaseItem: model.BaseItem{Name: "Salmon Bowl", BasePrice: 0, Width: 1, Length: 1, Height: 2}, Type: "prepared food", WeightGrams: 460, BaseExpiryDays: 2, UsesTotal: 4, Temperature: 70, Nutrition: model.Nutrition{Food: 40, Energy: 17, Confidence: 1, Strength: 3}}
	ChickenBroccoli = model.Food{BaseItem: model.BaseItem{Name: "Chicken Broccoli", BasePrice: 0, Width: 1, Length: 2, Height: 1}, Type: "prepared food", WeightGrams: 270, BaseExpiryDays: 2, UsesTotal: 3, Temperature: 70, Nutrition: model.Nutrition{Food: 30, Energy: 12, Confidence: 3, Strength: 5}}
	Salad = model.Food{BaseItem: model.BaseItem{Name: "Salad", BasePrice: 0, Width: 1, Length: 1, Height: 2}, Type: "prepared food", WeightGrams: 320, BaseExpiryDays: 2, UsesTotal: 3, Temperature: 4, Nutrition: model.Nutrition{Food: 15, Energy: 5, Confidence: 5}}
	Omelette = model.Food{BaseItem: model.BaseItem{Name: "Omelette", BasePrice: 0, Width: 1, Length: 1, Height: 1}, Type: "prepared food", WeightGrams: 100, BaseExpiryDays: 2, UsesTotal: 2, Temperature: 70, Nutrition: model.Nutrition{Food: 20, Energy: 10, Confidence: 2, Strength: 1}}
	MashedPotato = model.Food{BaseItem: model.BaseItem{Name: "Mashed Potato", BasePrice: 0, Width: 1, Length: 1, Height: 1}, Type: "prepared food", WeightGrams: 250, BaseExpiryDays: 2, UsesTotal: 2, Temperature: 70, Nutrition: model.Nutrition{Food: 25, Energy: 15, Confidence: 1, Strength: 1}}
	Toast = model.Food{BaseItem: model.BaseItem{Name: "Toast", BasePrice: 0, Width: 1, Length: 2, Height: 1}, Type: "prepared food", WeightGrams: 70, BaseExpiryDays: 1, UsesTotal: 2, Temperature: 50, Nutrition: model.Nutrition{Food: 14, Energy: 10, Hygiene: -1}}
	GrilledCheese = model.Food{BaseItem: model.BaseItem{Name: "Grilled Cheese", BasePrice: 0, Width: 1, Length: 2, Height: 1}, Type: "prepared food", WeightGrams: 110, BaseExpiryDays: 1, UsesTotal: 2, Temperature: 60, Nutrition: model.Nutrition{Food: 18, Energy: 9, Hygiene: -1}}

	// ── Processed ingredients ──
	ChoppedOnion = model.Food{BaseItem: model.BaseItem{Name: "Chopped Onion", BasePrice: 0, Width: 1, Length: 1, Height: 1}, Type: "raw veg", WeightGrams: 140, BaseExpiryDays: 3, UsesTotal: 2, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 3, Energy: 1}}
	SlicedOnion = model.Food{BaseItem: model.BaseItem{Name: "Sliced Onion", BasePrice: 0, Width: 1, Length: 1, Height: 1}, Type: "raw veg", WeightGrams: 140, BaseExpiryDays: 3, UsesTotal: 2, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 3, Energy: 1}}
	MincedGarlic = model.Food{BaseItem: model.BaseItem{Name: "Minced Garlic", BasePrice: 0, Width: 1, Length: 1, Height: 1}, Type: "raw veg", WeightGrams: 8, BaseExpiryDays: 3, UsesTotal: 4, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 1}}
	ChoppedCarrot = model.Food{BaseItem: model.BaseItem{Name: "Chopped Carrot", BasePrice: 0, Width: 1, Length: 1, Height: 1}, Type: "raw veg", WeightGrams: 70, BaseExpiryDays: 3, UsesTotal: 2, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 4, Energy: 2}}
	SlicedCarrot = model.Food{BaseItem: model.BaseItem{Name: "Sliced Carrot", BasePrice: 0, Width: 1, Length: 1, Height: 1}, Type: "raw veg", WeightGrams: 70, BaseExpiryDays: 3, UsesTotal: 2, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 4, Energy: 2}}
	SlicedTomato = model.Food{BaseItem: model.BaseItem{Name: "Sliced Tomato", BasePrice: 0, Width: 1, Length: 1, Height: 1}, Type: "raw veg", WeightGrams: 110, BaseExpiryDays: 2, UsesTotal: 2, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 3, Energy: 2}}
	SlicedCheese = model.Food{BaseItem: model.BaseItem{Name: "Sliced Cheese", BasePrice: 0, Width: 1, Length: 1, Height: 1}, Type: "dairy", WeightGrams: 35, BaseExpiryDays: 10, UsesTotal: 5, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 5, Energy: 2}}
	GratedCheese = model.Food{BaseItem: model.BaseItem{Name: "Grated Cheese", BasePrice: 0, Width: 1, Length: 1, Height: 1}, Type: "dairy", WeightGrams: 35, BaseExpiryDays: 5, UsesTotal: 5, Temperature: 4, CanBeMixed: true, Nutrition: model.Nutrition{Food: 5, Energy: 2}}
	RiceFlour = model.Food{BaseItem: model.BaseItem{Name: "Rice Flour", BasePrice: 0, Width: 1, Length: 1, Height: 1}, Type: "raw carb", WeightGrams: 95, BaseExpiryDays: 180, UsesTotal: 6, Temperature: 20, CanBeMixed: true}

	model.RegisterFood(
		Apple, Banana, Orange, Tomato, Avocado,
		Broccoli, Carrot, Onion, Garlic, Lettuce,
		Potato, Rice, Pasta, Oatmeal,
		Egg, ChickenBreast, BeefMince, SalmonFillet, Tofu,
		Milk, Cheese, Butter, Yogurt,
		CookingOil, SoySauce, Salt, Pepper, Bread,
		CookedEgg, CookedChicken, CookedBeef, CookedSalmon,
		CookedBroccoli, CookedCarrot, CookedPotato, CookedRice, CookedPasta,
		EggFriedRice, ChickenRiceBowl, Bolognese, SalmonBowl, ChickenBroccoli,
		Salad, Omelette, MashedPotato, Toast, GrilledCheese,
		ChoppedOnion, SlicedOnion, MincedGarlic, ChoppedCarrot, SlicedCarrot,
		SlicedTomato, SlicedCheese, GratedCheese, RiceFlour,
	)
}

var All = []model.Food{
	Apple, Banana, Orange, Tomato, Avocado,
	Broccoli, Carrot, Onion, Garlic, Lettuce,
	Potato, Egg, ChickenBreast, BeefMince, SalmonFillet, Tofu,
	Milk, Cheese, Butter, Yogurt,
	Rice, Bread, Pasta, Oatmeal,
	CookingOil, SoySauce, Salt, Pepper,
}
