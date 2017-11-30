package shaders

import "github.com/go-gl/gl/v4.5-core/gl"

// shader is a base struct of all other shader programs.
type shader struct {
	uint32
}

// Program returns the opengl program id of this vertex shader.
func (s *shader) Program() uint32 {
	return s.uint32
}

// Use binds this program to be used.
func (s *shader) Use() {
	gl.UseProgram(s.uint32)
}
