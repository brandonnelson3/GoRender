package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// IVec3 is a integer vector with 3 elements.
type IVec3 [3]int32

// IVector3 is a wrapper around a program/uniform for binding.
type IVector3 struct {
	program uint32
	uniform int32
}

// NewIVector3 instantiates a IVector3 for the provided program and uniform location.
func NewIVector3(p uint32, u int32) *IVector3 {
	return &IVector3{p, u}
}

// Set sets this Vector3 to the provided data, and updates the uniform data.
func (m *IVector3) Set(nv IVec3) {
	gl.ProgramUniform3iv(m.program, m.uniform, 1, &nv[0])
}
