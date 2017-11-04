package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// Float is a wrapper around a program/uniform for binding.
type Float struct {
	program uint32
	uniform int32
}

// NewFloat instantiates a Float for the provided program and uniform location.
func NewFloat(p uint32, u int32) *Float {
	return &Float{p, u}
}

// Set sets this Float to the provided data, and updates the uniform data.
func (m *Float) Set(f float32) {
	gl.ProgramUniform1f(m.program, m.uniform, f)
}
