package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// IVec4 is a integer vector with 4 elements.
type IVec4 [4]int32

// IVector4 is a wrapper around a program/uniform for binding.
type IVector4 struct {
	program uint32
	uniform int32
}

// NewIVector4 instantiates a IVector4 for the provided program and uniform location.
func NewIVector4(p uint32, u int32) *IVector4 {
	return &IVector4{p, u}
}

// Set sets this Vector4 to the provided data, and updates the uniform data.
func (m *IVector4) Set(nv IVec4) {
	gl.ProgramUniform4iv(m.program, m.uniform, 1, &nv[0])
}
