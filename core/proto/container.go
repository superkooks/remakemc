package proto

import (
	"remakemc/core"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/google/uuid"
)

// Sets the player's held item as the index of the hotbar.
// Sent by clients
type PlayerHeldItem int

// Update the externally visible inventory slots of other entities, such as held items and armor.
// Sent by the server
// type EntityEquipment struct {
// 	core.EntityEquipment
// 	EntityID uuid.UUID
// }

// Updates the contents of the currently open screen.
// May also be sent with by the CONTAINER_OPEN event.
// Sent by the server
type ContainerContents struct {
	EntityID      uuid.UUID
	Slots         []core.ItemStack
	FloatingStack core.ItemStack
}

// Sent when the player clicks on a slot in a container
// Sent by clients
type ContainerClick struct {
	EntityID  uuid.UUID
	SlotIndex int

	// Which keys were pressed
	LeftClick  bool
	RightClick bool
	ShiftKey   glfw.Action
	NumberKey  int
}
