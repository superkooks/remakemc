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
	// The side width of the atlas
	Width int

	// The side width of the largest texture
	largestTexSide int

	textures map[string]*image.RGBA
	uvs      map[string]mgl32.Vec4
}

func NewAtlas() *Atlas {
	return &Atlas{
		textures: make(map[string]*image.RGBA),
		uvs:      make(map[string]mgl32.Vec4),
	}
}

func (a *Atlas) AddTex(i *image.RGBA, name string) {
	if a.largestTexSide < i.Rect.Dx() {
		a.largestTexSide = i.Rect.Dx()
	} else if a.largestTexSide < i.Rect.Dy() {
		a.largestTexSide = i.Rect.Dy()
	}

	a.textures[name] = i
}

// All textures MUST be in PNG format
// Filenames MUST be texname+".png"
// To maximize packing efficiency, ensure all textures are square and
// have the size side length
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
	atlasWidth := int(math.Ceil(math.Sqrt(float64(count)))) * a.largestTexSide
	a.Width = atlasWidth

	i := image.NewRGBA(image.Rect(0, 0, atlasWidth, atlasWidth))

	j := 0
	for k, v := range a.textures {
		x := (j * a.largestTexSide) % atlasWidth
		y := (j * a.largestTexSide) / atlasWidth * a.largestTexSide

		fmt.Println(k, x, y)

		draw.Draw(i, image.Rect(x, y, x+a.largestTexSide, y+a.largestTexSide), v, image.Pt(0, 0), draw.Over)

		a.uvs[k] = mgl32.Vec4{
			float32(x) / float32(i.Rect.Dx()),
			float32(y) / float32(i.Rect.Dy()),
			float32(x+v.Rect.Dx()) / float32(i.Rect.Dx()),
			float32(y+v.Rect.Dy()) / float32(i.Rect.Dy()),
		}

		j++
	}

	return i
}

// MUST be called after the atlas is finalized
func (a *Atlas) GetUV(name string) (start, end mgl32.Vec2) {
	v := a.uvs[name]
	// Correct uvs so they align on the centre of a texel
	v[0] += (1 / float32(a.Width)) / 2
	v[1] += (1 / float32(a.Width)) / 2
	v[2] -= (1 / float32(a.Width)) / 2
	v[3] -= (1 / float32(a.Width)) / 2

	return mgl32.Vec2{v[0], v[1]}, mgl32.Vec2{v[2], v[3]}
}
