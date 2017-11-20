package gfx

import (
	"github.com/brandonnelson3/GoRender/gfx/shaders"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Sky struct {
	pipeline, vao, vbo uint32

	vertexShader   *shaders.SkyVertexShader
	fragmentShader *shaders.SkyFragmentShader
}

func NewSky() (*Sky, error) {
	vertexShader, err := shaders.NewSkyVertexShader()
	if err != nil {
		return nil, err
	}
	fragmentShader, err := shaders.NewSkyFragmentShader()
	if err != nil {
		return nil, err
	}

	var pipeline uint32
	gl.CreateProgramPipelines(1, &pipeline)
	vertexShader.AddToPipeline(pipeline)
	fragmentShader.AddToPipeline(pipeline)
	gl.ValidateProgramPipeline(pipeline)
	gl.UseProgram(0)
	gl.BindProgramPipeline(pipeline)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	vertexShader.BindVertexAttributes()

	return &Sky{
		pipeline:       pipeline,
		vao:            vao,
		vbo:            vbo,
		vertexShader:   vertexShader,
		fragmentShader: fragmentShader,
	}, nil
}

// Render renders the sky. This should always be rendered before any other part of the scene.
func (sky *Sky) Render() {
	gl.Disable(gl.DEPTH_TEST)
	gl.BindProgramPipeline(sky.pipeline)
	view := mgl32.LookAtV(mgl32.Vec3{}, ActiveCamera.GetForward(), mgl32.Vec3{0, 1, 0})
	sky.vertexShader.View.Set(view)
	sky.vertexShader.Projection.Set(Window.GetProjection())
	sky.fragmentShader.DirectionalLightBuffer.Set(GetDirectionalLightBuffer())
	gl.BindVertexArray(sky.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, 2*3)
	gl.Enable(gl.DEPTH_TEST)
}

func (sky *Sky) Update() {
	vertices := []mgl32.Vec3{
		mgl32.Vec3{-1, 1, .1},
		mgl32.Vec3{1, -1, .1},
		mgl32.Vec3{1, 1, .1},
		mgl32.Vec3{-1, -1, .1},
		mgl32.Vec3{1, -1, .1},
		mgl32.Vec3{-1, 1, .1},
	}

	view := mgl32.LookAtV(mgl32.Vec3{}, ActiveCamera.GetForward(), mgl32.Vec3{0, 1, 0})
	transform := Window.GetProjection().Mul4(view).Transpose().Inv()
	for i, v := range vertices {
		vertices[i] = transformTransposed(v, transform)
	}

	gl.BindVertexArray(sky.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, sky.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*3*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)
}
