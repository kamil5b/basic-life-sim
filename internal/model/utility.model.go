package model

// Utility is a tool or cookware item that sits on a host (stove, desk, cutting board).
// It is NOT a RoomItem — it cannot be placed directly on the floor grid.
type Utility struct {
	BaseItem
	Ability      ProcessAbility // what this utility does (chop, grind, slice, etc.)
	CookSurface  *CookCapacity  // non-nil for pans/woks that can hold food
	UtilitySlots uint8          // how many sub-utilities can sit on this (e.g. cutting board → knife)
	Actions      []string
	DoAction     func(input string, stat *Stats, char *Character)
}

var UtilityRegistry = map[string]Utility{}

func RegisterUtility(list ...Utility) {
	for _, u := range list {
		UtilityRegistry[u.Name] = u
	}
}
