package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// UIVec2 is a integer vector with 2 elements.
type UIVec2 [2]uint32

// UIVector2 is a wrapper around a program/uniform for binding.
type UIVector2 struct {
	program uint32
	uniform int32
}

// NewUIVector2 instantiates a UIVector2 for the provided program and uniform location.
func NewUIVector2(p uint32, u int32) *UIVector2 {
	return &UIVector2{p, u}
}

// Set sets this UIVector2 to the provided data, and updates the uniform data.
func (m *UIVector2) Set(nv UIVec2) {
	gl.ProgramUniform2uiv(m.program, m.uniform, 1, &nv[0])
}
