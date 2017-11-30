package shaders

import "github.com/go-gl/gl/v4.5-core/gl"

type Shader struct {
	uint32
}

// Program returns the opengl program id of this vertex shader.
func (s *Shader) Program() uint32 {
	return s.uint32
}

// Use binds this program to be used.
func (s *Shader) Use() {
	gl.UseProgram(s.uint32)
}
