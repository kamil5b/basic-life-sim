package furniture

import "basic-life-sim/internal/model"

var SingleBed = model.RoomItem{
	Name:          "Single Bed",
	Type:          model.RoomItemFurniture,
	Category:      model.CategoryLow,
	WillBlockPath: true,
	Width:         1, Length: 2, Height: 1,
	BasePrice: 150,
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "sleep":
			char.Energy.Current = safeAdd(char.Energy.Current, 40)
		}
	},
}

var QueenBed = model.RoomItem{
	Name:          "Queen Bed",
	Type:          model.RoomItemFurniture,
	Category:      model.CategoryMedium,
	WillBlockPath: true,
	Width:         2, Length: 2, Height: 1,
	BasePrice: 350,
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "sleep":
			char.Energy.Current = safeAdd(char.Energy.Current, 50)
		}
	},
}

var KingBed = model.RoomItem{
	Name:          "King Bed",
	Type:          model.RoomItemFurniture,
	Category:      model.CategoryHigh,
	WillBlockPath: true,
	Width:         2, Length: 2, Height: 1,
	BasePrice: 700,
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "sleep":
			char.Energy.Current = safeAdd(char.Energy.Current, 60)
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
