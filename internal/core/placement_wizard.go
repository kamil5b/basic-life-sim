package core

import (
	"fmt"
	"image/color"

	"github.com/kamil5b/basic-life-sim/internal/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// gridCellAtOrigin converts screen pixels to grid coords using explicit origin.
func gridCellAtOrigin(px, py int, ox, oy float32, char *model.Character) (gx, gy int) {
	cs := int(rpCellSz)
	layout := char.CurrentHome.Type.Layout
	rows := len(layout)
	if rows == 0 {
		return -1, -1
	}
	cols := len(layout[0])
	gx = (px - int(ox)) / cs
	gy = (py - int(oy)) / cs
	if gx < 0 || gy < 0 || gx >= cols || gy >= rows {
		return -1, -1
	}
	return gx, gy
}

type pwStep int

const (
	pwStepGrid pwStep = iota
	pwStepZ
	pwStepDir
)

type placementWizard struct {
	// config (set at construction, never changed)
	char        *model.Character
	item        *model.RoomItem // item being placed/moved
	excludeIdx  int             // index of item to exclude from collision (-1 = none)
	gridOriginX float32
	gridOriginY float32
	panelOX     float32 // left panel X origin
	panelOY     float32 // left panel Y origin
	cancelLabel string
	onFinalize  func(x, y, z uint8, dir model.Direction) error
	onCancel    func()

	// mutable state
	x, y, z uint8
	dir     model.Direction
	hoverX  int
	hoverY  int
	err     string
	curStep pwStep
}

func newPlacementWizard(
	char *model.Character,
	item *model.RoomItem,
	excludeIdx int,
	gridOriginX, gridOriginY float32,
	panelOX, panelOY float32,
	cancelLabel string,
	onFinalize func(x, y, z uint8, dir model.Direction) error,
	onCancel func(),
) *placementWizard {
	return &placementWizard{
		char:        char,
		item:        item,
		excludeIdx:  excludeIdx,
		gridOriginX: gridOriginX,
		gridOriginY: gridOriginY,
		panelOX:     panelOX,
		panelOY:     panelOY,
		cancelLabel: cancelLabel,
		onFinalize:  onFinalize,
		onCancel:    onCancel,
		hoverX:      -1,
		hoverY:      -1,
		dir:         model.North,
		curStep:     pwStepGrid,
	}
}

func (w *placementWizard) reset() {
	w.x, w.y, w.z = 0, 0, 0
	w.dir = model.North
	w.hoverX, w.hoverY = -1, -1
	w.err = ""
	w.curStep = pwStepGrid
}

func (w *placementWizard) step() pwStep { return w.curStep }

// gridCellAt converts screen pixels to grid cell coords using the wizard's origin.
func (w *placementWizard) gridCellAt(px, py int) (gx, gy int) {
	ox := int(w.gridOriginX)
	oy := int(w.gridOriginY)
	cs := int(rpCellSz)
	layout := w.char.CurrentHome.Type.Layout
	rows := len(layout)
	if rows == 0 {
		return -1, -1
	}
	cols := len(layout[0])
	gx = (px - ox) / cs
	gy = (py - oy) / cs
	if gx < 0 || gy < 0 || gx >= cols || gy >= rows {
		return -1, -1
	}
	return gx, gy
}

// maxZ returns the highest valid Z anchor for the item given the room's max height.
func (w *placementWizard) maxZ() uint8 {
	mh := w.char.CurrentHome.Type.MaxHeight
	ih := w.item.Height
	if ih >= mh {
		return 0
	}
	return mh - ih
}

// dirFacesWall returns true if placing item at (w.x, w.y) facing dir is invalid.
// Items at excludeIdx are excluded from the occupied-cell set.
func (w *placementWizard) dirFacesWall(dir model.Direction) bool {
	home := w.char.CurrentHome
	layout := home.Type.Layout
	rows := len(layout)
	if rows == 0 {
		return false
	}
	cols := len(layout[0])

	for _, cell := range occupiedCells(w.x, w.y, *w.item, dir) {
		cx, cy := int(cell[0]), int(cell[1])
		if cy < 0 || cy >= rows || cx < 0 || cx >= cols {
			return true
		}
		if layout[cy][cx] != model.HomeCellFloor && layout[cy][cx] != model.HomeCellDoor {
			return true
		}
	}

	occupied := make(map[[2]uint8]bool)
	for i, pl := range home.RoomItems {
		if i == w.excludeIdx {
			continue
		}
		for _, cell := range occupiedCells(pl.X, pl.Y, pl.Item, pl.Direction) {
			occupied[cell] = true
		}
	}

	stepX, stepY := dirStepXY(dir)
	allCells := occupiedCells(w.x, w.y, *w.item, dir)
	footprintSet := make(map[[2]uint8]bool, len(allCells))
	for _, c := range allCells {
		footprintSet[c] = true
	}

	for _, cell := range allCells {
		cx, cy := int(cell[0]), int(cell[1])
		if footprintSet[[2]uint8{uint8(cx + stepX), uint8(cy + stepY)}] {
			continue
		}
		n1x, n1y := cx+stepX, cy+stepY
		if n1x < 0 || n1y < 0 || n1y >= rows || n1x >= cols {
			return true
		}
		if layout[n1y][n1x] == model.HomeCellWall {
			return true
		}
		if !w.item.NeedClearance {
			continue
		}
		if occupied[[2]uint8{uint8(n1x), uint8(n1y)}] {
			return true
		}

	}
	return false
}

// dirOfItemBelow returns the direction of the item directly below (z == w.z, top == w.z).
// Falls back to North if none found.
func (w *placementWizard) dirOfItemBelow() model.Direction {
	home := w.char.CurrentHome
	newCells := occupiedCells(w.x, w.y, *w.item, w.dir)
	for _, placed := range home.RoomItems {
		if placed.Z+placed.Item.Height != w.z {
			continue
		}
		below := occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction)
		if footprintsOverlap(newCells, below) {
			return placed.Direction
		}
	}
	return model.North
}

// Returns true if the wizard is fully done (finalized or cancelled).
func (w *placementWizard) update() bool {
	mx, my := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	// cancel button always available
	cx2, cy2, cw, ch := spCancelBtnRect()
	if clicked && isHovered(mx, my, cx2, cy2, cw, ch) {
		switch w.curStep {
		case pwStepGrid:
			w.onCancel()
			return true
		case pwStepZ:
			w.curStep = pwStepGrid
			w.err = ""
		case pwStepDir:
			w.curStep = pwStepZ
			w.err = ""
		}
		return false
	}

	switch w.curStep {
	case pwStepGrid:
		w.hoverX, w.hoverY = w.gridCellAt(mx, my)
		if clicked {
			gx, gy := w.gridCellAt(mx, my)
			if gx >= 0 {
				w.x = uint8(gx)
				w.y = uint8(gy)
				w.z = 0
				w.err = ""
				if err := canPlaceOnGrid(w.char.CurrentHome, *w.item, w.x, w.y); err != nil {
					w.err = err.Error()
					return false
				}
				w.curStep = pwStepZ
			}
		}

	case pwStepZ:
		maxZ := w.maxZ()
		if clicked {
			mh := int(w.char.CurrentHome.Type.MaxHeight)
			// Precompute supported Z levels.
			supportedZ := make(map[uint8]bool)
			for _, placed := range w.char.CurrentHome.RoomItems {
				topZ := placed.Z + placed.Item.Height
				below := occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction)
				newCells := occupiedCells(w.x, w.y, *w.item, w.dir)
				if footprintsOverlap(newCells, below) {
					supportedZ[topZ] = true
				}
			}
			for z := 0; z < mh; z++ {
				cx2, cy2, cw, ch := spZCellRect(z, mh)
				if isHovered(mx, my, cx2, cy2, cw, ch) {
					candidate := uint8(z)
					validAnchor := w.item.CanOverhang || candidate == 0 || supportedZ[candidate]
					if validAnchor && candidate <= maxZ {
						w.z = candidate
					}
					break
				}
			}
			zbx, zby, zbw, zbh := spZConfirmBtnRect()
			if isHovered(mx, my, zbx, zby, zbw, zbh) {
				w.err = ""
				if w.z > 0 {
					// Direction is locked to the item below — skip dir step.
					dir := w.dirOfItemBelow()
					w.dir = dir
					if err := w.onFinalize(w.x, w.y, w.z, dir); err != nil {
						w.err = err.Error()
						w.curStep = pwStepGrid
					}
					return true
				}
				w.curStep = pwStepDir
			}
		}

	case pwStepDir:
		dirs := []model.Direction{model.North, model.East, model.South, model.West}
		for i, d := range dirs {
			bx, by, bw, bh := spDirBtnRect(i)
			if isHovered(mx, my, bx, by, bw, bh) {
				w.dir = d
				w.err = ""
			}
			if clicked && isHovered(mx, my, bx, by, bw, bh) {
				if !w.dirFacesWall(d) {
					if err := w.onFinalize(w.x, w.y, w.z, d); err != nil {
						w.err = err.Error()
						w.curStep = pwStepGrid
					}
					return true
				}
				w.err = "Can't face a wall — pick another direction."
			}
		}
	}
	return false
}

// DrawLeftPanel draws step instructions + Z picker or direction picker inside the left panel.
func (w *placementWizard) drawLeftPanel(dst *ebiten.Image, mx, my int) {
	lx := float64(w.panelOX) + 20

	switch w.curStep {
	case pwStepZ:
		mh := int(w.char.CurrentHome.Type.MaxHeight)
		maxZ := int(w.maxZ())
		itemH := int(w.item.Height)
		canOverhang := w.item.CanOverhang
		selZ := int(w.z)

		drawText(dst, "Step 2: Pick height (Z)", lx, float64(w.panelOY)+16, fontS, colorMuted)
		drawText(dst, w.item.Name, lx, float64(w.panelOY)+40, fontM, colorAccent)
		drawText(dst, fmt.Sprintf("pos (%d,%d)", w.x, w.y), lx, float64(w.panelOY)+66, fontS, colorMuted)

		labelX := float64(zColX) + float64(zCellW) + 8
		// Precompute which Z levels have a supporting item directly below.
		supportedZ := make(map[int]bool)
		for _, placed := range w.char.CurrentHome.RoomItems {
			topZ := int(placed.Z + placed.Item.Height)
			if topZ > mh {
				continue
			}
			below := occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction)
			newCells := occupiedCells(w.x, w.y, *w.item, w.dir)
			if footprintsOverlap(newCells, below) {
				supportedZ[topZ] = true
			}
		}
		for z := mh - 1; z >= 0; z-- {
			cx2, cy2, cw, ch := spZCellRect(z, mh)
			occ := z >= selZ && z < selZ+itemH
			validAnchor := canOverhang || z == 0 || supportedZ[z]
			clickable := validAnchor && uint8(z) <= uint8(maxZ)

			var bg color.RGBA
			switch {
			case occ && z == selZ:
				bg = colorAccent
			case occ:
				bg = colorItem
			case !clickable:
				bg = color.RGBA{40, 40, 50, 255}
			case isHovered(mx, my, cx2, cy2, cw, ch):
				bg = colorHighlight
			default:
				bg = colorPanel
			}
			border := colorBorder
			if !clickable {
				border = color.RGBA{50, 50, 60, 255}
			} else if z == selZ {
				border = colorAccent
			}
			fillRect(dst, cx2, cy2, cw, ch, bg)
			strokeRect(dst, cx2, cy2, cw, ch, border)

			zv := fmt.Sprintf("%d", z)
			tc := colorMuted
			if occ {
				tc = colorBg
			}
			zw, _ := text.Measure(zv, fontS, 0)
			drawText(dst, zv, float64(cx2)+float64(cw)/2-zw/2, float64(cy2)+float64(ch)/2-7, fontS, tc)

			row := mh - 1 - z
			annY := float64(zColY) + float64(row)*float64(zCellH)
			var ann string
			switch {
			case occ && z == selZ:
				ann = "← anchor"
			case occ:
				ann = "[X]"
			case !clickable:
				ann = "(locked)"
			default:
				ann = "[ ]"
			}
			annColor := colorMuted
			if occ {
				annColor = colorAccent
			}
			drawText(dst, ann, labelX, annY+float64(zCellH)/2-7, fontS, annColor)
		}

		legendY := float64(zColY) + float64(mh)*float64(zCellH) + 6
		drawText(dst, fmt.Sprintf("Item height: %d  MaxH: %d", itemH, mh), float64(zColX), legendY, fontS, colorMuted)
		if canOverhang {
			drawText(dst, "Can overhang: yes", float64(zColX), legendY+16, fontS, colorGreen)
		}
		cbx, cby, cbw, cbh := spZConfirmBtnRect()
		drawButton(dst, "Confirm Z →", cbx, cby, cbw, cbh, fontM, isHovered(mx, my, cbx, cby, cbw, cbh), true)

	case pwStepDir:
		drawText(dst, "Step 3: Choose facing direction", lx, float64(w.panelOY)+16, fontS, colorMuted)
		drawText(dst, w.item.Name, lx, float64(w.panelOY)+40, fontM, colorAccent)
		drawText(dst, fmt.Sprintf("pos (%d,%d)  z=%d", w.x, w.y, w.z), lx, float64(w.panelOY)+66, fontS, colorMuted)

		dirs := []model.Direction{model.North, model.East, model.South, model.West}
		for i, d := range dirs {
			bx, by, bw, bh := spDirBtnRect(i)
			blocked := w.dirFacesWall(d)
			hov := isHovered(mx, my, bx, by, bw, bh) && !blocked
			drawButton(dst, dirBtnLabels[i], bx, by, bw, bh, fontM, hov, !blocked)
			if blocked {
				drawText(dst, "wall", float64(bx)+float64(bw)/2-10, float64(by+bh)+2, fontS, colorRed)
			}
		}
		if w.err != "" {
			drawTextWrapped(dst, w.err,
				float64(w.panelOX)+20, float64(w.panelOY)+240,
				float64(spListW)-24, 18, fontS, colorRed)
		}
	}
}

// DrawGhost draws the ghost footprint overlay on the grid.
// ox, oy are the grid pixel origin.
func (w *placementWizard) drawGhost(dst *ebiten.Image, ox, oy float32) {
	if w.item == nil {
		return
	}

	mx, my := ebiten.CursorPosition()
	home := w.char.CurrentHome
	layout := home.Type.Layout
	rows := len(layout)

	var hx, hy int
	switch w.curStep {
	case pwStepGrid:
		hx, hy = w.gridCellAt(mx, my)
	default:
		hx, hy = int(w.x), int(w.y)
	}

	if hx < 0 {
		if w.err != "" {
			drawText(dst, w.err, float64(ox), float64(oy)+float64(float32(rows)*rpCellSz)+6, fontS, colorRed)
		}
		return
	}

	ghostDir := w.dir
	if w.curStep == pwStepGrid {
		// pick first valid direction so ghost can be shown
		savedX, savedY := w.x, w.y
		w.x, w.y = uint8(hx), uint8(hy)
		ghostDir = model.North
		for _, d := range []model.Direction{model.North, model.East, model.South, model.West} {
			if !w.dirFacesWall(d) {
				ghostDir = d
				break
			}
		}
		w.x, w.y = savedX, savedY
	}

	gCols, gRows := itemFootprint(*w.item, ghostDir)
	for dr := uint8(0); dr < gRows; dr++ {
		for dc := uint8(0); dc < gCols; dc++ {
			gcx := ox + float32(hx+int(dc))*rpCellSz
			gcy := oy + float32(hy+int(dr))*rpCellSz
			fillRect(dst, gcx+1, gcy+1, rpCellSz-3, rpCellSz-3, color.RGBA{100, 200, 100, 120})
		}
	}
	if w.curStep == pwStepDir {
		drawFacingArrow(dst, ox, oy, hx, hy, ghostDir)
	}

	if w.err != "" {
		drawText(dst, w.err, float64(ox), float64(oy)+float64(float32(rows)*rpCellSz)+6, fontS, colorRed)
	}
}
