package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// UIVec3 is a integer vector with 3 elements.
type UIVec3 [3]uint32

// UIVector3 is a wrapper around a program/uniform for binding.
type UIVector3 struct {
	program uint32
	uniform int32
}

// NewUIVector3 instantiates a UIVector3 for the provided program and uniform location.
func NewUIVector3(p uint32, u int32) *UIVector3 {
	return &UIVector3{p, u}
}

// Set sets this UIVector3 to the provided data, and updates the uniform data.
func (m *UIVector3) Set(nv UIVec3) {
	gl.ProgramUniform3uiv(m.program, m.uniform, 1, &nv[0])
}
