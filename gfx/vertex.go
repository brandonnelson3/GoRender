package gfx

import "github.com/go-gl/mathgl/mgl32"

// Vertex is a Vertex.
type Vertex struct {
	Vert, Norm mgl32.Vec3
	UV         mgl32.Vec2
}

// LineVertex is a Vertex.
type LineVertex struct {
	Vert, Color mgl32.Vec3
}

// PipVertex is a Vertex for the PIP.
type PipVertex struct {
	Vert, UV mgl32.Vec2
}
