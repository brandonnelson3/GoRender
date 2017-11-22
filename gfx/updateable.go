package gfx

import "github.com/brandonnelson3/GoRender/gfx/shaders"

type Updateable interface {
	Update(*shaders.DepthVertexShader, *shaders.ColorVertexShader)
}
