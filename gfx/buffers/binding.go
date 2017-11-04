package buffers

import "github.com/go-gl/gl/v4.5-core/gl"

// Binding is a wrapper around a shader buffer layout location.
type Binding struct {
	uint32
}

// NewBinding instantiates a Binding for the provided buffer layout location.
func NewBinding(l uint32) *Binding {
	return &Binding{l}
}

// Set Binds this Binding to the provided buffer.
func (b *Binding) Set(buf uint32) {
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, b.uint32, buf)
}
