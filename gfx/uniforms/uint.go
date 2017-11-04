package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// UInt is a wrapper around a program/uniform for binding.
type UInt struct {
	program uint32
	uniform int32
}

// NewUInt instantiates a UInt for the provided program and uniform location.
func NewUInt(p uint32, u int32) *UInt {
	return &UInt{p, u}
}

// Set sets this UInt to the provided data, and updates the uniform data.
func (m *UInt) Set(i uint32) {
	gl.ProgramUniform1ui(m.program, m.uniform, i)
}
