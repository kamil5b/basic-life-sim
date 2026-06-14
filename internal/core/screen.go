package core

import "github.com/yohamta/furex/v2"

// Screen is implemented by every game screen.
type Screen interface {
	BuildView(g *Game) *furex.View
}
