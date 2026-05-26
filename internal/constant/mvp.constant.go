package constant

import (
	"basic-life-sim/internal/model"
)

var (
	// LEVEL - 1: INDEKOS (8x8)

	// Shared Bathroom, Shared Living Room, Shared Kitchen
	IndekosLayout = model.HomeLayout{
		{0, 0, 0, 0, 0, 2, 0, 0},
		{0, 1, 1, 1, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 1, 1, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
	// Private Bathroom, Shared Living Room, Shared Kitchen
	IndekosWithBathroomLayout = model.HomeLayout{
		{0, 0, 0, 0, 0, 2, 0, 0},
		{0, 1, 1, 1, 0, 1, 1, 0},
		{0, 1, 1, 1, 2, 1, 1, 0},
		{0, 0, 0, 0, 0, 1, 1, 0},
		{0, 1, 1, 1, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 1, 1, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}

	// LEVEL 2: SMALL APARTMENT (10 x 10)

	// Private Bathroom, Private Living Room, Private Kitchen
	SmallApartmentStudioLayout = model.HomeLayout{
		{0, 0, 0, 0, 0, 0, 0, 2, 0, 0},
		{0, 1, 1, 1, 1, 0, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 2, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 0, 1, 1, 1, 0},
		{0, 0, 0, 0, 0, 0, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 1, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 1, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 1, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 1, 1, 1, 1, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	// Private Bathroom, Private Living Room, Private Kitchen, Private Bedroom
	SmallApartment1BedroomLayout = model.HomeLayout{
		{0, 0, 0, 0, 0, 0, 0, 2, 0, 0},
		{0, 1, 1, 1, 1, 0, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 2, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 0, 1, 1, 1, 0},
		{0, 0, 0, 0, 0, 0, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 0, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 2, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 0, 1, 1, 1, 0},
		{0, 1, 1, 1, 1, 0, 1, 1, 1, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
)

const (
	HomeNameTypeIndekos                      = "Indekos"
	HomeNameTypeIndekosWithBathroom          = "Indekos with Bathroom"
	HomeNameTypeSmallApartmentStudio         = "Small Apartment Studio"
	HomeNameTypeSmallApartment1BedroomLayout = "Small Apartment 1 Bedroom"
)

var (
	Indekos = model.HomeType{
		Name:           HomeNameTypeIndekos,
		Layout:         IndekosLayout,
		Level:          1,
		MaxHeight:      3,
		SharedBathroom: true,
		SharedKitchen:  true,
	}
	IndekosWithBathroom = model.HomeType{
		Name:           HomeNameTypeIndekosWithBathroom,
		Layout:         IndekosWithBathroomLayout,
		MaxHeight:      3,
		Level:          1,
		SharedBathroom: false,
		SharedKitchen:  true,
	}
	SmallApartmentStudio = model.HomeType{
		Name:           HomeNameTypeSmallApartmentStudio,
		Layout:         SmallApartmentStudioLayout,
		MaxHeight:      5,
		Level:          2,
		SharedBathroom: false,
		SharedKitchen:  false,
	}
	SmallApartment1Bedroom = model.HomeType{
		Name:           HomeNameTypeSmallApartment1BedroomLayout,
		Layout:         SmallApartment1BedroomLayout,
		Level:          2,
		MaxHeight:      5,
		SharedBathroom: false,
		SharedKitchen:  false,
	}
)
