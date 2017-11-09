package gfx

import (
	"fmt"
	"math"
	"time"

	"github.com/brandonnelson3/GoRender/input"
	"github.com/brandonnelson3/GoRender/messagebus"

	"github.com/go-gl/gl/v4.5-core/gl"
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

	redColor    = mgl32.Vec3{1, 0, 0}
	greenColor  = mgl32.Vec3{0, 1, 0}
	blueColor   = mgl32.Vec3{0, 0, 1}
	yellowColor = mgl32.Vec3{1, 1, 0}

	cascadeColors = []mgl32.Vec3{redColor, blueColor, greenColor}

	whiteColor = mgl32.Vec3{1, 1, 1}
)

type camera struct {
	position           mgl32.Vec3
	direction          mgl32.Vec3
	horizontalAngle    float32
	verticalAngle      float32
	sensitivity        float32
	speed              float32
	frustumRenderable  *Renderable
	cascadeRenderables []*Renderable
}

// InitCameras instantiates new cameras into the package first and third person package variables.
func InitCameras() {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	Renderer.lineVertexShader.BindVertexAttributes()

	FirstPerson = &camera{
		position:        mgl32.Vec3{0, 9, 0},
		horizontalAngle: 0,
		verticalAngle:   0,
		sensitivity:     0.001,
		speed:           20,
	}
	FirstPerson.frustumRenderable = &Renderable{
		vao:         vao,
		vbo:         vbo,
		renderStyle: gl.LINES,
		vertCount:   24,
	}

	ThirdPerson = &camera{
		position:        mgl32.Vec3{-10, 10, -10},
		horizontalAngle: -pi4,
		verticalAngle:   -pi4,
		sensitivity:     0.001,
		speed:           20,
	}
	ActiveCamera = FirstPerson
	messagebus.RegisterType("key", func(m *messagebus.Message) {
		direction := mgl32.Vec3{0, 0, 0}
		pressedKeys := m.Data1.([]glfw.Key)
		for _, key := range pressedKeys {
			switch key {
			case glfw.KeyW:
				direction = direction.Add(ActiveCamera.GetForward())
			case glfw.KeyS:
				direction = direction.Sub(ActiveCamera.GetForward())
			case glfw.KeyD:
				direction = direction.Add(ActiveCamera.GetRight())
			case glfw.KeyA:
				direction = direction.Sub(ActiveCamera.GetRight())
			}
		}
		ActiveCamera.direction = direction
	})
	messagebus.RegisterType("key", func(m *messagebus.Message) {
		pressedKeysThisFrame := m.Data2.([]glfw.Key)
		for _, key := range pressedKeysThisFrame {
			switch key {
			case glfw.KeyC:
				if ActiveCamera == FirstPerson {
					ActiveCamera = ThirdPerson
				} else {
					ActiveCamera = FirstPerson
				}
			}
		}
	})
	messagebus.RegisterType("mouse", func(m *messagebus.Message) {
		mouseInput := m.Data1.(input.MouseInput)
		ActiveCamera.verticalAngle -= ActiveCamera.sensitivity * float32(mouseInput.Y-float64(Window.Height)/2)
		if ActiveCamera.verticalAngle < -pi2 {
			ActiveCamera.verticalAngle = float32(-pi2 + 0.0001)
		}
		if ActiveCamera.verticalAngle > pi2 {
			ActiveCamera.verticalAngle = float32(pi2 - 0.0001)
		}
		ActiveCamera.horizontalAngle -= ActiveCamera.sensitivity * float32(mouseInput.X-float64(Window.Width)/2)
		for ActiveCamera.horizontalAngle < 0 {
			ActiveCamera.horizontalAngle += float32(2 * math.Pi)
		}
		for ActiveCamera.horizontalAngle > float32(2*math.Pi) {
			ActiveCamera.horizontalAngle -= float32(2 * math.Pi)
		}
	})

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
	if c == FirstPerson {
		cornerVerticies := []mgl32.Vec3{
			// Front
			mgl32.Vec3{-1, 1, 0},
			mgl32.Vec3{1, 1, 0},
			mgl32.Vec3{1, 1, 0},
			mgl32.Vec3{1, -1, 0},
			mgl32.Vec3{1, -1, 0},
			mgl32.Vec3{-1, -1, 0},
			mgl32.Vec3{-1, -1, 0},
			mgl32.Vec3{-1, 1, 0},

			// Back
			mgl32.Vec3{-1, 1, 1},
			mgl32.Vec3{1, 1, 1},
			mgl32.Vec3{1, 1, 1},
			mgl32.Vec3{1, -1, 1},
			mgl32.Vec3{1, -1, 1},
			mgl32.Vec3{-1, -1, 1},
			mgl32.Vec3{-1, -1, 1},
			mgl32.Vec3{-1, 1, 1},

			// Sides
			mgl32.Vec3{-1, 1, 0},
			mgl32.Vec3{-1, 1, 1},
			mgl32.Vec3{1, 1, 0},
			mgl32.Vec3{1, 1, 1},
			mgl32.Vec3{1, -1, 0},
			mgl32.Vec3{1, -1, 1},
			mgl32.Vec3{-1, -1, 0},
			mgl32.Vec3{-1, -1, 1},
		}

		verticies := []LineVertex{}

		for j := 0; j < 3; j++ {
			transform := Window.GetShadowCascadePerspectiveProjection(j).Mul4(c.GetView()).Transpose().Inv().Transpose()
			for _, v := range cornerVerticies {
				vert := transform.Mul4x1(v.Vec4(1))
				verticies = append(verticies, LineVertex{mgl32.Vec3{vert[0] / vert[3], vert[1] / vert[3], vert[2] / vert[3]}, cascadeColors[j]})
			}
		}

		transform := Window.GetProjection().Mul4(c.GetView()).Transpose().Inv().Transpose()
		for _, v := range cornerVerticies {
			vert := transform.Mul4x1(v.Vec4(1))
			verticies = append(verticies, LineVertex{mgl32.Vec3{vert[0] / vert[3], vert[1] / vert[3], vert[2] / vert[3]}, whiteColor})
		}
		c.frustumRenderable.vertCount = int32(len(verticies))

		gl.BindVertexArray(c.frustumRenderable.vao)
		gl.BindBuffer(gl.ARRAY_BUFFER, c.frustumRenderable.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(verticies)*6*4, gl.Ptr(verticies), gl.DYNAMIC_DRAW)
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

// RenderFrustum renders the frustum for this camera.
func (c *camera) RenderFrustum() {
	gl.BindVertexArray(c.frustumRenderable.vao)
	gl.DrawArrays(c.frustumRenderable.renderStyle, 0, c.frustumRenderable.vertCount-24)
}
