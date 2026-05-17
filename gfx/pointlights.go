package gfx

import (
	"sort"
	"sync"
	"unsafe"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	// MaximumPointLights is the maximum number of lights that the pointlight system is prepared to handle.
	MaximumPointLights = 1024

	// MaxPointLightShadows is the maximum number of point lights that receive shadow cubemaps.
	MaxPointLightShadows = 4

	// pointShadowMapSize is the resolution (width = height) of each cubemap face.
	pointShadowMapSize = 512

	// pointShadowFarPlane is the far plane distance used for point light shadow projection.
	PointShadowFarPlane = 25.0
)

var (
	// PointLights are the current pointlights in the scene.
	PointLights         [MaximumPointLights]PointLight
	numPointLights      = uint32(0)
	nextPointLightIndex = uint32(0)
	mu                  sync.Mutex

	lightBuffer, visibleLightIndicesBuffer uint32

	// Point light shadow cubemaps: one FBO shared, and a single Texture Array holding all cubemaps.
	pointShadowFBO   uint32
	pointShadowArray uint32

	// shadowLightIndices holds, for each shadow slot, the global PointLights index of the
	// light whose cubemap is stored in that slot. -1 means unused.
	shadowLightIndices [MaxPointLightShadows]int

	// pointShadowLightPositions is a flat float32 slice of the shadow-casting light positions,
	// indexed by shadow slot. Kept in sync with shadowLightIndices.
	pointShadowLightPositions [MaxPointLightShadows * 3]float32

	// Caching states to prevent re-rendering static cubemaps
	lastShadowLightIndices    [MaxPointLightShadows]int
	lastShadowLightPositions  [MaxPointLightShadows]mgl32.Vec3
	shadowSlotRenderedObjects [MaxPointLightShadows][]RenderedObjectState
)

// RenderedObjectState holds the cached bounds of a rendered object to detect movement.
type RenderedObjectState struct {
	Renderable Renderable
	Min, Max   mgl32.Vec3
}

// PointLight represents all of the data about a PointLight.
type PointLight struct {
	Color     mgl32.Vec3
	Intensity float32
	Position  mgl32.Vec3
	Radius    float32
}

// VisibleIndex is a wrapper around an index.
type VisibleIndex struct {
	index int32
}

// InitPointLights sets up buffer space for light culling calculations and storage.
func InitPointLights() {
	// Prepare light buffers
	gl.GenBuffers(1, &lightBuffer)
	gl.GenBuffers(1, &visibleLightIndicesBuffer)

	// Bind light buffer
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, lightBuffer)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, MaximumPointLights*int(unsafe.Sizeof(PointLight{})), unsafe.Pointer(&PointLights), gl.DYNAMIC_DRAW)

	// Bind visible light indices buffer
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, visibleLightIndicesBuffer)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, int(getTotalNumTiles())*int(unsafe.Sizeof(VisibleIndex{}))*MaximumPointLights, nil, gl.STATIC_DRAW)

	// Unbind for safety.
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, 0)

	initPointLightShadows()
}

// initPointLightShadows allocates the FBO and a single texture array used for point light shadows.
func initPointLightShadows() {
	for i := range shadowLightIndices {
		shadowLightIndices[i] = -1
		lastShadowLightIndices[i] = -1
		shadowSlotRenderedObjects[i] = nil
	}

	gl.GenFramebuffers(1, &pointShadowFBO)
	gl.GenTextures(1, &pointShadowArray)

	gl.BindTexture(gl.TEXTURE_CUBE_MAP_ARRAY, pointShadowArray)
	// 4 layers, 6 faces each = 24 faces total.
	gl.TexImage3D(
		gl.TEXTURE_CUBE_MAP_ARRAY,
		0, gl.DEPTH_COMPONENT32F,
		pointShadowMapSize, pointShadowMapSize, MaxPointLightShadows*6,
		0, gl.DEPTH_COMPONENT, gl.FLOAT, nil,
	)

	gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_COMPARE_MODE, gl.COMPARE_REF_TO_TEXTURE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_COMPARE_FUNC, gl.LEQUAL)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP_ARRAY, 0)
}

// ResizePointLightBuffers reallocates the visible light indices buffer based on the current window size.
func ResizePointLightBuffers() {
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, visibleLightIndicesBuffer)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, int(getTotalNumTiles())*int(unsafe.Sizeof(VisibleIndex{}))*MaximumPointLights, nil, gl.STATIC_DRAW)
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, 0)
}

// ResetPointLights clears all point lights from the scene.
// Intended for render-test scenes that want full control over lighting.
func ResetPointLights() {
	mu.Lock()
	PointLights = [MaximumPointLights]PointLight{}
	numPointLights = 0
	nextPointLightIndex = 0
	mu.Unlock()

	for i := range lastShadowLightIndices {
		lastShadowLightIndices[i] = -1
		shadowSlotRenderedObjects[i] = nil
	}

	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, lightBuffer)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, MaximumPointLights*int(unsafe.Sizeof(PointLight{})), unsafe.Pointer(&PointLights), gl.DYNAMIC_DRAW)
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, 0)
}

// GetNumPointLights returns the number of PointLights that are currently in the scene.
func GetNumPointLights() uint32 {
	return numPointLights
}

// AddPointLight adds a PointLight to the scene with the given attributes.
func AddPointLight(position, color mgl32.Vec3, intensity, radius float32) {
	mu.Lock()

	PointLights[nextPointLightIndex].Color = color
	PointLights[nextPointLightIndex].Intensity = intensity
	PointLights[nextPointLightIndex].Position = position
	PointLights[nextPointLightIndex].Radius = radius

	numPointLights++
	nextPointLightIndex++

	if numPointLights >= MaximumPointLights {
		numPointLights = MaximumPointLights - 1
	}
	if nextPointLightIndex >= MaximumPointLights {
		nextPointLightIndex = 0
	}

	mu.Unlock()

	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, lightBuffer)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, MaximumPointLights*int(unsafe.Sizeof(PointLight{})), unsafe.Pointer(&PointLights), gl.DYNAMIC_DRAW)

	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, 0)
}

// GetPointLightBuffer retrieves the private lightBuffer variable.
func GetPointLightBuffer() uint32 {
	return lightBuffer
}

// GetPointLightVisibleLightIndicesBuffer retrieves the private visibleLightIndicesBuffer variable.
func GetPointLightVisibleLightIndicesBuffer() uint32 {
	return visibleLightIndicesBuffer
}

// GetPointShadowFBO returns the shared FBO used for rendering point light shadow cubemaps.
func GetPointShadowFBO() uint32 {
	return pointShadowFBO
}

// GetPointShadowArray returns the texture array ID containing all point shadow cubemaps.
func GetPointShadowArray() uint32 {
	return pointShadowArray
}

// UpdatePointLightShadowSlots selects the closest MaxPointLightShadows lights to
// cameraPos that are also within the camera's view, and updates shadowLightIndices + pointShadowLightPositions.
// Returns the number of active shadow slots.
func UpdatePointLightShadowSlots(cameraPos mgl32.Vec3, frustum *Frustum) int {
	mu.Lock()
	n := int(numPointLights)
	mu.Unlock()

	if n == 0 {
		for i := range shadowLightIndices {
			shadowLightIndices[i] = -1
		}
		return 0
	}

	// Build a sorted list of (distance, index) pairs.
	type lightDist struct {
		idx  int
		dist float32
	}
	candidates := make([]lightDist, 0, n)
	for i := 0; i < n; i++ {
		p := PointLights[i].Position
		// If a light's shadow sphere doesn't intersect the camera frustum, it's not visible.
		if frustum != nil && !frustum.IsSphereIn(p, PointShadowFarPlane) {
			continue
		}
		candidates = append(candidates, lightDist{i, p.Sub(cameraPos).Len()})
	}
	sort.Slice(candidates, func(a, b int) bool {
		return candidates[a].dist < candidates[b].dist
	})

	count := len(candidates)
	if count > MaxPointLightShadows {
		count = MaxPointLightShadows
	}
	for slot := 0; slot < MaxPointLightShadows; slot++ {
		if slot < count {
			idx := candidates[slot].idx
			shadowLightIndices[slot] = idx
			p := PointLights[idx].Position
			pointShadowLightPositions[slot*3+0] = p[0]
			pointShadowLightPositions[slot*3+1] = p[1]
			pointShadowLightPositions[slot*3+2] = p[2]
		} else {
			shadowLightIndices[slot] = -1
		}
	}
	return count
}

// GetShadowLightIndices returns the current shadow slot → light index mapping.
func GetShadowLightIndices() []int {
	return shadowLightIndices[:]
}

// GetPointShadowLightPositions returns a flat float32 slice of shadow light positions
// (3 floats per slot, MaxPointLightShadows slots).
func GetPointShadowLightPositions() *[MaxPointLightShadows * 3]float32 {
	return &pointShadowLightPositions
}

// BuildPointLightCubemapMatrices returns the 6 face view-projection matrices for the
// given light position, using PointShadowFarPlane as the far plane.
func BuildPointLightCubemapMatrices(lightPos mgl32.Vec3) [6]mgl32.Mat4 {
	proj := mgl32.Perspective(mgl32.DegToRad(90), 1.0, 0.05, PointShadowFarPlane)
	return [6]mgl32.Mat4{
		proj.Mul4(mgl32.LookAtV(lightPos, lightPos.Add(mgl32.Vec3{1, 0, 0}), mgl32.Vec3{0, -1, 0})),
		proj.Mul4(mgl32.LookAtV(lightPos, lightPos.Add(mgl32.Vec3{-1, 0, 0}), mgl32.Vec3{0, -1, 0})),
		proj.Mul4(mgl32.LookAtV(lightPos, lightPos.Add(mgl32.Vec3{0, 1, 0}), mgl32.Vec3{0, 0, 1})),
		proj.Mul4(mgl32.LookAtV(lightPos, lightPos.Add(mgl32.Vec3{0, -1, 0}), mgl32.Vec3{0, 0, -1})),
		proj.Mul4(mgl32.LookAtV(lightPos, lightPos.Add(mgl32.Vec3{0, 0, 1}), mgl32.Vec3{0, -1, 0})),
		proj.Mul4(mgl32.LookAtV(lightPos, lightPos.Add(mgl32.Vec3{0, 0, -1}), mgl32.Vec3{0, -1, 0})),
	}
}

// IsShadowSlotDirty returns true if the shadow slot needs to be re-rendered.
// It also updates the last-rendered state for that slot.
func IsShadowSlotDirty(slot int, intersecting []Renderable) bool {
	lightIdx := shadowLightIndices[slot]
	if lightIdx == -1 {
		lastShadowLightIndices[slot] = -1
		shadowSlotRenderedObjects[slot] = nil
		return false
	}

	currentPos := PointLights[lightIdx].Position
	dirty := lastShadowLightIndices[slot] != lightIdx ||
		lastShadowLightPositions[slot].Sub(currentPos).LenSqr() > 0.0001 ||
		len(intersecting) != len(shadowSlotRenderedObjects[slot])

	if !dirty {
		for i, currentObj := range intersecting {
			prevObj := shadowSlotRenderedObjects[slot][i]
			if prevObj.Renderable != currentObj {
				dirty = true
				break
			}
			min, max := currentObj.GetBounds()
			if prevObj.Min != min || prevObj.Max != max {
				dirty = true
				break
			}
		}
	}

	if dirty {
		lastShadowLightIndices[slot] = lightIdx
		lastShadowLightPositions[slot] = currentPos
		shadowSlotRenderedObjects[slot] = make([]RenderedObjectState, len(intersecting))
		for i, currentObj := range intersecting {
			min, max := currentObj.GetBounds()
			shadowSlotRenderedObjects[slot][i] = RenderedObjectState{
				Renderable: currentObj,
				Min:        min,
				Max:        max,
			}
		}
	}
	return dirty
}
