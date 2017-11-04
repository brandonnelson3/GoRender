package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Vector2 is a wrapper around a mgl32.Vec2, and a program/uniform for binding.
type Vector2 struct {
	program uint32
	uniform int32
}

// NewVector2 instantiates a 0 vector for the provided program and uniform location.
func NewVector2(p uint32, u int32) *Vector2 {
	return &Vector2{p, u}
}

// Set Sets this Vector2 to the provided data, and updates the uniform data.
func (m *Vector2) Set(nv mgl32.Vec2) {
	gl.ProgramUniform2fv(m.program, m.uniform, 1, &nv[0])
}
