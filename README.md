# Basic Life Simulator

A life simulation game built with Go + [Ebitengine](https://github.com/hajimehoshi/ebiten). Live in a room, buy furniture, store food, cook, eat, manage needs, and survive.

## Quick Start

```bash
go run .
```

- **Requires Go 1.25+** (see `go.mod`)
- Ebitengine handles window creation — no extra dependencies.

## Gameplay

Start a new game, choose name/gender/home, then manage your character's life:

| Feature | How |
|---------|-----|
| Buy items | Click door → Shop → pick category → Buy & Place |
| Place items | Choose grid cell → pick Z height → pick direction |
| Move items | Select room item → ✦ Move → pick new position |
| Store food | Click fridge → 📦 Organize → place in 3D slot grid |
| Eat food | Click floor food → Eat |
| Cook food | Place food on stove surface → cook |
| Save/Load | Escape or ☰ Menu → Save Game / Load Game |
| Speed control | HUD speed buttons (0.5×–100×) |

## Controls

- **Mouse**: click grid cells, items, buttons
- **Escape**: toggle menu overlay / return from shop
- **HUD ☰ Menu**: same as Escape

## Project Structure

```
basic-life-sim/
├── main.go                    # Entry point
├── internal/
│   ├── core/                  # Game engine, UI, logic
│   │   ├── game.go            # Ebitengine Game, Layout, Update, Draw
│   │   ├── screen.go          # Screen interface (BuildView)
│   │   ├── main_screen.go     # HUD, content, menu overlay
│   │   ├── title.go           # Title screen (flexbox)
│   │   ├── newgame.go         # New game wizard (flexbox)
│   │   ├── room_panel.go      # Room grid + item list + floor + doors
│   │   ├── shop_panel.go      # Shop catalog + buy flow
│   │   ├── placement_wizard.go # Item placement (grid→Z→direction)
│   │   ├── storage_wizard.go  # Fridge 3D slot organizer
│   │   ├── ui.go              # Drawing primitives
│   │   ├── logic.go           # Spatial helpers, nutrition
│   │   ├── save.go            # Save/load to disk
│   │   ├── fonts.go           # Font initialization
│   │   └── time.go            # Game clock
│   ├── model/                 # Data models (Character, Home, Items)
│   ├── buyable/               # Item catalog definitions
│   ├── constant/              # Home types, layouts
│   └── experience/            # Experience definitions
└── saves/                     # Save files (*.gob)
```

## UI Architecture

Uses [furex](https://github.com/yohamta/furex) flexbox engine for layout:

```
Game.rootView (furex.Row)
├── HUDView (220px fixed) — character info, needs, speed, menu button
└── ContentView (flex) — room or shop based on mode
```

- Title/NewGame screens use separate root views with `furex.Column` + `JustifyCenter`
- All drawing uses frame-relative coordinates (`frame.Min.X/Y` as origin)
- No global layout constants
