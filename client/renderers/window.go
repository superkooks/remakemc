package renderers

import (
	"log"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

var Win *glfw.Window
var FOVDegrees float32 = 70

var cachedAspectRatio float32
var isFocused bool

func InitAll(width, height int) {
	// Init GLFW window
	Win = initGlfw(width, height)

	// Init GL
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	// Init renderers
	initGUI()
	initChunk()
	initSelector()
	initEntity()
}

func GetAspectRatio() float32 {
	// TODO Expire cache
	if cachedAspectRatio == 0 {
		winX, winY := Win.GetSize()
		cachedAspectRatio = float32(winX) / float32(winY)
	}
	return cachedAspectRatio
}

func initGlfw(width, height int) *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLDebugContext, glfw.True)

	window, err := glfw.CreateWindow(width, height, "RemakeMC", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	window.SetFocusCallback(windowFocusCallback)

	return window
}

func IsWindowFocused() bool {
	return isFocused
}

func windowFocusCallback(w *glfw.Window, focused bool) {
	isFocused = focused
}
