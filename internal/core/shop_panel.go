package core

import (
	buyablefood "basic-life-sim/internal/buyable/food"
	roomitem "basic-life-sim/internal/buyable/room-item"
	"basic-life-sim/internal/model"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type spMode int

const (
	spModeCatalog   spMode = iota // browse catalog + select item
	spModePlaceGrid               // click a cell on the room grid
	spModePlaceZ                  // pick Z height with +/- buttons
	spModePlaceDir                // pick a facing direction
)

type spCategory int

const (
	spCatAppliance spCategory = iota
	spCatFurniture
	spCatHygiene
	spCatFood
)

var spCatLabels = []string{"Appliances", "Furniture", "Hygiene", "Food"}

type shopPanel struct {
	char    *model.Character
	main    *mainScreen
	cat     spCategory
	selItem int // index in current catalog, -1 = none
	mode    spMode

	// placement state
	pendingItem *model.RoomItem
	pendingFood *model.Food
	// foodX/Y used only for floor-food fallback (food skips the 3-step wizard)
	foodX, foodY uint8
	wizard       *placementWizard
}

func newShopPanel(char *model.Character, main *mainScreen) *shopPanel {
	return &shopPanel{char: char, main: main, selItem: -1}
}

func (p *shopPanel) currentCatalog() []model.RoomItem {
	switch p.cat {
	case spCatAppliance:
		return roomitem.Appliances
	case spCatFurniture:
		return roomitem.Furniture
	case spCatHygiene:
		return roomitem.Hygiene
	}
	return nil
}

func (p *shopPanel) update() {
	mx, my := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	switch p.mode {
	// ── catalog browse ────────────────────────────────────────────────────────
	case spModeCatalog:
		if !clicked {
			return
		}
		// category tabs
		for i := range spCatLabels {
			tx, ty, tw, th := spCatTabRect(i)
			if isHovered(mx, my, tx, ty, tw, th) {
				p.cat = spCategory(i)
				p.selItem = -1
				return
			}
		}
		// item rows
		catalog := p.visibleCatalogLen()
		for i := 0; i < catalog; i++ {
			rx, ry, rw, rh := spItemRowRect(i)
			if isHovered(mx, my, rx, ry, rw, rh) {
				p.selItem = i
				return
			}
		}
		// buy button
		bx, by, bw, bh := spBuyBtnRect()
		if isHovered(mx, my, bx, by, bw, bh) && p.selItem >= 0 {
			p.tryBeginPlace()
		}

	// ── click on grid cell (food only — items use wizard) ────────────────────
	case spModePlaceGrid:
		if p.wizard != nil {
			// wizard owns cancel + all 3 steps for items
			p.wizard.update()
			if p.wizard != nil {
				switch p.wizard.step() {
				case pwStepZ:
					p.mode = spModePlaceZ
				case pwStepDir:
					p.mode = spModePlaceDir
				}
			}
			return
		}
		// food path: simple one-click grid placement
		cx2, cy2, cw, ch := spCancelBtnRect()
		if clicked && isHovered(mx, my, cx2, cy2, cw, ch) {
			p.cancelPlace()
			return
		}
		if clicked {
			gx, gy := gridCellAtOrigin(mx, my, spGridOriginX(), spGridOriginY(), p.char)
			if gx >= 0 {
				p.foodX = uint8(gx)
				p.foodY = uint8(gy)
				p.finalizeFoodPlace()
			}
		}

	// ── pick Z / Dir: delegate to wizard ─────────────────────────────────────
	case spModePlaceZ, spModePlaceDir:
		if p.wizard != nil {
			p.wizard.update()
		}
	}
}

// tryBeginPlace validates affordability then switches to grid placement mode.
func (p *shopPanel) tryBeginPlace() {
	char := p.char
	if p.cat == spCatFood {
		f := buyablefood.All[p.selItem]
		if char.CurrentStats.Money < f.BasePrice {
			p.main.setMessage(fmt.Sprintf("Not enough money. Need $%.2f", f.BasePrice))
			return
		}
		p.pendingFood = &buyablefood.All[p.selItem]
		p.pendingItem = nil
		p.wizard = nil
	} else {
		cat := p.currentCatalog()
		item := cat[p.selItem]
		if char.CurrentStats.Money < item.BasePrice {
			p.main.setMessage(fmt.Sprintf("Not enough money. Need $%.2f", item.BasePrice))
			return
		}
		p.pendingItem = &cat[p.selItem]
		p.pendingFood = nil
		p.wizard = newPlacementWizard(
			p.char, p.pendingItem, -1,
			spGridOriginX(), spGridOriginY(),
			"✕ Cancel",
			func(x, y, z uint8, dir model.Direction) error {
				return p.finalizeItemPlace(x, y, z, dir)
			},
			p.cancelPlace,
		)
	}
	p.mode = spModePlaceGrid
}

func (p *shopPanel) cancelPlace() {
	p.pendingItem = nil
	p.pendingFood = nil
	p.wizard = nil
	p.mode = spModeCatalog

	p.main.setMessage("Placement cancelled.")
}

func (p *shopPanel) finalizeItemPlace(x, y, z uint8, dir model.Direction) error {
	if p.pendingItem == nil {
		p.mode = spModeCatalog
		return nil
	}
	char := p.char
	item := *p.pendingItem

	if err := canPlace(char.CurrentHome, item, x, y, z, dir); err != nil {
		p.main.setMessage("Can't place: " + err.Error())
		return err
	}

	char.CurrentStats.Money -= item.BasePrice
	char.CurrentHome.RoomItems = append(char.CurrentHome.RoomItems, model.PlacedRoomItem{
		X: x, Y: y, Z: z,
		Direction: dir,
		Item:      item,
	})
	p.main.setMessage(fmt.Sprintf("Placed %s at (%d,%d,z=%d). Money: $%.2f", item.Name, x, y, z, char.CurrentStats.Money))
	p.pendingItem = nil
	p.wizard = nil
	p.mode = spModeCatalog
	return nil
}

func (p *shopPanel) finalizeFoodPlace() {
	if p.pendingFood == nil {
		p.mode = spModeCatalog
		return
	}
	char := p.char
	f := *p.pendingFood

	// auto-slot into fridge first
	fridges := findFridges(char)
	if len(fridges) > 0 {
		placed := &char.CurrentHome.RoomItems[fridges[0]]
		cap := placed.Item.Storage
		slot, ok := nextFreeSlot(placed, cap, f)
		if ok {
			char.CurrentStats.Money -= f.BasePrice
			mult := cap.MultiplierAt(slot[0], slot[1], slot[2])
			placed.Stored = append(placed.Stored, model.StoredFood{
				SlotX: slot[0], SlotY: slot[1], SlotZ: slot[2],
				PurchaseDate:   char.CurrentDate,
				MultiplierUsed: mult,
				UsesRemaining:  f.UsesTotal,
				Food:           f,
			})
			expiry := model.ExpiryDate(char.CurrentDate, f.BaseExpiryDays, mult)
			ey, em, ed := expiry.Unpack()
			p.main.setMessage(fmt.Sprintf("Bought %s → fridge, exp %04d-%02d-%02d. Money: $%.2f",
				f.Name, ey, em, ed, char.CurrentStats.Money))
			p.pendingFood = nil
			p.mode = spModeCatalog
			return
		}
	}

	// fall back to clicked floor cell
	char.CurrentStats.Money -= f.BasePrice
	char.CurrentHome.FloorFood = append(char.CurrentHome.FloorFood, model.PlacedFood{
		X: p.foodX, Y: p.foodY, Z: 0,
		PurchaseDate:  char.CurrentDate,
		UsesRemaining: f.UsesTotal,
		Food:          f,
	})
	expiry := model.ExpiryDate(char.CurrentDate, f.BaseExpiryDays, 1)
	ey, em, ed := expiry.Unpack()
	p.main.setMessage(fmt.Sprintf("Bought %s → floor (%d,%d), exp %04d-%02d-%02d. Money: $%.2f",
		f.Name, p.foodX, p.foodY, ey, em, ed, char.CurrentStats.Money))
	p.pendingFood = nil
	p.mode = spModeCatalog
}

// ── layout helpers ────────────────────────────────────────────────────────────

const spListW = float32(340) // catalog list occupies the left slice of the panel

func spCatTabRect(i int) (x, y, w, h float32) {
	w = (spListW - 8) / float32(len(spCatLabels))
	h = 30
	x = panelX + 4 + float32(i)*w
	y = panelY + 4
	return
}

func spItemRowRect(i int) (x, y, w, h float32) {
	return panelX + 4, panelY + 42 + float32(i)*32, spListW - 8, 28
}

func spBuyBtnRect() (x, y, w, h float32) {
	return panelX + 4, float32(ScreenH) - 60, 160, 38
}

// Cancel sits top-right of the panel, never overlapping action buttons.
func spCancelBtnRect() (x, y, w, h float32) {
	return panelX + spListW - 100, panelY + 4, 96, 28
}

var dirBtnLabels = []string{"↑ N", "→ E", "↓ S", "← W"}

// Z column visualiser — drawn inside the left panel area (panelX+4 to panelX+spListW)
const (
	zCellW = float32(48)
	zCellH = float32(38)
	zColX  = panelX + 60 // inset from left panel edge
	zColY  = panelY + 60 // below the title
)

func spZCellRect(zLevel, maxHeight int) (x, y, w, h float32) {
	// z=0 is at the bottom, z=maxHeight-1 at the top
	row := maxHeight - 1 - zLevel
	return zColX, zColY + float32(row)*zCellH, zCellW, zCellH - 2
}

func spZConfirmBtnRect() (x, y, w, h float32) {
	// below the Z column, in left panel
	return panelX + 60, float32(ScreenH) - 80, 180, 38
}

func spDirBtnRect(i int) (x, y, w, h float32) {
	// 2×2 grid inside left panel
	w, h = 88, 44
	col := float32(i % 2)
	row := float32(i / 2)
	x = panelX + 20 + col*(w+8)
	y = panelY + 120 + row*(h+10)
	return
}

// spGridOriginX/Y return the top-left pixel of the room grid in the shop panel.
// We render it to the right of the catalog list, reusing rpCellSz.
func spGridOriginX() float32 { return panelX + spListW + 16 }
func spGridOriginY() float32 { return panelY + 30 }

func (p *shopPanel) visibleCatalogLen() int {
	if p.cat == spCatFood {
		return len(buyablefood.All)
	}
	cat := p.currentCatalog()
	if cat == nil {
		return 0
	}
	return len(cat)
}

// ── draw ──────────────────────────────────────────────────────────────────────

func (p *shopPanel) draw(dst *ebiten.Image) {
	mx, my := ebiten.CursorPosition()

	// Left panel: catalog in catalog/grid mode; wizard draws Z/dir picker
	switch p.mode {
	case spModeCatalog, spModePlaceGrid:
		p.drawCatalog(dst, mx, my)
	case spModePlaceZ, spModePlaceDir:
		if p.wizard != nil {
			p.wizard.drawLeftPanel(dst, mx, my)
		}
	}

	// Right panel: room grid always visible
	p.drawPlacementGrid(dst, mx, my)

	// Cancel button in placement modes (top-right of left panel)
	if p.mode == spModePlaceGrid || p.mode == spModePlaceZ || p.mode == spModePlaceDir {
		cx2, cy2, cw, ch := spCancelBtnRect()
		drawButton(dst, "✕ Cancel", cx2, cy2, cw, ch, fontS, isHovered(mx, my, cx2, cy2, cw, ch), true)
	}
}

func (p *shopPanel) drawCatalog(dst *ebiten.Image, mx, my int) {
	char := p.char

	// category tabs
	for i, lbl := range spCatLabels {
		tx, ty, tw, th := spCatTabRect(i)
		active := spCategory(i) == p.cat
		hov := isHovered(mx, my, tx, ty, tw, th)
		bg := colorPanel
		border := colorBorder
		tc := colorMuted
		if active {
			bg = colorSelected
			border = colorAccent
			tc = colorText
		} else if hov {
			bg = colorHighlight
		}
		fillRect(dst, tx, ty, tw, th, bg)
		strokeRect(dst, tx, ty, tw, th, border)
		lw, _ := text.Measure(lbl, fontS, 0)
		drawText(dst, lbl, float64(tx)+float64(tw)/2-lw/2, float64(ty)+8, fontS, tc)
	}

	if p.cat == spCatFood {
		for i, f := range buyablefood.All {
			p.drawItemRow(dst, mx, my, i, f.Name,
				fmt.Sprintf("$%.2f  [%s]  uses:%d  exp:%dd", f.BasePrice, f.Type, f.UsesTotal, f.BaseExpiryDays),
				char.CurrentStats.Money >= f.BasePrice)
		}
	} else {
		cat := p.currentCatalog()
		for i, item := range cat {
			p.drawItemRow(dst, mx, my, i, item.Name,
				fmt.Sprintf("$%.2f  %dx%d", item.BasePrice, item.Width, item.Length),
				char.CurrentStats.Money >= item.BasePrice)
		}
	}

	// buy button (only in catalog mode)
	if p.mode == spModeCatalog {
		bx, by, bw, bh := spBuyBtnRect()
		enabled := p.selItem >= 0
		drawButton(dst, "Buy & Place →", bx, by, bw, bh, fontM,
			isHovered(mx, my, bx, by, bw, bh) && enabled, enabled)
		drawText(dst, fmt.Sprintf("$%.2f", char.CurrentStats.Money),
			float64(bx)+float64(bw)+12, float64(by)+10, fontM, colorGreen)
	}

	// detail row for selected item
	if p.selItem >= 0 {
		var actions []string
		if p.cat == spCatFood {
			if p.selItem < len(buyablefood.All) {
				f := buyablefood.All[p.selItem]
				actions = []string{fmt.Sprintf("Type: %s  |  Uses: %d  |  Expires in %d days", f.Type, f.UsesTotal, f.BaseExpiryDays)}
			}
		} else {
			cat := p.currentCatalog()
			if cat != nil && p.selItem < len(cat) {
				item := cat[p.selItem]
				actions = item.Actions
			}
		}
		iy := float64(panelY) + 42 + float64(p.visibleCatalogLen())*32 + 8
		for _, a := range actions {
			drawText(dst, "  "+a, float64(panelX)+8, iy, fontS, colorMuted)
			iy += 18
		}
	}
}

func (p *shopPanel) drawItemRow(dst *ebiten.Image, mx, my, i int, name, detail string, affordable bool) {
	rx, ry, rw, rh := spItemRowRect(i)
	sel := i == p.selItem
	hov := isHovered(mx, my, rx, ry, rw, rh)
	bg := colorPanel
	if sel {
		bg = colorSelected
	} else if hov {
		bg = colorHighlight
	}
	fillRect(dst, rx, ry, rw, rh, bg)
	strokeRect(dst, rx, ry, rw, rh, colorBorder)
	tc := colorText
	if !affordable {
		tc = colorMuted
	}
	drawText(dst, name, float64(rx)+8, float64(ry)+7, fontS, tc)
	dw, _ := text.Measure(detail, fontS, 0)
	drawText(dst, detail, float64(rx)+float64(rw)-dw-8, float64(ry)+7, fontS, colorMuted)
}

func (p *shopPanel) drawPlacementGrid(dst *ebiten.Image, mx, my int) {
	home := p.char.CurrentHome
	layout := home.Type.Layout
	if len(layout) == 0 {
		return
	}
	ox := spGridOriginX()
	oy := spGridOriginY()
	rows := len(layout)
	cols := len(layout[0])

	// header
	label := "Step 1: Click a floor cell to place"
	if p.mode == spModeCatalog {
		label = "Room"
	} else if p.mode == spModePlaceZ && p.wizard != nil {
		label = fmt.Sprintf("Placing at (%d,%d) — pick Z on the left", p.wizard.x, p.wizard.y)
	} else if p.mode == spModePlaceDir && p.wizard != nil {
		label = fmt.Sprintf("Placing at (%d,%d) z=%d — pick direction on the left", p.wizard.x, p.wizard.y, p.wizard.z)
	}
	drawText(dst, label, float64(ox), float64(oy)-18, fontS, colorMuted)

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			cx := ox + float32(c)*rpCellSz
			cy := oy + float32(r)*rpCellSz
			var bg color.RGBA
			switch layout[r][c] {
			case model.HomeCellWall:
				bg = colorWall
			case model.HomeCellFloor:
				bg = colorFloor
			case model.HomeCellDoor:
				bg = colorDoor
			}
			fillRect(dst, cx, cy, rpCellSz-1, rpCellSz-1, bg)
		}
	}

	// existing items overlay
	for _, placed := range home.RoomItems {
		for _, cell := range occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction) {
			cx := ox + float32(cell[0])*rpCellSz
			cy := oy + float32(cell[1])*rpCellSz
			fillRect(dst, cx+1, cy+1, rpCellSz-3, rpCellSz-3, colorItem)
		}
		lx := ox + float32(placed.X)*rpCellSz + 4
		ly := oy + float32(placed.Y)*rpCellSz + 4
		drawText(dst, string([]rune(placed.Item.Name)[0:1]), float64(lx), float64(ly), fontS, colorBg)
		drawFacingArrow(dst, ox, oy, int(placed.X), int(placed.Y), placed.Direction)
	}

	// floor food overlay
	for _, ff := range home.FloorFood {
		cx := ox + float32(ff.X)*rpCellSz
		cy := oy + float32(ff.Y)*rpCellSz
		fillRect(dst, cx+8, cy+8, rpCellSz-17, rpCellSz-17, colorFoodFloor)
	}

	// placement mode: ghost footprint
	if p.mode == spModePlaceGrid || p.mode == spModePlaceZ || p.mode == spModePlaceDir {
		if p.wizard != nil {
			p.wizard.drawGhost(dst, ox, oy)
		} else if p.pendingFood != nil {
			// food: single cell highlight at cursor
			hx, hy := gridCellAtOrigin(mx, my, ox, oy, p.char)
			if hx >= 0 {
				gcx := ox + float32(hx)*rpCellSz
				gcy := oy + float32(hy)*rpCellSz
				fillRect(dst, gcx+1, gcy+1, rpCellSz-3, rpCellSz-3, color.RGBA{100, 200, 100, 120})
			}
		}
	}

	// hover cursor cell highlight
	if p.mode == spModePlaceGrid {
		hx2, hy2 := gridCellAtOrigin(mx, my, ox, oy, p.char)
		if hx2 >= 0 {
			cx := ox + float32(hx2)*rpCellSz
			cy := oy + float32(hy2)*rpCellSz
			strokeRect(dst, cx, cy, rpCellSz-1, rpCellSz-1, colorAccent)
		}
	}
}
