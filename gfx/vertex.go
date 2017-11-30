package gfx

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Vertex is a Vertex.
type Vertex struct {
	Vert, Norm mgl32.Vec3
	UV         mgl32.Vec2
}

// BindVertexAttributes binds the attributes per vertex.
func BindVertexAttributes(s uint32) {
	vertAttrib := uint32(gl.GetAttribLocation(s, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))
	normAttrib := uint32(gl.GetAttribLocation(s, gl.Str("norm\x00")))
	gl.EnableVertexAttribArray(normAttrib)
	gl.VertexAttribPointer(normAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(12))
	uvAttrib := uint32(gl.GetAttribLocation(s, gl.Str("uv\x00")))
	gl.EnableVertexAttribArray(uvAttrib)
	gl.VertexAttribPointer(uvAttrib, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(24))
}

// SkyVertex is a Vertex for the sky.
type SkyVertex struct {
	Vert mgl32.Vec2
}

// BindSkyVertexAttributes binds the attributes per vertex.
func BindSkyVertexAttributes(s uint32) {
	vertAttrib := uint32(gl.GetAttribLocation(s, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
}

// LineVertex is a Vertex.
type LineVertex struct {
	Vert, Color mgl32.Vec3
}

// BindLineVertexAttributes binds the attributes per vertex.
func BindLineVertexAttributes(s uint32) {
	vertAttrib := uint32(gl.GetAttribLocation(s, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	colorAttrib := uint32(gl.GetAttribLocation(s, gl.Str("color\x00")))
	gl.EnableVertexAttribArray(colorAttrib)
	gl.VertexAttribPointer(colorAttrib, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(12))
}

// PipVertex is a Vertex for the PIP.
type PipVertex struct {
	Vert, UV mgl32.Vec2
}

// BindPipVertexAttributes binds the attributes per vertex.
func BindPipVertexAttributes(s uint32) {
	posAttrib := uint32(gl.GetAttribLocation(s, gl.Str("pos\x00")))
	gl.EnableVertexAttribArray(posAttrib)
	gl.VertexAttribPointer(posAttrib, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	uvAttrib := uint32(gl.GetAttribLocation(s, gl.Str("uv\x00")))
	gl.EnableVertexAttribArray(uvAttrib)
	gl.VertexAttribPointer(uvAttrib, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(8))
}
