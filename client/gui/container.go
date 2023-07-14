package gui

import (
	"fmt"
	"remakemc/client/renderers"
	"remakemc/core"

	"github.com/go-gl/mathgl/mgl32"
)

func RenderSlots(slots []core.Slot) {
	// Render the items in their slots
	for _, v := range slots {
		if !v.GetStack().IsEmpty() {
			// Render item
			start, end := v.GetBox()
			item := core.ItemRegistry[v.GetStack().Item]
			item.RenderType.RenderItem(item, start, end)

			if v.GetStack().Count > 1 {
				// Get the right-bottom-most corner
				var p mgl32.Vec2
				copy(p[:], end[:])
				mgl32.SetMax(&p[0], &start[0])
				mgl32.SetMin(&p[1], &start[1])

				// Render count
				RenderText(
					p,
					fmt.Sprint(v.GetStack().Count),
					Anchor{Horizontal: 1, Vertical: -1},
				)
			}
		}
	}

	// Convert the location of the cursor into OpenGL coordinates
	xpos, ypos := renderers.Win.GetCursorPos()
	width, height := renderers.Win.GetSize()
	cursorX := float32(xpos/float64(width)*2 - 1)
	cursorY := float32(-ypos/float64(height)*2 + 1)

	// See if we are over any slots
	for _, v := range slots {
		start, end := v.GetBox()
		if start.X() < cursorX && end.X() > cursorX && start.Y() < cursorY && end.Y() > cursorY {
			// Render the hover effect
			renderers.RenderGUIElement(SlotHighlight, start, end)
		}
	}
}

func RenderFloating(stack core.ItemStack, slotSize float32) {
	if stack.IsEmpty() {
		return
	}

	// Convert the location of the cursor into OpenGL coordinates
	xpos, ypos := renderers.Win.GetCursorPos()
	width, height := renderers.Win.GetSize()
	cursorX := float32(xpos/float64(width)*2 - 1)
	cursorY := float32(-ypos/float64(height)*2 + 1)

	// Render the item
	i := core.ItemRegistry[stack.Item]
	start, end := AnchorAt(mgl32.Vec2{cursorX, cursorY}, mgl32.Vec2{slotSize, slotSize}, Anchor{Horizontal: 0, Vertical: 0})
	i.RenderType.RenderItem(i, start, end)

	if stack.Count > 1 {
		// Get the right-bottom-most corner
		var p mgl32.Vec2
		copy(p[:], end[:])
		mgl32.SetMax(&p[0], &start[0])
		mgl32.SetMin(&p[1], &start[1])

		// Render the count
		RenderText(p, fmt.Sprint(stack.Count), Anchor{Horizontal: 1, Vertical: -1})
	}
}
