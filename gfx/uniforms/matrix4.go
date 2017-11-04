package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Matrix4 is a wrapper around a mgl32.Mat4, and a program/uniform for binding.
type Matrix4 struct {
	program uint32
	uniform int32
}

// NewMatrix4 instantiates an identity matrix for the provided program and uniform location.
func NewMatrix4(p uint32, u int32) *Matrix4 {
	return &Matrix4{p, u}
}

// Set sets this Matrix4 to the provided data, and updates the uniform data.
func (m *Matrix4) Set(nm mgl32.Mat4) {
	gl.ProgramUniformMatrix4fv(m.program, m.uniform, 1, false, &nm[0])
}
