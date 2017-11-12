package gfx

import (
	"unsafe"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	directionalLight DirectionalLight

	directionalLightBuffer uint32
)

// DirectionalLight represents all of the data about the DirectionaLight in the scene.
type DirectionalLight struct {
	Color      mgl32.Vec3
	Brightness float32
	Direction  mgl32.Vec3
}

// InitDirectionalLights sets up buffer space for storage of Directional Light data.
func InitDirectionalLights() {
	directionalLight = DirectionalLight{
		Color:      mgl32.Vec3{1, 1, .8},
		Brightness: 0.35,
		Direction:  mgl32.Vec3{0.001, -1, 0.001}.Normalize(),
	}

	// Prepare light buffer
	gl.GenBuffers(1, &directionalLightBuffer)

	// Bind light buffer
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, directionalLightBuffer)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, int(unsafe.Sizeof(directionalLight)), unsafe.Pointer(&directionalLight), gl.DYNAMIC_DRAW)

	// Unbind for safety.
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, 0)
}

// GetDirectionalLightBuffer retrieves the private directionalLightBuffer variable.
func GetDirectionalLightBuffer() uint32 {
	return directionalLightBuffer
}

// GetDirectionalLightDirection returns the private directional light's direction.
func GetDirectionalLightDirection() mgl32.Vec3 {
	return directionalLight.Direction
}
