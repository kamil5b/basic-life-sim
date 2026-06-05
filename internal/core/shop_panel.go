package core

import (
	"fmt"
	"image/color"

	buyablefood "github.com/kamil5b/basic-life-sim/internal/buyable/food"
	roomitem "github.com/kamil5b/basic-life-sim/internal/buyable/room-item"
	"github.com/kamil5b/basic-life-sim/internal/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type spMode int

const (
	spModeCatalog      spMode = iota // browse catalog + select item
	spModePlaceGrid                  // click a cell on the room grid
	spModePlaceZ                     // pick Z height with +/- buttons
	spModePlaceDir                   // pick a facing direction
	spModeStackConfirm               // confirm mix vs separate stack
)

type spCategory int

type spSort int

const (
	spSortName spSort = iota
	spSortPrice
	spSortVolume
)

const itemsPerPage = 10

const (
	spCatAppliance spCategory = iota
	spCatFurniture
	spCatHygiene
	spCatFood
	spCatUtility
	spCatStorage
)

var spCatLabels = []string{"Appliances", "Furniture", "Hygiene", "Food", "Utilities", "Storage"}

type shopPanel struct {
	char    *model.Character
	main    *mainScreen
	cat     spCategory
	selItem int // index in full (unsorted) catalog, -1 = none
	mode    spMode

	// pagination + sort
	page    int
	sortBy  spSort
	sortAsc bool

	// placement state
	pendingItem    *model.RoomItem
	pendingFood    *model.Food
	pendingUtility *model.Utility
	// foodX/Y used only for floor-food fallback (food skips the 3-step wizard)
	foodX, foodY  uint8
	wizard        *placementWizard
	stackFloorIdx int // index in FloorFood of the food being stacked onto (-1 = none)

	// fridge placement state
	fridgeWiz *fridgeWizard
}

func newShopPanel(char *model.Character, main *mainScreen) *shopPanel {
	return &shopPanel{
		char:          char,
		main:          main,
		selItem:       -1,
		stackFloorIdx: -1,
	}
}

func (p *shopPanel) currentCatalog() []model.RoomItem {
	switch p.cat {
	case spCatAppliance:
		return roomitem.Appliances
	case spCatFurniture:
		return roomitem.Furniture
	case spCatHygiene:
		return roomitem.Hygiene
	case spCatStorage:
		return roomitem.Storage
	}
	return nil
}

func (p *shopPanel) update() {
	if p.fridgeWiz != nil {
		if p.fridgeWiz.update() {
			p.fridgeWiz = nil
			p.pendingFood = nil
			p.mode = spModeCatalog
		}
		return
	}

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
				p.page = 0
				return
			}
		}
		// sort buttons
		sortY := panelY + 36
		for si := 0; si < 3; si++ {
			sx := panelX + 4 + float32(si)*54
			if isHovered(mx, my, sx, sortY, 50, 16) {
				if spSort(si) == p.sortBy {
					p.sortAsc = !p.sortAsc
				} else {
					p.sortBy = spSort(si)
					p.sortAsc = true
				}
				p.page = 0
				p.selItem = -1
				return
			}
		}
		// page prev/next
		pageY := panelY + 52
		if isHovered(mx, my, panelX+4, pageY, 28, 22) && p.page > 0 {
			p.page--
			p.selItem = -1
			return
		}
		nx2 := panelX + spListW - 40
		if isHovered(mx, my, nx2, pageY, 28, 22) && p.page < p.totalPages()-1 {
			p.page++
			p.selItem = -1
			return
		}
		// item rows (page-relative → real index)
		page := p.catalogPage()
		rowBaseY := float32(panelY) + 78
		for pi, idx := range page {
			rx := float32(panelX) + 4
			ry := rowBaseY + float32(pi)*30
			rw := spListW - 8
			rh := float32(26)
			if isHovered(mx, my, rx, ry, rw, rh) {
				p.selItem = idx
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

	case spModeStackConfirm:
		if !clicked {
			return
		}
		f := *p.pendingFood
		char := p.char
		// Mix button
		mx2, my2, mw, mh := spStackMixBtnRect()
		if isHovered(mx, my, mx2, my2, mw, mh) && p.stackFloorIdx >= 0 {
			existing := &char.CurrentHome.FloorFood[p.stackFloorIdx]
			char.CurrentStats.Money -= f.BasePrice
			existing.UsesRemaining += f.UsesTotal
			p.main.setMessage(fmt.Sprintf("Mixed %s. Total uses: %d. Money: $%.2f",
				f.Name, existing.UsesRemaining, char.CurrentStats.Money))
			p.pendingFood = nil
			p.stackFloorIdx = -1
			p.mode = spModeCatalog
			return
		}
		// Place Separately button
		sx, sy, sw, sh := spStackSepBtnRect()
		if isHovered(mx, my, sx, sy, sw, sh) {
			nextZ := nextFoodZ(char, p.foodX, p.foodY)
			char.CurrentStats.Money -= f.BasePrice
			char.CurrentHome.FloorFood = append(char.CurrentHome.FloorFood, model.PlacedFood{
				X: p.foodX, Y: p.foodY, Z: nextZ,
				PurchaseDate:  char.CurrentDate,
				UsesRemaining: f.UsesTotal,
				Food:          f,
			})
			p.main.setMessage(fmt.Sprintf("Placed %s separately at z=%d. Money: $%.2f",
				f.Name, nextZ, char.CurrentStats.Money))
			p.pendingFood = nil
			p.stackFloorIdx = -1
			p.mode = spModeCatalog
			return
		}
		// Cancel
		cx2, cy2, cw, ch := spCancelBtnRect()
		if isHovered(mx, my, cx2, cy2, cw, ch) {
			p.cancelPlace()
			p.stackFloorIdx = -1
		}

	// ── pick Z / Dir: delegate to wizard ─────────────────────────────────────────────
	case spModePlaceZ, spModePlaceDir:
		if p.wizard != nil {
			p.wizard.update()
		}
	}
}

// tryBeginPlace validates affordability then switches to grid placement mode.
func (p *shopPanel) tryBeginPlace() {
	char := p.char
	if p.cat == spCatUtility {
		u := roomitem.Utilities[p.selItem]
		if char.CurrentStats.Money < u.BasePrice {
			p.main.setMessage(fmt.Sprintf("Not enough money. Need $%.2f", u.BasePrice))
			return
		}
		char.CurrentStats.Money -= u.BasePrice
		p.pendingUtility = &u
		// Utilities go to inventory — place on host later via Room panel.
		p.main.setMessage(fmt.Sprintf("Bought %s — place it on a stove/desk via Room tab. Money: $%.2f", u.Name, char.CurrentStats.Money))
		p.selItem = -1
		return
	}
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
	p.pendingUtility = nil
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

	// Check if clicked cell overlaps a fridge → store inside it
	fridgeAt := -1
	for _, fri := range findFridges(char) {
		placed := char.CurrentHome.RoomItems[fri]
		for _, cell := range occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction) {
			if cell[0] == p.foodX && cell[1] == p.foodY {
				fridgeAt = fri
				break
			}
		}
		if fridgeAt >= 0 {
			break
		}
	}

	if fridgeAt >= 0 {
		// Open the fridge interior UI so the player picks the slot manually.
		placed := char.CurrentHome.RoomItems[fridgeAt]
		if placed.Item.Storage.TotalSlots() <= usedSlotCount(&char.CurrentHome.RoomItems[fridgeAt]) {
			p.main.setMessage("Fridge is full! Choose another cell.")
			return
		}
		p.fridgeWiz = newFridgeWizard(
			char, fridgeAt,
			spGridOriginX(), spGridOriginY(), panelX+8,
			&f, true,
			p.main.setMessage,
			func() {
				p.fridgeWiz = nil
				p.main.setMessage("Cancelled. Click another cell.")
			},
			func() {
				p.fridgeWiz = nil
				p.pendingFood = nil
				p.mode = spModeCatalog
			},
		)
		return
	}

	// Check if there's already floor food at this cell
	stackIdx := -1
	for i, ff := range char.CurrentHome.FloorFood {
		if ff.X == p.foodX && ff.Y == p.foodY {
			stackIdx = i
			break
		}
	}

	if stackIdx >= 0 && f.CanBeMixed && char.CurrentHome.FloorFood[stackIdx].Food.Name == f.Name {
		// Same food type and mixable → ask user
		p.stackFloorIdx = stackIdx
		p.mode = spModeStackConfirm
		return
	}

	// Place on floor, stacking Z on top of any existing food
	nextZ := nextFoodZ(char, p.foodX, p.foodY)
	char.CurrentStats.Money -= f.BasePrice
	char.CurrentHome.FloorFood = append(char.CurrentHome.FloorFood, model.PlacedFood{
		X: p.foodX, Y: p.foodY, Z: nextZ,
		PurchaseDate:  char.CurrentDate,
		UsesRemaining: f.UsesTotal,
		Food:          f,
	})
	expiry := model.ExpiryDate(char.CurrentDate, f.BaseExpiryDays, 1)
	ey, em, ed := expiry.Unpack()
	p.main.setMessage(fmt.Sprintf("Bought %s → floor (%d,%d,z=%d), exp %04d-%02d-%02d. Money: $%.2f",
		f.Name, p.foodX, p.foodY, nextZ, ey, em, ed, char.CurrentStats.Money))
	p.pendingFood = nil
	p.stackFloorIdx = -1
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

func spStackMixBtnRect() (x, y, w, h float32) {
	return panelX + 4, panelY + 80, 160, 44
}

func spStackSepBtnRect() (x, y, w, h float32) {
	return panelX + 4, panelY + 136, 200, 44
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

func (p *shopPanel) fullLen() int {
	switch p.cat {
	case spCatFood:
		return len(buyablefood.All)
	case spCatUtility:
		return len(roomitem.Utilities)
	default:
		cat := p.currentCatalog()
		if cat == nil {
			return 0
		}
		return len(cat)
	}
}

func (p *shopPanel) totalPages() int {
	n := p.fullLen()
	if n == 0 {
		return 1
	}
	return (n + itemsPerPage - 1) / itemsPerPage
}

// catalogPage returns the indices into the full catalog for the current page, sorted.
func (p *shopPanel) catalogPage() []int {
	n := p.fullLen()
	if n == 0 {
		return nil
	}
	// Build index slice
	idxs := make([]int, n)
	for i := range idxs {
		idxs[i] = i
	}
	// Sort
	p.sortIndices(idxs)
	// Paginate
	start := p.page * itemsPerPage
	if start >= n {
		p.page = (n - 1) / itemsPerPage
		start = p.page * itemsPerPage
	}
	end := start + itemsPerPage
	if end > n {
		end = n
	}
	return idxs[start:end]
}

func (p *shopPanel) sortIndices(idxs []int) {
	get := func(i int) (name string, price float64, vol uint) {
		switch p.cat {
		case spCatFood:
			f := buyablefood.All[i]
			return f.Name, f.BasePrice, uint(f.Width) * uint(f.Length) * uint(f.Height)
		case spCatUtility:
			u := roomitem.Utilities[i]
			return u.Name, u.BasePrice, uint(u.Width) * uint(u.Length) * uint(u.Height)
		default:
			cat := p.currentCatalog()
			if cat != nil && i < len(cat) {
				it := cat[i]
				return it.Name, it.BasePrice, uint(it.Width) * uint(it.Length) * uint(it.Height)
			}
		}
		return
	}

	// Bubble sort — small N, fine for UI
	for a := 0; a < len(idxs)-1; a++ {
		for b := a + 1; b < len(idxs); b++ {
			na, pa, va := get(idxs[a])
			nb, pb, vb := get(idxs[b])
			less := false
			switch p.sortBy {
			case spSortName:
				less = na < nb
			case spSortPrice:
				less = pa < pb
			case spSortVolume:
				less = va < vb
			}
			if !p.sortAsc {
				less = !less
			}
			if less {
				idxs[a], idxs[b] = idxs[b], idxs[a]
			}
		}
	}
}

func (p *shopPanel) visibleCatalogLen() int {
	return len(p.catalogPage())
}

// ── draw ──────────────────────────────────────────────────────────────────────

func (p *shopPanel) draw(dst *ebiten.Image) {
	if p.fridgeWiz != nil {
		p.fridgeWiz.draw(dst)
		return
	}

	mx, my := ebiten.CursorPosition()

	// Left panel: catalog in catalog/grid mode; wizard draws Z/dir picker
	switch p.mode {
	case spModeCatalog, spModePlaceGrid:
		p.drawCatalog(dst, mx, my)
	case spModeStackConfirm:
		p.drawStackConfirm(dst, mx, my)
	case spModePlaceZ, spModePlaceDir:
		if p.wizard != nil {
			p.wizard.drawLeftPanel(dst, mx, my)
		}
	}

	// Right panel: room grid always visible
	p.drawPlacementGrid(dst, mx, my)

	// Cancel button in placement modes (top-right of left panel)
	if p.mode == spModePlaceGrid || p.mode == spModePlaceZ || p.mode == spModePlaceDir || p.mode == spModeStackConfirm {
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

	// sort bar
	sortNames := []string{"Name", "Price", "Vol"}
	sortY := panelY + 36
	for si, sn := range sortNames {
		sx := panelX + 4 + float32(si)*54
		sw := float32(50)
		sh := float32(16)
		active := spSort(si) == p.sortBy
		hov := isHovered(mx, my, sx, sortY, sw, sh)
		bg := colorPanel
		if active {
			bg = colorSelected
		} else if hov {
			bg = colorHighlight
		}
		fillRect(dst, sx, sortY, sw, sh, bg)
		strokeRect(dst, sx, sortY, sw, sh, colorBorder)
		arr := ""
		if active {
			if p.sortAsc {
				arr = "▲"
			} else {
				arr = "▼"
			}
		}
		drawText(dst, arr+sn, float64(sx)+2, float64(sortY)+4, fontS, colorText)
	}

	// page controls
	pageY := panelY + 52
	totalP := p.totalPages()
	// prev
	px := panelX + 4
	drawButton(dst, "◀", px, pageY, 28, 22, fontS, p.page > 0 && isHovered(mx, my, px, pageY, 28, 22), p.page > 0)
	// page label
	pageLabel := fmt.Sprintf("Pg %d/%d", p.page+1, totalP)
	drawText(dst, pageLabel, float64(panelX)+40, float64(pageY)+5, fontS, colorText)
	// next
	nx := panelX + float32(spListW) - 40
	drawButton(dst, "▶", nx, pageY, 28, 22, fontS, p.page < totalP-1 && isHovered(mx, my, nx, pageY, 28, 22), p.page < totalP-1)

	// items (page-relative)
	page := p.catalogPage()
	rowBaseY := panelY + 78
	for pi, idx := range page {
		if p.cat == spCatFood {
			f := buyablefood.All[idx]
			p.drawItemRowAt(dst, mx, my, pi, idx, f.Name,
				fmt.Sprintf("$%.2f  [%s]  uses:%d  exp:%dd", f.BasePrice, f.Type, f.UsesTotal, f.BaseExpiryDays),
				char.CurrentStats.Money >= f.BasePrice, rowBaseY)
		} else if p.cat == spCatUtility {
			u := roomitem.Utilities[idx]
			detail := fmt.Sprintf("$%.2f  %dx%d", u.BasePrice, u.Width, u.Length)
			if u.Ability != "" {
				detail += fmt.Sprintf("  ab:%s", u.Ability)
			}
			if u.CookSurface != nil {
				detail += fmt.Sprintf("  cook:%d", u.CookSurface.Slots)
			}
			p.drawItemRowAt(dst, mx, my, pi, idx, u.Name, detail,
				char.CurrentStats.Money >= u.BasePrice, rowBaseY)
		} else {
			cat := p.currentCatalog()
			if cat != nil && idx < len(cat) {
				item := cat[idx]
				p.drawItemRowAt(dst, mx, my, pi, idx, item.Name,
					fmt.Sprintf("$%.2f  %dx%d", item.BasePrice, item.Width, item.Length),
					char.CurrentStats.Money >= item.BasePrice, rowBaseY)
			}
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

	// detail row for selected item (below items)
	if p.selItem >= 0 {
		var actions []string
		if p.cat == spCatFood {
			if p.selItem < len(buyablefood.All) {
				f := buyablefood.All[p.selItem]
				actions = []string{fmt.Sprintf("Type: %s  |  Uses: %d  |  Expires in %d days", f.Type, f.UsesTotal, f.BaseExpiryDays)}
			}
		} else if p.cat == spCatUtility {
			if p.selItem < len(roomitem.Utilities) {
				u := roomitem.Utilities[p.selItem]
				actions = u.Actions
			}
		} else {
			cat := p.currentCatalog()
			if cat != nil && p.selItem < len(cat) {
				item := cat[p.selItem]
				actions = item.Actions
			}
		}
		iy := float64(panelY) + 78 + float64(itemsPerPage)*32 + 8
		for _, a := range actions {
			drawText(dst, "  "+a, float64(panelX)+8, iy, fontS, colorMuted)
			iy += 18
		}
	}
}

func (p *shopPanel) drawItemRow(dst *ebiten.Image, mx, my, i int, name, detail string, affordable bool) {
	rx, ry, rw, rh := spItemRowRect(i)
	p.drawRow(dst, mx, my, i, name, detail, affordable, rx, ry, rw, rh)
}

func (p *shopPanel) drawItemRowAt(dst *ebiten.Image, mx, my, pi, realIdx int, name, detail string, affordable bool, baseY float32) {
	rx := panelX + 4
	ry := baseY + float32(pi)*30
	rw := spListW - 8
	rh := float32(26)
	// Use realIdx for selection check
	p.drawRow(dst, mx, my, realIdx, name, detail, affordable, rx, ry, rw, rh)
}

func (p *shopPanel) drawRow(dst *ebiten.Image, mx, my, idx int, name, detail string, affordable bool, rx, ry, rw, rh float32) {
	sel := idx == p.selItem
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

func (p *shopPanel) drawStackConfirm(dst *ebiten.Image, mx, my int) {
	if p.pendingFood == nil || p.stackFloorIdx < 0 {
		return
	}
	f := *p.pendingFood
	existing := p.char.CurrentHome.FloorFood[p.stackFloorIdx]
	lx := float64(panelX) + 12
	drawText(dst, "Stack Food", lx, float64(panelY)+16, fontM, colorAccent)
	drawText(dst, fmt.Sprintf("Buying: %s (uses: %d)", f.Name, f.UsesTotal),
		lx, float64(panelY)+44, fontS, colorText)
	drawText(dst, fmt.Sprintf("On floor: %s (uses: %d)", existing.Food.Name, existing.UsesRemaining),
		lx, float64(panelY)+62, fontS, colorMuted)
	drawText(dst, "Combine into one stack?", lx, float64(panelY)+82, fontS, colorText)

	mx2, my2, mw, mh := spStackMixBtnRect()
	drawButton(dst, "Mix Together", mx2, my2, mw, mh, fontM, isHovered(mx, my, mx2, my2, mw, mh), true)

	sx, sy, sw, sh := spStackSepBtnRect()
	drawButton(dst, "Place Separately", sx, sy, sw, sh, fontM, isHovered(mx, my, sx, sy, sw, sh), true)

	cx2, cy2, cw, ch := spCancelBtnRect()
	drawButton(dst, "✕ Cancel", cx2, cy2, cw, ch, fontS, isHovered(mx, my, cx2, cy2, cw, ch), true)
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
