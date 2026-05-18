package gfx

import (
	"log"

	"github.com/brandonnelson3/GoRender/gfx/shaders"
	"github.com/brandonnelson3/GoRender/messagebus"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	hudVao        uint32
	hudTexture    uint32
	hudShader     *shaders.HudShader
	hudEnabled    bool = true // Default to showing it as requested
	hudInitialized bool
)

// InitHUD loads the transparent HUD texture and sets up the full-screen quad VAO/VBO.
func InitHUD() {
	if hudInitialized {
		return
	}

	var err error
	hudShader, err = shaders.NewHudShader()
	if err != nil {
		log.Printf("[HUD Overlay] Failed to compile HUD shader: %v\n", err)
		return
	}

	// Load our transparent HUD texture
	hudTexture, err = LoadTexture("slideframes/hud.png")
	if err != nil {
		log.Printf("[HUD Overlay] Failed to load slideframes/hud.png: %v\n", err)
		return
	}
	log.Println("[HUD Overlay] Successfully loaded slideframes/hud.png texture.")

	// Define full screen quad vertices.
	// We want to stretch/fit it to the entire window.
	topLeft  := mgl32.Vec2{0.0, 0.0}
	topRight := mgl32.Vec2{float32(Window.Width), 0.0}
	botLeft  := mgl32.Vec2{0.0, float32(Window.Height)}
	botRight := mgl32.Vec2{float32(Window.Width), float32(Window.Height)}

	// UV coordinate mapping to map the Go image top-left to UV top-left (0, 1) in OpenGL
	planeVertices := []PipVertex{
		{topLeft,  mgl32.Vec2{0, 0}},
		{botRight, mgl32.Vec2{1, 1}},
		{topRight, mgl32.Vec2{1, 0}},
		{topLeft,  mgl32.Vec2{0, 0}},
		{botLeft,  mgl32.Vec2{0, 1}},
		{botRight, mgl32.Vec2{1, 1}},
	}

	gl.GenVertexArrays(1, &hudVao)
	gl.BindVertexArray(hudVao)

	var hudVbo uint32
	gl.GenBuffers(1, &hudVbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, hudVbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(planeVertices)*4*4, gl.Ptr(planeVertices), gl.STATIC_DRAW)

	BindPipVertexAttributes(hudShader.Program())
	gl.BindVertexArray(0)

	// Register key callback to toggle the HUD using the 'H' key
	messagebus.RegisterType("key", func(m *messagebus.Message) {
		pressedThisFrame := m.Data2.([]glfw.Key)
		for _, key := range pressedThisFrame {
			if key == glfw.KeyH {
				hudEnabled = !hudEnabled
				log.Printf("[HUD Overlay] Toggled Google Slides HUD. Active: %v\n", hudEnabled)
			}
		}
	})

	hudInitialized = true
}

// RenderHUD draws the transparent full-screen HUD texture over the screen.
func RenderHUD() {
	if !hudEnabled || RenderTestMode {
		return
	}

	if !hudInitialized {
		InitHUD()
	}

	gl.Disable(gl.DEPTH_TEST)
	gl.Disable(gl.CULL_FACE)

	// Enable standard alpha blending just in case it was modified
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	hudShader.Use()
	hudShader.Projection.Set(mgl32.Ortho(0.0, float32(Window.Width), float32(Window.Height), 0.0, -1.0, 1.0))
	
	// Bind HUD texture to unit 8
	hudShader.HudTexture.Set(gl.TEXTURE8, 8, hudTexture)

	gl.BindVertexArray(hudVao)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)

	gl.ActiveTexture(gl.TEXTURE8)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.BindVertexArray(0)

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
}
