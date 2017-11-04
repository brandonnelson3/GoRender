package gfx

import (
	"github.com/brandonnelson3/GoRender/messagebus"

	"github.com/go-gl/glfw/v3.1/glfw"
)

var (
	// Window is the Global window
	Window w
)

// Window is GoRender's primary Window representation. This class is a wrapper around an opengl glfw window, and GoRender specific functionality.
type w struct {
	width, height       int
	nearPlane, farPlane float32
	fieldOfViewDegrees  float32

	*glfw.Window
}

// CreateWindow instantiates and opens a new window with opengl. This is stored in the global package variable.
func CreateWindow(title string, width, height int, near, far, fov float32) {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 5)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(width, height, title, nil, nil)
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
