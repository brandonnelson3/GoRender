package gfx

import "github.com/go-gl/mathgl/mgl32"

// Vertex is a Vertex.
type Vertex struct {
	Vert, Norm mgl32.Vec3
	UV         mgl32.Vec2
}
