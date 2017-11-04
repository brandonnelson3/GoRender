package gfx

import "github.com/go-gl/gl/v4.5-core/gl"

var (
	Renderer r
)

type r struct {
}

func InitRenderer() {
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.MULTISAMPLE)
	gl.DepthMask(true)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	Renderer = r{}
}

func (renderer *r) Render(renderables []*Renderable) {
	for _, renderable := range renderables {
		renderable.Render()
	}
}
