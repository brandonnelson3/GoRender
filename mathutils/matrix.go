package mathutils

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

func PerspectiveZO(fovy, aspect, near, far float32) mgl32.Mat4 {
	// fovy = (fovy * math.Pi) / 180.0 // convert from degrees to radians
	nmf, f := near-far, float32(1./math.Tan(float64(fovy)/2.0))

	return mgl32.Mat4{f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, far / nmf, -1,
		0, 0, (far * near) / nmf, 0}
}
