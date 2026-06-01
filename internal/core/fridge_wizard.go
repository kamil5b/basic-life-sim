package core

import (
	"fmt"
	"image/color"

	"github.com/kamil5b/basic-life-sim/internal/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const fwCellSz = float32(56)

// fridgeWizard is a reusable component for managing food inside a fridge.
// It is used both when buying food (shop context, pendingFood != nil) and
// when organising an existing fridge (room context, pendingFood == nil).
type fridgeWizard struct {
	char      *model.Character
	fridgeIdx int // index into char.CurrentHome.RoomItems

	// navigation
	currentZ uint8

	// slot selection
	selSlotX int // -1 = none
	selSlotY int

	// within-fridge move mode (room context only)
	moveMode      bool
	movingFoodIdx int // index into placed.Stored; -1 = none

	// shop context: food pending purchase (nil = room context)
	pendingFood *model.Food

	// pixel origin of the fridge grid (right area of the panel)
	gridX, gridY float32
	// x for left-panel info text
	infoX float32

	// callbacks
	setMsg   func(string)
	onCancel func()
	onDone   func()
}

func newFridgeWizard(
	char *model.Character,
	fridgeIdx int,
	gridX, gridY, infoX float32,
	pendingFood *model.Food,
	setMsg func(string),
	onCancel func(),
	onDone func(),
) *fridgeWizard {
	return &fridgeWizard{
		char:          char,
		fridgeIdx:     fridgeIdx,
		currentZ:      0,
		selSlotX:      -1,
		selSlotY:      -1,
		movingFoodIdx: -1,
		pendingFood:   pendingFood,
		gridX:         gridX,
		gridY:         gridY,
		infoX:         infoX,
		setMsg:        setMsg,
		onCancel:      onCancel,
		onDone:        onDone,
	}
}

// ── button layout ─────────────────────────────────────────────────────────────

func (w *fridgeWizard) backBtnRect() (x, y, width, height float32) {
	return panelX + 4, float32(ScreenH) - float32(tabH) - 50, 100, 32
}

// action button 1: "Take Out" (room) or "✓ Place Here" (shop)
func (w *fridgeWizard) action1BtnRect() (x, y, width, height float32) {
	return panelX + 110, float32(ScreenH) - float32(tabH) - 50, 140, 32
}

// action button 2: "✦ Move / ✦ Cancel" (room only)
func (w *fridgeWizard) action2BtnRect() (x, y, width, height float32) {
	return panelX + 258, float32(ScreenH) - float32(tabH) - 50, 130, 32
}

func (w *fridgeWizard) zUpBtnRect() (x, y, width, height float32) {
	return w.gridX, w.gridY - 30, 44, 24
}

func (w *fridgeWizard) zDownBtnRect() (x, y, width, height float32) {
	return w.gridX + 50, w.gridY - 30, 44, 24
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (w *fridgeWizard) gridCellAt(px, py int, storage *model.StorageCapacity) (gx, gy int) {
	ox, oy, cs := int(w.gridX), int(w.gridY), int(fwCellSz)
	gx = (px - ox) / cs
	gy = (py - oy) / cs
	if gx < 0 || gy < 0 || gx >= int(storage.Width) || gy >= int(storage.Length) {
		return -1, -1
	}
	return gx, gy
}

func (w *fridgeWizard) findFoodAtSel(placed *model.PlacedRoomItem) int {
	for i, s := range placed.Stored {
		if s.SlotZ == w.currentZ && int(s.SlotX) == w.selSlotX && int(s.SlotY) == w.selSlotY {
			return i
		}
	}
	return -1
}

// ── update ────────────────────────────────────────────────────────────────────

// update processes input. Returns true when the wizard is finished (done or cancelled).
func (w *fridgeWizard) update() bool {
	mx, my := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if !clicked {
		return false
	}

	placed := &w.char.CurrentHome.RoomItems[w.fridgeIdx]
	storage := placed.Item.Storage

	// Back / Cancel
	bx, by, bw, bh := w.backBtnRect()
	if isHovered(mx, my, bx, by, bw, bh) {
		w.onCancel()
		return true
	}

	// Z navigation
	zUpX, zUpY, zUpW, zUpH := w.zUpBtnRect()
	if isHovered(mx, my, zUpX, zUpY, zUpW, zUpH) && int(w.currentZ) < int(storage.Height)-1 {
		w.currentZ++
		w.selSlotX = -1
		return false
	}
	zDnX, zDnY, zDnW, zDnH := w.zDownBtnRect()
	if isHovered(mx, my, zDnX, zDnY, zDnW, zDnH) && w.currentZ > 0 {
		w.currentZ--
		w.selSlotX = -1
		return false
	}

	if w.pendingFood != nil {
		// ── shop context ──────────────────────────────────────────────────────
		if w.selSlotX >= 0 {
			a1x, a1y, a1w, a1h := w.action1BtnRect()
			if isHovered(mx, my, a1x, a1y, a1w, a1h) {
				slotX, slotY := uint8(w.selSlotX), uint8(w.selSlotY)
				f := *w.pendingFood
				occupied := occupiedFoodSlots(placed)
				if canFitFood(occupied, storage, f, slotX, slotY, w.currentZ) {
					mult := storage.MultiplierAt(slotX, slotY, w.currentZ)
					w.char.CurrentStats.Money -= f.BasePrice
					placed.Stored = append(placed.Stored, model.StoredFood{
						SlotX: slotX, SlotY: slotY, SlotZ: w.currentZ,
						PurchaseDate:   w.char.CurrentDate,
						MultiplierUsed: mult,
						UsesRemaining:  f.UsesTotal,
						Food:           f,
					})
					expiry := model.ExpiryDate(w.char.CurrentDate, f.BaseExpiryDays, mult)
					ey, em, ed := expiry.Unpack()
					w.setMsg(fmt.Sprintf("Bought %s → slot (%d,%d,%d), exp %04d-%02d-%02d. Money: $%.2f",
						f.Name, slotX, slotY, w.currentZ, ey, em, ed, w.char.CurrentStats.Money))
					w.onDone()
					return true
				}
				w.setMsg("Slot occupied or too small for this food.")
				return false
			}
		}
	} else {
		// ── room context ──────────────────────────────────────────────────────
		if w.selSlotX >= 0 && !w.moveMode {
			// Take Out
			a1x, a1y, a1w, a1h := w.action1BtnRect()
			if isHovered(mx, my, a1x, a1y, a1w, a1h) {
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
					w.char.CurrentHome.FloorFood = append(w.char.CurrentHome.FloorFood, model.PlacedFood{
						X: placed.X, Y: placed.Y, Z: 0,
						PurchaseDate:  w.char.CurrentDate,
						UsesRemaining: s.UsesRemaining,
						Food:          newFood,
					})
					expiry := model.ExpiryDate(w.char.CurrentDate, newFood.BaseExpiryDays, 1)
					ey, em, ed := expiry.Unpack()
					w.setMsg(fmt.Sprintf("Took out %s. New exp %04d-%02d-%02d", s.Food.Name, ey, em, ed))
					w.selSlotX = -1
					w.onDone()
				}
				return false
			}
		}
		// Move toggle
		if w.selSlotX >= 0 {
			a2x, a2y, a2w, a2h := w.action2BtnRect()
			if isHovered(mx, my, a2x, a2y, a2w, a2h) {
				if !w.moveMode {
					idx := w.findFoodAtSel(placed)
					if idx >= 0 {
						w.movingFoodIdx = idx
						w.moveMode = true
						w.setMsg("Click a target slot to move the food there.")
					}
				} else {
					w.moveMode = false
					w.movingFoodIdx = -1
				}
				return false
			}
		}
	}

	// Grid cell click
	gx, gy := w.gridCellAt(mx, my, storage)
	if gx < 0 {
		return false
	}
	if w.moveMode && w.movingFoodIdx >= 0 {
		sf := &placed.Stored[w.movingFoodIdx]
		occupied := occupiedFoodSlots(placed)
		for dz := uint8(0); dz < sf.Food.Height; dz++ {
			for dy := uint8(0); dy < sf.Food.Length; dy++ {
				for dx := uint8(0); dx < sf.Food.Width; dx++ {
					delete(occupied, [3]uint8{sf.SlotX + dx, sf.SlotY + dy, sf.SlotZ + dz})
				}
			}
		}
		slotX, slotY := uint8(gx), uint8(gy)
		if canFitFood(occupied, storage, sf.Food, slotX, slotY, w.currentZ) {
			newMult := storage.MultiplierAt(slotX, slotY, w.currentZ)
			sf.SlotX, sf.SlotY, sf.SlotZ = slotX, slotY, w.currentZ
			sf.MultiplierUsed = newMult
			w.setMsg(fmt.Sprintf("Moved %s to slot (%d,%d,%d)", sf.Food.Name, slotX, slotY, w.currentZ))
		} else {
			w.setMsg("Slot occupied or out of bounds.")
		}
		w.moveMode = false
		w.movingFoodIdx = -1
		w.selSlotX = -1
	} else {
		w.selSlotX = gx
		w.selSlotY = gy
		w.moveMode = false
		w.movingFoodIdx = -1
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

	// ── left panel info ────────────────────────────────────────────────────────
	used := usedSlotCount(&char.CurrentHome.RoomItems[w.fridgeIdx])
	total := storage.TotalSlots()

	if w.pendingFood != nil {
		f := *w.pendingFood
		drawText(dst, "Place in "+placed.Item.Name, lx, float64(panelY)+14, fontM, colorAccent)
		drawText(dst, fmt.Sprintf("Buying: %s  $%.2f  uses:%d  exp:%dd",
			f.Name, f.BasePrice, f.UsesTotal, f.BaseExpiryDays),
			lx, float64(panelY)+38, fontS, colorText)
		drawText(dst, fmt.Sprintf("Slots: %d/%d used  |  Layer Z=%d / %d",
			used, total, w.currentZ, int(storage.Height)-1),
			lx, float64(panelY)+58, fontS, colorMuted)
		drawText(dst, "❄ = cold zone (longer expiry)", lx, float64(panelY)+76, fontS, color.RGBA{80, 160, 255, 255})

		if w.selSlotX >= 0 {
			slotX, slotY := uint8(w.selSlotX), uint8(w.selSlotY)
			mult := storage.MultiplierAt(slotX, slotY, w.currentZ)
			expiry := model.ExpiryDate(char.CurrentDate, f.BaseExpiryDays, mult)
			ey, em, ed := expiry.Unpack()
			occupied := occupiedFoodSlots(&char.CurrentHome.RoomItems[w.fridgeIdx])
			canPlace := canFitFood(occupied, storage, f, slotX, slotY, w.currentZ)
			col := colorGreen
			msg := fmt.Sprintf("Slot (%d,%d,%d)  %.1fx  exp:%04d-%02d-%02d", slotX, slotY, w.currentZ, mult, ey, em, ed)
			if !canPlace {
				col = colorRed
				msg = fmt.Sprintf("Slot (%d,%d,%d) — OCCUPIED", slotX, slotY, w.currentZ)
			}
			drawText(dst, msg, lx, float64(panelY)+100, fontS, col)

			a1x, a1y, a1w, a1h := w.action1BtnRect()
			drawButton(dst, "✓ Place Here", a1x, a1y, a1w, a1h, fontM,
				isHovered(mx, my, a1x, a1y, a1w, a1h) && canPlace, canPlace)
		} else {
			drawText(dst, "↗ Click a slot on the right to select it", lx, float64(panelY)+100, fontS, colorMuted)
		}
	} else {
		drawText(dst, placed.Item.Name+" Interior", lx, float64(panelY)+4, fontM, colorAccent)
		drawText(dst, fmt.Sprintf("%d/%d slots used  |  Layer Z=%d / %d",
			used, total, w.currentZ, int(storage.Height)-1),
			lx, float64(panelY)+26, fontS, colorMuted)

		if w.selSlotX >= 0 && w.selSlotX < int(storage.Width) {
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
				drawText(dst, fmt.Sprintf("Selected: %s", sf.Food.Name), lx, float64(panelY)+50, fontS, col)
				drawText(dst, fmt.Sprintf("uses:%d  slot:(%d,%d,%d)  exp:%04d-%02d-%02d",
					sf.UsesRemaining, sf.SlotX, sf.SlotY, sf.SlotZ, ey, em, ed),
					lx, float64(panelY)+68, fontS, colorMuted)
				if w.moveMode {
					drawText(dst, "→ Click a target slot", lx, float64(panelY)+88, fontS, colorYellow)
				}
				// Take Out
				a1x, a1y, a1w, a1h := w.action1BtnRect()
				takeEnabled := !w.moveMode
				drawButton(dst, "Take Out", a1x, a1y, a1w, a1h, fontS,
					isHovered(mx, my, a1x, a1y, a1w, a1h) && takeEnabled, takeEnabled)
				// Move
				a2x, a2y, a2w, a2h := w.action2BtnRect()
				moveLabel := "✦ Move"
				if w.moveMode {
					moveLabel = "✦ Cancel"
				}
				drawButton(dst, moveLabel, a2x, a2y, a2w, a2h, fontS,
					isHovered(mx, my, a2x, a2y, a2w, a2h), true)
			} else {
				drawText(dst, fmt.Sprintf("Empty slot (%d,%d,%d)", w.selSlotX, w.selSlotY, w.currentZ),
					lx, float64(panelY)+50, fontS, colorMuted)
			}
		} else {
			drawText(dst, "Click a slot to select food", lx, float64(panelY)+50, fontS, colorMuted)
		}
	}

	// Back / Cancel
	bx, by, bw, bh := w.backBtnRect()
	label := "← Back"
	if w.pendingFood != nil {
		label = "✕ Cancel"
	}
	drawButton(dst, label, bx, by, bw, bh, fontS, isHovered(mx, my, bx, by, bw, bh), true)

	// ── fridge grid ────────────────────────────────────────────────────────────
	ox, oy, cs := w.gridX, w.gridY, fwCellSz

	// Z nav
	zUpX, zUpY, zUpW, zUpH := w.zUpBtnRect()
	zDnX, zDnY, zDnW, zDnH := w.zDownBtnRect()
	drawButton(dst, "Z ▲", zUpX, zUpY, zUpW, zUpH, fontS,
		isHovered(mx, my, zUpX, zUpY, zUpW, zUpH), int(w.currentZ) < int(storage.Height)-1)
	drawButton(dst, "Z ▼", zDnX, zDnY, zDnW, zDnH, fontS,
		isHovered(mx, my, zDnX, zDnY, zDnW, zDnH), w.currentZ > 0)
	drawText(dst, fmt.Sprintf("Layer Z=%d / %d", w.currentZ, int(storage.Height)-1),
		float64(ox)+100, float64(oy)-22, fontS, colorMuted)

	occupied := occupiedFoodSlots(&char.CurrentHome.RoomItems[w.fridgeIdx])

	for row := 0; row < int(storage.Length); row++ {
		for col := 0; col < int(storage.Width); col++ {
			cx := ox + float32(col)*cs
			cy := oy + float32(row)*cs
			slotKey := [3]uint8{uint8(col), uint8(row), w.currentZ}
			isOccupied := occupied[slotKey]
			isSel := w.selSlotX == col && w.selSlotY == row
			isCold := false
			for _, cz := range storage.ColdZones {
				if cz.Contains(uint8(col), uint8(row), w.currentZ) {
					isCold = true
				}
			}
			hov := isHovered(mx, my, cx, cy, cs-1, cs-1)

			var bg color.RGBA
			switch {
			case isSel && !isOccupied:
				bg = colorSelected
			case isSel && isOccupied:
				bg = color.RGBA{80, 20, 20, 255}
			case isOccupied:
				bg = color.RGBA{40, 55, 100, 255}
			case hov && w.moveMode:
				bg = color.RGBA{50, 100, 50, 200}
			case hov:
				bg = colorHighlight
			case isCold:
				bg = color.RGBA{20, 40, 80, 255}
			default:
				bg = color.RGBA{28, 35, 50, 255}
			}
			borderCol := colorBorder
			if isSel {
				borderCol = colorAccent
			} else if isCold {
				borderCol = color.RGBA{60, 120, 220, 255}
			}
			fillRect(dst, cx, cy, cs-1, cs-1, bg)
			strokeRect(dst, cx, cy, cs-1, cs-1, borderCol)

			coordLabel := fmt.Sprintf("%d,%d", col, row)
			if isCold {
				coordLabel += " ❄"
			}
			drawText(dst, coordLabel, float64(cx)+4, float64(cy)+4, fontS, colorMuted)

			if !isOccupied {
				mult := storage.MultiplierAt(uint8(col), uint8(row), w.currentZ)
				drawText(dst, fmt.Sprintf("%.1fx", mult), float64(cx)+4, float64(cy)+22, fontS, colorGreen)
			} else {
				for _, s := range placed.Stored {
					if s.SlotZ == w.currentZ && int(s.SlotX) == col && int(s.SlotY) == row {
						exp := model.IsExpired(s.PurchaseDate, s.Food.BaseExpiryDays, s.MultiplierUsed, char.CurrentDate)
						tc := colorText
						if exp {
							tc = colorRed
						}
						name := s.Food.Name
						if len(name) > 8 {
							name = name[:8]
						}
						drawText(dst, name, float64(cx)+4, float64(cy)+22, fontS, tc)
						if exp {
							drawText(dst, "EXP", float64(cx)+4, float64(cy)+38, fontS, colorRed)
						}
						break
					}
				}
			}
		}
	}
}
