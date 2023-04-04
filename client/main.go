package client

import (
	"fmt"
	"remakemc/client/gui"
	"remakemc/client/renderers"
	"remakemc/config"
	"remakemc/core"
	"remakemc/core/blocks"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func Start() {
	runtime.LockOSThread()

	// Don't let allocated memory exceed 125% of in-use memory
	debug.SetGCPercent(25)

	// Initialize texture atlas
	blocks.Grass.RenderType.Init()
	blocks.Dirt.RenderType.Init()
	blocks.Stone.RenderType.Init()
	blocks.Cobblestone.RenderType.Init()

	// Initialize OpenGL and GLFW
	renderers.InitAll(config.App.Client.DefaultWidth, config.App.Client.DefaultHeight)
	defer glfw.Terminate()

	// Initialize gui elements
	gui.Init()

	// Initialize terrain
	t := time.Now()
	dim := &core.Dimension{
		Chunks: make(map[core.Vec3]*core.Chunk),
	}
	for x := -16; x < 128+16; x += 16 {
		for z := -16; z < 128+16; z += 16 {
			GenTerrainColumn(core.NewVec3(x, 0, z), dim)
		}
	}
	fmt.Println("Generated initial terrain in", time.Since(t))
	t = time.Now()
	for _, v := range dim.Chunks {
		renderers.MakeChunkVAO(dim, v)
	}
	fmt.Println("Generated initial terrain VAOs in", time.Since(t))

	// Initialize player
	player := NewPlayer(mgl32.Vec3{5.1, 70, 5.1})
	renderers.Win.SetInputMode(glfw.CursorMode, glfw.CursorHidden)

	// Initialize inputs
	mouseOne := new(core.Debounced)
	mouseTwo := new(core.Debounced)

	// Game loop
	previousTime := glfw.GetTime()
	var frames int
	var cumulativeTime float64
	for !renderers.Win.ShouldClose() {
		// Get delta time
		time := glfw.GetTime()
		deltaTime := time - previousTime
		previousTime = time

		// Calculate frame rate
		frames++
		cumulativeTime += deltaTime
		if frames == 60 {
			fmt.Println(1/(cumulativeTime/60), "fps")
			frames = 0
			cumulativeTime = 0
		}

		// Clear window
		gl.ClearColor(79.0/255, 167.0/255, 234.0/255, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Process user input and recalculate view matrix
		player.ProcessMousePosition(deltaTime)
		player.PhysicsUpdate(deltaTime, dim)
		view := mgl32.LookAtV(
			player.CameraPos(),                       // Camera is at ... in World Space
			player.CameraPos().Add(player.LookDir()), // and looks at
			mgl32.Vec3{0, 1, 0},                      // Head is up
		)

		// Mining
		if renderers.Win.GetMouseButton(glfw.MouseButton1) == glfw.Press && mouseOne.Invoke() {
			core.TraceRay(player.LookDir(), player.CameraPos(), 16, func(v, _ mgl32.Vec3) (stop bool) {
				block := dim.GetBlockAt(core.NewVec3FromFloat(v))
				if block.Type != nil {
					block.Type = nil
					dim.SetBlockAt(block, core.NewVec3FromFloat(v))
					renderers.UpdateRequiredMeshes(dim, core.NewVec3FromFloat(v))
					return true
				} else {
					return false
				}
			})
		} else if renderers.Win.GetMouseButton(glfw.MouseButton1) == glfw.Release {
			mouseOne.Reset()
		}

		// Placing
		if renderers.Win.GetMouseButton(glfw.MouseButton2) == glfw.Press && mouseTwo.Invoke() {
			core.TraceRay(player.LookDir(), player.CameraPos(), 16, func(v, h mgl32.Vec3) (stop bool) {
				block := dim.GetBlockAt(core.NewVec3FromFloat(v))
				if block.Type != nil {
					placePos := core.NewVec3FromFloat(v)

					switch core.FaceFromSubvoxel(h) {
					case core.FaceTop:
						placePos = placePos.Add(core.Vec3{Y: 1})
					case core.FaceBottom:
						placePos = placePos.Add(core.Vec3{Y: -1})
					case core.FaceLeft:
						placePos = placePos.Add(core.Vec3{X: -1})
					case core.FaceRight:
						placePos = placePos.Add(core.Vec3{X: 1})
					case core.FaceFront:
						placePos = placePos.Add(core.Vec3{Z: 1})
					case core.FaceBack:
						placePos = placePos.Add(core.Vec3{Z: -1})
					}

					dim.SetBlockAt(core.Block{
						Type: blocks.Cobblestone,
					}, placePos)
					renderers.UpdateRequiredMeshes(dim, placePos)
					return true
				} else {
					return false
				}
			})
		} else if renderers.Win.GetMouseButton(glfw.MouseButton2) == glfw.Release {
			mouseTwo.Reset()
		}

		// Render terrain
		renderers.RenderChunks(dim, view)

		// Find selector position and render
		core.TraceRay(player.LookDir(), player.CameraPos(), 16, func(v, _ mgl32.Vec3) (stop bool) {
			block := dim.GetBlockAt(core.NewVec3FromFloat(v))
			if block.Type != nil {
				renderers.RenderSelector(core.NewVec3FromFloat(v).ToFloat(), view)
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
