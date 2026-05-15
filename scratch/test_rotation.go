package main

import (
	"fmt"
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

func main() {
	v := mgl32.Vec3{1, 0, 0}
	angles := []float32{0, math.Pi / 4, math.Pi / 2, math.Pi, 3 * math.Pi / 2}
	for _, a := range angles {
		f := mgl32.Rotate3DY(a).Mul3x1(v)
		fmt.Printf("Angle %.2f: Forward [%.2f, %.2f, %.2f]\n", a, f.X(), f.Y(), f.Z())
	}
}
