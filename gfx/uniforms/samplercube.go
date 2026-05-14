package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// SamplerCube is a wrapper around a cubemap texture uniform.
type SamplerCube struct {
	program uint32
	uniform int32
}

// NewSamplerCube instantiates a SamplerCube for the provided program and uniform location.
func NewSamplerCube(p uint32, u int32) *SamplerCube {
	return &SamplerCube{p, u}
}

// Set binds a cubemap texture to the given texture unit/slot.
func (m *SamplerCube) Set(texture int, slot int32, samplerID uint32) {
	gl.ActiveTexture(uint32(texture))
	gl.ProgramUniform1i(m.program, m.uniform, slot)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, samplerID)
}
