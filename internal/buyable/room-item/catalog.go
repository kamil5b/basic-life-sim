package roomitem

import (
	"github.com/kamil5b/basic-life-sim/internal/buyable/room-item/appliance"
	"github.com/kamil5b/basic-life-sim/internal/buyable/room-item/furniture"
	"github.com/kamil5b/basic-life-sim/internal/buyable/room-item/hygiene"
	"github.com/kamil5b/basic-life-sim/internal/buyable/room-item/storage"
	"github.com/kamil5b/basic-life-sim/internal/buyable/room-item/utility"
	"github.com/kamil5b/basic-life-sim/internal/model"
)

var Appliances = []model.RoomItem{
	appliance.MiniRefrigerator,
	appliance.StandardRefrigerator,
	appliance.SingleBurnerStove,
	appliance.GasStove,
	appliance.SmallTV,
	appliance.SmartTV,
}

var Furniture = []model.RoomItem{
	furniture.SingleBed,
	furniture.QueenBed,
	furniture.KingBed,
	furniture.BasicChair,
	furniture.GamingChair,
	furniture.BasicDesk,
	furniture.ComputerDesk,
}

var Hygiene = []model.RoomItem{
	hygiene.BasicShower,
	hygiene.PremiumShower,
	hygiene.BasicSink,
	hygiene.VanitySink,
	hygiene.BasicToilet,
	hygiene.BidetToilet,
}

var Utilities = utility.All

func init() {
	for _, item := range Appliances {
		model.RegisterRoomItem(item)
	}
	for _, item := range Furniture {
		model.RegisterRoomItem(item)
	}
	for _, item := range Hygiene {
		model.RegisterRoomItem(item)
	}
	for _, item := range Storage {
		model.RegisterRoomItem(item)
	}
}

var Storage = []model.RoomItem{
	storage.BasicCupboard,
	storage.DoubleCupboard,
	storage.TallPantry,
	storage.WallShelfSmall,
	storage.WallShelfMedium,
	storage.WallShelfLong,
	storage.BasicRack,
	storage.WideRack,
	storage.ShelfDesk,
	storage.RackDesk,
}
