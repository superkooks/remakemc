package gui

import (
	"remakemc/client/renderers"

	"github.com/go-gl/mathgl/mgl32"
)

// Specify the direction to anchor using or -1, 0, 1
// for left/bottom, center, or right/top respectively
type Anchor struct {
	Vertical   int
	Horizontal int
}

// Anchors the corner of a gui element at pos, returning the start and end for that element.
// It also takes the box, in uniform coordinates, to scale by the aspect ratio.
func AnchorAt(pos mgl32.Vec2, box mgl32.Vec2, anchor Anchor) (boxMin, boxMax mgl32.Vec2) {
	switch anchor.Horizontal {
	case -1:
		boxMin[0] = pos.X()
	case 0:
		boxMin[0] = pos.X() - box.X()/2
	case 1:
		boxMin[0] = pos.X() - box.X()
	default:
		panic("invalid anchor")
	}

	switch anchor.Vertical {
	case 1:
		boxMin[1] = pos.Y() - box.Y()
	case 0:
		boxMin[1] = pos.Y() - box.Y()/2
	case -1:
		boxMin[1] = pos.Y()
	default:
		panic("invalid anchor")
	}

	boxMax = boxMin.Add(box)

	switch anchor.Vertical {
	case 1:
		boxMin[1] -= box[1] * (renderers.GetAspectRatio() - 1)
	case 0:
		boxMax[1] += box[1] * (renderers.GetAspectRatio() - 1) / 2
		boxMin[1] -= box[1] * (renderers.GetAspectRatio() - 1) / 2
	case -1:
		boxMax[1] += box[1] * (renderers.GetAspectRatio() - 1)
	}

	return
}

func RenderWithAnchor(e renderers.GUIElem, pos mgl32.Vec2, box mgl32.Vec2, anchor Anchor) {
	start, end := AnchorAt(pos, box, anchor)
	renderers.RenderGUIElement(e, start, end)
}
