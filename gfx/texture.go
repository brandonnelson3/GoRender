package gfx

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/png" // Required for image/png to work..
	"strings"

	"github.com/brandonnelson3/GoRender/loader"
	"github.com/go-gl/gl/v4.5-core/gl"
)

const (
	pngExt = ".png"
	jpgExt = ".jpg"
	tgaExt = ".tga"
)

// LoadTexture loads the texture in the provided file, based on the file extension.
func LoadTexture(file string) (uint32, error) {
	if strings.HasSuffix(file, pngExt) {
		return fromPng(file)
	}
	return 0, fmt.Errorf("Attempted to load texture from unsupported file type: %v", file)
}

// fromPng builds a texture from the provided Png file.
func fromPng(file string) (uint32, error) {
	r, err := loader.Load(file)
	if err != nil {
		return 0, err
	}
	img, _, err := image.Decode(r)
	if err != nil {
		return 0, err
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture, nil
}
