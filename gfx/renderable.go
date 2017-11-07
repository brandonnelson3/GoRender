package gfx

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Renderable is a object wrapping around something that is renderable. This struct contains all of the settings and elements needed to render something.
type Renderable struct {
	vao uint32

	Position        mgl32.Vec3
	Rotation, Scale mgl32.Mat4

	renderStyle uint32
	vertCount   int32

	// This renderable will choose not to draw if this is false.
	enabled bool
}

// NewRenderable instantiates a Renderable for the given verticies.
func NewRenderable(verticies []Vertex) *Renderable {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(verticies)*8*4, gl.Ptr(verticies), gl.STATIC_DRAW)

	Renderer.colorVertexShader.BindVertexAttributes()

	return &Renderable{
		vao:         vao,
		Position:    mgl32.Vec3{0, 0, 0},
		Rotation:    mgl32.Ident4(),
		Scale:       mgl32.Ident4(),
		renderStyle: gl.TRIANGLES,
		vertCount:   int32(len(verticies)),
		enabled:     true,
	}
}

// GetModelMatrix returns this renderable's final model transform matrix.
func (r *Renderable) GetModelMatrix() mgl32.Mat4 {
	return r.Scale.Mul4(r.Rotation).Mul4(mgl32.Translate3D(r.Position.X(), r.Position.Y(), r.Position.Z()))
}

// Render bind's this renderable's VAO and draws.
func (r *Renderable) Render() {
	if r.enabled {
		gl.BindVertexArray(r.vao)
		gl.DrawArrays(r.renderStyle, 0, r.vertCount)
	}
}

// PlaneVertices is the vertex list for a Plane.
var PlaneVertices = []Vertex{
	{mgl32.Vec3{-1000.0, 0, -1000.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{0, 0}},
	{mgl32.Vec3{1000.0, 0, -1000.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{0, 50}},
	{mgl32.Vec3{-1000.0, 0, 1000.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{50, 0}},
	{mgl32.Vec3{1000.0, 0, -1000.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{0, 50}},
	{mgl32.Vec3{1000.0, 0, 1000.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{50, 50}},
	{mgl32.Vec3{-1000.0, 0, 1000.0}, mgl32.Vec3{0, 1.0, 0}, mgl32.Vec2{50, 0}},
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
