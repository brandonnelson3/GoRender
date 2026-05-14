package uniforms

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

// SamplerCubeArray binds an array of cubemap textures to consecutive texture units.
// The GLSL uniform is declared as: uniform samplerCube name[N];
type SamplerCubeArray struct {
	program uint32
	uniform int32
}

// NewSamplerCubeArray instantiates a SamplerCubeArray for the provided program and uniform location.
func NewSamplerCubeArray(p uint32, u int32) *SamplerCubeArray {
	return &SamplerCubeArray{p, u}
}

// Set binds each cubemap in samplerIDs to successive texture units starting at
// baseTexture (e.g. gl.TEXTURE7), sets the uniform array to the matching slot
// indices starting at baseSlot (e.g. 7), and activates each texture unit.
func (m *SamplerCubeArray) Set(baseTexture int, baseSlot int32, samplerIDs []uint32) {
	slots := make([]int32, len(samplerIDs))
	for i, id := range samplerIDs {
		gl.ActiveTexture(uint32(baseTexture) + uint32(i))
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, id)
		slots[i] = baseSlot + int32(i)
	}
	if len(slots) > 0 {
		gl.ProgramUniform1iv(m.program, m.uniform, int32(len(slots)), &slots[0])
	}
}
