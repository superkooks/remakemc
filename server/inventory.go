package server

import (
	"remakemc/core"
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
					HeldItemType: c.Hotbar[h].Item,
				},
			}
		}
	}
}

func (c *Client) HandleContainerContents(m proto.ContainerContents) {
	// TODO validate items not being duplicated

	// TODO Update only clients who are looking at this container

	// TODO Add support for non-inventories (proto level)

	copy(c.Hotbar[:], m.Slots[:9])
	copy(c.Inventory[:], m.Slots[9:])

	c.SendQueue <- proto.CONTAINER_CONTENTS
	c.SendQueue <- m
}
