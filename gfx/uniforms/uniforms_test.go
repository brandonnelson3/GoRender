package uniforms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFloat(t *testing.T) {
	u := NewFloat(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewFloatArray(t *testing.T) {
	u := NewFloatArray(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewInt(t *testing.T) {
	u := NewInt(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewIVector2(t *testing.T) {
	u := NewIVector2(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewIVector3(t *testing.T) {
	u := NewIVector3(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewIVector4(t *testing.T) {
	u := NewIVector4(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewMatrix4(t *testing.T) {
	u := NewMatrix4(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewMatrix4Array(t *testing.T) {
	u := NewMatrix4Array(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewSampler2D(t *testing.T) {
	u := NewSampler2D(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewUInt(t *testing.T) {
	u := NewUInt(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewUIVector2(t *testing.T) {
	u := NewUIVector2(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewUIVector3(t *testing.T) {
	u := NewUIVector3(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewUIVector4(t *testing.T) {
	u := NewUIVector4(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewVector2(t *testing.T) {
	u := NewVector2(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewVector3(t *testing.T) {
	u := NewVector3(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}

func TestNewVector4(t *testing.T) {
	u := NewVector4(1, 2)
	assert.NotNil(t, u)
	assert.Equal(t, uint32(1), u.program)
	assert.Equal(t, int32(2), u.uniform)
}
