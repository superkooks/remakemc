package server

import (
	"remakemc/core"
	"remakemc/core/container"
	"remakemc/core/proto"
)

func (c *Client) HandlePlayerHeldItem(h proto.PlayerHeldItem) {
	if h < 0 || h > 8 {
		// TODO Invalid
		return
	}

	c.HotbarSlotSelected = int(h)

	// Update other clients about the client's new held item
	for _, v := range clients {
		if v != c {
			v.SendQueue <- proto.ENTITY_EQUIPMENT
			v.SendQueue <- proto.EntityEquipment{
				EntityID: c.Position.EntityID,
				EntityEquipment: core.EntityEquipment{
					HeldItemType: c.Inventory.Slots[h].GetStack().Item,
				},
			}
		}
	}
}

func (c *Client) HandleContainerClick(m proto.ContainerClick) {
	// TODO Update only clients who are looking at this container

	// TODO Add support for non-inventories (proto level)

	i := c.Inventory
	hovered := i.Slots[m.SlotIndex]

	if m.LeftClick {
		if i.GetFloating().IsEmpty() && !hovered.GetStack().IsEmpty() {
			// Take the stack from the slot
			s, ok := hovered.TakeStack(false)
			if ok {
				i.SetFloating(s)
			}
		} else if !i.GetFloating().IsEmpty() {
			// Place the stack in the slot
			i.SetFloating(hovered.PutStack(i.GetFloating()))
		}

	} else if m.RightClick {
		if i.GetFloating().IsEmpty() && !hovered.GetStack().IsEmpty() {
			// Take half the stack from the slot
			s, ok := hovered.TakeStack(true)
			if ok {
				i.SetFloating(s)
			}

		} else if !i.GetFloating().IsEmpty() {
			if hovered.GetStack().Item == i.Floating.Item || hovered.GetStack().IsEmpty() {
				// Place one item in the slot
				m := hovered.PutStack(core.ItemStack{Item: i.GetFloating().Item, Count: 1})
				if m.IsEmpty() {
					f := i.GetFloating()
					f.Count--
					i.SetFloating(f)
				}
			} else {
				// Exchange items
				i.SetFloating(hovered.PutStack(i.GetFloating()))
			}
		}
	}

	c.SendQueue <- proto.CONTAINER_CONTENTS
	c.SendQueue <- proto.ContainerContents{
		Slots:         container.GetStacksFromSlots(i.Slots),
		FloatingStack: i.Floating,
	}
}
