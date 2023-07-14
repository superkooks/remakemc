package container

import (
	"remakemc/client/gui"
	"remakemc/client/renderers"
	"remakemc/core"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
)

type Inventory struct {
	EntityID uuid.UUID
	Slots    []core.Slot
	Floating core.ItemStack

	slotSize float32
}

func (c *Inventory) Init(withBoxes bool, entityID uuid.UUID) {
	c.EntityID = entityID

	// Generate slots
	iwidth := float32(0.8)
	iheight := iwidth / 170 * 166

	slotAdvance := iwidth / 170 * 18
	slotContained := iwidth / 170 * 16
	c.slotSize = slotContained

	// Genereate hotbar slots
	for i := 0; i < 9; i++ {
		var start, end mgl32.Vec2
		if withBoxes {
			start, end = gui.AnchorAt(
				mgl32.Vec2{-iwidth/2 + iwidth/170*5 + slotAdvance*float32(i), (-iheight/2 + iwidth/170*8) * renderers.GetAspectRatio()},
				mgl32.Vec2{slotContained, slotContained},
				gui.Anchor{Horizontal: -1, Vertical: -1},
			)
		}

		c.Slots = append(c.Slots, &core.InventorySlot{
			Start: start, End: end,
		})
	}

	// Generate inventory
	for j := 0; j < 3; j++ {
		for i := 0; i < 9; i++ {
			var start, end mgl32.Vec2
			if withBoxes {
				start, end = gui.AnchorAt(
					mgl32.Vec2{-iwidth/2 + iwidth/170*5 + slotAdvance*float32(i), (-iheight/2 + iwidth/170*66 - slotAdvance*float32(j)) * renderers.GetAspectRatio()},
					mgl32.Vec2{slotContained, slotContained},
					gui.Anchor{Horizontal: -1, Vertical: -1},
				)
			}

			c.Slots = append(c.Slots, &core.InventorySlot{
				Start: start, End: end,
			})
		}
	}
}

func (c *Inventory) GetEntityID() uuid.UUID {
	return c.EntityID
}

func (c *Inventory) GetSlots() []core.Slot {
	return c.Slots
}

func (c *Inventory) GetFloating() core.ItemStack {
	return c.Floating
}

func (c *Inventory) SetFloating(s core.ItemStack) {
	c.Floating = s
}

func (c *Inventory) Render() {
	renderers.TintScreen(mgl32.Vec4{0, 0, 0, 0.8})

	iwidth := float32(0.8)
	iheight := iwidth / 170 * 166

	gui.RenderWithAnchor(gui.Inventory, mgl32.Vec2{0, 0}, mgl32.Vec2{iwidth, iheight}, gui.Anchor{Horizontal: 0, Vertical: 0})

	gui.RenderSlots(c.GetSlots())
	gui.RenderFloating(c.GetFloating(), c.slotSize)
}
