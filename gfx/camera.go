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

	cascadeColors = []mgl32.Vec3{redColor, greenColor, blueColor, yellowColor}

	whiteColor = mgl32.Vec3{1, 1, 1}
)

type camera struct {
	position        mgl32.Vec3
	direction       mgl32.Vec3
	horizontalAngle float32
	verticalAngle   float32
	sensitivity     float32
	speed           float32
	shadowMatrices  [NumberOfCascades]mgl32.Mat4

	// Frustum rendering done internally without Renderable.
	vao, vbo                       uint32
	renderFrustum                  bool
	renderCascade1                 bool
	renderCascade2                 bool
	renderCascade3                 bool
	renderCascadeCenters           bool
	renderCascade1ShadowFrustum    bool
	renderCascade2ShadowFrustum    bool
	renderCascade3ShadowFrustum    bool
	renderCascadeShadowFrustumEyes bool
}

// InitCameras instantiates new cameras into the package first and third person package variables.
func InitCameras() {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	BindLineVertexAttributes(Renderer.lineShader.Program())
	gl.BindVertexArray(0)

	FirstPerson = &camera{
		position:                       mgl32.Vec3{0, 40, 0},
		horizontalAngle:                0,
		verticalAngle:                  0,
		sensitivity:                    0.001,
		speed:                          20,
		vao:                            vao,
		vbo:                            vbo,
		renderCascade1:                 false,
		renderCascade2:                 false,
		renderCascade3:                 false,
		renderCascadeCenters:           false,
		renderFrustum:                  false,
		renderCascade1ShadowFrustum:    false,
		renderCascade2ShadowFrustum:    false,
		renderCascade3ShadowFrustum:    false,
		renderCascadeShadowFrustumEyes: false,
	}

	ThirdPerson = &camera{
		position:        mgl32.Vec3{-10, 60, -10},
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
			case glfw.KeyKP0:
				FirstPerson.renderFrustum = !FirstPerson.renderFrustum
			case glfw.KeyKP1:
				FirstPerson.renderCascade1 = !FirstPerson.renderCascade1
			case glfw.KeyKP2:
				FirstPerson.renderCascade2 = !FirstPerson.renderCascade2
			case glfw.KeyKP3:
				FirstPerson.renderCascade3 = !FirstPerson.renderCascade3
			case glfw.KeyKPDecimal:
				FirstPerson.renderCascadeCenters = !FirstPerson.renderCascadeCenters
			case glfw.KeyKP4:
				FirstPerson.renderCascade1ShadowFrustum = !FirstPerson.renderCascade1ShadowFrustum
			case glfw.KeyKP5:
				FirstPerson.renderCascade2ShadowFrustum = !FirstPerson.renderCascade2ShadowFrustum
			case glfw.KeyKP6:
				FirstPerson.renderCascade3ShadowFrustum = !FirstPerson.renderCascade3ShadowFrustum
			case glfw.KeyKPAdd:
				FirstPerson.renderCascadeShadowFrustumEyes = !FirstPerson.renderCascadeShadowFrustumEyes
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

func transform(p mgl32.Vec3, m mgl32.Mat4) mgl32.Vec3 {
	p1 := m.Mul4x1(p.Vec4(1))
	return mgl32.Vec3{p1.X() / p1.W(), p1.Y() / p1.W(), p1.Z() / p1.W()}
}

func transformTransposed(p mgl32.Vec3, m mgl32.Mat4) mgl32.Vec3 {
	p1 := m.Transpose().Mul4x1(p.Vec4(1))
	return mgl32.Vec3{p1.X() / p1.W(), p1.Y() / p1.W(), p1.Z() / p1.W()}
}

// Update is called every frame to execute this frame's movement.
func (c *camera) Update(d float64) {
	if c.direction.X() != 0 || c.direction.Y() != 0 || c.direction.Z() != 0 {
		delta := c.direction.Normalize().Mul(float32(d) * c.speed)
		c.position = c.position.Add(delta)
		c.direction = mgl32.Vec3{0, 0, 0}
	}
	if c == FirstPerson {
		cornerVertices := []mgl32.Vec3{
			mgl32.Vec3{-1, 1, -1},
			mgl32.Vec3{1, 1, -1},
			mgl32.Vec3{1, -1, -1},
			mgl32.Vec3{-1, -1, -1},
			mgl32.Vec3{-1, 1, 1},
			mgl32.Vec3{1, 1, 1},
			mgl32.Vec3{1, -1, 1},
			mgl32.Vec3{-1, -1, 1},
		}

		lineIndices := []int{
			// Near
			0, 1, 1, 2, 2, 3, 3, 0,
			// Far
			4, 5, 5, 6, 6, 7, 7, 4,
			// Sides
			0, 4, 1, 5, 2, 6, 3, 7,
		}

		vertices := []LineVertex{}
		for j := 0; j < NumberOfCascades; j++ {
			lightViewProjection := Window.GetShadowCascadePerspectiveProjection(j).Mul4(c.GetView()).Transpose().Inv()

			cascadeCornerVertices := [8]mgl32.Vec3{}
			cascadeCenter := mgl32.Vec3{}
			for i, v := range cornerVertices {
				cascadeCornerVertices[i] = transformTransposed(v, lightViewProjection)
				cascadeCenter = cascadeCenter.Add(cascadeCornerVertices[i])
			}
			cascadeCenter = cascadeCenter.Mul(.125)

			// Note this is using the second "value" of the lineIndices index.
			for _, i := range lineIndices {
				if i < 4 {
					vertices = append(vertices, LineVertex{cascadeCornerVertices[i], whiteColor})
				} else {
					vertices = append(vertices, LineVertex{cascadeCornerVertices[i], cascadeColors[j]})
				}
			}
			vertices = append(vertices, LineVertex{cascadeCenter, cascadeColors[j]})
			radius := cascadeCornerVertices[0].Sub(cascadeCornerVertices[6]).Len() / 2.0
			texelsPerUnit := float32(shadowMapSize) / (radius * 2.0)
			scalar := mgl32.Scale3D(texelsPerUnit, texelsPerUnit, texelsPerUnit)
			lookat := mgl32.LookAtV(mgl32.Vec3{0, 0, 0}, GetDirectionalLightDirection().Mul(-1), mgl32.Vec3{0, 1, 0}).Mul4(scalar)
			lookatInv := lookat.Inv()

			cascadeCenter = transform(cascadeCenter, lookat)
			cascadeCenter = mgl32.Vec3{
				float32(math.Floor(float64(cascadeCenter.X()))),
				float32(math.Floor(float64(cascadeCenter.Y()))),
				cascadeCenter.Z(),
			}
			cascadeCenter = transform(cascadeCenter, lookatInv)

			eye := cascadeCenter.Sub(GetDirectionalLightDirection())
			vertices = append(vertices, LineVertex{eye, yellowColor})

			lightViewMatrix := mgl32.LookAtV(eye, cascadeCenter, mgl32.Vec3{0, 1, 0})
			frustumOrthoMatrix := mgl32.Ortho(-radius, radius, -radius, radius, -6*radius, 6*radius)

			c.shadowMatrices[j] = frustumOrthoMatrix.Mul4(lightViewMatrix)
			lightViewProjection = c.shadowMatrices[j].Transpose().Inv()

			// Shadow Frustum Vert Calculation
			cascadeCornerVertices = [8]mgl32.Vec3{}
			for i, v := range cornerVertices {
				cascadeCornerVertices[i] = transformTransposed(v, lightViewProjection)
			}

			// Note this is using the second "value" of the lineIndices index.
			for _, i := range lineIndices {
				if i < 4 {
					vertices = append(vertices, LineVertex{cascadeCornerVertices[i], whiteColor})
				} else {
					vertices = append(vertices, LineVertex{cascadeCornerVertices[i], cascadeColors[j]})
				}
			}
		}

		lightViewProjection := Window.GetProjection().Mul4(c.GetView()).Transpose().Inv().Transpose()
		cascadeCornerVertices := [8]mgl32.Vec3{}
		for i, v := range cornerVertices {
			temp := lightViewProjection.Mul4x1(v.Vec4(1))
			cascadeCornerVertices[i] = mgl32.Vec3{temp.X() / temp.W(), temp.Y() / temp.W(), temp.Z() / temp.W()}
		}
		// Note this is using the second "value" of the lineIndices index.
		for _, i := range lineIndices {
			vertices = append(vertices, LineVertex{cascadeCornerVertices[i], whiteColor})
		}
		gl.BindVertexArray(c.vao)
		gl.BindBuffer(gl.ARRAY_BUFFER, c.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*6*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)
		gl.BindVertexArray(0)
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
	gl.BindVertexArray(c.vao)
	if c.renderCascade1 {
		gl.DrawArrays(gl.LINES, 0, 24)
		if c.renderCascadeCenters {
			gl.DrawArrays(gl.POINTS, 24, 1)
		}
	}
	if c.renderCascade1ShadowFrustum {
		if c.renderCascadeShadowFrustumEyes {
			gl.DrawArrays(gl.POINTS, 25, 1)
		}
		gl.DrawArrays(gl.LINES, 26, 24)
	}

	if c.renderCascade2 {
		gl.DrawArrays(gl.LINES, 50, 24)
		if c.renderCascadeCenters {
			gl.DrawArrays(gl.POINTS, 74, 1)
		}
	}
	if c.renderCascade2ShadowFrustum {
		if c.renderCascadeShadowFrustumEyes {
			gl.DrawArrays(gl.POINTS, 75, 1)
		}
		gl.DrawArrays(gl.LINES, 76, 24)
	}

	if c.renderCascade3 {
		gl.DrawArrays(gl.LINES, 100, 24)
		if c.renderCascadeCenters {
			gl.DrawArrays(gl.POINTS, 124, 1)
		}
	}
	if c.renderCascade3ShadowFrustum {
		if c.renderCascadeShadowFrustumEyes {
			gl.DrawArrays(gl.POINTS, 125, 1)
		}
		gl.DrawArrays(gl.LINES, 126, 24)
	}

	if c.renderFrustum {
		gl.DrawArrays(gl.LINES, 150, 24)
	}

	gl.BindVertexArray(0)
}
