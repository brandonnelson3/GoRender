// Package rendertest defines named test scenes for GoRender's render-test mode.
//
// Each Scene specifies a fixed output resolution and a Setup function that
// deterministically configures the renderer (camera position/angles, lighting,
// etc.) before a single frame is captured.
package rendertest

import (
	"github.com/brandonnelson3/GoRender/gfx"
	"github.com/go-gl/mathgl/mgl32"
)

// Scene is a named, self-contained rendering scenario.
type Scene struct {
	// Name is used as the output file basename (e.g. "sky_noon" → "sky_noon.png").
	Name string
	// Width and Height define the offscreen FBO resolution for this scene.
	Width, Height int32
	// Setup is called once, with the GL context current, before Render is called.
	// It must configure all global gfx state (camera, lights, etc.) deterministically.
	Setup func()
}

// All is the canonical list of render-test scenes.
// Add new scenarios here; they will automatically be picked up by -rendertest mode.
var All = []Scene{
	{
		Name:   "sky_noon",
		Width:  640,
		Height: 360,
		Setup: func() {
			// Camera looking due-north at a slight downward angle.
			gfx.FirstPerson.SetPose(mgl32.Vec3{0, 5, 0}, 0, -0.1)
			gfx.ActiveCamera = gfx.FirstPerson

			// Sun directly overhead/forward — noon lighting.
			gfx.ResetDirectionalLight(mgl32.Vec3{1, 1, 1}, 1.0, mgl32.Vec3{0, -1, 0}.Normalize())
		},
	},
	{
		Name:   "sky_sunset",
		Width:  640,
		Height: 360,
		Setup: func() {
			// Camera looking west into the sunset.
			gfx.FirstPerson.SetPose(mgl32.Vec3{0, 5, 0}, 1.5708 /* π/2, west */, 0)
			gfx.ActiveCamera = gfx.FirstPerson

			// Low-angle warm sun from the west.
			gfx.ResetDirectionalLight(mgl32.Vec3{1, 0.5, 0.2}, 1.0, mgl32.Vec3{-1, -0.1, 0}.Normalize())
		},
	},
}
