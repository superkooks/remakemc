package gui

import (
	"fmt"
	"remakemc/client/renderers"
	"remakemc/core"

	"github.com/go-gl/mathgl/mgl32"
)

type slot struct {
	Start, End mgl32.Vec2
	Index      int
	Hotbar     bool
}

// Returns the index (or -1) of the hovered slot
func RenderInventory(inventoryItems [27]core.ItemStack, hotbarItems [9]core.ItemStack, floatingStack core.ItemStack) (inventoryHover int, hotbarHover int) {
	// Render a black tint over the screen
	renderers.TintScreen(mgl32.Vec4{0, 0, 0, 0.8})

	// Render the inventory interface
	iwidth := float32(0.8)
	iheight := iwidth / 170 * 166

	RenderWithAnchor(inventory, mgl32.Vec2{0, 0}, mgl32.Vec2{iwidth, iheight}, Anchor{0, 0})

	slotAdvance := iwidth / 170 * 18
	slotContained := iwidth / 170 * 16

	// Render the items in the hotbar
	var hotbarSlots []slot
	for i := 0; i < 9; i++ {
		start, end := AnchorAt(
			mgl32.Vec2{-iwidth/2 + iwidth/170*5 + slotAdvance*float32(i), (-iheight/2 + iwidth/170*8) * renderers.GetAspectRatio()},
			mgl32.Vec2{slotContained, slotContained},
			Anchor{Horizontal: -1, Vertical: -1},
		)

		hotbarSlots = append(hotbarSlots, slot{
			Start: start, End: end, Index: i, Hotbar: true,
		})

		if hotbarItems[i].IsEmpty() {
			continue
		}

		item := core.ItemRegistry[hotbarItems[i].Item]
		item.RenderType.RenderItem(item, start, end)

		if hotbarItems[i].Count > 1 {
			RenderText(
				mgl32.Vec2{-iwidth/2 + iwidth/170*5 + slotAdvance*float32(i) + slotContained, (-iheight/2 + iwidth/170*8) * renderers.GetAspectRatio()},
				fmt.Sprint(hotbarItems[i].Count),
				Anchor{Horizontal: 1, Vertical: -1},
			)
		}
	}

	// Render the items in the inventory
	var inventorySlots []slot
	for j := 0; j < 3; j++ {
		for i := 0; i < 9; i++ {
			start, end := AnchorAt(
				mgl32.Vec2{-iwidth/2 + iwidth/170*5 + slotAdvance*float32(i), (-iheight/2 + iwidth/170*66 - slotAdvance*float32(j)) * renderers.GetAspectRatio()},
				mgl32.Vec2{slotContained, slotContained},
				Anchor{Horizontal: -1, Vertical: -1},
			)

			inventorySlots = append(inventorySlots, slot{
				Start: start, End: end, Index: i + j*9,
			})

			if inventoryItems[i+9*j].IsEmpty() {
				continue
			}

			item := core.ItemRegistry[inventoryItems[i+9*j].Item]
			item.RenderType.RenderItem(item, start, end)

			if inventoryItems[i+9*j].Count > 1 {
				RenderText(
					mgl32.Vec2{-iwidth/2 + iwidth/170*5 + slotAdvance*float32(i) + slotContained, (-iheight/2 + iwidth/170*66 - slotAdvance*float32(j)) * renderers.GetAspectRatio()},
					fmt.Sprint(inventoryItems[i+9*j].Count),
					Anchor{Horizontal: 1, Vertical: -1},
				)
			}
		}
	}

	// Calculate which slot is hovered over
	var hovered slot
	var isHovering bool
	var cursorX, cursorY float32
	{
		xpos, ypos := renderers.Win.GetCursorPos()

		// Convert the location of the cursor into OpenGL coordinates
		width, height := renderers.Win.GetSize()
		cursorX = float32(xpos/float64(width)*2 - 1)
		cursorY = float32(-ypos/float64(height)*2 + 1)

		// See if we are over any slots
		for _, v := range inventorySlots {
			if v.Start.X() < cursorX && v.End.X() > cursorX && v.Start.Y() < cursorY && v.End.Y() > cursorY {
				hovered = v
				isHovering = true
			}
		}

		for _, v := range hotbarSlots {
			if v.Start.X() < cursorX && v.End.X() > cursorX && v.Start.Y() < cursorY && v.End.Y() > cursorY {
				hovered = v
				isHovering = true
			}
		}
	}

	// Draw a highlight over the hovered slot
	if isHovering {
		renderers.RenderGUIElement(slotHighlight, hovered.Start, hovered.End)
	}

	// Render the floating stack
	if !floatingStack.IsEmpty() {
		i := core.ItemRegistry[floatingStack.Item]

		start, end := AnchorAt(mgl32.Vec2{cursorX, cursorY}, mgl32.Vec2{slotContained, slotContained}, Anchor{0, 0})
		i.RenderType.RenderItem(i, start, end)

		if floatingStack.Count > 1 {
			RenderText(mgl32.Vec2{cursorX + slotContained/2, cursorY - slotContained/2*renderers.GetAspectRatio()}, fmt.Sprint(floatingStack.Count), Anchor{Horizontal: 1, Vertical: -1})
		}
	}

	// Return the hovered index
	if isHovering {
		if !hovered.Hotbar {
			return hovered.Index, -1
		} else {
			return -1, hovered.Index
		}
	} else {
		return -1, -1
	}
}
