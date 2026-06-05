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

	// rotation: 0-5 cycles through all 6 axis-aligned orientations
	rotationIdx int

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
	w := &fridgeWizard{
		char:          char,
		fridgeIdx:     fridgeIdx,
		selZ:          -1,
		selSlotX:      -1,
		selSlotY:      -1,
		rotationIdx:   0,
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
	// Auto-rotate to a valid orientation that fits the fridge's Z capacity.
	if pendingFood != nil {
		w.ensureValidRotation(char.CurrentHome.RoomItems[fridgeIdx].Item.Storage)
	}
	return w
}

// ── top-surface helpers ───────────────────────────────────────────────────────

func (w *fridgeWizard) firstFreeTopCell() (x, y, topZ uint8, found bool) {
	placed := w.char.CurrentHome.RoomItems[w.fridgeIdx]
	topZ = placed.Z + placed.Item.Height
	cells := occupiedCells(placed.X, placed.Y, placed.Item, placed.Direction)
	taken := make(map[[2]uint8]bool)
	for _, fi := range w.char.CurrentHome.FloorItems {
		if fi.Z >= topZ {
			taken[[2]uint8{fi.X, fi.Y}] = true
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
	sx, sy, sz := uint8(w.selSlotX), uint8(w.selSlotY), uint8(w.selZ)
	for i, s := range placed.Stored {
		if sx >= s.SlotX && sx < s.SlotX+s.Food.Width &&
			sy >= s.SlotY && sy < s.SlotY+s.Food.Length &&
			sz >= s.SlotZ && sz < s.SlotZ+s.Food.Height {
			return i
		}
	}
	return -1
}

// rotationTable lists all 6 axis-aligned permutations of (Width, Length, Height).
// Each entry is [widthSrc, lengthSrc, heightSrc] — indices into the original [W, L, H] array.
var rotationTable = [6][3]int{
	{0, 1, 2}, // original
	{1, 0, 2}, // swap W↔L
	{0, 2, 1}, // swap L↔H
	{2, 1, 0}, // swap W↔H
	{1, 2, 0}, // W→L, L→H, H→W
	{2, 0, 1}, // W→H, L→W, H→L
}

// rotationLabel returns a short human-readable name for each rotation.
var rotationLabel = [6]string{
	"W×L×H", "L×W×H", "W×H×L", "H×L×W", "L×H×W", "H×W×L",
}

// applyRotation returns a copy of f with its dimensions permuted by rotIdx.
func applyRotation(f model.Food, rotIdx int) model.Food {
	dims := [3]uint8{f.Width, f.Length, f.Height}
	p := rotationTable[rotIdx%6]
	f.Width = dims[p[0]]
	f.Length = dims[p[1]]
	f.Height = dims[p[2]]
	return f
}

// ── rotation / Z validity helpers ───────────────────────────────────────────

// activeFoodBase returns the un-rotated food being placed/moved, or nil.
func (w *fridgeWizard) activeFoodBase(placed *model.PlacedRoomItem) *model.Food {
	if w.pendingFood != nil {
		return w.pendingFood
	}
	if w.moveMode && w.movingFoodIdx >= 0 && placed != nil {
		f := placed.Stored[w.movingFoodIdx].Food
		return &f
	}
	return nil
}

// activeFood returns the food in the current rotation, or zero-value if none.
func (w *fridgeWizard) activeFood(placed *model.PlacedRoomItem) (model.Food, bool) {
	base := w.activeFoodBase(placed)
	if base == nil {
		return model.Food{}, false
	}
	return applyRotation(*base, w.rotationIdx), true
}

// rotationFitsZ returns true if applyRotation(food, rotIdx) has Height <= maxZ.
func rotationFitsZ(food model.Food, rotIdx int, maxZ uint8) bool {
	return applyRotation(food, rotIdx).Height <= maxZ
}

// ensureValidRotation adjusts rotationIdx to the first orientation whose Height
// fits within storage.Height, if the current one is invalid.
func (w *fridgeWizard) ensureValidRotation(storage *model.StorageCapacity) {
	base := w.activeFoodBase(nil) // pending food only in ctor context
	if base == nil {
		if w.pendingFood == nil {
			return
		}
		base = w.pendingFood
	}
	if rotationFitsZ(*base, w.rotationIdx, storage.Height) {
		return
	}
	for i := 0; i < 6; i++ {
		if rotationFitsZ(*base, i, storage.Height) {
			w.rotationIdx = i
			w.selSlotX = -1
			w.selSlotY = -1
			return
		}
	}
}

// nextValidRotationIdx cycles to the next rotation that fits within storage.Height.
func (w *fridgeWizard) nextValidRotationIdx(storage *model.StorageCapacity, base model.Food) int {
	for i := 1; i <= 6; i++ {
		idx := (w.rotationIdx + i) % 6
		if rotationFitsZ(base, idx, storage.Height) {
			return idx
		}
	}
	return w.rotationIdx // no change if nothing fits (shouldn't happen)
}

// isZSelectableFor returns true if z is a valid TOP anchor for a food of the given height.
// The anchor is the highest Z the food occupies; it extends downward to z-foodH+1,
// so z must be >= foodH-1 (enough room below).
func isZSelectableFor(z int, foodH, fridgeH uint8) bool {
	return z >= int(foodH)-1
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

	// Rotate button — available in phase 2; only cycles valid orientations
	if w.selZ >= 0 {
		base, hasFood := w.activeFood(placed)
		if hasFood {
			rx, ry, rw, rh := w.rotateBtnRect(storage)
			if isHovered(mx, my, rx, ry, rw, rh) {
				w.rotationIdx = w.nextValidRotationIdx(storage, base)
				w.selSlotX = -1
				w.selSlotY = -1
				// Re-validate selZ: if food now doesn't fit starting there, reset
				if w.selZ >= 0 {
					f := applyRotation(*w.activeFoodBase(placed), w.rotationIdx)
					if !isZSelectableFor(w.selZ, f.Height, storage.Height) {
						w.selZ = -1
					}
				}
				return false
			}
		}
	}

	// ── actions that require full selection ───────────────────────────────────
	if w.fullySelected() {
		slotX, slotY := uint8(w.selSlotX), uint8(w.selSlotY)
		// selZ is the TOP anchor; actual bottom SlotZ = selZ - foodH + 1
		computeSlotZ := func(foodH uint8) uint8 {
			v := int(w.selZ) - int(foodH) + 1
			if v < 0 {
				v = 0
			}
			return uint8(v)
		}

		if w.pendingFood != nil {
			// Place Here
			a1x, a1y, a1w, a1h := w.action1BtnRect()
			if isHovered(mx, my, a1x, a1y, a1w, a1h) {
				f := applyRotation(*w.pendingFood, w.rotationIdx)
				slotZ := computeSlotZ(f.Height)
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
					w.setMsg(fmt.Sprintf("Placed %s [%s] → (x=%d,y=%d,z=%d), exp %04d-%02d-%02d",
						f.Name, rotationLabel[w.rotationIdx%6], slotX, slotY, slotZ, ey, em, ed))
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
						w.char.CurrentHome.FloorItems = append(w.char.CurrentHome.FloorItems, model.FloorItem{
							X: fx, Y: fy, Z: fz,
							Kind:          model.FloorKindFood,
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
						w.rotationIdx = 0
						w.ensureValidRotation(storage)
						w.setMsg("Pick Z then XY for the target slot. Use Rotate to reorient.")
						// reset selection so user picks a target
						w.selZ = -1
						w.selSlotX = -1
						w.selSlotY = -1
					}
				} else {
					w.moveMode = false
					w.movingFoodIdx = -1
					w.rotationIdx = 0
				}
				return false
			}
		}
	}

	// ── phase 2: XY grid click ────────────────────────────────────────────────
	if w.selZ >= 0 {
		gx, gy := w.xyCellAt(mx, my, storage)
		if gx >= 0 {
			w.selSlotX = gx
			w.selSlotY = gy

			// If in move mode and fully selected, execute the move immediately
			if w.moveMode && w.movingFoodIdx >= 0 {
				sf := &placed.Stored[w.movingFoodIdx]
				mf := applyRotation(sf.Food, w.rotationIdx)
				occupied := occupiedFoodSlots(placed)
				// Remove source slots from occupied map
				for dz := uint8(0); dz < sf.Food.Height; dz++ {
					for dy := uint8(0); dy < sf.Food.Length; dy++ {
						for dx := uint8(0); dx < sf.Food.Width; dx++ {
							delete(occupied, [3]uint8{sf.SlotX + dx, sf.SlotY + dy, sf.SlotZ + dz})
						}
					}
				}
				tX, tY := uint8(gx), uint8(gy)
				tZv := int(w.selZ) - int(mf.Height) + 1
				if tZv < 0 {
					tZv = 0
				}
				tZ := uint8(tZv)
				if canFitFood(occupied, storage, mf, tX, tY, tZ) {
					sf.Food = mf
					sf.SlotX, sf.SlotY, sf.SlotZ = tX, tY, tZ
					sf.MultiplierUsed = storage.MultiplierAt(tX, tY, tZ)
					w.setMsg(fmt.Sprintf("Moved %s [%s] to (x=%d,y=%d,z=%d)",
						mf.Name, rotationLabel[w.rotationIdx%6], tX, tY, tZ))
					w.moveMode = false
					w.movingFoodIdx = -1
					w.rotationIdx = 0
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
		// Check if food in current rotation fits starting at this Z
		f, hasF := w.activeFood(placed)
		if !hasF || isZSelectableFor(z, f.Height, storage.Height) {
			w.selZ = z
			w.selSlotX = -1
			w.selSlotY = -1
		}
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
		f := applyRotation(*w.pendingFood, w.rotationIdx)
		drawText(dst, "Place in "+placed.Item.Name, lx, float64(panelY)+14, fontM, colorAccent)
		drawText(dst, fmt.Sprintf("Buying: %s  $%.2f  uses:%d",
			w.pendingFood.Name, w.pendingFood.BasePrice, w.pendingFood.UsesTotal),
			lx, float64(panelY)+38, fontS, colorText)
		drawText(dst, fmt.Sprintf("Orientation: %s  (%dx%dx%d W×L×H)  %d/%d slots",
			rotationLabel[w.rotationIdx%6], f.Width, f.Length, f.Height, used, total),
			lx, float64(panelY)+56, fontS, colorMuted)
		drawText(dst, "❄ = cold zone  Step 1: pick Z  Step 2: pick XY",
			lx, float64(panelY)+74, fontS, colorMuted)

		if w.fullySelected() {
			slotX, slotY := uint8(w.selSlotX), uint8(w.selSlotY)
			slotZBottom := uint8(int(w.selZ) - int(f.Height) + 1)
			mult := storage.MultiplierAt(slotX, slotY, slotZBottom)
			expiry := model.ExpiryDate(char.CurrentDate, f.BaseExpiryDays, mult)
			ey, em, ed := expiry.Unpack()
			occupied := occupiedFoodSlots(&char.CurrentHome.RoomItems[w.fridgeIdx])
			canPlace := canFitFood(occupied, storage, f, slotX, slotY, slotZBottom)
			col := colorGreen
			zDesc := fmt.Sprintf("z=%d", w.selZ)
			if f.Height > 1 {
				zDesc = fmt.Sprintf("z%d–%d", slotZBottom, w.selZ)
			}
			msg := fmt.Sprintf("(x=%d,y=%d,%s)  %.1fx  exp:%04d-%02d-%02d",
				slotX, slotY, zDesc, mult, ey, em, ed)
			if !canPlace {
				col = colorRed
				msg = fmt.Sprintf("(x=%d,y=%d,%s) — does not fit", slotX, slotY, zDesc)
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
			mf := applyRotation(sf.Food, w.rotationIdx)
			drawText(dst, fmt.Sprintf("Moving: %s", mf.Name),
				lx, float64(panelY)+66, fontS, colorYellow)
			drawText(dst, fmt.Sprintf("Orientation: %s  (%dx%dx%d W×L×H)",
				rotationLabel[w.rotationIdx%6], mf.Width, mf.Length, mf.Height),
				lx, float64(panelY)+84, fontS, colorMuted)
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

	// compute active food height for Z validity and span display
	actF, hasActF := w.activeFood(&placed)
	var actFoodH uint8
	if hasActF {
		actFoodH = actF.Height
	}

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

		// Z validity for the active food
		validZ := !hasActF || isZSelectableFor(int(z), actFoodH, storage.Height)

		// Top-anchor: selZ is the TOP of the food; food spans from selZ down to selZ-foodH+1.
		inSpan := hasActF && w.selZ >= 0 && int(z) <= w.selZ && int(z) > w.selZ-int(actFoodH)
		isSelStart := w.selZ == int(z) // the top anchor cell

		hov := isHovered(mx, my, ox, cy, cs-1, cs-1) && validZ

		var bg color.RGBA
		switch {
		case isSelStart:
			bg = colorSelected
		case inSpan:
			bg = color.RGBA{40, 80, 160, 255} // blue tint = part of food span
		case !validZ:
			bg = color.RGBA{22, 22, 28, 255} // dimmed = can't start food here
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
		if isSelStart {
			borderCol = colorAccent
		} else if inSpan {
			borderCol = colorAccent
		} else if !validZ {
			borderCol = color.RGBA{35, 35, 42, 255}
		} else if hasCold {
			borderCol = color.RGBA{60, 120, 220, 255}
		}
		fillRect(dst, ox, cy, cs-1, cs-1, bg)
		strokeRect(dst, ox, cy, cs-1, cs-1, borderCol)

		zLabel := fmt.Sprintf("z%d", z)
		if hasCold {
			zLabel += "❄"
		}
		if !validZ {
			zLabel += "✗"
		}
		textCol := colorMuted
		if !validZ {
			textCol = color.RGBA{55, 55, 60, 255}
		}
		drawText(dst, zLabel, float64(ox)+3, float64(cy)+3, fontS, textCol)
		if !validZ {
			// no occupancy detail for invalid cells
		} else if isSelStart && actFoodH > 1 {
			// show span: top (selZ) down to selZ-foodH+1
			bottomZ := int(w.selZ) - int(actFoodH) + 1
			drawText(dst, fmt.Sprintf("z%d–%d", bottomZ, w.selZ),
				float64(ox)+3, float64(cy)+20, fontS, colorAccent)
		} else if usedSlots > 0 {
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
	rotLabel := fmt.Sprintf("↺ %s", rotationLabel[w.rotationIdx%6])
	showRotate := w.pendingFood != nil || (w.moveMode && w.movingFoodIdx >= 0)
	if showRotate {
		// Count how many valid rotations exist for this fridge
		validRotCount := 0
		var baseForRot model.Food
		if w.pendingFood != nil {
			baseForRot = *w.pendingFood
		} else if w.moveMode && w.movingFoodIdx >= 0 {
			baseForRot = placed.Stored[w.movingFoodIdx].Food
		}
		for i := 0; i < 6; i++ {
			if rotationFitsZ(baseForRot, i, storage.Height) {
				validRotCount++
			}
		}
		canRotate := validRotCount > 1
		drawButton(dst, rotLabel, rx, ry, rw, rh, fontS, isHovered(mx, my, rx, ry, rw, rh) && canRotate, canRotate)
		if !canRotate {
			drawText(dst, "(only flat fits)", float64(rx), float64(ry)+float64(rh)+2, fontS, colorMuted)
		}
	}

	// determine ghost footprint for pending/moving food
	var ghostW, ghostL int
	if w.pendingFood != nil {
		f := applyRotation(*w.pendingFood, w.rotationIdx)
		ghostW, ghostL = int(f.Width), int(f.Length)
	} else if w.moveMode && w.movingFoodIdx >= 0 {
		sf := placed.Stored[w.movingFoodIdx]
		mf := applyRotation(sf.Food, w.rotationIdx)
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
