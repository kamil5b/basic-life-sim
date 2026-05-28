package model

type Food struct {
	Name                  string
	Type                  string // cooked and prepared food, cooked protein, raw protein, fruit, raw vegetable, cooked vegetable, liquid, raw carb, cooked carb, herb
	BasePrice             float64
	Width, Length, Height uint8   // space this item occupies inside a storage container
	BaseExpiryDays        uint16  // days until expiry at room temperature
	UsesTotal             uint8   // how many times it can be eaten before it's gone
	Temperature           float64 // in celsius
	CanBeCooked           bool
	CanBeMixed            bool
	EatAction             func(stat *Stats, char *Character) // function that applies the effects of eating the food
}
