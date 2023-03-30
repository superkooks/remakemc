package core

import (
	"math"
	"sort"

	"github.com/go-gl/mathgl/mgl32"
)

// Vec3 is an integer vector, whereas mgl32.Vec3 is a float vector
type Vec3 struct {
	X, Y, Z int
}

func NewVec3(x, y, z int) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

func NewVec3FromFloat(v mgl32.Vec3) Vec3 {
	return Vec3{
		X: FloorFloat32(v.X()),
		Y: FloorFloat32(v.Y()),
		Z: FloorFloat32(v.Z()),
	}
}

func (v Vec3) ToFloat() mgl32.Vec3 {
	return mgl32.Vec3{float32(v.X), float32(v.Y), float32(v.Z)}
}

func (v Vec3) Add(u Vec3) Vec3 {
	return Vec3{X: v.X + u.X, Y: v.Y + u.Y, Z: v.Z + u.Z}
}

func (v Vec3) Sub(u Vec3) Vec3 {
	return Vec3{X: v.X - u.X, Y: v.Y - u.Y, Z: v.Z - u.Z}
}

func (v Vec3) Mul(c int) Vec3 {
	return Vec3{X: v.X * c, Y: v.Y * c, Z: v.Z * c}
}

func (v Vec3) Div(c int) Vec3 {
	return Vec3{X: v.X / c, Y: v.Y / c, Z: v.Z / c}
}

func FloorFloat32(f float32) int {
	return int(math.Floor(float64(f)))
}

// Trace a ray from the starting pos in dir, to a maxiumum of reach.
// A callback will be made for each intersection in the voxel grid with
// the coordinates of the block, and a subvoxel hit vector.
//
// Funny story, there doesn't seem to be a very good algorithm for this
// online (at least) that I could get to work. So I had to come up with
// this one myself.
//
// It should be faster than iterating along the line and looking for
// intersections because we don't need to test 2000 points
func TraceRay(dir mgl32.Vec3, pos mgl32.Vec3, reach float32,
	callback func(block mgl32.Vec3, hit mgl32.Vec3) (stop bool)) {
	// Determine length along ray required to step each axis
	//    /|
	// t / | t*v[y]
	//  /__|
	//   1 = t*v[x]
	//
	// t = 1/v[x]
	lenStepX := 1 / dir.X()
	lenStepY := 1 / dir.Y()
	lenStepZ := 1 / dir.Z()

	// Step each axis by one unit from the border until max reach, adding each
	// scalar to the list
	var steps []float64
	if !math.IsNaN(float64(lenStepX)) {
		// Substep to border of voxel, and get step dir
		stepDir := 1
		substep := float32(math.Ceil(float64(pos.X()))) - pos.X()
		if dir.X() < 0 {
			substep = float32(math.Floor(float64(pos.X()))) - pos.X()
			stepDir = -1
		}

		// X
		for i := 0; ; {
			t := (substep + float32(i)) * lenStepX
			if t > reach {
				break
			}

			steps = append(steps, float64(t))

			i += stepDir
		}
	}
	if !math.IsNaN(float64(lenStepY)) {
		// Substep to border of voxel, and get step dir
		stepDir := 1
		substep := float32(math.Ceil(float64(pos.Y()))) - pos.Y()
		if dir.Y() < 0 {
			substep = float32(math.Floor(float64(pos.Y()))) - pos.Y()
			stepDir = -1
		}

		// Y
		for i := 0; ; {
			t := (substep + float32(i)) * lenStepY
			if t > reach {
				break
			}

			steps = append(steps, float64(t))

			i += stepDir
		}
	}
	if !math.IsNaN(float64(lenStepZ)) {
		// Substep to border of voxel, and get step dir
		stepDir := 1
		substep := float32(math.Ceil(float64(pos.Z()))) - pos.Z()
		if dir.Z() < 0 {
			substep = float32(math.Floor(float64(pos.Z()))) - pos.Z()
			stepDir = -1
		}

		// Z
		for i := 0; ; {
			t := (substep + float32(i)) * lenStepZ
			if t > reach {
				break
			}

			steps = append(steps, float64(t))

			i += stepDir
		}
	}

	// Sort steps and interate
	sort.Float64s(steps)
	for _, t := range steps {
		p1 := mgl32.Vec3{
			pos.X() + float32(t)*dir.X(),
			pos.Y() + float32(t)*dir.Y(),
			pos.Z() + float32(t)*dir.Z(),
		}

		p2 := mgl32.Vec3{
			RoundThenFloor(p1.X()),
			RoundThenFloor(p1.Y()),
			RoundThenFloor(p1.Z()),
		}

		var p3 mgl32.Vec3
		copy(p3[:], p2[:])

		// If the direction is negative, we need to subtract one
		// when at an intersection on that axis
		if dir.X() < 0 && mgl32.FloatEqual(pos.X()+float32(t)*dir.X(), p3[0]) {
			p3[0]--
		}
		if dir.Y() < 0 && mgl32.FloatEqual(pos.Y()+float32(t)*dir.Y(), p3[1]) {
			p3[1]--
		}
		if dir.Z() < 0 && mgl32.FloatEqual(pos.Z()+float32(t)*dir.Z(), p3[2]) {
			p3[2]--
		}

		if callback(p3, p1.Sub(p3)) {
			return
		}
	}
}

func FaceFromSubvoxel(sv mgl32.Vec3) BlockFace {
	if mgl32.FloatEqualThreshold(sv.Y(), 1, 0.001) {
		return FaceTop
	}
	if mgl32.FloatEqualThreshold(sv.Y(), 0, 0.001) {
		return FaceBottom
	}
	if mgl32.FloatEqualThreshold(sv.X(), 0, 0.001) {
		return FaceLeft
	}
	if mgl32.FloatEqualThreshold(sv.X(), 1, 0.001) {
		return FaceRight
	}
	if mgl32.FloatEqualThreshold(sv.Z(), 1, 0.001) {
		return FaceFront
	}
	if mgl32.FloatEqualThreshold(sv.Z(), 0, 0.001) {
		return FaceBack
	}

	panic("FaceFromSubvoxel: vector not on face")
}
