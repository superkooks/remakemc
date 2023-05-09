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

// Font size during gui rendering
const FONT_SIZE = 0.03

var FontAtlas *renderers.Atlas
var textElems = make(map[rune]renderers.GUIElem)
var fontRatio float32 // the ratio between width and height

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

	face, err := opentype.NewFace(decoded, &opentype.FaceOptions{Size: 64, DPI: 72})
	if err != nil {
		panic(err)
	}

	// Get the max and min heights
	var maxDescender fixed.Int26_6
	var maxAscender fixed.Int26_6
	var maxAdvance fixed.Int26_6
	for i := 0x21; i < 0x7f; i++ {
		bounds, advance, ok := face.GlyphBounds(rune(i))
		if !ok {
			panic("could not find bounds for glyph")
		}

		if bounds.Min.Y < maxDescender {
			maxDescender = bounds.Min.Y
		}

		if bounds.Max.Y > maxAscender {
			maxAscender = bounds.Max.Y
		}

		if advance > maxAdvance {
			maxAdvance = advance
		}
	}

	// Fixed font size
	fontRatio = float32(maxAdvance>>6) / 64
	dstRect := image.Rect(
		0, 0,
		int(maxAdvance>>6), int(maxAscender-maxDescender)>>6,
	)

	// Add each glyph to atlas
	for i := 0x21; i < 0x7f; i++ {
		// Create images
		dstimg := image.NewRGBA(dstRect)
		srcimg := image.NewRGBA(dstRect)
		for x := 0; x < srcimg.Rect.Dx(); x++ {
			for y := 0; y < srcimg.Rect.Dy(); y++ {
				srcimg.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
			}
		}

		// Draw glyph to dstimg
		drect, mask, maskpoint, _, ok := face.Glyph(
			fixed.Point26_6{
				X: 0,
				Y: -maxDescender,
			},
			rune(i),
		)
		if !ok {
			panic("could not find bounds for glyph")
		}
		draw.DrawMask(dstimg, drect, srcimg, image.Pt(0, 0), mask, maskpoint, draw.Src)

		// Post processing: Convert brightness into alpha
		for x := 0; x < dstimg.Rect.Dx(); x++ {
			for y := 0; y < dstimg.Rect.Dy(); y++ {
				dstimg.SetRGBA(x, y, color.RGBA{255, 255, 255, dstimg.RGBAAt(x, y).R})
			}
		}

		// Add image to atlas
		FontAtlas.AddTex(dstimg, string(rune(i)))
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
		0, 0, 1,
		1, 0, 1,
		1, 1, 1,
		1, 1, 1,
		0, 1, 1,
		0, 0, 1,
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

		// Create new gui elem
		textElems[rune(i)] = renderers.GUIElem{
			VAO:       vao,
			Tex:       tex,
			VertCount: 6,
		}
	}
}

func RenderText(pos mgl32.Vec2, text string, anchor Anchor) {
	start, end := AnchorAt(
		pos,
		mgl32.Vec2{
			float32(len(text)) * FONT_SIZE,
			FONT_SIZE / fontRatio,
		},
		anchor,
	)

	letterbox := mgl32.Vec2{
		end.Sub(start)[0] / float32(len(text)),
		end.Sub(start)[1],
	}

	for k, v := range text {
		renderers.RenderGUIElement(
			textElems[v],
			start.Add(mgl32.Vec2{end.Sub(start)[0] / float32(len(text)) * float32(k), 0}),
			start.Add(mgl32.Vec2{end.Sub(start)[0] / float32(len(text)) * float32(k), 0}).Add(letterbox),
		)
	}
}
