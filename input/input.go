package input

import (
	"fmt"

	"github.com/brandonnelson3/GoRender/messagebus"

	"github.com/go-gl/glfw/v3.1/glfw"
)

const (
	keyRange = glfw.KeyLast
)

var (
	down          [keyRange]bool
	downThisFrame [keyRange]bool
)

// MouseInput is message data which is sent for every mouse cursor position callback.
type MouseInput struct {
	X, Y float64
}

// Update calls all of the currently pressed keys.
func Update() {
	pressedKeys := make([]glfw.Key, 0, 10)
	pressedKeysThisFrame := make([]glfw.Key, 0, 10)

	// glfw.KeySpace is the lowest key.
	for i := glfw.KeySpace; i < keyRange; i++ {
		if down[i] {
			pressedKeys = append(pressedKeys, i)
		}
		if downThisFrame[i] {
			pressedKeysThisFrame = append(pressedKeysThisFrame, i)
			downThisFrame[i] = false
		}
	}
	if len(pressedKeys) > 0 {
		messagebus.SendSync(&messagebus.Message{Type: "key", Data1: pressedKeys, Data2: pressedKeysThisFrame})
	}
}

// KeyCallBack is the function bound to handle key events from OpenGL.
func KeyCallBack(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		down[key] = true
		downThisFrame[key] = true
	}
	if action == glfw.Release {
		down[key] = false
	}
}

// MouseButtonCallback is the function bound to handle mouse button events from OpenGL.
func MouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	fmt.Printf("Got mouse press: %v\n", button)
}

// CursorPosCallback is the function bound to handle mouse movement events from OpenGL.
func CursorPosCallback(w *glfw.Window, x, y float64) {
	messagebus.SendSync(&messagebus.Message{Type: "mouse", Data1: MouseInput{x, y}})
}
