package client

import (
	"remakemc/client/renderers"
	"remakemc/core"

	"github.com/go-gl/glfw/v3.2/glfw"
)

var containerOpen bool
var openContainer core.Container

func OpenContainer(c core.Container) {
	containerOpen = true
	renderers.Win.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	// renderers.Win.SetScrollCallback(nil)
	openContainer = c
}

func CloseContainer() {
	containerOpen = false
	renderers.Win.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	// renderers.Win.SetScrollCallback(player.ScrollCallback)
	openContainer = nil
}
