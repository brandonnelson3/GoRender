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
		Color:      mgl32.Vec3{1, 1, 1},
		Brightness: 1.0,
		Direction:  mgl32.Vec3{1, -1, 0}.Normalize(),
	}

	// Prepare light buffer
	gl.GenBuffers(1, &directionalLightBuffer)

	// Bind light buffer
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, directionalLightBuffer)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, int(unsafe.Sizeof(directionalLight)), unsafe.Pointer(&directionalLight), gl.DYNAMIC_DRAW)

	// Unbind for safety.
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, 0)
}

// UpdateDirectionalLight updates the global directional light by calling the provided function and applying the result of the function.
func UpdateDirectionalLight(f func(dL DirectionalLight) DirectionalLight) {
	directionalLight = f(directionalLight)

	// Bind light buffer
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, directionalLightBuffer)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, int(unsafe.Sizeof(directionalLight)), unsafe.Pointer(&directionalLight), gl.DYNAMIC_DRAW)

	// Unbind for safety.
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, 0)
}

// ResetDirectionalLight replaces the directional light with the given values.
// It is intended for deterministic setup in render-test mode.
func ResetDirectionalLight(color mgl32.Vec3, brightness float32, direction mgl32.Vec3) {
	UpdateDirectionalLight(func(_ DirectionalLight) DirectionalLight {
		return DirectionalLight{Color: color, Brightness: brightness, Direction: direction}
	})
}

// GetDirectionalLightBuffer retrieves the private directionalLightBuffer variable.
func GetDirectionalLightBuffer() uint32 {
	return directionalLightBuffer
}

// GetDirectionalLightDirection returns the private directional light's direction.
func GetDirectionalLightDirection() mgl32.Vec3 {
	return directionalLight.Direction
}
