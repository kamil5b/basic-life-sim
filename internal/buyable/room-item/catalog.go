package roomitem

import (
	"github.com/kamil5b/basic-life-sim/internal/buyable/room-item/appliance"
	"github.com/kamil5b/basic-life-sim/internal/buyable/room-item/furniture"
	"github.com/kamil5b/basic-life-sim/internal/buyable/room-item/hygiene"
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
