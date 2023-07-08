package gui

import (
	"fmt"
	"remakemc/client/renderers"
	"remakemc/core"

	"github.com/go-gl/mathgl/mgl32"
)

func RenderInventory(inventoryItems [27]core.ItemStack, hotbarItems [9]core.ItemStack) {
	// Render a black tint over the screen
	renderers.TintScreen(mgl32.Vec4{0, 0, 0, 0.8})

	// Render the inventory interface
	iwidth := float32(0.8)
	iheight := iwidth / 170 * 166

	RenderWithAnchor(inventory, mgl32.Vec2{0, 0}, mgl32.Vec2{iwidth, iheight}, Anchor{0, 0})

	slotAdvance := iwidth / 170 * 18
	slotContained := iwidth / 170 * 16

	// Render the items in the hotbar
	for i := 0; i < 9; i++ {
		if hotbarItems[i].IsEmpty() {
			continue
		}

		start, end := AnchorAt(
			mgl32.Vec2{-iwidth/2 + iwidth/170*5 + slotAdvance*float32(i), (-iheight/2 + iwidth/170*8) * renderers.GetAspectRatio()},
			mgl32.Vec2{slotContained, slotContained},
			Anchor{Horizontal: -1, Vertical: -1},
		)

		item := core.ItemRegistry[hotbarItems[i].Item]
		item.RenderType.RenderItem(item, start, end)

		RenderText(
			mgl32.Vec2{-iwidth/2 + iwidth/170*5 + slotAdvance*float32(i) + slotContained, (-iheight/2 + iwidth/170*8) * renderers.GetAspectRatio()},
			fmt.Sprint(hotbarItems[i].Count),
			Anchor{Horizontal: 1, Vertical: -1},
		)
	}

	// Render the items in the inventory
	for j := 0; j < 3; j++ {
		for i := 0; i < 9; i++ {
			if inventoryItems[i+9*j].IsEmpty() {
				continue
			}

			start, end := AnchorAt(
				mgl32.Vec2{-iwidth/2 + iwidth/170*5 + slotAdvance*float32(i), (-iheight/2 + iwidth/170*66 - slotAdvance*float32(j)) * renderers.GetAspectRatio()},
				mgl32.Vec2{slotContained, slotContained},
				Anchor{Horizontal: -1, Vertical: -1},
			)

			item := core.ItemRegistry[inventoryItems[i+9*j].Item]
			item.RenderType.RenderItem(item, start, end)

			RenderText(
				mgl32.Vec2{-iwidth/2 + iwidth/170*5 + slotAdvance*float32(i) + slotContained, (-iheight/2 + iwidth/170*66 - slotAdvance*float32(j)) * renderers.GetAspectRatio()},
				fmt.Sprint(inventoryItems[i+9*j].Count),
				Anchor{Horizontal: 1, Vertical: -1},
			)
		}
	}
}
