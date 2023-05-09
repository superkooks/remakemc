package gui

import (
	"github.com/go-gl/mathgl/mgl32"
)

func RenderGame() {
	RenderWithAnchor(crosshair, mgl32.Vec2{0, 0}, mgl32.Vec2{0.03, 0.03}, Anchor{0, 0})
}
