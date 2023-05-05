package gui

import (
	"image"
	"image/color"
	"image/draw"
	"io"
	"remakemc/client/assets"
	"remakemc/client/renderers"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

var FontAtlas *renderers.Atlas

var textElems = make(map[rune]renderers.GUIElem)

// Reads the hardcoded otf font from assets and creates the texture atlas
// for that font.
func ReadOTF() {
	FontAtlas = renderers.NewAtlas()

	// Read and parse font face
	f, err := assets.Files.Open("font.otf")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	decoded, err := sfnt.Parse(b)
	if err != nil {
		panic(err)
	}

	face, err := opentype.NewFace(decoded, &opentype.FaceOptions{Size: 32, DPI: 72})
	if err != nil {
		panic(err)
	}

	// Add each glyph to atlas
	for i := 0x21; i < 0x7f; i++ {
		bounds, _, ok := face.GlyphBounds(rune(i))
		if !ok {
			panic("could not find bounds for glyph")
		}

		// Fixed font size
		dstRect := image.Rect(
			0, 0,
			18, 32,
		)
		img := image.NewRGBA(dstRect)

		srcimg := image.NewRGBA(dstRect)
		for x := 0; x < srcimg.Rect.Dx(); x++ {
			for y := 0; y < srcimg.Rect.Dy(); y++ {
				srcimg.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
			}
		}

		// Draw glyph to dst
		drect, mask, maskpoint, _, ok := face.Glyph(fixed.Point26_6{X: -bounds.Min.X, Y: -bounds.Min.Y}, rune(i))
		if !ok {
			panic("could not find bounds for glyph")
		}
		draw.DrawMask(img, drect, srcimg, image.Pt(0, 0), mask, maskpoint, draw.Src)

		// Add image to atlas
		FontAtlas.AddTex(img, string(rune(i)))
	}

	// Generate texture
	var tex uint32
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	// Assign texture data
	i := FontAtlas.Finalize()
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(i.Bounds().Dx()), int32(i.Bounds().Dy()), 0,
		gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(i.Pix))

	// Create buffer for vertices
	verts := renderers.GlBufferFrom([]float32{
		// -1, -1, -1,
		// 1, -1, -1,
		// 1, 1, -1,
		// 1, 1, -1,
		// -1, 1, -1,
		// -1, -1, -1,
		-0.015, -0.015 / 18 * 32, 0.015,
		0.015, -0.015 / 18 * 32, 0.015,
		0.015, 0.015 / 18 * 32, 0.015,
		0.015, 0.015 / 18 * 32, 0.015,
		-0.015, 0.015 / 18 * 32, 0.015,
		-0.015, -0.015 / 18 * 32, 0.015,
	})

	for i := 0x21; i < 0x7f; i++ {
		// Create buffer for uvs
		startUV, endUV := FontAtlas.GetUV(string(rune(i)))
		uvs := renderers.GlBufferFrom([]float32{
			startUV.X(), endUV.Y(),
			endUV.X(), endUV.Y(),
			endUV.X(), startUV.Y(),
			endUV.X(), startUV.Y(),
			startUV.X(), startUV.Y(),
			startUV.X(), endUV.Y(),
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

		textElems[rune(i)] = renderers.GUIElem{
			VAO:       vao,
			Tex:       tex,
			VertCount: 6,
		}
	}
}

func RenderText(pos mgl32.Vec2, text string) {
	for _, v := range text {
		renderers.RenderGUIElement(textElems[v])
	}
}
