package renderers

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"remakemc/client/assets"

	"github.com/go-gl/mathgl/mgl32"
)

var BlockAtlas = NewAtlas()

type Atlas struct {
	// The side width of the largest texture
	LargestTexSide int

	textures map[string]*image.RGBA
	uvs      map[string]mgl32.Vec4
}

func NewAtlas() *Atlas {
	return &Atlas{
		textures: make(map[string]*image.RGBA),
		uvs:      make(map[string]mgl32.Vec4),
	}
}

// All textures MUST have an aspect ratio of 1
func (a *Atlas) AddTex(i *image.RGBA, name string) {
	if i.Rect.Dx() != i.Rect.Dy() {
		panic("textures in atlas must have an aspect ratio of 1")
	}

	if a.LargestTexSide < i.Rect.Dx() {
		a.LargestTexSide = i.Rect.Dx()
	}

	a.textures[name] = i
}

// All textures MUST have an aspect ratio of 1 and be in PNG format
// All textures widths SHOULD be a power of 2
// Filenames MUST be texname+".png"
func (a *Atlas) AddTexFromAssets(texname string) {
	f, err := assets.Files.Open(texname + ".png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	i, err := png.Decode(f)
	if err != nil {
		panic(err)
	}

	a.AddTex(i.(*image.RGBA), texname)
}

// Finalize the atlas, rendering all textures onto one image
func (a *Atlas) Finalize() *image.RGBA {
	count := len(a.textures)
	atlasWidth := int(math.Ceil(math.Sqrt(float64(count)))) * a.LargestTexSide

	i := image.NewRGBA(image.Rect(0, 0, atlasWidth, atlasWidth))

	j := 0
	for k, v := range a.textures {
		x := (j * a.LargestTexSide) % atlasWidth
		y := (j * a.LargestTexSide) / atlasWidth * a.LargestTexSide

		fmt.Println(k, x, y)

		draw.Draw(i, image.Rect(x, y, x+a.LargestTexSide, y+a.LargestTexSide), v, image.Pt(0, 0), draw.Over)

		a.uvs[k] = mgl32.Vec4{
			float32(x) / float32(i.Rect.Dx()),
			float32(y) / float32(i.Rect.Dy()),
			float32(x+a.LargestTexSide) / float32(i.Rect.Dx()),
			float32(y+a.LargestTexSide) / float32(i.Rect.Dy()),
		}

		j++
	}

	return i
}

// MUST be called after the atlas is finalized
func (a *Atlas) GetUV(name string) (start, end mgl32.Vec2) {
	v := a.uvs[name]
	return mgl32.Vec2{v[0], v[1]}, mgl32.Vec2{v[2], v[3]}
}
