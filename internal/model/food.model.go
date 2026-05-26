package model

type Food struct {
	Name        string
	Type        string  // cooked and prepared food, cooked protein, raw protein, fruit, raw vegetable, cooked vegetable, liquid, raw carb, cooked carb, herb
	Temperature float64 // in celcius
	CanBeCooked bool
	CanBeMixed  bool
	SpoiledDate CompactDate
	EatAction   func(stat *Stats) // function that applies the effects of eating the food to the player's stats
}
