package model

// HomeCell represents a single cell in a home layout grid.
type HomeCell uint8

const (
	HomeCellWall  HomeCell = 0
	HomeCellFloor HomeCell = 1
	HomeCellDoor  HomeCell = 2
)

// HomeLayout is the main foundation grid for the home.
// HomeCellWall  - not available for placement
// HomeCellFloor - available for placement
// HomeCellDoor  - not available for placement, used as exit point
type HomeLayout [][]HomeCell

type HomeType struct {
	Name           string
	Layout         HomeLayout
	MaxHeight      uint8
	Level          uint8
	SharedBathroom bool // if true, there will be random events where the bathroom is occupied
	SharedKitchen  bool // if true, there will be random events where the kitchen is occupied
	UpfrontCost    float64
	MonthlyCost    float64
}

// RoomItemType categorises what kind of room item this is.
type RoomItemType uint8

const (
	RoomItemFurniture RoomItemType = iota
	RoomItemAppliance
	RoomItemHygiene
	RoomItemNeeds
)

// RoomItemCategory represents the quality/price tier of a room item.
type RoomItemCategory uint8

const (
	CategoryLow RoomItemCategory = iota
	CategoryMediumLow
	CategoryMedium
	CategoryMediumHigh
	CategoryHigh
)

// Direction represents the cardinal facing of a placed room item.
type Direction uint8

const (
	North Direction = iota
	East
	South
	West
)

// Stats holds purely numeric character state that items and experiences mutate.
type Stats struct {
	Money float64
}

// NeedStat holds the current value and maximum cap for a single character need.
type NeedStat struct {
	Current uint16
	Max     uint16
}

type RoomItem struct {
	Name                  string
	Type                  RoomItemType
	Category              RoomItemCategory
	WillBlockPath         bool
	Width, Length, Height uint8
	BasePrice             float64
	Actions               []string // available action verbs passed to DoAction
	DoAction              func(input string, stat *Stats, char *Character)
}

type PlacedRoomItem struct {
	X         uint8
	Y         uint8
	Z         uint8
	Direction Direction
	Item      RoomItem
}

type Home struct {
	Type      HomeType
	RoomItems []PlacedRoomItem
}

type ExperienceType uint8

const (
	UniStudent ExperienceType = iota
	PartTimer
	FullTimer
)

type ExperienceQualification struct {
	Type     ExperienceType
	Major    []string
	Duration uint16 // in months, 0 means any duration
}

type Experience struct {
	Name            string
	Description     string
	Type            ExperienceType
	Major           string
	Qualifications  []ExperienceQualification                                                  // empty means no qualifications required
	ApplyExperience func(pastExperiences []TakenExperience, stat *Stats, char *Character) bool // returns true if accepted
}

type TakenExperience struct {
	Experience
	StartDate CompactDate
	EndDate   *CompactDate // nil means still ongoing
}

type Character struct {
	Name        string
	IsMale      bool
	CurrentHome Home
	Experiences []TakenExperience
	Age         uint8

	Food       NeedStat
	Energy     NeedStat
	Hygiene    NeedStat
	Confidence NeedStat
	Strength   NeedStat

	CurrentStats Stats
}
