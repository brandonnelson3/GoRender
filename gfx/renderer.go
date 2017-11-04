package gfx

import (
	"log"

	"github.com/brandonnelson3/GoRender/gfx/shaders"
	"github.com/go-gl/gl/v4.5-core/gl"
)

var (
	Renderer r
)

type r struct {
	lightCullingShader  *shaders.LightCullingShader
	colorVertexShader   *shaders.ColorVertexShader
	colorFragmentShader *shaders.ColorFragmentShader
}

func InitRenderer() {
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.MULTISAMPLE)
	gl.DepthMask(true)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	lcs, err := shaders.NewLightCullingShader()
	if err != nil {
		log.Fatalf("Failed to compile LightCullingShader: %v", err)
	}

	cvs, err := shaders.NewColorVertexShader()
	if err != nil {
		log.Fatalf("Failed to compile ColorVertexShader: %v", err)
	}

	cfs, err := shaders.NewColorFragmentShader()
	if err != nil {
		log.Fatalf("Failed to compile ColorFragmentShader: %v", err)
	}

	Renderer = r{
		lightCullingShader:  lcs,
		colorVertexShader:   cvs,
		colorFragmentShader: cfs,
	}
}

func (renderer *r) Render(renderables []*Renderable) {
	for _, renderable := range renderables {
		renderable.Render()
	}
}
