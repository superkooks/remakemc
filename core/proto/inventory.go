package proto

import (
	"remakemc/core"

	"github.com/google/uuid"
)

// Sets the player's held item as the index of the hotbar.
// Sent by clients
type PlayerHeldItem int

// Update the externally visible inventory slots of other entities, such as held items and armor.
// Sent by the server
type EntityEquipment struct {
	core.EntityEquipment
	EntityID uuid.UUID
}

// Updates the contents of the currently open screen.
// Sent by clients and the server
type ContainerContents struct {
	Slots         []core.ItemStack
	FloatingStack core.ItemStack
}
