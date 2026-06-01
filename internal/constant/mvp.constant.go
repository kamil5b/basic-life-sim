package constant

import (
	"github.com/kamil5b/basic-life-sim/internal/model"
)

var (
	// LEVEL 1: INDEKOS (8x8)

	// Shared Bathroom, Shared Living Room, Shared Kitchen
	IndekosLayout = model.HomeLayout{
		{model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellDoor, model.HomeCellWall, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall},
	}
	// Private Bathroom, Shared Living Room, Shared Kitchen
	IndekosWithBathroomLayout = model.HomeLayout{
		{model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellDoor, model.HomeCellWall, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellDoor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall},
	}

	// LEVEL 2: SMALL APARTMENT (10x10)

	// Private Bathroom, Private Living Room, Private Kitchen
	SmallApartmentStudioLayout = model.HomeLayout{
		{model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellDoor, model.HomeCellWall, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellDoor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall},
	}

	// Private Bathroom, Private Living Room, Private Kitchen, Private Bedroom
	SmallApartment1BedroomLayout = model.HomeLayout{
		{model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellDoor, model.HomeCellWall, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellDoor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellDoor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellFloor, model.HomeCellWall},
		{model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall, model.HomeCellWall},
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
		UpfrontCost:    500,
		MonthlyCost:    300,
	}
	IndekosWithBathroom = model.HomeType{
		Name:           HomeNameTypeIndekosWithBathroom,
		Layout:         IndekosWithBathroomLayout,
		MaxHeight:      3,
		Level:          1,
		SharedBathroom: false,
		SharedKitchen:  true,
		UpfrontCost:    750,
		MonthlyCost:    450,
	}
	SmallApartmentStudio = model.HomeType{
		Name:           HomeNameTypeSmallApartmentStudio,
		Layout:         SmallApartmentStudioLayout,
		MaxHeight:      5,
		Level:          2,
		SharedBathroom: false,
		SharedKitchen:  false,
		UpfrontCost:    2000,
		MonthlyCost:    800,
	}
	SmallApartment1Bedroom = model.HomeType{
		Name:           HomeNameTypeSmallApartment1BedroomLayout,
		Layout:         SmallApartment1BedroomLayout,
		Level:          2,
		MaxHeight:      5,
		SharedBathroom: false,
		SharedKitchen:  false,
		UpfrontCost:    3000,
		MonthlyCost:    1100,
	}
)

var (
	Level1HomeTypes = []model.HomeType{
		Indekos,
		IndekosWithBathroom,
	}
	Level2HomeTypes = []model.HomeType{
		SmallApartmentStudio,
		SmallApartment1Bedroom,
	}
)
