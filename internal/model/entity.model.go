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
	RoomItemStorage
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

// ColdZone defines a sub-region within a storage container that has a higher expiry multiplier.
type ColdZone struct {
	OriginX, OriginY, OriginZ uint8
	Width, Length, Height     uint8
	ExpiryMultiplier          float32
}

// Contains reports whether slot (x,y,z) falls inside this cold zone.
func (cz ColdZone) Contains(x, y, z uint8) bool {
	return x >= cz.OriginX && x < cz.OriginX+cz.Width &&
		y >= cz.OriginY && y < cz.OriginY+cz.Length &&
		z >= cz.OriginZ && z < cz.OriginZ+cz.Height
}

// StorageCapacity defines the internal 3-D dimensions of a storage container (e.g. a fridge).
// Each slot is 1×1×1 and holds exactly one food item.
type StorageCapacity struct {
	Width, Length, Height uint8
	ExpiryMultiplier      float32    // base multiplier for all slots in this container
	ColdZones             []ColdZone // optional sub-regions with a higher multiplier
}

// TotalSlots returns the total number of food slots available.
func (s StorageCapacity) TotalSlots() int {
	return int(s.Width) * int(s.Length) * int(s.Height)
}

// MultiplierAt returns the effective expiry multiplier for a given slot.
// If the slot falls inside a cold zone, the highest applicable zone multiplier is used.
func (s StorageCapacity) MultiplierAt(x, y, z uint8) float32 {
	best := s.ExpiryMultiplier
	for _, cz := range s.ColdZones {
		if cz.Contains(x, y, z) && cz.ExpiryMultiplier > best {
			best = cz.ExpiryMultiplier
		}
	}
	return best
}

// CookCapacity describes a cooking appliance's surface — how many food items it can hold at once.
type CookCapacity struct {
	Slots uint8 // max number of food items on the surface simultaneously
}

type RoomItem struct {
	Name                  string
	Type                  RoomItemType
	Category              RoomItemCategory
	WillBlockPath         bool
	CanOverhang           bool // if true, item may be placed at z>0 without requiring support below
	NeedClearance         bool // if true, requires 1 free cell in front; if false, can be placed flush against other items
	Width, Length, Height uint8
	BasePrice             float64
	Storage               *StorageCapacity // non-nil for items that can store food (fridge)
	CookSurface           *CookCapacity    // non-nil for items that can cook food (stove, microwave)
	UtilitySlots          uint8            // how many utilities can sit on this item (stove, counter)
	Actions               []string         // available action verbs passed to DoAction
	DoAction              func(input string, stat *Stats, char *Character)
}

// StoredFood is a food item occupying a slot inside a storage container.
type StoredFood struct {
	SlotX, SlotY, SlotZ uint8
	PurchaseDate        CompactDate // date the food was bought, used to compute expiry
	MultiplierUsed      float32     // effective multiplier at the slot this food occupies
	UsesRemaining       uint8       // decrements on each eat; item is removed at 0
	Food                Food
}

// InventoryItem holds fields shared by StoredItem and FloorItem.
type InventoryItem struct {
	Kind           FloorItemKind
	PurchaseDate   CompactDate // food only: date bought
	MultiplierUsed float32     // food only: effective expiry multiplier
	UsesRemaining  uint8       // food: decrements on eat; utility: unused
	Food           Food        // set when Kind == FloorKindFood
	Utility        Utility     // set when Kind == FloorKindUtility
}

// Name returns the display name of whichever kind is stored.
func (it InventoryItem) Name() string {
	if it.Kind == FloorKindFood {
		return it.Food.Name
	}
	return it.Utility.Name
}

// Dims returns the dimensions of whatever is stored.
func (it InventoryItem) Dims() (w, l, h uint8) {
	if it.Kind == FloorKindFood {
		return it.Food.Width, it.Food.Length, it.Food.Height
	}
	return it.Utility.Width, it.Utility.Length, it.Utility.Height
}

// StoredItem is a food or utility occupying a slot inside a storage container.
type StoredItem struct {
	SlotX, SlotY, SlotZ uint8
	Item                InventoryItem
}

// Dims returns the dimensions of whatever is stored.
func (s StoredItem) Dims() (w, l, h uint8) {
	return s.Item.Dims()
}

// NewStoredItem creates a StoredItem with Kind-appropriate defaults.
func NewStoredItem(kind FloorItemKind, food Food, utility Utility, slotX, slotY, slotZ uint8, purchaseDate CompactDate, multiplier float32, uses uint8) StoredItem {
	return StoredItem{
		SlotX: slotX,
		SlotY: slotY,
		SlotZ: slotZ,
		Item: InventoryItem{
			Kind:           kind,
			PurchaseDate:   purchaseDate,
			MultiplierUsed: multiplier,
			UsesRemaining:  uses,
			Food:           food,
			Utility:        utility,
		},
	}
}

// SurfaceFood is a food item resting on a cooking appliance surface (placed there by the player).
type SurfaceFood struct {
	PurchaseDate  CompactDate
	UsesRemaining uint8
	Cooked        bool // true once the cook action has been applied
	Food          Food
}

type PlacedRoomItem struct {
	X         uint8
	Y         uint8
	Z         uint8
	Direction Direction
	Item      RoomItem
	Stored    []StoredItem    // food & utilities stored inside
	OnSurface []SurfaceFood   // food resting on the surface (stove/microwave)
	Utilities []PlacedUtility // utilities placed on this item
}

// PlacedUtility — a utility item placed on a host (e.g. pan on stove)
type PlacedUtility struct {
	X, Y, Z   uint8
	Direction Direction
	Item      Utility
	OnSurface []SurfaceFood // food placed inside the pan/wok
}

// PlacedFood is a food item sitting directly in the room (not inside a fridge).
type PlacedFood struct {
	X, Y, Z       uint8
	PurchaseDate  CompactDate
	UsesRemaining uint8
	Food          Food
}

// FloorItemKind tags what a FloorItem holds.
type FloorItemKind uint8

const (
	FloorKindFood FloorItemKind = iota
	FloorKindUtility
)

// FloorItem is a food or utility placed directly on the floor grid.
type FloorItem struct {
	X, Y, Z uint8
	Item    InventoryItem
}

type Home struct {
	Type       HomeType
	RoomItems  []PlacedRoomItem
	FloorItems []FloorItem // food & utilities on the floor
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
	CurrentDate CompactDate

	Food       NeedStat
	Energy     NeedStat
	Hygiene    NeedStat
	Confidence NeedStat
	Strength   NeedStat

	CurrentStats Stats
}
