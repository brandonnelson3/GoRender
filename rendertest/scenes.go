// Package rendertest defines named test scenes for GoRender's render-test mode.
//
// Each Scene specifies a fixed output resolution and a Setup function that
// deterministically configures the renderer (camera position/angles, lighting,
// objects, etc.) before a single frame is captured.
package rendertest

import (
	"log"
	"math"

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
	// It must configure all global gfx state (camera, lights, etc.) deterministically
	// and return the list of renderables to include in the frame.
	Setup func() []gfx.Renderable
}

// All is the canonical list of render-test scenes.
// Add new scenarios here; they will automatically be picked up by -rendertest mode.
var All = []Scene{
	{
		Name:   "sky_noon",
		Width:  640,
		Height: 360,
		Setup: func() []gfx.Renderable {
			// Camera looking due-north at a slight downward angle.
			gfx.FirstPerson.SetPose(mgl32.Vec3{0, 5, 0}, 0, -0.1)
			gfx.ActiveCamera = gfx.FirstPerson

			// Sun directly overhead/forward — noon lighting.
			gfx.ResetDirectionalLight(mgl32.Vec3{1, 1, 1}, 1.0, mgl32.Vec3{0, -1, 0}.Normalize())
			gfx.ResetPointLights()

			return nil
		},
	},
	{
		Name:   "sky_sunset",
		Width:  640,
		Height: 360,
		Setup: func() []gfx.Renderable {
			// Camera looking west into the sunset.
			gfx.FirstPerson.SetPose(mgl32.Vec3{0, 5, 0}, float32(math.Pi/2) /* west */, 0)
			gfx.ActiveCamera = gfx.FirstPerson

			// Low-angle warm sun from the west.
			gfx.ResetDirectionalLight(mgl32.Vec3{1, 0.5, 0.2}, 1.0, mgl32.Vec3{-1, -0.1, 0}.Normalize())
			gfx.ResetPointLights()

			return nil
		},
	},
	{
		Name:   "corner_room",
		Width:  640,
		Height: 360,
		Setup:  setupCornerRoom,
	},
}

// setupCornerRoom builds a small interior corner:
//   - Sand-textured floor
//   - Brick-textured left wall  (XZ plane, running along X)
//   - Brick-textured right wall (YZ plane, running along Z), 90° to the left wall
//   - Crate sitting in the corner where the two walls meet
//   - Single red point light above the crate
//
// The corner is at the origin; the room opens toward +X and +Z.
// Camera is placed diagonally looking back into the corner.
func setupCornerRoom() []gfx.Renderable {
	// Camera inside the room at (9,4,9), looking toward the corner at origin.
	// Forward = Rotate3DY(θ)*(1,0,0) = (cosθ, 0, -sinθ).
	// For -X,-Z direction: cosθ = -√2/2 → θ = 3π/4 (135°).
	gfx.FirstPerson.SetPose(
		mgl32.Vec3{9, 4, 9},
		float32(3*math.Pi/4), // 135° — toward -X,-Z (the corner)
		-0.2,                 // slight downward tilt to see the cube
	)
	gfx.ActiveCamera = gfx.FirstPerson

	// Pure indoor: no directional sun. The renderer's built-in ambient (.2,.2,.2)
	// provides the base fill; the red point light is the scene's main light.
	gfx.ResetDirectionalLight(mgl32.Vec3{0, 0, 0}, 0.0, mgl32.Vec3{0, -1, 0})

	gfx.ResetPointLights()
	gfx.AddPointLight(
		mgl32.Vec3{1.5, 5, 1.5}, // above the crate in the corner
		mgl32.Vec3{1, 0.95, 0.85}, // soft warm white
		1.0,
		22.0, // large radius so the red wash reaches all walls
	)

	floorTex := mustLoadTexture("assets/sand.png")
	wallTex := mustLoadTexture("assets/brick_wall.png")
	crateTex := mustLoadTexture("assets/crate1_diffuse.png")

	const (
		roomW = float32(16)
		roomD = float32(16)
		roomH = float32(8)
	)

	// Floor — Y=0, normal +Y. CCW from above: (0,0,0)→(0,0,D)→(W,0,D)
	floorN := mgl32.Vec3{0, 1, 0}
	floorVerts := []gfx.Vertex{
		{Vert: mgl32.Vec3{0, 0, 0}, Norm: floorN, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{0, 0, roomD}, Norm: floorN, UV: mgl32.Vec2{0, 4}},
		{Vert: mgl32.Vec3{roomW, 0, roomD}, Norm: floorN, UV: mgl32.Vec2{4, 4}},
		{Vert: mgl32.Vec3{0, 0, 0}, Norm: floorN, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{roomW, 0, roomD}, Norm: floorN, UV: mgl32.Vec2{4, 4}},
		{Vert: mgl32.Vec3{roomW, 0, 0}, Norm: floorN, UV: mgl32.Vec2{4, 0}},
	}

	// Ceiling — Y=roomH, normal -Y (faces down into the room, blocks the sky).
	// CCW from below: (0,H,0)→(W,H,0)→(W,H,D)
	ceilN := mgl32.Vec3{0, -1, 0}
	ceilVerts := []gfx.Vertex{
		{Vert: mgl32.Vec3{0, roomH, 0}, Norm: ceilN, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{roomW, roomH, 0}, Norm: ceilN, UV: mgl32.Vec2{4, 0}},
		{Vert: mgl32.Vec3{roomW, roomH, roomD}, Norm: ceilN, UV: mgl32.Vec2{4, 4}},
		{Vert: mgl32.Vec3{0, roomH, 0}, Norm: ceilN, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{roomW, roomH, roomD}, Norm: ceilN, UV: mgl32.Vec2{4, 4}},
		{Vert: mgl32.Vec3{0, roomH, roomD}, Norm: ceilN, UV: mgl32.Vec2{0, 4}},
	}

	// Left wall — Z=0, normal +Z. CCW from +Z: (0,0,0)→(W,0,0)→(W,H,0)
	leftN := mgl32.Vec3{0, 0, 1}
	leftVerts := []gfx.Vertex{
		{Vert: mgl32.Vec3{0, 0, 0}, Norm: leftN, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{roomW, 0, 0}, Norm: leftN, UV: mgl32.Vec2{4, 0}},
		{Vert: mgl32.Vec3{roomW, roomH, 0}, Norm: leftN, UV: mgl32.Vec2{4, 2}},
		{Vert: mgl32.Vec3{0, 0, 0}, Norm: leftN, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{roomW, roomH, 0}, Norm: leftN, UV: mgl32.Vec2{4, 2}},
		{Vert: mgl32.Vec3{0, roomH, 0}, Norm: leftN, UV: mgl32.Vec2{0, 2}},
	}

	// Right wall — X=0, normal +X. CCW from +X: (0,0,D)→(0,0,0)→(0,H,0)
	rightN := mgl32.Vec3{1, 0, 0}
	rightVerts := []gfx.Vertex{
		{Vert: mgl32.Vec3{0, 0, roomD}, Norm: rightN, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{0, 0, 0}, Norm: rightN, UV: mgl32.Vec2{4, 0}},
		{Vert: mgl32.Vec3{0, roomH, 0}, Norm: rightN, UV: mgl32.Vec2{4, 2}},
		{Vert: mgl32.Vec3{0, 0, roomD}, Norm: rightN, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{0, roomH, 0}, Norm: rightN, UV: mgl32.Vec2{4, 2}},
		{Vert: mgl32.Vec3{0, roomH, roomD}, Norm: rightN, UV: mgl32.Vec2{0, 2}},
	}

	cubeVerts := makeCube(mgl32.Vec3{0, 0, 0}, 2)

	return []gfx.Renderable{
		gfx.NewVAORenderable(floorVerts, floorTex),
		gfx.NewVAORenderable(ceilVerts, floorTex),
		gfx.NewVAORenderable(leftVerts, wallTex),
		gfx.NewVAORenderable(rightVerts, wallTex),
		gfx.NewVAORenderable(cubeVerts, crateTex),
	}
}

// makeCube builds a unit-cube with side length `size`, with its minimum corner
// at `origin`. All 6 faces are included with correct outward normals and UV tiling.
func makeCube(origin mgl32.Vec3, size float32) []gfx.Vertex {
	o := origin
	s := size
	verts := []gfx.Vertex{
		// Bottom (-Y)
		{Vert: mgl32.Vec3{o.X(), o.Y(), o.Z()}, Norm: mgl32.Vec3{0, -1, 0}, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y(), o.Z()}, Norm: mgl32.Vec3{0, -1, 0}, UV: mgl32.Vec2{1, 0}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y(), o.Z() + s}, Norm: mgl32.Vec3{0, -1, 0}, UV: mgl32.Vec2{1, 1}},
		{Vert: mgl32.Vec3{o.X(), o.Y(), o.Z()}, Norm: mgl32.Vec3{0, -1, 0}, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y(), o.Z() + s}, Norm: mgl32.Vec3{0, -1, 0}, UV: mgl32.Vec2{1, 1}},
		{Vert: mgl32.Vec3{o.X(), o.Y(), o.Z() + s}, Norm: mgl32.Vec3{0, -1, 0}, UV: mgl32.Vec2{0, 1}},
		// Top (+Y)
		{Vert: mgl32.Vec3{o.X(), o.Y() + s, o.Z()}, Norm: mgl32.Vec3{0, 1, 0}, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y() + s, o.Z() + s}, Norm: mgl32.Vec3{0, 1, 0}, UV: mgl32.Vec2{1, 1}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y() + s, o.Z()}, Norm: mgl32.Vec3{0, 1, 0}, UV: mgl32.Vec2{1, 0}},
		{Vert: mgl32.Vec3{o.X(), o.Y() + s, o.Z()}, Norm: mgl32.Vec3{0, 1, 0}, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{o.X(), o.Y() + s, o.Z() + s}, Norm: mgl32.Vec3{0, 1, 0}, UV: mgl32.Vec2{0, 1}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y() + s, o.Z() + s}, Norm: mgl32.Vec3{0, 1, 0}, UV: mgl32.Vec2{1, 1}},
		// Front (+Z)
		{Vert: mgl32.Vec3{o.X(), o.Y(), o.Z() + s}, Norm: mgl32.Vec3{0, 0, 1}, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y(), o.Z() + s}, Norm: mgl32.Vec3{0, 0, 1}, UV: mgl32.Vec2{1, 0}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y() + s, o.Z() + s}, Norm: mgl32.Vec3{0, 0, 1}, UV: mgl32.Vec2{1, 1}},
		{Vert: mgl32.Vec3{o.X(), o.Y(), o.Z() + s}, Norm: mgl32.Vec3{0, 0, 1}, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y() + s, o.Z() + s}, Norm: mgl32.Vec3{0, 0, 1}, UV: mgl32.Vec2{1, 1}},
		{Vert: mgl32.Vec3{o.X(), o.Y() + s, o.Z() + s}, Norm: mgl32.Vec3{0, 0, 1}, UV: mgl32.Vec2{0, 1}},
		// Back (-Z)
		{Vert: mgl32.Vec3{o.X() + s, o.Y(), o.Z()}, Norm: mgl32.Vec3{0, 0, -1}, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{o.X(), o.Y(), o.Z()}, Norm: mgl32.Vec3{0, 0, -1}, UV: mgl32.Vec2{1, 0}},
		{Vert: mgl32.Vec3{o.X(), o.Y() + s, o.Z()}, Norm: mgl32.Vec3{0, 0, -1}, UV: mgl32.Vec2{1, 1}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y(), o.Z()}, Norm: mgl32.Vec3{0, 0, -1}, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{o.X(), o.Y() + s, o.Z()}, Norm: mgl32.Vec3{0, 0, -1}, UV: mgl32.Vec2{1, 1}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y() + s, o.Z()}, Norm: mgl32.Vec3{0, 0, -1}, UV: mgl32.Vec2{0, 1}},
		// Left (-X)
		{Vert: mgl32.Vec3{o.X(), o.Y(), o.Z()}, Norm: mgl32.Vec3{-1, 0, 0}, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{o.X(), o.Y(), o.Z() + s}, Norm: mgl32.Vec3{-1, 0, 0}, UV: mgl32.Vec2{1, 0}},
		{Vert: mgl32.Vec3{o.X(), o.Y() + s, o.Z() + s}, Norm: mgl32.Vec3{-1, 0, 0}, UV: mgl32.Vec2{1, 1}},
		{Vert: mgl32.Vec3{o.X(), o.Y(), o.Z()}, Norm: mgl32.Vec3{-1, 0, 0}, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{o.X(), o.Y() + s, o.Z() + s}, Norm: mgl32.Vec3{-1, 0, 0}, UV: mgl32.Vec2{1, 1}},
		{Vert: mgl32.Vec3{o.X(), o.Y() + s, o.Z()}, Norm: mgl32.Vec3{-1, 0, 0}, UV: mgl32.Vec2{0, 1}},
		// Right (+X)
		{Vert: mgl32.Vec3{o.X() + s, o.Y(), o.Z() + s}, Norm: mgl32.Vec3{1, 0, 0}, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y(), o.Z()}, Norm: mgl32.Vec3{1, 0, 0}, UV: mgl32.Vec2{1, 0}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y() + s, o.Z()}, Norm: mgl32.Vec3{1, 0, 0}, UV: mgl32.Vec2{1, 1}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y(), o.Z() + s}, Norm: mgl32.Vec3{1, 0, 0}, UV: mgl32.Vec2{0, 0}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y() + s, o.Z()}, Norm: mgl32.Vec3{1, 0, 0}, UV: mgl32.Vec2{1, 1}},
		{Vert: mgl32.Vec3{o.X() + s, o.Y() + s, o.Z() + s}, Norm: mgl32.Vec3{1, 0, 0}, UV: mgl32.Vec2{0, 1}},
	}
	return verts
}

func mustLoadTexture(path string) uint32 {
	tex, err := gfx.LoadTexture(path)
	if err != nil {
		log.Fatalf("rendertest: failed to load texture %q: %v", path, err)
	}
	return tex
}
