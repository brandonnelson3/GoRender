package mathutils

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// PerspectiveZO is a function exists because the mgl32.Perspective function is a [-1, 1]
// implementation. This engine is designed with a [0, 1] in mind and therefore this function
// is an implementation as such.
func PerspectiveZO(fovy, aspect, near, far float32) mgl32.Mat4 {
	nmf, f := near-far, float32(1./math.Tan(float64(fovy)/2.0))

	return mgl32.Mat4{f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, far / nmf, -1,
		0, 0, (far * near) / nmf, 0}
}
