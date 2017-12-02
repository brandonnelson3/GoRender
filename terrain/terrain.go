package terrain

import (
	"sync"
	"time"

	perlin "github.com/aquilax/go-perlin"
	"github.com/brandonnelson3/GoRender/gfx"
	"github.com/brandonnelson3/GoRender/gfx/shaders"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	cellsize     = int32(64)
	cellsizep1   = cellsize + 1
	cellsizep1p2 = cellsizep1 + 2

	worldSize   = 6
	worldSizem1 = worldSize - 1
	worldSizet2 = worldSize * 2
	worldTotal  = (2 * worldSize) * (2 * worldSize) * (2 * worldSize)
)

var (
	halfCell = mgl32.Vec3{float32(cellsize) / 2.0, 0, float32(cellsize) / 2.0}
)

type cellId struct {
	x, z int32
}

func (lhs *cellId) Equal(rhs cellId) bool {
	return lhs.x == rhs.x && lhs.z == rhs.z
}

type cell struct {
	id cellId

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
		colorShader.Model.Set(mgl32.Translate3D(float32(c.id.x*cellsize+1), 0, float32(c.id.z*cellsize+1)))
		gl.DrawElements(gl.TRIANGLES, c.numIndices, gl.UNSIGNED_INT, nil)
		gl.BindVertexArray(0)
	}
}

func (c *cell) RenderDepth(depthShader *shaders.DepthShader) {
	if c.vao != 0 {
		gl.BindVertexArray(c.vao)
		depthShader.Model.Set(mgl32.Translate3D(float32(c.id.x*cellsize+1), 0, float32(c.id.z*cellsize+1)))
		gl.DrawElements(gl.TRIANGLES, c.numIndices, gl.UNSIGNED_INT, nil)
		gl.BindVertexArray(0)
	}
}

type Terrain struct {
	mu   sync.Mutex
	data map[cellId]*cell

	noise *perlin.Perlin

	diffuse uint32
}

func NewTerrain() *Terrain {
	diffuseTexture, err := gfx.LoadTexture("assets/sand.png")
	if err != nil {
		panic(err)
	}

	t := &Terrain{
		data:    make(map[cellId]*cell),
		noise:   perlin.NewPerlin(2, 2, 3, int64(0)),
		diffuse: diffuseTexture,
	}

	for x := int32(1 - worldSize); x <= worldSize; x++ {
		for y := int32(1 - worldSize); y <= worldSize; y++ {
			for z := int32(1 - worldSize); z <= worldSize; z++ {
				go t.generate(x, z)
			}
		}
	}

	return t
}

func calculateNormal(pos1, pos2, pos3 mgl32.Vec3) mgl32.Vec3 {
	a := pos2.Sub(pos1)
	b := pos3.Sub(pos1)
	return a.Cross(b)
}

func calculateIndice(x, z uint32) uint32 {
	return z*uint32(cellsizep1) + x
}

func isCellInWorld(cell, centroidCell cellId) bool {
	if cell.x < centroidCell.x-worldSizem1 {
		return false
	}
	if cell.z < centroidCell.z-worldSizem1 {
		return false
	}

	if cell.x > centroidCell.x+worldSize {
		return false
	}
	if cell.z > centroidCell.z+worldSize {
		return false
	}
	return true
}

func (t *Terrain) GenerateCell(id cellId) *cell {
	var grid [cellsizep1p2][cellsizep1p2]mgl32.Vec3
	for x := int32(0); x < cellsizep1p2; x++ {
		for z := int32(0); z < cellsizep1p2; z++ {
			h := float32((t.noise.Noise2D(float64(id.x*cellsize+x)/100.0, float64(id.z*cellsize+z)/100.0)+1)/2.0) * 50
			grid[x][z] = mgl32.Vec3{float32(x), h, float32(z)}
		}
	}

	var verts []gfx.Vertex
	for x := int32(1); x <= cellsizep1; x++ {
		for z := int32(1); z <= cellsizep1; z++ {
			verts = append(verts, gfx.Vertex{grid[x][z], mgl32.Vec3{}, mgl32.Vec2{float32(x) / 5.0, float32(z) / 5.0}})
		}
	}

	var indices []uint32
	for z := uint32(0); z < uint32(cellsize); z++ {
		for x := uint32(0); x < uint32(cellsize); x++ {
			i1 := calculateIndice(x, z)
			i2 := calculateIndice(x+1, z)
			i3 := calculateIndice(x, z+1)
			i4 := calculateIndice(x+1, z+1)
			if x%2 == 0 && z%2 == 0 || x%2 == 1 && z%2 == 1 {
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
		id:         id,
		verts:      verts,
		indices:    indices,
		numIndices: int32(len(indices)),
	}
}

func (t *Terrain) generate(x, z int32) {
	lastCell := cellId{-1, -1}
	for {
		// No point in checking more often then every 100ms.
		<-time.After(100 * time.Millisecond)

		// Positions are shifted by half a cell from cell positions since cell positions are in the lower left corner.
		pos := gfx.FirstPerson.GetPosition().Sub(halfCell)

		// If this is the same cell as last iteration bail.
		thisCell := cellId{int32(pos.X())/cellsize + x, int32(pos.Z())/cellsize + z}
		if lastCell.Equal(thisCell) {
			continue
		}
		lastCell = thisCell

		// If this cell is already present in the world bail.
		t.mu.Lock()
		_, newOk := t.data[thisCell]
		t.mu.Unlock()
		if newOk {
			continue
		}

		// This is a new cell not currently present in the world. Generate then insert.
		c := t.GenerateCell(thisCell)
		t.mu.Lock()
		t.data[thisCell] = c
		t.mu.Unlock()
	}
}

func (t *Terrain) GetHeight(x, z float32) float32 {
	return float32((t.noise.Noise2D(float64(x)/100.0, float64(z)/100.0)+1)/2.0) * 50
}

func (t *Terrain) Update(colorShader *shaders.ColorShader) {
	pos := gfx.FirstPerson.GetPosition()
	pos = pos.Sub(halfCell)
	centroidCell := cellId{int32(pos.X()) / cellsize, int32(pos.Z()) / cellsize}

	t.mu.Lock()
	defer t.mu.Unlock()
	for _, c := range t.data {
		if !isCellInWorld(c.id, centroidCell) {
			delete(t.data, c.id)
			continue
		}
		c.Update(colorShader)
	}
}

func (t *Terrain) Render(colorShader *shaders.ColorShader) {
	colorShader.Diffuse.Set(gl.TEXTURE0, 0, t.diffuse)

	t.mu.Lock()
	defer t.mu.Unlock()
	for _, c := range t.data {
		c.Render(colorShader)
	}
}

func (t *Terrain) RenderDepth(depthShader *shaders.DepthShader) {
	depthShader.Diffuse.Set(gl.TEXTURE0, 0, t.diffuse)

	t.mu.Lock()
	defer t.mu.Unlock()
	for _, c := range t.data {
		c.RenderDepth(depthShader)
	}
}
