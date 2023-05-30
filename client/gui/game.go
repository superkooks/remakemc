package gui

import (
	"github.com/go-gl/mathgl/mgl32"
)

func RenderGame() {
	RenderWithAnchor(crosshair, mgl32.Vec2{0, 0}, mgl32.Vec2{0.03, 0.03}, Anchor{0, 0})

	// Render hotbar
	{
		slotWidth := float32(1) / 9
		selectedSlot := 6
		RenderWithAnchor(hotbar,
			mgl32.Vec2{0, -1 + (0.1319-0.1209)/2},
			mgl32.Vec2{1, 0.1209},
			Anchor{Horizontal: 0, Vertical: -1},
		)
		RenderWithAnchor(
			hotbarSelected,
			mgl32.Vec2{-0.5 + (0.1209 - 0.1319) + slotWidth*float32(selectedSlot), -1},
			// mgl32.Vec2{-0.5 + slotWidth*float32(selectedSlot), -1},
			mgl32.Vec2{0.1319, 0.1319},
			Anchor{Horizontal: -1, Vertical: -1},
		)
	}
}
