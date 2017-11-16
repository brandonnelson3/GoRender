package gfx

import (
	"github.com/brandonnelson3/GoRender/messagebus"

	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	// Window is the Global window
	Window w

	// shadowSplits is the percents of the full view spectrum for each shadow cascade.
	// The 0th cascade is effective shadowSplits[0] to shadowSplits[1], therefore there
	// should be n+1 elements in this list where n is the number of cascades.
	shadowSplits = []float32{0.1, 15, 100, 500, 1000}
)

// Window is GoRender's primary Window representation. This class is a wrapper around an opengl glfw window, and GoRender specific functionality.
type w struct {
	Width, Height       uint32
	nearPlane, farPlane float32
	fieldOfViewDegrees  float32

	*glfw.Window
}

// CreateWindow instantiates and opens a new window with opengl. This is stored in the global package variable.
func CreateWindow(title string, width, height uint32, near, far, fov float32) {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 5)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(int(width), int(height), title, nil, nil)
	if err != nil {
		panic(err)
	}
	Window = w{width, height, near, far, fov, window}

	messagebus.RegisterType("key", handleEscape)
}

func handleEscape(m *messagebus.Message) {
	pressedKeys := m.Data1.([]glfw.Key)

	for _, key := range pressedKeys {
		if key == glfw.KeyEscape {
			Window.SetShouldClose(true)
		}
	}
}

// RecenterCursor recenters the mouse in this window.
func (window *w) RecenterCursor() {
	window.SetCursorPos(float64(window.Width)/2, float64(window.Height)/2)
}

// GetProjection returns the projection matrix.
func (window *w) GetProjection() mgl32.Mat4 {
	return mgl32.Perspective(mgl32.DegToRad(window.fieldOfViewDegrees), float32(window.Width)/float32(window.Height), window.nearPlane, window.farPlane)
}

func getPortionOfRange(near, far, nearPortion, farPortion float32) (float32, float32) {
	delta := far - near
	return near + delta*nearPortion, near + delta*farPortion
}

// GetShadowCascadePerspectiveProjection returns the i-th cascade's frustum specific perspective projection matrix.
func (window *w) GetShadowCascadePerspectiveProjection(i int) mgl32.Mat4 {
	return mgl32.Perspective(mgl32.DegToRad(window.fieldOfViewDegrees), float32(window.Width)/float32(window.Height), shadowSplits[i], shadowSplits[i+1])
}

// GetNearFar returns a mgl32.Vec2 consisting of the near and far planes for the given i-th cascade.
func (window *w) GetNearFar(i int) mgl32.Vec2 {
	n, f := getPortionOfRange(window.nearPlane, window.farPlane, shadowSplits[i], shadowSplits[i+1])
	return mgl32.Vec2{n, f}
}
