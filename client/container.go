package client

import (
	"remakemc/client/renderers"
	"remakemc/core"
	"remakemc/core/proto"

	"github.com/go-gl/glfw/v3.2/glfw"
)

func ProcessContainerInteraction(c core.Container) {
	// Convert the location of the cursor into OpenGL coordinates
	xpos, ypos := renderers.Win.GetCursorPos()
	width, height := renderers.Win.GetSize()
	cursorX := float32(xpos/float64(width)*2 - 1)
	cursorY := float32(-ypos/float64(height)*2 + 1)

	// Determine hovered slot
	var hovered core.Slot
	var slotIndex int
	for k, v := range c.GetSlots() {
		start, end := v.GetBox()
		if start.X() < cursorX && end.X() > cursorX && start.Y() < cursorY && end.Y() > cursorY {
			hovered = v
			slotIndex = k
		}
	}

	if hovered == nil {
		// TODO drop items (if outside inventory)
		return
	}

	// Left click
	if renderers.Win.GetMouseButton(glfw.MouseButton1) == glfw.Press && mouseOne.Invoke() {
		serverWrite <- proto.CONTAINER_CLICK
		serverWrite <- proto.ContainerClick{
			EntityID:  c.GetEntityID(),
			SlotIndex: slotIndex,
			LeftClick: true,
		}
	}

	// Right click
	if renderers.Win.GetMouseButton(glfw.MouseButton2) == glfw.Press && mouseTwo.Invoke() {
		serverWrite <- proto.CONTAINER_CLICK
		serverWrite <- proto.ContainerClick{
			EntityID:   c.GetEntityID(),
			SlotIndex:  slotIndex,
			RightClick: true,
		}
	}

	// TODO painting mode
}
