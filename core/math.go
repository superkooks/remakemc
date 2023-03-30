package core

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// Useful for our chunk system
// https://en.wikipedia.org/wiki/Modulo#/media/File:Divmod_floored.svg
// https://en.wikipedia.org/wiki/Modulo#Implementing_other_modulo_definitions_using_truncation
func FlooredDivision(x int, y int) int {
	q := x / y
	r := x % y

	if (r > 0 && y < 0) || (r < 0 && y > 0) {
		q--
		r += y
	}

	return q
}

func FlooredRemainder(x int, y int) int {
	q := x / y
	r := x % y

	if (r > 0 && y < 0) || (r < 0 && y > 0) {
		q--
		r += y
	}

	return r
}

func RoundThenFloor(f float32) float32 {
	return float32(math.Floor(float64(mgl32.Round(f, 3))))
}
