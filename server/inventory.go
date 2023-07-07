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

	// Update other clients about the user's new held item
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
