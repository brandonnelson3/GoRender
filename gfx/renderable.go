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
	Render(*shaders.ColorShader, *Frustum)
	RenderDepth(*shaders.DepthShader, *Frustum)
	RenderPointLightDepth(*shaders.PointLightShadowShader, *Frustum)
	GetBounds() (mgl32.Vec3, mgl32.Vec3) // World space Min, Max
}

// VAORenderable is a object wrapping around something that is renderable on top of a vao.
type VAORenderable struct {
	vao, vbo uint32

	Position        mgl32.Vec3
	Rotation, Scale mgl32.Mat4

	renderStyle uint32
	portions    []RenderablePortion

	// LocalMin and LocalMax are the axis-aligned bounding box in model space.
	LocalMin, LocalMax mgl32.Vec3
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

	BindVertexAttributes(Renderer.colorShader.Program())

	gl.BindVertexArray(0)

	min, max := calculateBounds(verticies)

	return &VAORenderable{
		vao:         vao,
		vbo:         vbo,
		Position:    mgl32.Vec3{},
		Rotation:    mgl32.Ident4(),
		Scale:       mgl32.Ident4(),
		renderStyle: gl.TRIANGLES,
		portions:    []RenderablePortion{{0, int32(len(verticies)), diffuse}},
		LocalMin:    min,
		LocalMax:    max,
	}
}

func calculateBounds(verts []Vertex) (mgl32.Vec3, mgl32.Vec3) {
	if len(verts) == 0 {
		return mgl32.Vec3{}, mgl32.Vec3{}
	}
	min := verts[0].Vert
	max := verts[0].Vert
	for _, v := range verts {
		for i := 0; i < 3; i++ {
			if v.Vert[i] < min[i] {
				min[i] = v.Vert[i]
			}
			if v.Vert[i] > max[i] {
				max[i] = v.Vert[i]
			}
		}
	}
	return min, max
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

	BindVertexAttributes(Renderer.colorShader.Program())

	gl.BindVertexArray(0)

	min, max := calculateBounds(verticies)

	return &VAORenderable{
		vao:         vao,
		vbo:         vbo,
		Position:    mgl32.Vec3{},
		Rotation:    mgl32.Ident4(),
		Scale:       mgl32.Ident4(),
		renderStyle: gl.TRIANGLES,
		portions:    portions,
		LocalMin:    min,
		LocalMax:    max,
	}
}

// getModelMatrix returns this renderable's final model transform matrix.
func (r *VAORenderable) getModelMatrix() mgl32.Mat4 {
	return mgl32.Translate3D(r.Position.X(), r.Position.Y(), r.Position.Z()).Mul4(r.Scale.Mul4(r.Rotation))
}

// Render bind's this renderable's VAO and draws.
func (r *VAORenderable) Render(colorShader *shaders.ColorShader, frustum *Frustum) {
	if frustum != nil {
		min, max := r.GetBounds()
		if !frustum.IsBoxIn(min, max) {
			return
		}
	}
	gl.BindVertexArray(r.vao)
	colorShader.Model.Set(r.getModelMatrix())
	for _, p := range r.portions {
		colorShader.Diffuse.Set(gl.TEXTURE0, 0, p.diffuse)
		gl.DrawArrays(r.renderStyle, p.startIndex, p.numIndex)
	}
}

// RenderDepth binds this renderable's VAO and draws for depth.
func (r *VAORenderable) RenderDepth(depthShader *shaders.DepthShader, frustum *Frustum) {
	if frustum != nil {
		min, max := r.GetBounds()
		if !frustum.IsBoxIn(min, max) {
			return
		}
	}
	gl.BindVertexArray(r.vao)
	depthShader.Model.Set(r.getModelMatrix())
	for _, p := range r.portions {
		depthShader.Diffuse.Set(gl.TEXTURE0, 0, p.diffuse)
		gl.DrawArrays(r.renderStyle, p.startIndex, p.numIndex)
	}
}

// RenderPointLightDepth binds this renderable's VAO and draws for point light shadow depth.
func (r *VAORenderable) RenderPointLightDepth(shader *shaders.PointLightShadowShader, frustum *Frustum) {
	if frustum != nil {
		min, max := r.GetBounds()
		if !frustum.IsBoxIn(min, max) {
			return
		}
	}
	gl.BindVertexArray(r.vao)
	shader.Model.Set(r.getModelMatrix())
	for _, p := range r.portions {
		shader.Diffuse.Set(gl.TEXTURE0, 0, p.diffuse)
		gl.DrawArrays(r.renderStyle, p.startIndex, p.numIndex)
	}
}

// GetBounds returns the world-space axis-aligned bounding box.
func (r *VAORenderable) GetBounds() (mgl32.Vec3, mgl32.Vec3) {
	m := r.getModelMatrix()
	corners := [8]mgl32.Vec3{
		{r.LocalMin.X(), r.LocalMin.Y(), r.LocalMin.Z()},
		{r.LocalMax.X(), r.LocalMin.Y(), r.LocalMin.Z()},
		{r.LocalMin.X(), r.LocalMax.Y(), r.LocalMin.Z()},
		{r.LocalMax.X(), r.LocalMax.Y(), r.LocalMin.Z()},
		{r.LocalMin.X(), r.LocalMin.Y(), r.LocalMax.Z()},
		{r.LocalMax.X(), r.LocalMin.Y(), r.LocalMax.Z()},
		{r.LocalMin.X(), r.LocalMax.Y(), r.LocalMax.Z()},
		{r.LocalMax.X(), r.LocalMax.Y(), r.LocalMax.Z()},
	}

	worldMin := mgl32.Vec3{1e9, 1e9, 1e9}
	worldMax := mgl32.Vec3{-1e9, -1e9, -1e9}

	for _, c := range corners {
		p := m.Mul4x1(c.Vec4(1))
		for i := 0; i < 3; i++ {
			if p[i] < worldMin[i] {
				worldMin[i] = p[i]
			}
			if p[i] > worldMax[i] {
				worldMax[i] = p[i]
			}
		}
	}

	return worldMin, worldMax
}

func (r *VAORenderable) Copy() *VAORenderable {
	temp := *r
	return &temp
}
