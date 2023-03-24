package client

import (
	"fmt"
	"remakemc/client/gui"
	"remakemc/client/renderers"
	"remakemc/config"
	"remakemc/core"
	"remakemc/core/blocks"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func Start() {
	runtime.LockOSThread()

	renderers.InitAll(config.App.Client.DefaultWidth, config.App.Client.DefaultHeight)
	defer glfw.Terminate()

	gui.Init()

	dim := oneChunkDim(&blocks.Grass)
	fmt.Println(dim.Chunks)

	t := time.Now()
	for !renderers.Win.ShouldClose() {
		// Clear window
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		camPos := mgl32.Vec3{-5, 18 + float32(time.Since(t).Seconds()), -5}
		view := mgl32.LookAtV(
			camPos,               // Camera is at ... in World Space
			mgl32.Vec3{0, 16, 0}, // and looks at
			mgl32.Vec3{0, 1, 0},  // Head is up
		)

		winX, winY := renderers.Win.GetSize()
		renderers.RenderChunk(dim.Chunks[core.Vec3{}], view, float32(winX)/float32(winY))

		gui.RenderGame()

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

	dim.Chunks[core.NewVec3(0, 0, 0)] = chk

	renderers.MakeChunkVAO(dim, chk)

	return dim
}
