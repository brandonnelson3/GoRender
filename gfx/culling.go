package gfx

import "github.com/go-gl/mathgl/mgl32"

// Plane represents a plane in 3D space in the form Ax + By + Cz + D = 0.
type Plane struct {
	Normal   mgl32.Vec3
	Distance float32
}

// Normalize normalizes the plane's normal and adjusts the distance accordingly.
func (p *Plane) Normalize() {
	length := p.Normal.Len()
	p.Normal = p.Normal.Mul(1.0 / length)
	p.Distance /= length
}

// DistanceToPoint returns the signed distance from the plane to a point.
func (p *Plane) DistanceToPoint(point mgl32.Vec3) float32 {
	return p.Normal.Dot(point) + p.Distance
}

// Frustum represents the 6 planes of a viewing frustum.
type Frustum struct {
	Planes [6]Plane
}

// NewFrustumFromMatrix extracts frustum planes from a projection-view matrix.
// The matrix should be in the order: projection * view.
func NewFrustumFromMatrix(m mgl32.Mat4) *Frustum {
	f := &Frustum{}

	// Left Plane
	f.Planes[0] = Plane{
		Normal:   mgl32.Vec3{m[3] + m[0], m[7] + m[4], m[11] + m[8]},
		Distance: m[15] + m[12],
	}
	// Right Plane
	f.Planes[1] = Plane{
		Normal:   mgl32.Vec3{m[3] - m[0], m[7] - m[4], m[11] - m[8]},
		Distance: m[15] - m[12],
	}
	// Bottom Plane
	f.Planes[2] = Plane{
		Normal:   mgl32.Vec3{m[3] + m[1], m[7] + m[5], m[11] + m[9]},
		Distance: m[15] + m[13],
	}
	// Top Plane
	f.Planes[3] = Plane{
		Normal:   mgl32.Vec3{m[3] - m[1], m[7] - m[5], m[11] - m[9]},
		Distance: m[15] - m[13],
	}
	// Near Plane
	f.Planes[4] = Plane{
		Normal:   mgl32.Vec3{m[3] + m[2], m[7] + m[6], m[11] + m[10]},
		Distance: m[15] + m[14],
	}
	// Far Plane
	f.Planes[5] = Plane{
		Normal:   mgl32.Vec3{m[3] - m[2], m[7] - m[6], m[11] - m[10]},
		Distance: m[15] - m[14],
	}

	for i := range f.Planes {
		f.Planes[i].Normalize()
	}

	return f
}

// IsBoxIn returns true if the axis-aligned bounding box is partially or fully inside the frustum.
func (f *Frustum) IsBoxIn(min, max mgl32.Vec3) bool {
	for i := 0; i < 6; i++ {
		// Check the "positive" vertex of the box relative to the plane normal.
		// If the most-positive vertex is behind the plane, the box is outside.
		p := min
		if f.Planes[i].Normal.X() >= 0 {
			p = mgl32.Vec3{max.X(), p.Y(), p.Z()}
		}
		if f.Planes[i].Normal.Y() >= 0 {
			p = mgl32.Vec3{p.X(), max.Y(), p.Z()}
		}
		if f.Planes[i].Normal.Z() >= 0 {
			p = mgl32.Vec3{p.X(), p.Y(), max.Z()}
		}

		if f.Planes[i].DistanceToPoint(p) < 0 {
			return false
		}
	}
	return true
}

// IsSphereIn returns true if the sphere is partially or fully inside the frustum.
func (f *Frustum) IsSphereIn(center mgl32.Vec3, radius float32) bool {
	for i := 0; i < 6; i++ {
		if f.Planes[i].DistanceToPoint(center) < -radius {
			return false
		}
	}
	return true
}
