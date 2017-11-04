package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Vector3 is a wrapper around a mgl32.Vec3, and a program/uniform for binding.
type Vector3 struct {
	program uint32
	uniform int32
}

// NewVector3 instantiates a 0 vector for the provided program and uniform location.
func NewVector3(p uint32, u int32) *Vector3 {
	return &Vector3{p, u}
}

// Set Sets this Vector3 to the provided data, and updates the uniform data.
func (m *Vector3) Set(nv mgl32.Vec3) {
	gl.ProgramUniform3fv(m.program, m.uniform, 1, &nv[0])
}
