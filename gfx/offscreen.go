package gfx

import (
	"fmt"
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

// OffscreenFBO is a fixed-size framebuffer for rendering to texture.
// It is used by the render-test mode to capture deterministic output
// independent of the actual window size or swapchain.
type OffscreenFBO struct {
	fbo, colorTex, depthRBO uint32
	Width, Height            int32
}

// NewOffscreenFBO creates a new offscreen framebuffer at the given resolution.
// Must be called after a GL context is current.
func NewOffscreenFBO(width, height int32) (*OffscreenFBO, error) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	// Color texture
	var colorTex uint32
	gl.GenTextures(1, &colorTex)
	gl.BindTexture(gl.TEXTURE_2D, colorTex)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, width, height, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, colorTex, 0)

	// Depth+stencil renderbuffer
	var depthRBO uint32
	gl.GenRenderbuffers(1, &depthRBO)
	gl.BindRenderbuffer(gl.RENDERBUFFER, depthRBO)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH24_STENCIL8, width, height)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, depthRBO)

	if status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); status != gl.FRAMEBUFFER_COMPLETE {
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		return nil, fmt.Errorf("offscreen FBO incomplete: status 0x%x", status)
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	return &OffscreenFBO{
		fbo:      fbo,
		colorTex: colorTex,
		depthRBO: depthRBO,
		Width:    width,
		Height:   height,
	}, nil
}

// Bind sets this FBO as the active draw/read framebuffer and sets the viewport.
func (o *OffscreenFBO) Bind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, o.fbo)
	gl.Viewport(0, 0, o.Width, o.Height)
}

// Unbind restores the default framebuffer.
func (o *OffscreenFBO) Unbind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

// ReadPixels reads the color attachment and returns the image, flipped vertically
// (OpenGL origin is bottom-left, image convention is top-left).
func (o *OffscreenFBO) ReadPixels() *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, int(o.Width), int(o.Height)))
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, o.fbo)
	gl.ReadPixels(0, 0, o.Width, o.Height, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, 0)
	flipNRGBAVertical(img)
	return img
}

// flipNRGBAVertical flips the image rows in-place (OpenGL bottom-left → top-left).
func flipNRGBAVertical(img *image.NRGBA) {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	bytesPerRow := w * 4
	for top, bot := 0, h-1; top < bot; top, bot = top+1, bot-1 {
		topSlice := img.Pix[top*bytesPerRow : (top+1)*bytesPerRow]
		botSlice := img.Pix[bot*bytesPerRow : (bot+1)*bytesPerRow]
		for i := range topSlice {
			topSlice[i], botSlice[i] = botSlice[i], topSlice[i]
		}
	}
}

// Handle returns the underlying OpenGL framebuffer object ID.
func (o *OffscreenFBO) Handle() uint32 {
	return o.fbo
}

// Delete frees the GL resources associated with this FBO.
func (o *OffscreenFBO) Delete() {
	gl.DeleteFramebuffers(1, &o.fbo)
	gl.DeleteTextures(1, &o.colorTex)
	gl.DeleteRenderbuffers(1, &o.depthRBO)
}
