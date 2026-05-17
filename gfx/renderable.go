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

	// Instancing support
	InstanceTransforms []mgl32.Mat4
	instanceVBO        uint32
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

func (r *VAORenderable) setupInstancing() {
	if len(r.InstanceTransforms) == 0 || r.instanceVBO != 0 {
		return
	}

	gl.GenBuffers(1, &r.instanceVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.instanceVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(r.InstanceTransforms)*16*4, gl.Ptr(r.InstanceTransforms), gl.STATIC_DRAW)

	gl.BindVertexArray(r.vao)
	for i := uint32(0); i < 4; i++ {
		loc := uint32(3) + i
		gl.EnableVertexAttribArray(loc)
		gl.VertexAttribPointer(loc, 4, gl.FLOAT, false, 16*4, gl.PtrOffset(int(i*4*4)))
		gl.VertexAttribDivisor(loc, 1)
	}
	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}

// Render bind's this renderable's VAO and draws.
func (r *VAORenderable) Render(colorShader *shaders.ColorShader, frustum *Frustum) {
	if frustum != nil {
		min, max := r.GetBounds()
		if !frustum.IsBoxIn(min, max) {
			return
		}
	}

	isInstanced := len(r.InstanceTransforms) > 0
	if isInstanced {
		r.setupInstancing()
		colorShader.IsInstanced.Set(1)
	} else {
		colorShader.IsInstanced.Set(0)
		colorShader.Model.Set(r.getModelMatrix())
	}

	gl.BindVertexArray(r.vao)
	for _, p := range r.portions {
		colorShader.Diffuse.Set(gl.TEXTURE0, 0, p.diffuse)
		if isInstanced {
			gl.DrawArraysInstanced(r.renderStyle, p.startIndex, p.numIndex, int32(len(r.InstanceTransforms)))
		} else {
			gl.DrawArrays(r.renderStyle, p.startIndex, p.numIndex)
		}
	}

	if isInstanced {
		colorShader.IsInstanced.Set(0)
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

	isInstanced := len(r.InstanceTransforms) > 0
	if isInstanced {
		r.setupInstancing()
		depthShader.IsInstanced.Set(1)
	} else {
		depthShader.IsInstanced.Set(0)
		depthShader.Model.Set(r.getModelMatrix())
	}

	gl.BindVertexArray(r.vao)
	for _, p := range r.portions {
		depthShader.Diffuse.Set(gl.TEXTURE0, 0, p.diffuse)
		if isInstanced {
			gl.DrawArraysInstanced(r.renderStyle, p.startIndex, p.numIndex, int32(len(r.InstanceTransforms)))
		} else {
			gl.DrawArrays(r.renderStyle, p.startIndex, p.numIndex)
		}
	}

	if isInstanced {
		depthShader.IsInstanced.Set(0)
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

	isInstanced := len(r.InstanceTransforms) > 0
	if isInstanced {
		r.setupInstancing()
		shader.IsInstanced.Set(1)
	} else {
		shader.IsInstanced.Set(0)
		shader.Model.Set(r.getModelMatrix())
	}

	gl.BindVertexArray(r.vao)
	for _, p := range r.portions {
		shader.Diffuse.Set(gl.TEXTURE0, 0, p.diffuse)
		if isInstanced {
			gl.DrawArraysInstanced(r.renderStyle, p.startIndex, p.numIndex, int32(len(r.InstanceTransforms)))
		} else {
			gl.DrawArrays(r.renderStyle, p.startIndex, p.numIndex)
		}
	}

	if isInstanced {
		shader.IsInstanced.Set(0)
	}
}

// GetBounds returns the world-space axis-aligned bounding box.
func (r *VAORenderable) GetBounds() (mgl32.Vec3, mgl32.Vec3) {
	if len(r.InstanceTransforms) > 0 {
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

		for _, m := range r.InstanceTransforms {
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
		}
		return worldMin, worldMax
	}

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
	temp.instanceVBO = 0
	return &temp
}
