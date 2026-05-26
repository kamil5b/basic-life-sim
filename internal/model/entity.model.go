package model

// HomeLayout is main foundation grid for the home
// 0 is wall / not available for things to place in
// 1 is floor / available for things to place in
// 2 is door / not available for things to place in, but also can be used as exit point
// There will be mechanics to place furniture and other things in the home, but still have to consider the layout of the home and where things can be placed.
// For example, you can't place a thing infront of a door, you can't clustered all the furniture in one corner of the home, you have to consider the flow of the home and how people will move around in it.
type HomeLayout [][]uint8

type HomeType struct {
	Name           string
	Layout         HomeLayout
	MaxHeight      uint8
	Level          uint8
	SharedBathroom bool // if true, there will be random events where the bathroom is occupied and the character can't use it, and they will have to wait until it's available again
	SharedKitchen  bool // if true, there will be random events where the kitchen is occupied and the character can't use it, and they will have to wait until it's available again
}

type Stats struct {
	Money               float64
	Food                uint16
	Energy              uint16
	Hygiene             uint16
	Confidence          uint16
	Strength            uint16
	AvailableNextAction []string
}

type RoomItem interface {
	GetName() string
	GetType() string
	GetCategory() string
	WillBlockPath() bool
	GetDimensions() (width, length, height uint8)
	DoAction(input string, stat *Stats)
}

type PlacedRoomItem struct {
	X         uint8
	Y         uint8
	Z         uint8
	Direction uint8 // 0 is north, 1 is east, 2 is south, 3 is west
	Item      RoomItem
}

type Home struct {
	Type      HomeType
	RoomItems []PlacedRoomItem
}

type ExperienceType uint8

const (
	UniStudent ExperienceType = iota // 0
	PartTimer                        // 1
	FullTimer                        // 2
)

type ExperienceQualification struct {
	Type     string
	Major    []string // if empty, then it can be any major
	Duration uint8    // in months, if 0, then it can be any duration
}

type Experience struct {
	Name            string
	Description     string
	Type            ExperienceType
	Major           string
	Qualifications  []ExperienceQualification                                 // if empty, then it can be any qualifications
	ApplyExperience func(pastExperiences []TakenExperience, stat *Stats) bool // returns true if accepted
}

type TakenExperience struct {
	Experience
	StartDate CompactDate
	EndDate   *CompactDate // If null, then it's still ongoing
}

type Character struct {
	Name        string
	IsMale      bool
	CurrentHome Home
	Experiences []TakenExperience
	Age         uint8

	// Needs
	MaxFood       uint16
	MaxEnergy     uint16
	MaxHygiene    uint16
	MaxConfidence uint16
	MaxStrength   uint16
	CurrentStats  Stats
}
