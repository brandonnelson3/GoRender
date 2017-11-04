package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Vector4 is a wrapper around a mgl32.Vec4, and a program/uniform for binding.
type Vector4 struct {
	program uint32
	uniform int32
}

// NewVector4 instantiates a 0 vector for the provided program and uniform location.
func NewVector4(p uint32, u int32) *Vector4 {
	return &Vector4{p, u}
}

// Set Sets this Vector4 to the provided data, and updates the uniform data.
func (m *Vector4) Set(nv mgl32.Vec4) {
	gl.ProgramUniform4fv(m.program, m.uniform, 1, &nv[0])
}
