package model

// BaseItem holds fields shared by Food and Utility for placement and storage.
type BaseItem struct {
	Name                  string
	BasePrice             float64
	Width, Length, Height uint8
}

// Size returns the item's dimensions.
func (b BaseItem) Size() (w, l, h uint8) {
	return b.Width, b.Length, b.Height
}

// Nutrition — exact stat changes per use, UI-displayable
type Nutrition struct {
	Food       uint16
	Energy     uint16
	Hygiene    int16
	Confidence int16
	Strength   uint16
}

// MixIngredient — one side of a recipe pair
type MixIngredient struct {
	FoodName string // exact name match
	FoodType string // or type match (e.g. "cooked protein")
}

// MixRecipe — what this food + partner produces
type MixRecipe struct {
	Ingredient MixIngredient
	ResultName string // lookup key in FoodRegistry
}

// ProcessAbility — what a utility can do to food (chop, grind, slice, etc.)
type ProcessAbility string

const (
	ProcessChop  ProcessAbility = "chop"
	ProcessGrind ProcessAbility = "grind"
	ProcessSlice ProcessAbility = "slice"
	ProcessBlend ProcessAbility = "blend"
	ProcessPeel  ProcessAbility = "peel"
)

// ProcessResult — what food becomes after a utility processes it
type ProcessResult struct {
	Ability    ProcessAbility
	ResultName string
}

type Food struct {
	BaseItem
	Type           string // cooked and prepared food, cooked protein, raw protein, fruit, raw vegetable, cooked vegetable, liquid, raw carb, cooked carb, herb, dairy, seasoning
	WeightGrams    uint16
	BaseExpiryDays uint16  // days until expiry at room temperature
	UsesTotal      uint8   // how many times it can be eaten/used before it's gone
	Temperature    float64 // in celsius

	// Cooking
	CanBeCooked  bool
	CookedResult string // "" means no transformation; lookup key in FoodRegistry

	// Mixing / Recipes
	CanBeMixed bool
	MixRecipes []MixRecipe

	// Processing (knife → chop, blender → blend)
	ProcessResults []ProcessResult

	// Nutrition
	Nutrition Nutrition

	// Special effect (nil unless special case like raw chicken)
	OnEat func(char *Character)
}

var FoodRegistry = map[string]Food{}

func RegisterFood(list ...Food) {
	for _, f := range list {
		FoodRegistry[f.Name] = f
	}
}

func ResolveCookedResult(f Food) *Food {
	if f.CookedResult == "" {
		return nil
	}
	if r, ok := FoodRegistry[f.CookedResult]; ok {
		return &r
	}
	return nil
}
