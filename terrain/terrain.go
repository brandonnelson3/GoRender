package terrain

import (
	perlin "github.com/aquilax/go-perlin"
	"github.com/brandonnelson3/GoRender/gfx"
	"github.com/brandonnelson3/GoRender/gfx/shaders"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	cellsize     = 64
	cellsizep1   = cellsize + 1
	cellsizep1p2 = cellsizep1 + 2
)

type cell struct {
	x, y, z int

	vao, vbo uint32
	numVerts int32

	verts []gfx.Vertex
}

func (c *cell) Update(depthVertexShader *shaders.DepthVertexShader, colorVertexShader *shaders.ColorVertexShader) {
	if c.vao == 0 {
		gl.GenVertexArrays(1, &c.vao)
		gl.BindVertexArray(c.vao)
		gl.GenBuffers(1, &c.vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, c.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(c.verts)*8*4, gl.Ptr(c.verts), gl.STATIC_DRAW)
		colorVertexShader.BindVertexAttributes()
		depthVertexShader.BindVertexAttributes()
	}
}

func (c *cell) Render(vertexShader *shaders.ColorVertexShader, fragmentShader *shaders.ColorFragmentShader) {
	if c.vao != 0 {
		gl.BindVertexArray(c.vao)
		vertexShader.Model.Set(mgl32.Ident4())
		gl.DrawArrays(gl.TRIANGLES, 0, c.numVerts)
	}
}

func (c *cell) RenderDepth(vertexShader *shaders.DepthVertexShader, fragmentShader *shaders.DepthFragmentShader) {
	if c.vao != 0 {
		gl.BindVertexArray(c.vao)
		vertexShader.Model.Set(mgl32.Ident4())
		gl.DrawArrays(gl.TRIANGLES, 0, c.numVerts)
	}
}

func calculateNormal(pos1, pos2, pos3 mgl32.Vec3) mgl32.Vec3 {
	a := pos2.Sub(pos1)
	b := pos3.Sub(pos1)
	return a.Cross(b)
}

func GenerateCell(perlin *perlin.Perlin, x, y, z int) *cell {
	var grid [cellsizep1p2][cellsizep1p2]mgl32.Vec3
	for xi := 0; xi < cellsizep1p2; xi++ {
		for zi := 0; zi < cellsizep1p2; zi++ {
			h := float32((perlin.Noise2D(float64(x*cellsize+xi)/10.0, float64(z*cellsize+zi)/10.0)+1)/2.0) * 4
			grid[xi][zi] = mgl32.Vec3{float32(x*cellsize + xi), h, float32(z*cellsize + zi)}
		}
	}

	var verts []gfx.Vertex
	for xi := 2; xi <= cellsizep1; xi++ {
		for zi := 2; zi <= cellsizep1; zi++ {
			vec00 := grid[xi][zi]
			vec01 := grid[xi][zi-1]
			vec10 := grid[xi-1][zi]
			vec11 := grid[xi-1][zi-1]

			if xi%2 == 0 && zi%2 == 0 || xi%2 == 1 && zi%2 == 1 {

				normal1 := calculateNormal(vec00, vec01, vec10)
				normal2 := calculateNormal(vec10, vec01, vec11)

				verts = append(verts,
					gfx.Vertex{vec00, normal1, mgl32.Vec2{0.0, 0.0}},
					gfx.Vertex{vec01, normal1, mgl32.Vec2{0.0, 1.0}},
					gfx.Vertex{vec10, normal1, mgl32.Vec2{1.0, 0.0}},
					gfx.Vertex{vec10, normal2, mgl32.Vec2{1.0, 0.0}},
					gfx.Vertex{vec01, normal2, mgl32.Vec2{0.0, 1.0}},
					gfx.Vertex{vec11, normal2, mgl32.Vec2{1.0, 1.0}},
				)
			} else {
				normal1 := calculateNormal(vec00, vec11, vec10)
				normal2 := calculateNormal(vec00, vec01, vec11)

				verts = append(verts,
					gfx.Vertex{vec00, normal1, mgl32.Vec2{0.0, 0.0}},
					gfx.Vertex{vec11, normal1, mgl32.Vec2{1.0, 1.0}},
					gfx.Vertex{vec10, normal1, mgl32.Vec2{1.0, 0.0}},
					gfx.Vertex{vec00, normal2, mgl32.Vec2{0.0, 0.0}},
					gfx.Vertex{vec01, normal2, mgl32.Vec2{0.0, 1.0}},
					gfx.Vertex{vec11, normal2, mgl32.Vec2{1.0, 1.0}},
				)
			}
		}
	}

	return &cell{
		x:        x,
		y:        y,
		z:        z,
		verts:    verts,
		numVerts: int32(len(verts)),
	}
}

type Terrain struct {
	c *cell

	noise *perlin.Perlin

	diffuse uint32
}

func NewTerrain() *Terrain {
	diffuseTexture, err := gfx.LoadTexture("assets/crate1_diffuse.png")
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

func (t *Terrain) Update(depthVertexShader *shaders.DepthVertexShader, colorVertexShader *shaders.ColorVertexShader) {
	t.c.Update(depthVertexShader, colorVertexShader)
}

func (t *Terrain) Render(vertexShader *shaders.ColorVertexShader, fragmentShader *shaders.ColorFragmentShader) {
	fragmentShader.Diffuse.Set(gl.TEXTURE0, 0, t.diffuse)
	t.c.Render(vertexShader, fragmentShader)
}

func (t *Terrain) RenderDepth(vertexShader *shaders.DepthVertexShader, fragmentShader *shaders.DepthFragmentShader) {
	fragmentShader.Diffuse.Set(gl.TEXTURE0, 0, t.diffuse)
	t.c.RenderDepth(vertexShader, fragmentShader)
}
