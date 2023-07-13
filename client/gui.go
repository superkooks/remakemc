package client

import (
	"remakemc/client/renderers"
	"remakemc/core/container"

	"github.com/go-gl/glfw/v3.2/glfw"
)

var containerOpen bool
var openContainer container.Container

func OpenContainer(c container.Container) {
	containerOpen = true
	renderers.Win.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	renderers.Win.SetScrollCallback(nil)
	openContainer = c
}

func CloseContainer() {
	containerOpen = false
	renderers.Win.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	renderers.Win.SetScrollCallback(player.ScrollCallback)
	openContainer = nil
}
