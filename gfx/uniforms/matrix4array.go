package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// Matrix4Array is a wrapper around an array of mgl32.Mat4, and a program/uniform for binding.
type Matrix4Array struct {
	program uint32
	uniform int32
}

// NewMatrix4Array instantiates a matrix array for the provided program and uniform location.
func NewMatrix4Array(p uint32, u int32) *Matrix4Array {
	return &Matrix4Array{p, u}
}

// Set sets this Matrix4 to the provided data, and updates the uniform data.
func (m *Matrix4Array) Set(first *float32, numberOfMatrices int32) {
	gl.ProgramUniformMatrix4fv(m.program, m.uniform, numberOfMatrices, false, first)
}
