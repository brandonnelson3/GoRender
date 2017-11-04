package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// UIVec4 is a integer vector with 4 elements.
type UIVec4 [4]uint32

// UIVector4 is a wrapper around a program/uniform for binding.
type UIVector4 struct {
	program uint32
	uniform int32
}

// NewUIVector4 instantiates a UIVector4 for the provided program and uniform location.
func NewUIVector4(p uint32, u int32) *UIVector4 {
	return &UIVector4{p, u}
}

// Set sets this UIVector4 to the provided data, and updates the uniform data.
func (m *UIVector4) Set(nv UIVec4) {
	gl.ProgramUniform4uiv(m.program, m.uniform, 1, &nv[0])
}
