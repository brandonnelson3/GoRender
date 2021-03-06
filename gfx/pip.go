package gfx

import (
	"github.com/brandonnelson3/GoRender/gfx/shaders"
	"github.com/brandonnelson3/GoRender/messagebus"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	planeSquareVao uint32

	pipShader *shaders.PipShader

	enabled  = false
	depthMap *uint32
	nearFar  mgl32.Vec2
)

func InitPip() {
	var err error
	pipShader, err = shaders.NewPipShader()
	if err != nil {
		panic(err)
	}

	// Square definition
	sizex := float32(480.0)
	sizey := float32(480.0)
	padding := uint32(50)
	topLeft := mgl32.Vec2{float32(Window.Width-padding) - sizex, float32(Window.Height-padding) - sizey}
	topRight := mgl32.Vec2{float32(Window.Width - padding), float32(Window.Height-padding) - sizey}
	botLeft := mgl32.Vec2{float32(Window.Width-padding) - sizex, float32(Window.Height - padding)}
	botRight := mgl32.Vec2{float32(Window.Width - padding), float32(Window.Height - padding)}
	planeVertices := []PipVertex{
		{topLeft, mgl32.Vec2{0, 1}},
		{botRight, mgl32.Vec2{1, 0}},
		{topRight, mgl32.Vec2{1, 1}},
		{topLeft, mgl32.Vec2{0, 1}},
		{botLeft, mgl32.Vec2{0, 0}},
		{botRight, mgl32.Vec2{1, 0}},
	}
	gl.GenVertexArrays(1, &planeSquareVao)
	gl.BindVertexArray(planeSquareVao)
	var planeSquareVbo uint32
	gl.GenBuffers(1, &planeSquareVbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, planeSquareVbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(planeVertices)*4*4, gl.Ptr(planeVertices), gl.STATIC_DRAW)
	BindPipVertexAttributes(pipShader.Program())
	gl.BindVertexArray(0)

	messagebus.RegisterType("key", func(m *messagebus.Message) {
		pressedKeys := m.Data1.([]glfw.Key)
		for _, key := range pressedKeys {
			switch key {
			case glfw.KeyHome:
				enabled = true
			case glfw.KeyEnd:
				enabled = false
			}
		}
	})
}

func UpdatePip(m *uint32, nf mgl32.Vec2) {
	depthMap = m
	nearFar = nf
}

func RenderPip() {
	if enabled {
		gl.Disable(gl.DEPTH_TEST)
		pipShader.Use()
		pipShader.Projection.Set(mgl32.Ortho(0.0, float32(Window.Width), float32(Window.Height), 0.0, -1.0, 1.0))
		pipShader.DepthMap.Set(gl.TEXTURE4, 4, *depthMap)
		pipShader.NearFar.Set(nearFar)
		gl.BindVertexArray(planeSquareVao)
		gl.DrawArrays(gl.TRIANGLES, 0, 2*3)
		gl.ActiveTexture(gl.TEXTURE4)
		gl.BindTexture(gl.TEXTURE_2D, 0)
		gl.BindVertexArray(0)
		gl.Enable(gl.DEPTH_TEST)
	}
}
