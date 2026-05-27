package hygiene

import "basic-life-sim/internal/model"

var BasicToilet = model.RoomItem{
	Name:          "Basic Toilet",
	Type:          "Hygiene",
	Category:      "Low",
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 60,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "use":
			stat.Hygiene = safeAdd(stat.Hygiene, 10)
		}
	},
}

var BidetToilet = model.RoomItem{
	Name:          "Bidet Toilet",
	Type:          "Hygiene",
	Category:      "Medium-High",
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 400,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "use":
			stat.Hygiene = safeAdd(stat.Hygiene, 20)
			stat.Confidence = safeAdd(stat.Confidence, 5)
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
