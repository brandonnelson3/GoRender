package gfx

import "github.com/go-gl/mathgl/mgl32"

const firstPersonCameraModelScale = float32(0.027)

// InitFirstPersonCameraModel loads the OBJ at path, builds a VAORenderable
// scaled to a realistic camera size (~0.56 world units wide), and stores it in
// FirstPersonCameraRenderable. The renderer will automatically draw it at
// FirstPerson's position whenever ThirdPerson is the active camera.
//
// Scale rationale: the Blender model is ~21 units wide (X: -10.5 to +10.5).
// A real 35mm film camera body is ~14 cm wide. At 1 world unit ≈ 0.25 m,
// the target width is 0.56 units → scale = 0.56/21 ≈ 0.027.
//
// If the model has already been loaded this call is a no-op (idempotent), so
// it is safe to call it from every scene setup function.
func InitFirstPersonCameraModel(path string) error {
	if FirstPersonCameraRenderable != nil {
		return nil // already loaded
	}
	obj, err := LoadObjFile(path)
	if err != nil {
		return err
	}
	r := obj.GetChunkedRenderable()
	s := firstPersonCameraModelScale
	r.Scale = mgl32.Scale3D(s, s, s)
	// Rotation and Position are updated every frame by the renderer.
	FirstPersonCameraRenderable = r
	return nil
}
