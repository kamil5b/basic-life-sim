# Technical Guide

## Architecture overview

The game uses Ebitengine for rendering and input, with [furex v2](https://github.com/yohamta/furex) for flexbox-based UI layout.

### Screen lifecycle

```
Game.SetScreen(s)
  → g.rootView = s.BuildView(g)     # create furex view tree
  → g.rootView.Layout()             # compute flexbox frames immediately
  → g.main = screen-specific state   # store ref for Update/Draw routing

Game.Update()
  → advance game time (main screen)
  → handle Escape key
  → g.rootView.UpdateWithSize(w, h) # process furex events + layout
  → g.updater()                     # screens that need per-frame logic (newgame text input)
  → g.main.updateContent()          # room/shop panel precision click detection

Game.Draw(screen)
  → screen.Fill(colorBg)
  → g.rootView.Draw(screen)         # each handler draws in its frame
```

### View tree (main screen)

```
Game.rootView (furex.Row)
├── HUDView (Width: 220, FlexShrink: 0)
│   └── HUDHandler { furex.Drawer, furex.MouseLeftButtonHandler }
│       Draws: name, age, date, time, speed buttons, needs, money, [☰ Menu]
│       Clicks: speed buttons, menu toggle
│
└── ContentView (FlexGrow: 1)
    └── ContentHandler { furex.Drawer }
        Stores frame → mainScreen.contentFrame
        Draws: room OR shop based on mainScreen.mode
        Draws: menu overlay if open
```

Title screen and new game screen create their own root views (not children of Game.rootView). `Game.SetScreen()` replaces the entire root view.

### Handler pattern

Every interactive UI element is a furex `View` with a handler implementing at least `furex.Drawer`:

```go
type myHandler struct { ... }
func (h *myHandler) Draw(screen *ebiten.Image, frame image.Rectangle, v *furex.View) { ... }
```

**Coarse click targets** (buttons, tabs) additionally implement `furex.MouseLeftButtonHandler`:

```go
func (h *myHandler) HandleJustPressedMouseButtonLeft(x, y int) bool { ... }
func (h *myHandler) HandleJustReleasedMouseButtonLeft(x, y int) {}
```

**Precision targets** (grid cells, item list rows, storage slots) use manual detection:

```go
mx, my := ebiten.CursorPosition()
if isHovered(mx, my, frame.Min.X+offsetX, frame.Min.Y+offsetY, w, h) { ... }
```

### Coordinate system

All drawing is relative to `frame.Min.X/Y` from the handler's furex frame. No global position constants exist.

```go
// Any panel accessing its origin:
ox := p.main.panelX()  // → float32(main.contentFrame.Min.X)
oy := p.main.panelY()  // → float32(main.contentFrame.Min.Y)
ow := p.main.panelW()  // → float32(main.contentFrame.Dx())
oh := p.main.panelH()  // → float32(main.contentFrame.Dy())
```

`main.contentFrame` is set by `ContentHandler.Draw()` each frame before delegating to room/shop panels.

### State transitions without view tree rebuilds

The main screen (room/shop/menu) does NOT rebuild its view tree on state change. Instead:

- `mainScreen.mode` switches between `modeRoom`/`modeShop` — `ContentHandler.Draw()` branches
- `menuOverlay.open` toggles — `ContentHandler.Draw()` draws overlay conditionally
- Room panel modes (`rpMode*`) — `roomPanel.draw()` branches internally

Title and new game screens DO rebuild their view tree when step/state changes, because the UI structure differs. They call `rebuild(g)` which creates a new `furex.View` tree, calls `Layout()`, and assigns to `g.rootView`.

**Important:** `Layout()` is used instead of `UpdateWithSize()` inside event handlers. `UpdateWithSize` triggers `View.Update()` which re-processes mouse events — this causes infinite recursion if called from within a mouse handler. `Layout()` only computes flexbox positions without processing input.

### Room panel mode machine

```
rpModeList          ← default: show item list + grid, track hover
    │
    ├─ click cell with items  → rpModeFiltered
    ├─ click door cell        → rpModeDoorPopup
    │
rpModeFiltered      ← show items at selected cell
    ├─ click item row         → rpModeAction (room item) / rpModeFloorItemAction (floor item)
    ├─ ← Back                → rpModeList
    │
rpModeAction        ← show room item actions
    ├─ ✦ Move                → rpModeMoveGrid (placement wizard)
    ├─ Sell / Trash          → back to rpModeList
    ├─ 📦 Organize           → storage wizard
    ├─ ← Back                → rpModeFiltered
    │
rpModeFloorItemAction ← show floor item actions
    ├─ Eat / Use             → back to rpModeFiltered
    ├─ ✦ Move (toggle)      → floorMoveMode=true (click cell → move/rpModeFloorFoodFridgeChoice)
    ├─ Trash                 → back to rpModeFiltered
    ├─ ← Back                → rpModeFiltered
    │
rpModeDoorPopup     ← door exit popup
    ├─ Go to Shop            → main.mode = modeShop
    ├─ click elsewhere       → rpModeList
```

### Door system

Doors are `HomeCellDoor` cells in the room layout. The `HomeType.Doors` field maps door cell positions to their type:

```go
type DoorDef struct {
    X, Y uint8
    Type DoorType  // DoorRoomExit or DoorInternal
}
```

- `DoorRoomExit` → shows "Go to Shop" popup, switches to shop mode
- `DoorInternal` → reserved for future room transitions

### Shop flow

```
spModeCatalog       ← browse, select item
    └─ Buy & Place  → spModePlaceGrid
        │
spModePlaceGrid     ← click grid cell
    ├─ room item    → placement wizard (pwStepGrid → pwStepZ → pwStepDir)
    ├─ food         → finalizeFoodPlace (storage choice / stack confirm / floor)
    └─ utility      → finalizeUtilityPlace (storage choice / floor)
```

### Storage wizard

Two-phase selection inside a 3D fridge/cupboard:

1. **Phase 1:** Click Z cell in left column (top-down: highest Z first)
2. **Phase 2:** Click XY slot in grid (shown after Z selected)

Supports rotation (6 axis-aligned orientations) via rotation table. Food expiry multiplier varies by slot (cold zones have higher multipliers).

### Save system

Saves use Go's `encoding/gob`. Function fields (`DoAction`, `OnEat`, `ApplyExperience`) are nil'd before encode and restored from registries after decode. Max 5 save slots (`slot1.gob` – `slot5.gob` in `saves/`).

### Files NOT modified during UI migration

- `ui.go` — drawing primitives (unchanged API)
- `logic.go` — spatial helpers, nutrition, food mutation
- `save.go` — save/load serialization
- `fonts.go` — font initialization
- `time.go` — game clock advancement
- All `internal/model/*`, `internal/buyable/*`, `internal/experience/*`, `internal/constant/*`
