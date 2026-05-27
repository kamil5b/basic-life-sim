package furniture

import "basic-life-sim/internal/model"

var BasicChair = model.RoomItem{
	Name:          "Basic Chair",
	Type:          "Furniture",
	Category:      "Low",
	WillBlockPath: false,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 30,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "sit":
			stat.Energy = safeAdd(stat.Energy, 5)
		}
	},
}

var GamingChair = model.RoomItem{
	Name:          "Gaming Chair",
	Type:          "Furniture",
	Category:      "Medium-High",
	WillBlockPath: false,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 250,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "sit":
			stat.Energy = safeAdd(stat.Energy, 10)
			stat.Confidence = safeAdd(stat.Confidence, 3)
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
