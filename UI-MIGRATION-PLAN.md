# furex-ui Migration + UI Revamp Plan

Migrate manual UI layout to [furex-ui](https://github.com/yohamta0/furex-ui) flexbox engine.
**Also**: remove tab bar, replace with in-world interactions (click door → shop, click food → eat, Escape/HUD → menu).

---

## Current State

| File | Lines | Role |
|------|-------|------|
| `internal/core/ui.go` | 186 | Drawing primitives: `fillRect`, `strokeRect`, `drawText`, `drawButton`, `listRow`, `needBar`, `drawFacingArrow` |
| `internal/core/game.go` | 43 | Root Ebitengine game, screen routing, layout constants |
| `internal/core/screen.go` | 10 | `Screen` interface: `Update`, `Draw` |
| `internal/core/main_screen.go` | 240 | Main screen shell: HUD panel, tab bar, panel routing |
| `internal/core/title.go` | 185 | Title screen: title, buttons, load screen |
| `internal/core/newgame.go` | 229 | New game wizard: name input, gender, home selection |
| `internal/core/menu_panel.go` | 255 | Menu panel: save, load, exit |
| `internal/core/food_panel.go` | 431 | Food list: eat, fridge, stove, floor actions, move grid |
| `internal/core/shop_panel.go` | 1248 | Shop: catalog, category/sort, placement wizard integration |
| `internal/core/placement_wizard.go` | 480 | Item placement state machine: grid pick → Z pick → direction pick |
| `internal/core/storage_wizard.go` | 973 | Fridge/storage 3D grid: slot XYZ pick, rotation, move |
| `internal/core/room_panel.go` | 1330 | Room view: grid draw, item list, actions, move, fridge |
| `internal/core/logic.go` | 558 | Spatial helpers, nutrition, food mutation — **no UI, stays unchanged** |

---

## UI Revamp — Interaction Design

### Before (current)

```
┌──────┬──────────────────────────────────────────────┐
│ HUD  │ [Room] [Shop] [Food] [Menu]   ← tab bar      │
│      ├──────────────────────────────────────────────┤
│      │                                              │
│      │   Tab content panel                          │
│      │                                              │
└──────┴──────────────────────────────────────────────┘
```

### After (revamp)

```
┌──────┬──────────────────────────────────────────────┐
│ HUD  │  Room view (always, full content area)       │
│      │  ┌──────────────────────────────────────┐    │
│      │  │ Item list │      Room grid           │    │
│      │  │           │   ██░░░░██               │    │
│      │  │ [Chair]   │   ░░[🚪]░░░░              │    │
│      │  │ [Stove]   │   ██░░██░░               │    │
│      │  │ [Fridge]  │   ░░░░░░░░               │    │
│ [☰] │  └──────────────────────────────────────┘    │
└──────┴──────────────────────────────────────────────┘
```

### Navigation flows

```
ROOM (default mode)
 │
 ├─ Click door cell ──→ Door popup (near door)
 │   If RoomExit door:  [Go to Shop] ──→ mode = shop (replaces content area)
 │   If Internal door:  reserved for future room transitions
 │                        └─ click away ──→ popup closes
 │
 ├─ Click floor item ──→ Item selected in left sidebar
 │   (floor items can be food OR utilities: fork, knife, plate, etc.)
 │                        └─ Action buttons: Eat (if food) | Move | Trash
 │
 ├─ Click "Move" on selected floor item ──→ Move mode
 │   └─ Click target room item (fridge/stove/cupboard) ──→ storage wizard
 │
 ├─ Click room item in list ──→ Item selected
 │                               └─ Action buttons: Move | Sell | Trash | Organize | Use
 │
 ├─ Click room item on grid (e.g. fridge, stove) ──→ Storage wizard
 │
 ├─ Escape key ──→ Menu overlay (modal on top of room)
 │                 └─ Save | Load | Exit
 │                 └─ Escape or [Back] ──→ close overlay
 │
 └─ HUD [☰ Menu] button ──→ same as Escape

SHOP mode (full content area)
 │
 ├─ Category tabs, sort controls, paged catalog
 ├─ Buy item ──→ Placement wizard (grid → Z → direction)
 │               └─ Confirm ──→ item placed, return to ROOM mode
 │
 └─ Escape or [← Back] ──→ return to ROOM mode
```

### What gets removed

| Removed | Replacement |
|---------|-------------|
| Tab bar (Room / Shop / Food / Menu) | In-world clicks + Escape + HUD button |
| `mainTab` type, `tabLabels`, `tabRect()`, tab routing | `mainMode` type: `modeRoom`, `modeShop` |
| `food_panel.go` as standalone tab panel | Absorbed into room: click floor item → context actions in left sidebar |
| `menu_panel.go` as standalone tab panel | Becomes `menuOverlay` — modal overlay toggled by Escape or HUD [☰] |

### Door modeling (model change)

Doors are `HomeCellDoor` cells in the layout grid. Multiple doors allowed, minimum 1.
One door is always the **Room Exit Door** — connects to outside world (Shop, shared bathroom/kitchen).
Other doors are internal (bathroom door, bedroom door).

**Model addition (`HomeType`):**
```go
type DoorDef struct {
    X, Y uint8
    Type DoorType  // RoomExit, Bathroom, Bedroom, etc.
}

type DoorType uint8
const (
    DoorRoomExit DoorType = iota
    DoorInternal
)
```

`HomeType.Doors []DoorDef` — maps door cells to their purpose. Clicking a `DoorRoomExit` shows the "Go to Shop" popup.
Internal doors: reserved for future room transitions (bathroom, bedroom).

---

## Architecture (furex-ui)

### View Tree

```
Game.rootView (furex.Row)
│
├── HUDView (Width: 220, FlexShrink: 0)
│   └── Handler: HUDHandler
│       - Character name, age, date, time
│       - Speed buttons (0.5x–100x)
│       - Money display
│       - Needs bars (Food, Energy, Hygiene, Confidence, Strength)
│       - Home type name
│       - [☰ Menu] button (bottom of HUD)
│
└── ContentView (FlexGrow: 1)
    └── Handler: ContentHandler
        - mainScreen.mode determines active view:
            modeRoom → room grid + item list + food context + door popup + storage wizard
            modeShop → shop catalog + placement wizard
        - Menu overlay: drawn on top when open (Escape / HUD [☰] toggles)
```

Title screen and new game screen use **separate root views**, set when `Game.SetScreen()` is called.

### Handler pattern

All panel handlers implement at minimum:

```go
type panelHandler interface {
    furex.Drawer  // Draw(screen *ebiten.Image, frame image.Rectangle, view *furex.View)
}
```

- **Coarse click targets** (HUD buttons, menu overlay, shop category tabs): implement `furex.MouseLeftButtonHandler` or `furex.MouseHandler`.
- **Precision targets** (grid cells, list items, Z-column cells, storage slots): keep manual `ebiten.CursorPosition()` + `isHovered()` relative to `frame.Min`.

### Coordinate translation

```
Global cursor:  mx, my = ebiten.CursorPosition()
Local cursor:   lx = mx - frame.Min.X
                ly = my - frame.Min.Y
```

All drawing uses `frame.Min.X/Y` as origin.

---

## What Changes (summary)

### Removed

- All `*Rect()` helper functions (~50+ across 8 files)
- Global layout constants: `hudW`, `tabH`, `panelX`, `panelY`, `panelW`, `panelH`
- Panel-specific layout constants: `spListW`, `zColX`, `zColY`, `fpListX`, `fpListY`, `fpBtnY`, `rpListX`, `rpGridX`, `rpGridY`, `fpGridOriginX/Y`, `mp*BtnRect`, `ngBtnX/Y`, etc.
- `mainTab` type, `tabLabels`, `tabRect()`, `tab` field, tab routing
- `food_panel.go` — absorbed into `room_panel.go`
- `menu_panel.go` — replaced by `menuOverlay` in `main_screen.go`

### Modified structure

| File | Change |
|------|--------|
| `game.go` | Add furex root view; remove `ScreenW`/`ScreenH` constants; `Layout()` returns window size |
| `screen.go` | `Screen` interface gains `BuildView(g) *furex.View` method |
| `main_screen.go` | Remove tab system; add `mainMode`, `menuOverlay`; build view tree |
| `room_panel.go` | Merge food interaction; add door popup; receive frame from handler |
| `shop_panel.go` | Receive frame; remove rect helpers; accessed via door not tab |
| `placement_wizard.go` | Receive sub-frame from shop; remove global rect references |
| `storage_wizard.go` | Receive sub-frame from room; remove global rect references |
| `title.go` | Build view tree via `BuildView()`; remove rect helpers |
| `newgame.go` | Build view tree via `BuildView()`; remove rect helpers |

### Unchanged

- `ui.go` — drawing primitives (called with handler-local coords instead of global)
- `logic.go` — game logic entirely
- `save.go` — save/load
- `fonts.go` — font init
- `time.go` — time advancement
- All `internal/model/*` — data models
- All `internal/buyable/*` — item/food definitions
- All `internal/experience/*` — experience definitions
- All `internal/constant/*` — constants
- `rpCellSz` and other visual scaling constants (not position-related)
- `ebiten.CursorPosition()` + `isHovered()` — kept for per-cell precision

---

## Phases

### Phase 1 — Scaffold

**Files:** `go.mod`, `game.go`, `screen.go`, `main_screen.go`

1. Add `github.com/yohamta/furex/v2` dependency
2. `Game` struct gains `rootView *furex.View` field
3. `Game.Layout(ow, oh int) (int, int)` returns `ow, oh` (furex handles resize); clamp minimum 800×600
4. `Game.Update()` → `g.rootView.Update()`
5. `Game.Draw(screen)` → `screen.Fill(colorBg)` then `g.rootView.Draw(screen)`
6. `Game.SetScreen(s Screen)` → builds new rootView via `s.BuildView(g)`
7. New `Screen` interface:
   ```go
   type Screen interface {
       BuildView(g *Game) *furex.View
   }
   ```
8. `mainScreen` refactored:
   ```go
   type mainMode int
   const (
       modeRoom mainMode = iota
       modeShop
   )

   type mainScreen struct {
       char        model.Character
       mode        mainMode
       room        *roomPanel
       shop        *shopPanel
       menuOverlay *menuOverlay
       message     string
       lastUpdate  time.Time
   }
   ```
9. `mainScreen.BuildView()` creates:
   - `HUDView` with `HUDHandler` (contains [☰ Menu] button)
   - `ContentView` with `ContentHandler` (delegates to room or shop based on mode)
10. `HUDHandler` implements `furex.Drawer` + `furex.MouseLeftButtonHandler` for speed buttons and menu button
11. `ContentHandler` implements `furex.Drawer` — calls `room.draw()` or `shop.draw()` with its frame
12. `menuOverlay` struct: draws save/load/exit on top of room when `open == true`; toggled by Escape or HUD [☰]

**Removed:**
- `ScreenW`, `ScreenH` global constants
- `hudW`, `tabH`, `panelX`, `panelY`, `panelW`, `panelH`
- `mainTab`, `tabLabels`, `tabRect()`, tab routing logic

---

### Phase 2 — Title Screen + New Game

**Files:** `title.go`, `newgame.go`

**title.go:**
- `titleScreen` implements `Screen.BuildView(g)`:
  ```
  Root (furex.Column, JustifyCenter, AlignItemsCenter)
  ├── TitleLabel ("BASIC LIFE SIMULATOR")
  ├── SubLabel ("A life simulation game")
  ├── [New Game] button  → g.SetScreen(newNewGameScreen())
  ├── [Load Game] button → load mode (rebuild tree)
  └── [Exit] button      → ebiten.Termination
  ```
- Load mode: replaces tree with save slot rows + Load/Delete buttons + Back button
- Remove: `titleButtonRect()`, `titleSaveRowRect()`, `titleLoadActionRect()`

**newgame.go:**
- `newGameScreen` implements `Screen.BuildView(g)`:
  ```
  Root (furex.Column, JustifyCenter, AlignItemsCenter)
  ├── StepLabel ("Enter name" / "Choose gender" / "Choose home")
  ├── Content (varies per step)
  │   ├── Step 1: text input box + [Continue] button
  │   ├── Step 2: [♂ Male] [♀ Female] row
  │   └── Step 3: home cards row + [Start Game] button
  └── Error label (shown when errMsg != "")
  ```
- Remove: `ngBtnX()`, `ngBtnY()`, `ngHomeCardRect()`

---

### Phase 3 — Room Panel + Food + Door + Storage Wizard

**Files:** `room_panel.go`, `food_panel.go` (absorbed), `storage_wizard.go`

This is the largest phase. Room panel absorbs food interaction and adds door interaction.

**roomPanel struct (updated):**
```go
type roomPanel struct {
    char    *model.Character
    main    *mainScreen

    // Grid interaction
    hoverCell    [2]int
    selCell      [2]int   // selected grid cell

    // Item list
    selItem      int      // selected room item index in item list

    // Floor item context (was foodPanel)
    floorSelIdx  int      // selected floor item index
    floorSources []floorItemSource
    floorMode    fpMode
    floorMoveIdx int

    // Door popup
    doorPopupOpen bool
    doorCell      [2]int  // which door was clicked

    // Item actions
    selAct       int      // selected action

    // Modes
    mode         rpMode   // list, filtered, action, moveGrid, moveZ, moveDir, etc.
    wizard       *placementWizard
    storageWiz   *storageWizard

    // Floor interaction
    floorMoveMode           bool
    floorFoodFridgeTarget   int
    floorFoodFridgeTargetXY [2]int
    utilTarget              int
}
```

**Layout (within ContentHandler frame):**
```
frame.Min.X ─────────────────────────── frame.Max.X
│ Item list (left 280px) │ Room grid (rest)           │
│ ┌────────────────────┐ │ ┌────────────────────────┐ │
│ │ [Chair]            │ │ │ ██░░░░██               │ │
│ │ [Stove]            │ │ │ ░░[🚪]░░░░              │ │
│ │ [Fridge]           │ │ │ ██░░██░░               │ │
│ │ ────────────────── │ │ │ ░░░░░░░░               │ │
│ │ Food context       │ │ │                        │ │
│ │ (when food sel)    │ │ │  [door popup]          │ │
│ │ [Eat][→Fridge]...  │ │ │                        │ │
│ └────────────────────┘ │ └────────────────────────┘ │
│ [Move] [Sell] [Trash]  │                            │
│ [Organize]             │                            │
└────────────────────────┴────────────────────────────┘
```

**Door interaction logic:**
- In room grid draw: door cells get distinct highlight when hovered
- Click door cell → look up door type from `home.Type.Doors`
- `DoorRoomExit`: show popup with "[Go to Shop]" button → `main.mode = modeShop`
- `DoorInternal`: reserved (no action yet)
- Click anywhere else → `doorPopupOpen = false`

**Floor item context (absorbed from food_panel.go):**
- Click floor item on grid (food or utility) → populate sources, select first
- Item info + action buttons draw in left sidebar, below item list or replacing it
- Actions depend on item kind:
  - Food: Eat, Move, Trash
  - Utility: Move, Trash, Use (if applicable)
- **Move flow (two-step):**
  1. Click "Move" → room enters move mode (item follows cursor)
  2. Click target room item (fridge/stove/cupboard) → storage wizard opens for that item
  3. Storage wizard: place item in slot grid → confirm → item moved

**Storage wizard:**
- Receives sub-frame (full ContentHandler frame)
- Draws 3D slot grid + rotation controls
- [Back] button or Escape closes it, returns to room
- Remove: `rpListX`, `rpListW`, `rpListY`, `rpGridX`, `rpGridY`, all `rp*BtnRect()`, `fpGridOriginX/Y`, `fpCancelBtnRect()`
- Keep: `rpCellSz`, `fwCellSz`

---

### Phase 4 — Shop Panel + Placement Wizard

**Files:** `shop_panel.go`, `placement_wizard.go`

**Layout (within ContentHandler frame):**
```
frame.Min.X ─────────────────────────── frame.Max.X
│ [← Back]  Shop Catalog                              │
│ ┌──────────────────────────────────────────────────┐│
│ │ Category: [Appliance][Furniture][Hygiene]...[All]││
│ │ Sort: [Name] [Price] [Volume]                    ││
│ ├──────────────────────────────────────────────────┤│
│ │ Row 0: [Item Name]  $price  dimensions  [Buy]   ││
│ │ Row 1: ...                                       ││
│ │ ...                                              ││
│ ├──────────────────────────────────────────────────┤│
│ │ [← Page]  Page 1/3  [→ Page]                     ││
│ └──────────────────────────────────────────────────┘│
│                                                     │
│ Placement mode (when buying):                       │
│ ┌────────────────────┬─────────────────────────────┐│
│ │ Left panel         │ Room grid (placement target)││
│ │ Step 2: Pick Z     │ ██░░░░██                    ││
│ │ [item name]        │ ░░░░░░░░                    ││
│ │ pos (x,y)          │ ██░░██░░                    ││
│ │                    │                             ││
│ │ Z levels:          │ [ghost overlay]             ││
│ │  2 ← anchor        │                             ││
│ │  1 [X]             │                             ││
│ │  0 [ ]             │                             ││
│ │ [Confirm Z →]      │                             ││
│ └────────────────────┴─────────────────────────────┘│
│ [✕ Cancel]                                          │
└─────────────────────────────────────────────────────┘
```

- `[← Back]` button or Escape → `main.mode = modeRoom`
- Category tabs, sort controls in horizontal rows at top
- Catalog rows: vertical list with paging
- Placement wizard receives sub-frames for left panel + grid area
- Remove: `spListW`, `spCatTabRect()`, `spItemRowRect()`, `spBuyBtnRect()`, `spCancelBtnRect()`, `spStackMixBtnRect()`, `spStackSepBtnRect()`, `spRecipeBtnRect()`, `spGridOriginX/Y`, `spZConfirmBtnRect()`, `spDirBtnRect()`, `zColX`, `zColY`, `zCellW`, `zCellH`
- Z-column picker: drawn using relative coords within left-panel sub-frame
- Direction picker: horizontal row of 4 buttons within left-panel sub-frame

---

### Phase 5 — Menu Overlay

**Files:** `main_screen.go` (new `menuOverlay` type), delete `menu_panel.go`

```go
type menuOverlay struct {
    main    *mainScreen
    open    bool
    mode    mpMode  // normal, save, load (reused from current menu_panel.go)
    selSlot int
    infos   []SaveInfo
}
```

- Drawn as semi-transparent overlay on top of room/shop content
- Centered panel with: Save Game / Load Game / Exit to Title
- Save/Load mode: fills overlay with slot rows + Save/Load/Delete buttons + Back
- Escape toggles `open`
- HUD [☰ Menu] button toggles `open`
- `open == true` → blocks clicks from passing through to room/shop

---

### Phase 6 — Cleanup

**Files:** all modified files

1. Remove all `*Rect()` helper functions (grep for `Rect(` across `internal/core/`)
2. Remove any remaining references to `ScreenW`, `ScreenH`, `hudW`, `tabH`, `panelX`, `panelY`, `panelW`, `panelH`
3. Verify all drawing uses `frame.Min.X/Y` as origin, never global constants
4. Delete `food_panel.go` (absorbed into room_panel.go)
5. Delete `menu_panel.go` (replaced by menuOverlay in main_screen.go)
6. Window resize: verify furex recalculates frames correctly; `Layout()` returns `max(ow, 800), max(oh, 600)`
7. Test full flow:
   - Title → New Game (name, gender, home) → Main
   - Click door → Shop → buy item → place → back to room
   - Click floor food → eat
   - Click fridge → storage wizard → interact → back
   - Click room item → move / sell / trash
   - Escape → menu → save / load / exit
   - HUD [☰] → menu → same
   - Resize window → layout adapts

---

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|-----------|
| Room grid drawing (~150 lines) | High | Keep as custom draw. Only origin changes. Test each cell type + overlay. |
| Floor item context absorbed into room panel | High | `food_panel.go` logic is self-contained. Move into `roomPanel` struct. Test food + utility actions. |
| Door popup position + hit testing | Medium | Popup drawn at fixed offset from door cell. Use frame-relative coords. |
| Placement wizard state machine | Medium | Wizard is self-contained. Receives sub-frames instead of globe origins. |
| Storage wizard 3D slot grid | Medium | Same treatment. Internal grid math unchanged. |
| Menu overlay click passthrough | Medium | When `open == true`, ContentHandler skips its own click handling. |
| Window resize + grid click mapping | Medium | `gridCellAt()` subtracts `frame.Min`. Test with varied window sizes. |
| Cursor coordinate translation errors | Medium | Audit every `isHovered()` site. Add `lx = mx - frame.Min.X` helper. |
| Handler lifecycle (persist across frames) | Low | Panel structs already persist in `mainScreen`. |
| Shop accessed while placement ghost active | Low | Door popup only available when room is in list/action mode (not move/place). |
| Performance (grid as furex views) | N/A | Grid cells are custom-drawn, not individual furex views. |

---

## Files NOT Changed

- `internal/core/logic.go`
- `internal/core/save.go`
- `internal/core/fonts.go`
- `internal/core/time.go`
- `internal/core/ui.go` (drawing primitives unchanged)
- `internal/model/*` (all data models)
- `internal/buyable/*` (all item/food definitions)
- `internal/experience/*` (all experience definitions)
- `internal/constant/*` (all constants)
- `main.go`
