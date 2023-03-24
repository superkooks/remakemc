package core

import (
	"math"

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
