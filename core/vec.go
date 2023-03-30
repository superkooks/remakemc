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

func TraceRay(dir mgl32.Vec3, pos mgl32.Vec3, reach float32, callback func(v mgl32.Vec3) (stop bool)) {
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
		p := mgl32.Vec3{
			RoundThenFloor(pos.X() + float32(t)*dir.X()),
			RoundThenFloor(pos.Y() + float32(t)*dir.Y()),
			RoundThenFloor(pos.Z() + float32(t)*dir.Z()),
		}

		// If the direction is negative, we need to subtract one
		// when at an intersection on that axis
		if dir.X() < 0 && mgl32.FloatEqual(pos.X()+float32(t)*dir.X(), p[0]) {
			p[0]--
		}
		if dir.Y() < 0 && mgl32.FloatEqual(pos.Y()+float32(t)*dir.Y(), p[1]) {
			p[1]--
		}
		if dir.Z() < 0 && mgl32.FloatEqual(pos.Z()+float32(t)*dir.Z(), p[2]) {
			p[2]--
		}

		if callback(p) {
			return
		}
	}
}
