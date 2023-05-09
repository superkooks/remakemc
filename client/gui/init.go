package gui

import (
	_ "embed"
	"image"
	"image/png"
	"remakemc/client/assets"
	"remakemc/client/renderers"

	"github.com/go-gl/gl/v4.1-core/gl"
)

func Init() {
	// Load all GUI assets
	initFromAssets("crosshair.png", &crosshair)
}

func initFromAssets(fileName string, target *renderers.GUIElem) {
	// Read texture from embedded assets
	f, err := assets.Files.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	i, err := png.Decode(f)
	if err != nil {
		panic(err)
	}

	// Generate texture
	var tex uint32
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	// Assign texture data
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(i.Bounds().Dx()), int32(i.Bounds().Dy()), 0,
		gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(i.(*image.NRGBA).Pix))

	// Create buffer for vertices
	verts := renderers.GlBufferFrom([]float32{
		0, 0, 1,
		1, 0, 1,
		1, 1, 1,
		1, 1, 1,
		0, 1, 1,
		0, 0, 1,
	})

	// Create buffer for uvs
	uvs := renderers.GlBufferFrom([]float32{
		0, 0,
		1, 0,
		1, 1,
		1, 1,
		0, 1,
		0, 0,
	})

	// Assign buffers to vertex array
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, verts)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil) // vec3

	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, uvs)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 0, nil) // vec2

	target.VAO = vao
	target.Tex = tex
	target.VertCount = 6
}

var crosshair renderers.GUIElem
