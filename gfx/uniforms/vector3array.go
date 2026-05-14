package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// Vector3Array is a wrapper around an array of vec3 uniforms, and a program/uniform for binding.
type Vector3Array struct {
	program uint32
	uniform int32
}

// NewVector3Array instantiates a Vector3Array for the provided program and uniform location.
func NewVector3Array(p uint32, u int32) *Vector3Array {
	return &Vector3Array{p, u}
}

// Set sets this Vector3Array uniform to the provided slice, and updates the uniform data.
// first must point to the first float32 of the contiguous data (e.g. &vec3s[0][0]).
func (m *Vector3Array) Set(first *float32, count int32) {
	gl.ProgramUniform3fv(m.program, m.uniform, count, first)
}
