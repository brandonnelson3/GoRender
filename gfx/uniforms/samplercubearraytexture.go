package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// SamplerCubeArrayTexture binds a single Texture Cube Map Array to a texture unit.
// The GLSL uniform is declared as: uniform samplerCubeArray name;
type SamplerCubeArrayTexture struct {
	program uint32
	uniform int32
}

// NewSamplerCubeArrayTexture instantiates a SamplerCubeArrayTexture for the provided program and uniform location.
func NewSamplerCubeArrayTexture(p uint32, u int32) *SamplerCubeArrayTexture {
	return &SamplerCubeArrayTexture{p, u}
}

// Set binds the texture array to the given texture unit and sets the uniform slot index.
func (m *SamplerCubeArrayTexture) Set(texture int, slot int32, samplerID uint32) {
	gl.ActiveTexture(uint32(texture))
	gl.BindTexture(gl.TEXTURE_CUBE_MAP_ARRAY, samplerID)
	gl.ProgramUniform1i(m.program, m.uniform, slot)
}
