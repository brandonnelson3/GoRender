package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// FloatArray is a wrapper around a program/uniform for binding.
type FloatArray struct {
	program uint32
	uniform int32
}

// NewFloatArray instantiates a FloatArray for the provided program and uniform location.
func NewFloatArray(p uint32, u int32) *FloatArray {
	return &FloatArray{p, u}
}

// Set sets this FloatArray to the provided data, and updates the uniform data.
func (m *FloatArray) Set(first *float32, numberOfFloats int32) {
	gl.ProgramUniform1fv(m.program, m.uniform, numberOfFloats, first)
}
