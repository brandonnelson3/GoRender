package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// Int is a wrapper around a program/uniform for binding.
type Int struct {
	program uint32
	uniform int32
}

// NewInt instantiates a Int for the provided program and uniform location.
func NewInt(p uint32, u int32) *Int {
	return &Int{p, u}
}

// Set sets this Int to the provided data, and updates the uniform data.
func (m *Int) Set(i int32) {
	gl.ProgramUniform1i(m.program, m.uniform, i)
}
