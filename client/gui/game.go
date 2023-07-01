package gui

import (
	"fmt"
	"remakemc/client/renderers"
	"remakemc/core"

	"github.com/go-gl/mathgl/mgl32"
)

func RenderGame(selectedHotbarSlot int, hotbarItems []core.ItemStack) {
	RenderWithAnchor(crosshair, mgl32.Vec2{0, 0}, mgl32.Vec2{0.03, 0.03}, Anchor{0, 0})

	// Render hotbar
	{
		hwidth := float32(0.75)

		slotWidth := hwidth / 182 * 22
		slotAdvance := hwidth / 182 * 20
		slotContained := hwidth / 182 * 19
		selectorWidth := hwidth / 182 * 24

		RenderWithAnchor(hotbar,
			mgl32.Vec2{0, -1 + (selectorWidth-slotWidth)/2*renderers.GetAspectRatio()},
			mgl32.Vec2{hwidth, slotWidth},
			Anchor{Horizontal: 0, Vertical: -1},
		)

		// Render selector
		RenderWithAnchor(
			hotbarSelected,
			mgl32.Vec2{-hwidth/2 - (selectorWidth-slotWidth)/2 + slotAdvance*float32(selectedHotbarSlot), -1},
			mgl32.Vec2{selectorWidth, selectorWidth},
			Anchor{Horizontal: -1, Vertical: -1},
		)

		// Render the items in the hotbar
		for i := 0; i < 9; i++ {
			if hotbarItems[i].IsEmpty() {
				continue
			}

			start, end := AnchorAt(
				mgl32.Vec2{-hwidth/2 + slotAdvance*float32(i) + (slotWidth-slotContained)/2, -1 + ((selectorWidth-slotWidth)/2+(slotWidth-slotContained)/2)*renderers.GetAspectRatio()},
				mgl32.Vec2{slotContained, slotContained},
				Anchor{Horizontal: -1, Vertical: -1},
			)

			item := core.ItemRegistry[hotbarItems[i].Item]
			item.RenderType.RenderItem(item, start, end)

			RenderText(
				mgl32.Vec2{-hwidth/2 + slotAdvance*float32(i+1), -1 + ((selectorWidth-slotWidth)/2+slotWidth/22*2)*renderers.GetAspectRatio()},
				fmt.Sprint(hotbarItems[i].Count),
				Anchor{Horizontal: 1, Vertical: -1},
			)
		}
	}
}
