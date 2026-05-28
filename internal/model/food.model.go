package model

type Food struct {
	Name                  string
	Type                  string // cooked and prepared food, cooked protein, raw protein, fruit, raw vegetable, cooked vegetable, liquid, raw carb, cooked carb, herb
	BasePrice             float64
	Width, Length, Height uint8   // space this item occupies inside a storage container or on a surface
	BaseExpiryDays        uint16  // days until expiry at room temperature
	UsesTotal             uint8   // how many times it can be eaten/used before it's gone
	Temperature           float64 // in celsius
	CanBeCooked           bool
	CanBeMixed            bool
	CookedResult          *Food                              // what this food becomes after cooking; nil means no transformation
	EatAction             func(stat *Stats, char *Character) // nil means the food is inedible raw (penalises stats)
}
