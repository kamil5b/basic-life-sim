package hygiene

import "basic-life-sim/internal/model"

var BasicSink = model.RoomItem{
	Name:          "Basic Sink",
	Type:          "Hygiene",
	Category:      "Low",
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 50,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "wash":
			stat.Hygiene = safeAdd(stat.Hygiene, 10)
		}
	},
}

var VanitySink = model.RoomItem{
	Name:          "Vanity Sink",
	Type:          "Hygiene",
	Category:      "Medium",
	WillBlockPath: true,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 180,
	DoAction: func(input string, stat *model.Stats) {
		switch input {
		case "wash":
			stat.Hygiene = safeAdd(stat.Hygiene, 15)
			stat.Confidence = safeAdd(stat.Confidence, 3)
		}
	},
}
