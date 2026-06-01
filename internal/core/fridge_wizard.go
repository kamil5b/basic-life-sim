package core

import (
	"fmt"
	"image/color"

	"github.com/kamil5b/basic-life-sim/internal/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const fwCellSz = float32(56)

// fridgeWizard manages food placement inside a fridge.
//
// Selection is two-phase:
//
//	Phase 1 — click a Z cell in the left column (Z=highest at top, Z=0 at bottom).
//	Phase 2 — click an XY cell in the grid that appears to the right.
//
// A Rotate button swaps Width↔Length of the pending/moving food.
//
// Used in shop context (pendingFood != nil) and room-organise context (nil).
type fridgeWizard struct {
	char      *model.Character
	fridgeIdx int

	// phase 1
	selZ int // -1 = not yet picked

	// phase 2
	selSlotX int // -1 = not yet picked
	selSlotY int

	// rotation (swaps food Width↔Length for placement)
	rotated bool

	// move mode (room context only)
	moveMode      bool
	movingFoodIdx int // index into placed.Stored; -1 = none

	// food to place (nil = room-organise)
	pendingFood *model.Food
	chargeMoney bool

	// pixel origin of the Z column
	gridX, gridY float32
	infoX        float32

	setMsg   func(string)
	onCancel func()
	onDone   func()
}

func newFridgeWizard(
	char *model.Character,
	fridgeIdx int,
	gridX, gridY, infoX float32,
	pendingFood *model.Food,
	chargeMoney bool,
	setMsg func(string),
	onCancel func(),
	onDone func(),
) *fridgeWizard {
	return &fridgeWizard{
		char:          char,
		fridgeIdx:     fridgeIdx,
		selZ:          -1,
		selSlotX:      -1,
		selSlotY:      -1,
		movingFoodIdx: -1,
		pendingFood:   pendingFood,
		chargeMoney:   chargeMoney,
		gridX:         gridX,
		gridY:         gridY,
		infoX:         infoX,
		setMsg:        setMsg,
		onCancel:      onCancel,
		onDone:        onDone,
	}
}

// ── top-surface helpers ───────────────────────────────────────────────────────

func (w *fridgeWizard) firstFreeTopCell() (x, y, topZ uint8, found bool) {
	placed := w.char.CurrentHome.RoomItems[w.fridgeIdx]
	topZ = placed.Z + placed.Item.Height
	cells := occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction)
	taken := make(map[[2]uint8]bool)
	for _, ff := range w.char.CurrentHome.FloorFood {
		if ff.Z >= topZ {
			taken[[2]uint8{ff.X, ff.Y}] = true
		}
	}
	for _, c := range cells {
		if !taken[c] {
			return c[0], c[1], topZ, true
		}
	}
	return 0, 0, topZ, false
}

func (w *fridgeWizard) hasFoodOnTop() bool {
	_, _, _, found := w.firstFreeTopCell()
	return !found
}

// ── selection helpers ─────────────────────────────────────────────────────────

func (w *fridgeWizard) fullySelected() bool {
	return w.selZ >= 0 && w.selSlotX >= 0 && w.selSlotY >= 0
}

func (w *fridgeWizard) findFoodAtSel(placed *model.PlacedRoomItem) int {
	if !w.fullySelected() {
		return -1
	}
	z := uint8(w.selZ)
	for i, s := range placed.Stored {
		if s.SlotZ == z && int(s.SlotX) == w.selSlotX && int(s.SlotY) == w.selSlotY {
			return i
		}
	}
	return -1
}

// effectiveFood returns the food with Width/Length swapped when rotated.
// Pass a Food value (for pending) or a copy from Stored.
func rotateFood(f model.Food, rotated bool) model.Food {
	if rotated {
		f.Width, f.Length = f.Length, f.Width
	}
	return f
}

func isColdAt(storage *model.StorageCapacity, x, y, z uint8) bool {
	for _, cz := range storage.ColdZones {
		if cz.Contains(x, y, z) {
			return true
		}
	}
	return false
}

// ── layout helpers ────────────────────────────────────────────────────────────

// xyGridOrigin returns the top-left pixel of the XY grid (right of the Z column).
func (w *fridgeWizard) xyGridOrigin() (ox, oy float32) {
	return w.gridX + fwCellSz + 10, w.gridY
}

func (w *fridgeWizard) rotateBtnRect(storage *model.StorageCapacity) (x, y, w2, h float32) {
	ox, oy := w.xyGridOrigin()
	return ox, oy + float32(storage.Length)*fwCellSz + 8, 110, 28
}

func (w *fridgeWizard) backBtnRect() (x, y, width, height float32) {
	return panelX + 4, float32(ScreenH) - float32(tabH) - 50, 100, 32
}

func (w *fridgeWizard) action1BtnRect() (x, y, width, height float32) {
	return panelX + 110, float32(ScreenH) - float32(tabH) - 50, 140, 32
}

func (w *fridgeWizard) action2BtnRect() (x, y, width, height float32) {
	return panelX + 258, float32(ScreenH) - float32(tabH) - 50, 130, 32
}

// ── coordinate converters ─────────────────────────────────────────────────────

// zColCellAt maps mouse → Z index (-1 if outside the Z column).
// Visual row 0 = highest Z (top shelf), last row = Z=0 (floor).
func (w *fridgeWizard) zColCellAt(px, py int, storage *model.StorageCapacity) int {
	ox, oy, cs := int(w.gridX), int(w.gridY), int(fwCellSz)
	if px < ox || px >= ox+cs {
		return -1
	}
	row := (py - oy) / cs
	if row < 0 || row >= int(storage.Height) {
		return -1
	}
	return int(storage.Height) - 1 - row
}

// xyCellAt maps mouse → (col, row) in the XY grid (-1,-1 if outside).
func (w *fridgeWizard) xyCellAt(px, py int, storage *model.StorageCapacity) (gx, gy int) {
	ox, oy := w.xyGridOrigin()
	cs := int(fwCellSz)
	col := (px - int(ox)) / cs
	row := (py - int(oy)) / cs
	if col < 0 || row < 0 || col >= int(storage.Width) || row >= int(storage.Length) {
		return -1, -1
	}
	return col, row
}

// ── update ────────────────────────────────────────────────────────────────────

func (w *fridgeWizard) update() bool {
	mx, my := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if !clicked {
		return false
	}

	placed := &w.char.CurrentHome.RoomItems[w.fridgeIdx]
	storage := placed.Item.Storage

	// Back / Cancel — always available
	bx, by, bw, bh := w.backBtnRect()
	if isHovered(mx, my, bx, by, bw, bh) {
		w.onCancel()
		return true
	}

	// Rotate button — available in phase 2
	if w.selZ >= 0 {
		rx, ry, rw, rh := w.rotateBtnRect(storage)
		if isHovered(mx, my, rx, ry, rw, rh) {
			w.rotated = !w.rotated
			w.selSlotX = -1
			w.selSlotY = -1
			return false
		}
	}

	// ── actions that require full selection ───────────────────────────────────
	if w.fullySelected() {
		slotX, slotY, slotZ := uint8(w.selSlotX), uint8(w.selSlotY), uint8(w.selZ)

		if w.pendingFood != nil {
			// Place Here
			a1x, a1y, a1w, a1h := w.action1BtnRect()
			if isHovered(mx, my, a1x, a1y, a1w, a1h) {
				f := rotateFood(*w.pendingFood, w.rotated)
				occupied := occupiedFoodSlots(placed)
				if canFitFood(occupied, storage, f, slotX, slotY, slotZ) {
					mult := storage.MultiplierAt(slotX, slotY, slotZ)
					if w.chargeMoney {
						w.char.CurrentStats.Money -= w.pendingFood.BasePrice
					}
					placed.Stored = append(placed.Stored, model.StoredFood{
						SlotX: slotX, SlotY: slotY, SlotZ: slotZ,
						PurchaseDate:   w.char.CurrentDate,
						MultiplierUsed: mult,
						UsesRemaining:  f.UsesTotal,
						Food:           f,
					})
					expiry := model.ExpiryDate(w.char.CurrentDate, f.BaseExpiryDays, mult)
					ey, em, ed := expiry.Unpack()
					rot := ""
					if w.rotated {
						rot = " [rotated]"
					}
					w.setMsg(fmt.Sprintf("Placed %s%s → (x=%d,y=%d,z=%d), exp %04d-%02d-%02d",
						f.Name, rot, slotX, slotY, slotZ, ey, em, ed))
					w.onDone()
					return true
				}
				w.setMsg("Slot occupied or food does not fit here.")
				return false
			}
		} else {
			// Take Out
			if !w.moveMode {
				a1x, a1y, a1w, a1h := w.action1BtnRect()
				if isHovered(mx, my, a1x, a1y, a1w, a1h) {
					if w.hasFoodOnTop() {
						w.setMsg("Can't take out — the fridge top is full. Remove food from on top first.")
						return false
					}
					foodIdx := w.findFoodAtSel(placed)
					if foodIdx >= 0 {
						s := placed.Stored[foodIdx]
						daysElapsed := model.DaysBetween(s.PurchaseDate, w.char.CurrentDate)
						fridgeDays := float32(daysElapsed) / s.MultiplierUsed
						remaining := float32(s.Food.BaseExpiryDays) - fridgeDays
						if remaining < 1 {
							remaining = 1
						}
						newFood := s.Food
						newFood.BaseExpiryDays = uint16(remaining)
						placed.Stored = append(placed.Stored[:foodIdx], placed.Stored[foodIdx+1:]...)
						fx, fy, fz, _ := w.firstFreeTopCell()
						w.char.CurrentHome.FloorFood = append(w.char.CurrentHome.FloorFood, model.PlacedFood{
							X: fx, Y: fy, Z: fz,
							PurchaseDate:  w.char.CurrentDate,
							UsesRemaining: s.UsesRemaining,
							Food:          newFood,
						})
						expiry := model.ExpiryDate(w.char.CurrentDate, newFood.BaseExpiryDays, 1)
						ey, em, ed := expiry.Unpack()
						w.setMsg(fmt.Sprintf("Took out %s → fridge top (%d,%d,z=%d). New exp %04d-%02d-%02d",
							s.Food.Name, fx, fy, fz, ey, em, ed))
						w.selZ = -1
						w.selSlotX = -1
						w.selSlotY = -1
						w.onDone()
					}
					return false
				}
			}
			// Move toggle
			a2x, a2y, a2w, a2h := w.action2BtnRect()
			if isHovered(mx, my, a2x, a2y, a2w, a2h) {
				if !w.moveMode {
					idx := w.findFoodAtSel(placed)
					if idx >= 0 {
						w.movingFoodIdx = idx
						w.moveMode = true
						w.rotated = false
						w.setMsg("Pick Z then XY for the target slot. Use Rotate to reorient.")
						// reset selection so user picks a target
						w.selZ = -1
						w.selSlotX = -1
						w.selSlotY = -1
					}
				} else {
					w.moveMode = false
					w.movingFoodIdx = -1
					w.rotated = false
				}
				return false
			}
		}
	}

	// ── phase 2: XY grid click ────────────────────────────────────────────────
	if w.selZ >= 0 {
		gx, gy := w.xyCellAt(mx, my, storage)
		if gx >= 0 {
			if w.moveMode && w.movingFoodIdx >= 0 && w.fullySelected() {
				// second XY click after a first one — treat as new target
			}
			w.selSlotX = gx
			w.selSlotY = gy

			// If in move mode and fully selected, execute the move immediately
			if w.moveMode && w.movingFoodIdx >= 0 {
				sf := &placed.Stored[w.movingFoodIdx]
				mf := rotateFood(sf.Food, w.rotated)
				occupied := occupiedFoodSlots(placed)
				// Remove source slots from occupied map
				for dz := uint8(0); dz < sf.Food.Height; dz++ {
					for dy := uint8(0); dy < sf.Food.Length; dy++ {
						for dx := uint8(0); dx < sf.Food.Width; dx++ {
							delete(occupied, [3]uint8{sf.SlotX + dx, sf.SlotY + dy, sf.SlotZ + dz})
						}
					}
				}
				tX, tY, tZ := uint8(gx), uint8(gy), uint8(w.selZ)
				if canFitFood(occupied, storage, mf, tX, tY, tZ) {
					sf.Food = mf
					sf.SlotX, sf.SlotY, sf.SlotZ = tX, tY, tZ
					sf.MultiplierUsed = storage.MultiplierAt(tX, tY, tZ)
					rot := ""
					if w.rotated {
						rot = " [rotated]"
					}
					w.setMsg(fmt.Sprintf("Moved %s%s to (x=%d,y=%d,z=%d)", mf.Name, rot, tX, tY, tZ))
					w.moveMode = false
					w.movingFoodIdx = -1
					w.rotated = false
					w.selZ = -1
					w.selSlotX = -1
					w.selSlotY = -1
				} else {
					w.setMsg("Target slot occupied or food does not fit.")
					w.selSlotX = -1
					w.selSlotY = -1
				}
			}
			return false
		}
	}

	// ── phase 1: Z column click ───────────────────────────────────────────────
	z := w.zColCellAt(mx, my, storage)
	if z >= 0 {
		w.selZ = z
		w.selSlotX = -1
		w.selSlotY = -1
	}

	return false
}

// ── draw ──────────────────────────────────────────────────────────────────────

func (w *fridgeWizard) draw(dst *ebiten.Image) {
	mx, my := ebiten.CursorPosition()
	char := w.char
	placed := char.CurrentHome.RoomItems[w.fridgeIdx]
	storage := placed.Item.Storage
	lx := float64(w.infoX)

	used := usedSlotCount(&char.CurrentHome.RoomItems[w.fridgeIdx])
	total := storage.TotalSlots()

	// ── left info panel ───────────────────────────────────────────────────────
	if w.pendingFood != nil {
		f := rotateFood(*w.pendingFood, w.rotated)
		rotTag := ""
		if w.rotated {
			rotTag = " [rotated]"
		}
		drawText(dst, "Place in "+placed.Item.Name, lx, float64(panelY)+14, fontM, colorAccent)
		drawText(dst, fmt.Sprintf("Buying: %s%s  $%.2f  uses:%d",
			w.pendingFood.Name, rotTag, w.pendingFood.BasePrice, w.pendingFood.UsesTotal),
			lx, float64(panelY)+38, fontS, colorText)
		drawText(dst, fmt.Sprintf("Size: %dx%d (W×L)  Slots: %d/%d used",
			f.Width, f.Length, used, total),
			lx, float64(panelY)+56, fontS, colorMuted)
		drawText(dst, "❄ = cold zone  Step 1: pick Z  Step 2: pick XY",
			lx, float64(panelY)+74, fontS, colorMuted)

		if w.fullySelected() {
			slotX, slotY, slotZ := uint8(w.selSlotX), uint8(w.selSlotY), uint8(w.selZ)
			mult := storage.MultiplierAt(slotX, slotY, slotZ)
			expiry := model.ExpiryDate(char.CurrentDate, f.BaseExpiryDays, mult)
			ey, em, ed := expiry.Unpack()
			occupied := occupiedFoodSlots(&char.CurrentHome.RoomItems[w.fridgeIdx])
			canPlace := canFitFood(occupied, storage, f, slotX, slotY, slotZ)
			col := colorGreen
			msg := fmt.Sprintf("(x=%d,y=%d,z=%d)  %.1fx  exp:%04d-%02d-%02d",
				slotX, slotY, slotZ, mult, ey, em, ed)
			if !canPlace {
				col = colorRed
				msg = fmt.Sprintf("(x=%d,y=%d,z=%d) — does not fit", slotX, slotY, slotZ)
			}
			drawText(dst, msg, lx, float64(panelY)+96, fontS, col)
			a1x, a1y, a1w, a1h := w.action1BtnRect()
			drawButton(dst, "✓ Place Here", a1x, a1y, a1w, a1h, fontM,
				isHovered(mx, my, a1x, a1y, a1w, a1h) && canPlace, canPlace)
		} else if w.selZ >= 0 {
			drawText(dst, fmt.Sprintf("Z=%d selected — now pick XY slot", w.selZ),
				lx, float64(panelY)+96, fontS, colorMuted)
		} else {
			drawText(dst, "Click a Z cell on the left column first",
				lx, float64(panelY)+96, fontS, colorMuted)
		}
	} else {
		drawText(dst, placed.Item.Name+" Interior", lx, float64(panelY)+4, fontM, colorAccent)
		drawText(dst, fmt.Sprintf("%d/%d slots used", used, total),
			lx, float64(panelY)+26, fontS, colorMuted)
		drawText(dst, "Step 1: pick Z  Step 2: pick XY",
			lx, float64(panelY)+44, fontS, colorMuted)

		if w.moveMode && w.movingFoodIdx >= 0 {
			sf := placed.Stored[w.movingFoodIdx]
			mf := rotateFood(sf.Food, w.rotated)
			rotTag := ""
			if w.rotated {
				rotTag = " [rotated]"
			}
			drawText(dst, fmt.Sprintf("Moving: %s%s (%dx%d)", mf.Name, rotTag, mf.Width, mf.Length),
				lx, float64(panelY)+66, fontS, colorYellow)
			drawText(dst, "Pick Z then XY target slot",
				lx, float64(panelY)+84, fontS, colorMuted)
		} else if w.fullySelected() {
			foodIdx := w.findFoodAtSel(&placed)
			if foodIdx >= 0 {
				sf := placed.Stored[foodIdx]
				exp := model.IsExpired(sf.PurchaseDate, sf.Food.BaseExpiryDays, sf.MultiplierUsed, char.CurrentDate)
				col := colorAccent
				if exp {
					col = colorRed
				}
				expiry := model.ExpiryDate(sf.PurchaseDate, sf.Food.BaseExpiryDays, sf.MultiplierUsed)
				ey, em, ed := expiry.Unpack()
				drawText(dst, fmt.Sprintf("Selected: %s", sf.Food.Name),
					lx, float64(panelY)+66, fontS, col)
				drawText(dst, fmt.Sprintf("uses:%d  (x=%d,y=%d,z=%d)  exp:%04d-%02d-%02d",
					sf.UsesRemaining, sf.SlotX, sf.SlotY, sf.SlotZ, ey, em, ed),
					lx, float64(panelY)+84, fontS, colorMuted)

				takeEnabled := !w.hasFoodOnTop()
				takeLabel := "Take Out"
				if !takeEnabled {
					takeLabel = "Take Out (blocked)"
				}
				a1x, a1y, a1w, a1h := w.action1BtnRect()
				drawButton(dst, takeLabel, a1x, a1y, a1w, a1h, fontS,
					isHovered(mx, my, a1x, a1y, a1w, a1h) && takeEnabled, takeEnabled)
				a2x, a2y, a2w, a2h := w.action2BtnRect()
				drawButton(dst, "✦ Move", a2x, a2y, a2w, a2h, fontS,
					isHovered(mx, my, a2x, a2y, a2w, a2h), true)
			} else {
				drawText(dst, fmt.Sprintf("Empty slot (x=%d,y=%d,z=%d)", w.selSlotX, w.selSlotY, w.selZ),
					lx, float64(panelY)+66, fontS, colorMuted)
			}
		} else if w.selZ >= 0 {
			drawText(dst, fmt.Sprintf("Z=%d — pick XY slot", w.selZ),
				lx, float64(panelY)+66, fontS, colorMuted)
		} else {
			drawText(dst, "Click a Z cell to start",
				lx, float64(panelY)+66, fontS, colorMuted)
		}
	}

	// Back / Cancel
	bx, by, bw, bh := w.backBtnRect()
	backLabel := "← Back"
	if w.pendingFood != nil {
		backLabel = "✕ Cancel"
	}
	drawButton(dst, backLabel, bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)

	// ── Z column ──────────────────────────────────────────────────────────────
	occupied := occupiedFoodSlots(&char.CurrentHome.RoomItems[w.fridgeIdx])
	ox, oy, cs := w.gridX, w.gridY, fwCellSz

	drawText(dst, "Z", float64(ox)+float64(cs)/2-4, float64(oy)-16, fontS, colorMuted)

	for zIdx := 0; zIdx < int(storage.Height); zIdx++ {
		z := uint8(int(storage.Height) - 1 - zIdx) // row 0 = highest Z
		cy := oy + float32(zIdx)*cs

		totalSlots := int(storage.Width) * int(storage.Length)
		usedSlots := 0
		hasExpired := false
		hasCold := false
		for x := 0; x < int(storage.Width); x++ {
			for y := 0; y < int(storage.Length); y++ {
				if occupied[[3]uint8{uint8(x), uint8(y), z}] {
					usedSlots++
					for _, s := range placed.Stored {
						if s.SlotX == uint8(x) && s.SlotY == uint8(y) && s.SlotZ == z {
							if model.IsExpired(s.PurchaseDate, s.Food.BaseExpiryDays, s.MultiplierUsed, char.CurrentDate) {
								hasExpired = true
							}
						}
					}
				}
				if isColdAt(storage, uint8(x), uint8(y), z) {
					hasCold = true
				}
			}
		}

		isSel := w.selZ == int(z)
		hov := isHovered(mx, my, ox, cy, cs-1, cs-1)

		var bg color.RGBA
		switch {
		case isSel:
			bg = colorSelected
		case usedSlots == totalSlots && hasExpired:
			bg = color.RGBA{80, 20, 20, 255}
		case usedSlots > 0 && hasExpired:
			bg = color.RGBA{60, 30, 30, 255}
		case usedSlots == totalSlots:
			bg = color.RGBA{40, 55, 100, 255}
		case usedSlots > 0:
			bg = color.RGBA{30, 45, 80, 255}
		case hov:
			bg = colorHighlight
		case hasCold:
			bg = color.RGBA{20, 40, 80, 255}
		default:
			bg = color.RGBA{28, 35, 50, 255}
		}
		borderCol := colorBorder
		if isSel {
			borderCol = colorAccent
		} else if hasCold {
			borderCol = color.RGBA{60, 120, 220, 255}
		}
		fillRect(dst, ox, cy, cs-1, cs-1, bg)
		strokeRect(dst, ox, cy, cs-1, cs-1, borderCol)

		zLabel := fmt.Sprintf("z%d", z)
		if hasCold {
			zLabel += "❄"
		}
		drawText(dst, zLabel, float64(ox)+3, float64(cy)+3, fontS, colorMuted)
		if usedSlots > 0 {
			tc := colorText
			if hasExpired {
				tc = colorRed
			}
			drawText(dst, fmt.Sprintf("%d/%d", usedSlots, totalSlots), float64(ox)+3, float64(cy)+20, fontS, tc)
		} else {
			drawText(dst, "free", float64(ox)+3, float64(cy)+20, fontS, colorGreen)
		}
	}

	// ── XY grid (shown after Z is selected) ──────────────────────────────────
	if w.selZ < 0 {
		return
	}
	selZU := uint8(w.selZ)
	xyOX, xyOY := w.xyGridOrigin()

	drawText(dst, "← X (width) →", float64(xyOX), float64(xyOY)-16, fontS, colorMuted)
	drawText(dst, "↓Y", float64(xyOX)-20, float64(xyOY), fontS, colorMuted)

	// Rotate button
	rx, ry, rw, rh := w.rotateBtnRect(storage)
	rotLabel := "↺ Rotate"
	if w.rotated {
		rotLabel = "↺ Rotated"
	}
	showRotate := w.pendingFood != nil || (w.moveMode && w.movingFoodIdx >= 0)
	if showRotate {
		drawButton(dst, rotLabel, rx, ry, rw, rh, fontS, isHovered(mx, my, rx, ry, rw, rh), true)
	}

	// determine ghost footprint for pending/moving food
	var ghostW, ghostL int
	if w.pendingFood != nil {
		f := rotateFood(*w.pendingFood, w.rotated)
		ghostW, ghostL = int(f.Width), int(f.Length)
	} else if w.moveMode && w.movingFoodIdx >= 0 {
		sf := placed.Stored[w.movingFoodIdx]
		mf := rotateFood(sf.Food, w.rotated)
		ghostW, ghostL = int(mf.Width), int(mf.Length)
	}

	hoverGX, hoverGY := w.xyCellAt(mx, my, storage)

	for y := 0; y < int(storage.Length); y++ {
		for x := 0; x < int(storage.Width); x++ {
			cx := xyOX + float32(x)*cs
			cy := xyOY + float32(y)*cs

			slotKey := [3]uint8{uint8(x), uint8(y), selZU}
			isOcc := occupied[slotKey]
			isSel := w.selSlotX == x && w.selSlotY == y
			cold := isColdAt(storage, uint8(x), uint8(y), selZU)

			// ghost highlight: cells within food footprint anchored at hover
			inGhost := false
			if hoverGX >= 0 && ghostW > 0 {
				inGhost = x >= hoverGX && x < hoverGX+ghostW &&
					y >= hoverGY && y < hoverGY+ghostL
			}

			hov := isHovered(mx, my, cx, cy, cs-1, cs-1)
			var bg color.RGBA
			switch {
			case isSel && !isOcc:
				bg = colorSelected
			case isSel && isOcc:
				bg = color.RGBA{80, 20, 20, 255}
			case inGhost && !isOcc:
				bg = color.RGBA{50, 120, 50, 180}
			case isOcc:
				bg = color.RGBA{40, 55, 100, 255}
			case hov:
				bg = colorHighlight
			case cold:
				bg = color.RGBA{20, 40, 80, 255}
			default:
				bg = color.RGBA{28, 35, 50, 255}
			}
			borderCol := colorBorder
			if isSel {
				borderCol = colorAccent
			} else if cold {
				borderCol = color.RGBA{60, 120, 220, 255}
			}
			fillRect(dst, cx, cy, cs-1, cs-1, bg)
			strokeRect(dst, cx, cy, cs-1, cs-1, borderCol)

			xyLabel := fmt.Sprintf("x%d", x)
			if y == 0 {
				// show x label only on top row
			} else {
				xyLabel = fmt.Sprintf("y%d", y)
			}
			xyLabel = fmt.Sprintf("x%d,y%d", x, y)
			if cold {
				xyLabel += "❄"
			}
			drawText(dst, xyLabel, float64(cx)+3, float64(cy)+3, fontS, colorMuted)

			if isOcc {
				for _, s := range placed.Stored {
					if s.SlotX == uint8(x) && s.SlotY == uint8(y) && s.SlotZ == selZU {
						exp := model.IsExpired(s.PurchaseDate, s.Food.BaseExpiryDays, s.MultiplierUsed, char.CurrentDate)
						tc := colorText
						if exp {
							tc = colorRed
						}
						name := s.Food.Name
						if len(name) > 7 {
							name = name[:7]
						}
						drawText(dst, name, float64(cx)+3, float64(cy)+20, fontS, tc)
						break
					}
				}
			} else {
				mult := storage.MultiplierAt(uint8(x), uint8(y), selZU)
				drawText(dst, fmt.Sprintf("%.1fx", mult), float64(cx)+3, float64(cy)+20, fontS, colorGreen)
			}
		}
	}
}
