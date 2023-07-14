package core

import "github.com/google/uuid"

type Container interface {
	// Init the slots, and anything else. With boxes generates the boxes for an aspect ratio.
	Init(withBoxes bool, entityID uuid.UUID)
	GetEntityID() uuid.UUID

	// Return the slots of the container
	GetSlots() []Slot

	// Get and set the floating itemstack
	GetFloating() ItemStack
	SetFloating(ItemStack)

	// Render the entire interface. You may use RenderSlots and RenderFloating as helpers.
	Render()
}

func GetStacksFromSlots(slots []Slot) (out []ItemStack) {
	for _, v := range slots {
		out = append(out, v.GetStack())
	}
	return
}

func SetSlotsFromStacks(stacks []ItemStack, slots []Slot) {
	if len(stacks) != len(slots) {
		panic("slots and stacks must have equal lengths")
	}

	for k := range slots {
		slots[k].SetStack(stacks[k])
	}
}
