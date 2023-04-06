package renderers

import (
	"remakemc/core"

	"github.com/go-gl/mathgl/mgl32"
)

type BasicOneTex struct {
	Tex string
}

func (t BasicOneTex) Init() {
	BlockAtlas.AddTexFromAssets(t.Tex)
}

func (t BasicOneTex) RenderFace(face core.BlockFace, pos mgl32.Vec3) (verts, normals, uvs []float32) {
	atlasStart, atlasEnd := BlockAtlas.GetUV(t.Tex)

	verts = makeFace(faceVertices[face], pos)
	normals = makeNormals(verts)
	uvs = makeUVs(faceUVs[face], atlasStart, atlasEnd)

	return
}

// Add a position to a face
func makeFace(face []float32, pos mgl32.Vec3) []float32 {
	newV := make([]float32, 3*6)
	copy(newV, face)
	for i := 0; i < 6; i++ {
		newV[i*3] += pos.X()
		newV[i*3+1] += pos.Y()
		newV[i*3+2] += pos.Z()
	}

	return newV
}

// Generate the normals for a face
func makeNormals(newV []float32) []float32 {
	normals := make([]float32, 18)
	for i := 0; i < 2; i++ {
		vecV := mgl32.Vec3{newV[3+i*9] - newV[0+i*9], newV[4+i*9] - newV[1+i*9], newV[5+i*9] - newV[2+i*9]}
		vecW := mgl32.Vec3{newV[6+i*9] - newV[0+i*9], newV[7+i*9] - newV[1+i*9], newV[8+i*9] - newV[2+i*9]}
		n := [3]float32(vecV.Cross(vecW))

		copy(normals[0+i*9:], n[:])
		copy(normals[3+i*9:], n[:])
		copy(normals[6+i*9:], n[:])
	}

	return normals
}

// Map the atlas UV to a face UV
func makeUVs(face []float32, start, end mgl32.Vec2) []float32 {
	uvs := make([]float32, len(face))
	for i := 0; i < len(face); i += 2 {
		if face[i] == 0 {
			uvs[i] = start.X()
		} else {
			uvs[i] = end.X()
		}

		if face[i+1] == 0 {
			uvs[i+1] = start.Y()
		} else {
			uvs[i+1] = end.Y()
		}
	}

	return uvs
}

// Vertices for faces
var faceVertices = map[core.BlockFace][]float32{
	core.FaceTop: {
		0, 1, 0,
		0, 1, 1,
		1, 1, 0,

		1, 1, 0,
		0, 1, 1,
		1, 1, 1,
	},

	core.FaceBottom: {
		0, 0, 0,
		1, 0, 0,
		0, 0, 1,

		1, 0, 0,
		1, 0, 1,
		0, 0, 1,
	},

	core.FaceLeft: {
		0, 0, 1,
		0, 1, 0,
		0, 0, 0,

		0, 0, 1,
		0, 1, 1,
		0, 1, 0,
	},

	core.FaceRight: {
		1, 0, 1,
		1, 0, 0,
		1, 1, 0,

		1, 0, 1,
		1, 1, 0,
		1, 1, 1,
	},

	core.FaceFront: {
		0, 0, 1,
		1, 0, 1,
		0, 1, 1,

		1, 0, 1,
		1, 1, 1,
		0, 1, 1,
	},

	core.FaceBack: {
		0, 0, 0,
		0, 1, 0,
		1, 0, 0,

		1, 0, 0,
		0, 1, 0,
		1, 1, 0,
	},
}

// UV maps for faces
var faceUVs = map[core.BlockFace][]float32{
	core.FaceTop: {
		0, 0,
		0, 1,
		1, 0,

		1, 0,
		0, 1,
		1, 1,
	},

	core.FaceBottom: {
		0, 0,
		1, 0,
		0, 1,

		1, 0,
		1, 1,
		0, 1,
	},

	core.FaceLeft: {
		0, 1,
		1, 0,
		0, 0,

		0, 1,
		1, 1,
		1, 0,
	},

	core.FaceRight: {
		1, 1,
		1, 0,
		0, 0,

		1, 1,
		0, 0,
		0, 1,
	},

	core.FaceFront: {
		1, 0,
		0, 0,
		1, 1,

		0, 0,
		0, 1,
		1, 1,
	},

	core.FaceBack: {
		0, 0,
		0, 1,
		1, 0,

		1, 0,
		0, 1,
		1, 1,
	},
}

// New vertices for each face, represented by a an index into the original
// face, that represent the alternative pair of triangles to render that face.
var flipMap = map[core.BlockFace][]int{
	core.FaceTop: {
		0, 1, 5, 0, 5, 2,
	},
	core.FaceBottom: {
		0, 1, 4, 0, 4, 2,
	},
	core.FaceLeft: {
		2, 4, 1, 2, 0, 4,
	},
	core.FaceRight: {
		1, 5, 0, 2, 5, 1,
	},
	core.FaceFront: {
		0, 1, 4, 0, 4, 2,
	},
	core.FaceBack: {
		0, 1, 5, 0, 5, 2,
	},
}
