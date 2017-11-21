package gfx

import (
	"github.com/brandonnelson3/GoRender/gfx/shaders"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// RenderablePortion allows rendering of a part of a vbo.
type RenderablePortion struct {
	startIndex, numIndex int32
	// TODO: This should be abstracted out to some form of "Material"
	diffuse uint32
}

type Renderable interface {
	Render(*shaders.ColorVertexShader, *shaders.ColorFragmentShader)
	RenderDepth(*shaders.DepthVertexShader, *shaders.DepthFragmentShader)
}

// VAORenderable is a object wrapping around something that is renderable on top of a vao.
type VAORenderable struct {
	vao, vbo uint32

	Position        mgl32.Vec3
	Rotation, Scale mgl32.Mat4

	renderStyle uint32
	portions    []RenderablePortion
}

// NewVAORenderable instantiates a Renderable for the given verticies of the normal Vertex Type.
func NewVAORenderable(verticies []Vertex, diffuse uint32) *VAORenderable {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(verticies)*8*4, gl.Ptr(verticies), gl.STATIC_DRAW)

	Renderer.colorVertexShader.BindVertexAttributes()

	return &VAORenderable{
		vao:         vao,
		vbo:         vbo,
		Position:    mgl32.Vec3{},
		Rotation:    mgl32.Ident4(),
		Scale:       mgl32.Ident4(),
		renderStyle: gl.TRIANGLES,
		portions:    []RenderablePortion{{0, int32(len(verticies)), diffuse}},
	}
}

// NewChunkedRenderable instantiates a Renderable for the given verticies of the normal Vertex Type.
func NewChunkedRenderable(verticies []Vertex, portions []RenderablePortion) *VAORenderable {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(verticies)*8*4, gl.Ptr(verticies), gl.STATIC_DRAW)

	Renderer.colorVertexShader.BindVertexAttributes()

	return &VAORenderable{
		vao:         vao,
		vbo:         vbo,
		Position:    mgl32.Vec3{},
		Rotation:    mgl32.Ident4(),
		Scale:       mgl32.Ident4(),
		renderStyle: gl.TRIANGLES,
		portions:    portions,
	}
}

// getModelMatrix returns this renderable's final model transform matrix.
func (r *VAORenderable) getModelMatrix() mgl32.Mat4 {
	return mgl32.Translate3D(r.Position.X(), r.Position.Y(), r.Position.Z()).Mul4(r.Scale.Mul4(r.Rotation))
}

// Render bind's this renderable's VAO and draws.
func (r *VAORenderable) Render(vertexShader *shaders.ColorVertexShader, fragmentShader *shaders.ColorFragmentShader) {
	gl.BindVertexArray(r.vao)
	vertexShader.Model.Set(r.getModelMatrix())
	for _, p := range r.portions {
		fragmentShader.Diffuse.Set(gl.TEXTURE0, 0, p.diffuse)
		gl.DrawArrays(r.renderStyle, p.startIndex, p.numIndex)
	}
}

// Render bind's this renderable's VAO and draws for depth.
func (r *VAORenderable) RenderDepth(vertexShader *shaders.DepthVertexShader, fragmentShader *shaders.DepthFragmentShader) {
	gl.BindVertexArray(r.vao)
	vertexShader.Model.Set(r.getModelMatrix())
	for _, p := range r.portions {
		fragmentShader.Diffuse.Set(gl.TEXTURE0, 0, p.diffuse)
		gl.DrawArrays(r.renderStyle, p.startIndex, p.numIndex)
	}
}

func (r *VAORenderable) Copy() *VAORenderable {
	temp := *r
	return &temp
}

// PlaneVertices is the vertex list for a Plane.
var PlaneVertices = []Vertex{
	{mgl32.Vec3{-10000.0, 0, -10000.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{0, 0}},
	{mgl32.Vec3{-10000.0, 0, 10000.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{0, 500}},
	{mgl32.Vec3{10000.0, 0, -10000.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{500, 0}},
	{mgl32.Vec3{-10000.0, 0, 10000.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{0, 500}},
	{mgl32.Vec3{10000.0, 0, 10000.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{500, 500}},
	{mgl32.Vec3{10000.0, 0, -10000.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{500, 0}},
}

// CubeVertices is the vertex list for a Cube.
var CubeVertices = []Vertex{
	//  X, Y, Z
	// Bottom
	{mgl32.Vec3{-1.0, -1.0, -1.0}, mgl32.Vec3{0, -1.0, 0}, mgl32.Vec2{0, 0}},
	{mgl32.Vec3{1.0, -1.0, -1.0}, mgl32.Vec3{0, -1.0, 0}, mgl32.Vec2{1, 0}},
	{mgl32.Vec3{-1.0, -1.0, 1.0}, mgl32.Vec3{0, -1.0, 0}, mgl32.Vec2{0, 1}},
	{mgl32.Vec3{1.0, -1.0, -1.0}, mgl32.Vec3{0, -1.0, 0}, mgl32.Vec2{1, 0}},
	{mgl32.Vec3{1.0, -1.0, 1.0}, mgl32.Vec3{0, -1.0, 0}, mgl32.Vec2{1, 1}},
	{mgl32.Vec3{-1.0, -1.0, 1.0}, mgl32.Vec3{0, -1.0, 0}, mgl32.Vec2{0, 1}},

	// Top
	{mgl32.Vec3{-1.0, 1.0, -1.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{0, 0}},
	{mgl32.Vec3{-1.0, 1.0, 1.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{0, 1}},
	{mgl32.Vec3{1.0, 1.0, -1.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{1, 0}},
	{mgl32.Vec3{1.0, 1.0, -1.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{1, 0}},
	{mgl32.Vec3{-1.0, 1.0, 1.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{0, 1}},
	{mgl32.Vec3{1.0, 1.0, 1.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{1, 1}},

	// Front
	{mgl32.Vec3{-1.0, -1.0, 1.0}, mgl32.Vec3{0, 0, 1.0}, mgl32.Vec2{1, 0}},
	{mgl32.Vec3{1.0, -1.0, 1.0}, mgl32.Vec3{0, 0, 1.0}, mgl32.Vec2{0, 0}},
	{mgl32.Vec3{-1.0, 1.0, 1.0}, mgl32.Vec3{0, 0, 1.0}, mgl32.Vec2{1, 1}},
	{mgl32.Vec3{1.0, -1.0, 1.0}, mgl32.Vec3{0, 0, 1.0}, mgl32.Vec2{0, 0}},
	{mgl32.Vec3{1.0, 1.0, 1.0}, mgl32.Vec3{0, 0, 1.0}, mgl32.Vec2{0, 1}},
	{mgl32.Vec3{-1.0, 1.0, 1.0}, mgl32.Vec3{0, 0, 1.0}, mgl32.Vec2{1, 1}},

	// Back
	{mgl32.Vec3{-1.0, -1.0, -1.0}, mgl32.Vec3{0, 0, -1.0}, mgl32.Vec2{0, 0}},
	{mgl32.Vec3{-1.0, 1.0, -1.0}, mgl32.Vec3{0, 0, -1.0}, mgl32.Vec2{0, 1}},
	{mgl32.Vec3{1.0, -1.0, -1.0}, mgl32.Vec3{0, 0, -1.0}, mgl32.Vec2{1, 0}},
	{mgl32.Vec3{1.0, -1.0, -1.0}, mgl32.Vec3{0, 0, -1.0}, mgl32.Vec2{1, 0}},
	{mgl32.Vec3{-1.0, 1.0, -1.0}, mgl32.Vec3{0, 0, -1.0}, mgl32.Vec2{0, 1}},
	{mgl32.Vec3{1.0, 1.0, -1.0}, mgl32.Vec3{0, 0, -1.0}, mgl32.Vec2{1, 1}},

	// Left
	{mgl32.Vec3{-1.0, -1.0, 1.0}, mgl32.Vec3{-1.0, 0, 0}, mgl32.Vec2{0, 1}},
	{mgl32.Vec3{-1.0, 1.0, -1.0}, mgl32.Vec3{-1.0, 0, 0}, mgl32.Vec2{1, 0}},
	{mgl32.Vec3{-1.0, -1.0, -1.0}, mgl32.Vec3{-1.0, 0, 0}, mgl32.Vec2{0, 0}},
	{mgl32.Vec3{-1.0, -1.0, 1.0}, mgl32.Vec3{-1.0, 0, 0}, mgl32.Vec2{0, 1}},
	{mgl32.Vec3{-1.0, 1.0, 1.0}, mgl32.Vec3{-1.0, 0, 0}, mgl32.Vec2{1, 1}},
	{mgl32.Vec3{-1.0, 1.0, -1.0}, mgl32.Vec3{-1.0, 0, 0}, mgl32.Vec2{1, 0}},

	// Right
	{mgl32.Vec3{1.0, -1.0, 1.0}, mgl32.Vec3{1.0, 0, 0}, mgl32.Vec2{1, 1}},
	{mgl32.Vec3{1.0, -1.0, -1.0}, mgl32.Vec3{1.0, 0, 0}, mgl32.Vec2{1, 0}},
	{mgl32.Vec3{1.0, 1.0, -1.0}, mgl32.Vec3{1.0, 0, 0}, mgl32.Vec2{0, 0}},
	{mgl32.Vec3{1.0, -1.0, 1.0}, mgl32.Vec3{1.0, 0, 0}, mgl32.Vec2{1, 1}},
	{mgl32.Vec3{1.0, 1.0, -1.0}, mgl32.Vec3{1.0, 0, 0}, mgl32.Vec2{0, 0}},
	{mgl32.Vec3{1.0, 1.0, 1.0}, mgl32.Vec3{1.0, 0, 0}, mgl32.Vec2{0, 1}},
}
