package camera

import (
	"fmt"
	"math"

	"github.com/brandonnelson3/GoRender/gfx"
	"github.com/brandonnelson3/GoRender/input"
	"github.com/brandonnelson3/GoRender/messagebus"

	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	pi2 = math.Pi / 2.0
	pi4 = math.Pi / 4.0
)

var (
	FirstPerson *camera
	ThirdPerson *camera
	Active      *camera
)

type camera struct {
	position        mgl32.Vec3
	direction       mgl32.Vec3
	horizontalAngle float32
	verticalAngle   float32
	sensitivity     float32
	speed           float32
}

// InitCameras instantiates new cameras into the package first and third person package variables.
func InitCameras() {
	FirstPerson = &camera{position: mgl32.Vec3{0, 0, 0}, horizontalAngle: 0, verticalAngle: 0, sensitivity: 0.001, speed: 20}
	ThirdPerson = &camera{position: mgl32.Vec3{0, 0, 0}, horizontalAngle: 0, verticalAngle: 0, sensitivity: 0.001, speed: 20}
	Active = FirstPerson
	messagebus.RegisterType("key", Active.handleMovement)
	messagebus.RegisterType("mouse", Active.handleMouse)
}

// Update is called every frame to execute this frame's movement.
func (c *camera) Update(d float64) {
	if c.direction.X() != 0 || c.direction.Y() != 0 || c.direction.Z() != 0 {
		delta := c.direction.Normalize().Mul(float32(d) * c.speed)
		c.position = c.position.Add(delta)
		c.direction = mgl32.Vec3{0, 0, 0}
	}
}

// GetPosition returns the position of this camera.
func (c *camera) GetPosition() mgl32.Vec3 {
	return c.position
}

// GetForward returns the forward unit vector for this camera.
func (c *camera) GetForward() mgl32.Vec3 {
	return mgl32.Rotate3DY(c.horizontalAngle).Mul3x1(mgl32.Rotate3DZ(c.verticalAngle).Mul3x1((mgl32.Vec3{1, 0, 0})))
}

// GetRight returns the right unit vector for this camera.
func (c *camera) GetRight() mgl32.Vec3 {
	return mgl32.Rotate3DY(c.horizontalAngle).Mul3x1(mgl32.Vec3{0, 0, 1})
}

// GetView returns the current view matrix for this camera.
func (c *camera) GetView() mgl32.Mat4 {
	return mgl32.LookAtV(c.position, c.position.Add(c.GetForward()), mgl32.Vec3{0, 1, 0})
}

func (c *camera) handleMovement(m *messagebus.Message) {
	direction := mgl32.Vec3{0, 0, 0}
	pressedKeys := m.Data1.([]glfw.Key)
	for _, key := range pressedKeys {
		switch key {
		case glfw.KeyW:
			direction = direction.Add(c.GetForward())
		case glfw.KeyS:
			direction = direction.Sub(c.GetForward())
		case glfw.KeyD:
			direction = direction.Add(c.GetRight())
		case glfw.KeyA:
			direction = direction.Sub(c.GetRight())
		}
	}
	pressedKeysThisFrame := m.Data2.([]glfw.Key)
	for _, key := range pressedKeysThisFrame {
		switch key {
		case glfw.KeyP:
			messagebus.SendAsync(&messagebus.Message{System: "Camera", Type: "log", Data1: fmt.Sprintf("position: mgl32.Vec3{%f, %f, %f}, horizontalAngle: %f, verticalAngle: %f", c.position.X(), c.position.Y(), c.position.Z(), c.horizontalAngle, c.verticalAngle)})
		}
	}
	c.direction = direction
}

func (c *camera) handleMouse(m *messagebus.Message) {
	mouseInput := m.Data1.(input.MouseInput)
	c.verticalAngle -= c.sensitivity * float32(mouseInput.Y-float64(gfx.Window.Height)/2)
	if c.verticalAngle < -pi2 {
		c.verticalAngle = float32(-pi2 + 0.0001)
	}
	if c.verticalAngle > pi2 {
		c.verticalAngle = float32(pi2 - 0.0001)
	}
	c.horizontalAngle -= c.sensitivity * float32(mouseInput.X-float64(gfx.Window.Width)/2)
	for c.horizontalAngle < 0 {
		c.horizontalAngle += float32(2 * math.Pi)
	}
}
