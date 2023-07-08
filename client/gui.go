package client

import (
	"remakemc/client/renderers"

	"github.com/go-gl/glfw/v3.2/glfw"
)

var inventoryOpen bool

func OpenInventory() {
	inventoryOpen = true
	renderers.Win.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	renderers.Win.SetScrollCallback(nil)
}

func CloseInventory() {
	inventoryOpen = false
	renderers.Win.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	renderers.Win.SetScrollCallback(player.ScrollCallback)
}
