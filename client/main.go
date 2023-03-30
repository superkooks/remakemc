package client

import (
	"fmt"
	"math/rand"
	"remakemc/client/gui"
	"remakemc/client/renderers"
	"remakemc/config"
	"remakemc/core"
	"remakemc/core/blocks"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func Start() {
	runtime.LockOSThread()

	// Initialize OpenGL and GLFW
	renderers.InitAll(config.App.Client.DefaultWidth, config.App.Client.DefaultHeight)
	defer glfw.Terminate()

	// Initialize gui elements
	gui.Init()

	// Initialize terrain
	dim := oneChunkDim(&blocks.Grass)
	fmt.Println(dim.Chunks)

	// Initialize player
	player := &Player{Speed: 10, Position: mgl32.Vec3{2, 2, -2}}
	renderers.Win.SetInputMode(glfw.CursorMode, glfw.CursorHidden)

	// Game loop
	previousTime := glfw.GetTime()
	for !renderers.Win.ShouldClose() {
		// Get delta time
		time := glfw.GetTime()
		deltaTime := time - previousTime
		previousTime = time

		// Clear window
		gl.ClearColor(79.0/255, 167.0/255, 234.0/255, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Process user input and recalculate view matrix
		player.ProcessMouse(deltaTime)
		player.ProcessKeyboard(deltaTime)
		view := mgl32.LookAtV(
			player.Position,                       // Camera is at ... in World Space
			player.Position.Add(player.LookVec()), // and looks at
			mgl32.Vec3{0, 1, 0},                   // Head is up
		)

		// Render terrain
		renderers.RenderChunk(dim.Chunks[core.Vec3{}], view)

		// Find selector position and render
		core.TraceRay(player.LookVec(), player.Position, 16, func(v mgl32.Vec3) (stop bool) {
			block := dim.GetBlockAt(core.NewVec3FromFloat(v))
			if block.Type != nil {
				renderers.RenderSelector(block.Position.ToFloat(), view)
				return true
			} else {
				return false
			}
		})

		// Render gui
		gui.RenderGame()

		// Update window
		glfw.PollEvents()
		renderers.Win.SwapBuffers()
	}
}

func oneChunkDim(typ *core.BlockType) *core.Dimension {
	dim := &core.Dimension{
		Chunks: make(map[core.Vec3]*core.Chunk),
	}

	chk := new(core.Chunk)
	for x := 0; x < 16; x++ {
		var b [][]core.Block
		for y := 0; y < 16; y++ {
			var a []core.Block
			for z := 0; z < 16; z++ {
				a = append(a, core.Block{Position: core.NewVec3(x, y, z), Type: typ})
			}

			b = append(b, a)
		}
		chk.Blocks = append(chk.Blocks, b)
	}

	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			for y := 0; y < int(rand.Float32()*6); y++ {
				chk.Blocks[x][15-y][z].Type = nil
			}
		}
	}

	dim.Chunks[core.NewVec3(0, 0, 0)] = chk

	renderers.MakeChunkVAO(dim, chk)

	return dim
}
