package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// IVec2 is a integer vector with 2 elements.
type IVec2 [2]int32

// IVector2 is a wrapper around a program/uniform for binding.
type IVector2 struct {
	program uint32
	uniform int32
}

// NewIVector2 instantiates a IVector2 for the provided program and uniform location.
func NewIVector2(p uint32, u int32) *IVector2 {
	return &IVector2{p, u}
}

// Set sets this Vector2 to the provided data, and updates the uniform data.
func (m *IVector2) Set(nv IVec2) {
	gl.ProgramUniform2iv(m.program, m.uniform, 1, &nv[0])
}
