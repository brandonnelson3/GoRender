package terrain

import (
	perlin "github.com/aquilax/go-perlin"
	"github.com/brandonnelson3/GoRender/gfx"
	"github.com/brandonnelson3/GoRender/gfx/shaders"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	cellsize     = uint32(64)
	cellsizep1   = cellsize + 1
	cellsizep1p2 = cellsizep1 + 2
)

type cell struct {
	x, y, z uint32

	vao, vbo, veb uint32
	numIndices    int32

	verts   []gfx.Vertex
	indices []uint32
}

func (c *cell) Update(colorShader *shaders.ColorShader) {
	if c.vao == 0 {
		gl.GenVertexArrays(1, &c.vao)
		gl.BindVertexArray(c.vao)
		gl.GenBuffers(1, &c.vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, c.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(c.verts)*8*4, gl.Ptr(c.verts), gl.STATIC_DRAW)
		gfx.BindVertexAttributes(colorShader.Program())
		gl.GenBuffers(1, &c.veb)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, c.veb)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(c.indices)*4, gl.Ptr(c.indices), gl.STATIC_DRAW)
		gl.BindVertexArray(0)
	}
}

func (c *cell) Render(colorShader *shaders.ColorShader) {
	if c.vao != 0 {
		gl.BindVertexArray(c.vao)
		colorShader.Model.Set(mgl32.Translate3D(float32(c.x*cellsizep1+1), float32(c.y*cellsizep1+1), float32(c.z*cellsizep1+1)))
		gl.DrawElements(gl.TRIANGLES, c.numIndices, gl.UNSIGNED_INT, nil)
		gl.BindVertexArray(0)
	}
}

func (c *cell) RenderDepth(depthShader *shaders.DepthShader) {
	if c.vao != 0 {
		gl.BindVertexArray(c.vao)
		depthShader.Model.Set(mgl32.Translate3D(float32(c.x*cellsizep1+1), float32(c.y*cellsizep1+1), float32(c.z*cellsizep1+1)))
		gl.DrawElements(gl.TRIANGLES, c.numIndices, gl.UNSIGNED_INT, nil)
		gl.BindVertexArray(0)
	}
}

func calculateNormal(pos1, pos2, pos3 mgl32.Vec3) mgl32.Vec3 {
	a := pos2.Sub(pos1)
	b := pos3.Sub(pos1)
	return a.Cross(b)
}

func calculateIndice(x, z uint32) uint32 {
	return z*cellsizep1 + x
}

func GenerateCell(perlin *perlin.Perlin, x, y, z uint32) *cell {
	var grid [cellsizep1p2][cellsizep1p2]mgl32.Vec3
	for xi := uint32(0); xi < cellsizep1p2; xi++ {
		for zi := uint32(0); zi < cellsizep1p2; zi++ {
			h := float32((perlin.Noise2D(float64(x*cellsize+xi)/10.0, float64(z*cellsize+zi)/10.0)+1)/2.0) * 4
			grid[xi][zi] = mgl32.Vec3{float32(x*cellsize + xi), h, float32(z*cellsize + zi)}
		}
	}

	var verts []gfx.Vertex
	for xi := uint32(1); xi <= cellsizep1; xi++ {
		for zi := uint32(1); zi <= cellsizep1; zi++ {
			verts = append(verts, gfx.Vertex{grid[xi][zi], mgl32.Vec3{}, mgl32.Vec2{float32(xi) / 5.0, float32(zi) / 5.0}})
		}
	}

	var indices []uint32
	for zi := uint32(0); zi < cellsize; zi++ {
		for xi := uint32(0); xi < cellsize; xi++ {
			i1 := calculateIndice(xi, zi)
			i2 := calculateIndice(xi+1, zi)
			i3 := calculateIndice(xi, zi+1)
			i4 := calculateIndice(xi+1, zi+1)
			if xi%2 == 0 && zi%2 == 0 || xi%2 == 1 && zi%2 == 1 {
				// 1-----2
				// |   / |
				// | /   |
				// 3-----4
				indices = append(indices, i3, i1, i2, i2, i4, i3)
			} else {
				// 1-----2
				// | \   |
				// |   \ |
				// 3-----4
				indices = append(indices, i1, i2, i4, i4, i3, i1)
			}
		}
	}

	for i := 0; i < len(indices); i += 3 {
		v1 := &verts[indices[i]]
		v2 := &verts[indices[i+1]]
		v3 := &verts[indices[i+2]]

		n := calculateNormal(v1.Vert, v2.Vert, v3.Vert)

		v1.Norm = v1.Norm.Add(n)
		v2.Norm = v2.Norm.Add(n)
		v3.Norm = v3.Norm.Add(n)
	}

	for i, v := range verts {
		verts[i].Norm = v.Norm.Normalize()
	}

	return &cell{
		x:          x,
		y:          y,
		z:          z,
		verts:      verts,
		indices:    indices,
		numIndices: int32(len(indices)),
	}
}

type Terrain struct {
	c *cell

	noise *perlin.Perlin

	diffuse uint32
}

func NewTerrain() *Terrain {
	diffuseTexture, err := gfx.LoadTexture("assets/sand.png")
	if err != nil {
		panic(err)
	}

	noise := perlin.NewPerlin(2, 2, 3, int64(0))

	return &Terrain{
		c:       GenerateCell(noise, 0, 0, 0),
		noise:   noise,
		diffuse: diffuseTexture,
	}
}

func (t *Terrain) GetHeight(x, z float32) float32 {
	return float32((t.noise.Noise2D(float64(x)/10.0, float64(z)/10.0)+1)/2.0) * 4
}

func (t *Terrain) Update(colorShader *shaders.ColorShader) {
	t.c.Update(colorShader)
}

func (t *Terrain) Render(colorShader *shaders.ColorShader) {
	colorShader.Diffuse.Set(gl.TEXTURE0, 0, t.diffuse)
	t.c.Render(colorShader)
}

func (t *Terrain) RenderDepth(depthShader *shaders.DepthShader) {
	depthShader.Diffuse.Set(gl.TEXTURE0, 0, t.diffuse)
	t.c.RenderDepth(depthShader)
}
