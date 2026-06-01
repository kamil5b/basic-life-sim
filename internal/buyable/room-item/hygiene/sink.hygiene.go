package hygiene

import "github.com/kamil5b/basic-life-sim/internal/model"

var BasicSink = model.RoomItem{
	Name:          "Basic Sink",
	Type:          model.RoomItemHygiene,
	Category:      model.CategoryLow,
	WillBlockPath: true,
	NeedClearance: true,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 50,
	Actions:   []string{"wash"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "wash":
			char.Hygiene.Current = safeAdd(char.Hygiene.Current, 10)
		}
	},
}

var VanitySink = model.RoomItem{
	Name:          "Vanity Sink",
	Type:          model.RoomItemHygiene,
	Category:      model.CategoryMedium,
	WillBlockPath: true,
	NeedClearance: true,
	Width:         1, Length: 1, Height: 1,
	BasePrice: 180,
	Actions:   []string{"wash"},
	DoAction: func(input string, stat *model.Stats, char *model.Character) {
		switch input {
		case "wash":
			char.Hygiene.Current = safeAdd(char.Hygiene.Current, 15)
			char.Confidence.Current = safeAdd(char.Confidence.Current, 3)
		}
	},
}
