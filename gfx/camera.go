package gfx

import (
	"fmt"
	"math"
	"time"

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
	// FirstPerson is the "main" camera in the scene.
	FirstPerson *camera
	// ThirdPerson is a "secondary" camera in the scene, mainly for observing the world around the FirstPerson camera.
	ThirdPerson *camera
	// ActiveCamera is either FirstPerson or ThirdPerson, depending on which is currently being used for rendering.
	ActiveCamera *camera
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
	FirstPerson = &camera{position: mgl32.Vec3{0, 9, 0}, horizontalAngle: 0, verticalAngle: 0, sensitivity: 0.001, speed: 20}
	ThirdPerson = &camera{position: mgl32.Vec3{0, 9, 0}, horizontalAngle: 0, verticalAngle: 0, sensitivity: 0.001, speed: 20}
	ActiveCamera = FirstPerson
	messagebus.RegisterType("key", ActiveCamera.handleMovement)
	messagebus.RegisterType("mouse", ActiveCamera.handleMouse)

	go updateConsoleOnTimer()
}

func updateConsoleOnTimer() {
	for range time.Tick(time.Millisecond * 100) {
		cameraPosition := ActiveCamera.GetPosition()
		cameraPositionValue := fmt.Sprintf("[%.2f, %.2f, %.2f]", cameraPosition.X(), cameraPosition.Y(), cameraPosition.Z())
		messagebus.SendAsync(&messagebus.Message{Type: "console", Data1: "camera_position", Data2: cameraPositionValue})

		cameraForward := ActiveCamera.GetForward()
		cameraForwardValue := fmt.Sprintf("[%.2f, %.2f, %.2f]", cameraForward.X(), cameraForward.Y(), cameraForward.Z())
		messagebus.SendAsync(&messagebus.Message{Type: "console", Data1: "camera_forward", Data2: cameraForwardValue})

		cameraAngleValue := fmt.Sprintf("[H: %.2f, V:%.2f]", ActiveCamera.horizontalAngle, ActiveCamera.verticalAngle)
		messagebus.SendAsync(&messagebus.Message{Type: "console", Data1: "camera_angle", Data2: cameraAngleValue})
	}
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
	c.direction = direction
}

func (c *camera) handleMouse(m *messagebus.Message) {
	mouseInput := m.Data1.(input.MouseInput)
	c.verticalAngle -= c.sensitivity * float32(mouseInput.Y-float64(Window.Height)/2)
	if c.verticalAngle < -pi2 {
		c.verticalAngle = float32(-pi2 + 0.0001)
	}
	if c.verticalAngle > pi2 {
		c.verticalAngle = float32(pi2 - 0.0001)
	}
	c.horizontalAngle -= c.sensitivity * float32(mouseInput.X-float64(Window.Width)/2)
	for c.horizontalAngle < 0 {
		c.horizontalAngle += float32(2 * math.Pi)
	}
	for c.horizontalAngle > float32(2*math.Pi) {
		c.horizontalAngle -= float32(2 * math.Pi)
	}
}
