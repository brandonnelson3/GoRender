package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// Sampler2D is a wrapper around a int32 which is the sampler texture id, and a program/uniform for binding.
type Sampler2D struct {
	program uint32
	uniform int32
}

// NewSampler2D instantiates a sampler2d for the provided program, and uniform location.
func NewSampler2D(p uint32, u int32) *Sampler2D {
	return &Sampler2D{p, u}
}

// Set sets this Sampler2D to the provided id, and updates the uniform data.
func (m *Sampler2D) Set(texture int, slot int32, samplerID uint32) {
	gl.ActiveTexture(uint32(texture))
	gl.ProgramUniform1i(m.program, m.uniform, slot)
	gl.BindTexture(gl.TEXTURE_2D, samplerID)
}
