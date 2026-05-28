package hygiene

import "basic-life-sim/internal/model"

var BasicShower = model.RoomItem{
	Name:          "Basic Shower",
	Type:          model.RoomItemHygiene,
	Category:      model.CategoryLow,
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 2,
	BasePrice: 150,
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "shower":
			char.Hygiene.Current = safeAdd(char.Hygiene.Current, 30)
			char.Energy.Current = safeAdd(char.Energy.Current, 5)
		}
	},
}

var PremiumShower = model.RoomItem{
	Name:          "Premium Shower",
	Type:          model.RoomItemHygiene,
	Category:      model.CategoryHigh,
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 2,
	BasePrice: 600,
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "shower":
			char.Hygiene.Current = safeAdd(char.Hygiene.Current, 40)
			char.Energy.Current = safeAdd(char.Energy.Current, 10)
			char.Confidence.Current = safeAdd(char.Confidence.Current, 5)
		}
	},
}

func safeAdd(v, delta uint16) uint16 {
	if v+delta < v {
		return ^uint16(0)
	}
	return v + delta
}

func safeSub(v, delta uint16) uint16 {
	if delta > v {
		return 0
	}
	return v - delta
}
