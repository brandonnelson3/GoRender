package gfx

import (
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"time"

	"github.com/go-gl/gl/v4.5-core/gl"
)

const (
	screenshotFilePath   = "screenshots"
	screenshotFileFormat = "2006-01-02-15-04-05"
)

func Screenshot() {
	path := filepath.Join(screenshotFilePath, time.Now().Format(screenshotFileFormat)+".jpg")
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 755)
	}
	screenshot := image.NewRGBA(image.Rect(0, 0, int(Window.Width), int(Window.Height)))
	gl.ReadPixels(0, 0, int32(Window.Width), int32(Window.Height), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(screenshot.Pix))
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	// Screenshots come out upside down, so this code flips the data before encoding it to the file.
	bytesPerPixel := 4
	bytesPerRow := int(Window.Width) * bytesPerPixel
	for rowFirstByteIndex1, rowFirstByteIndex2 := 0, (int(Window.Height)-1)*bytesPerRow; rowFirstByteIndex1 < int(len(screenshot.Pix)/2); rowFirstByteIndex1, rowFirstByteIndex2 = rowFirstByteIndex1+bytesPerRow, rowFirstByteIndex2-bytesPerRow {
		for pixelFirstByteIndexWithinRow := 0; pixelFirstByteIndexWithinRow < bytesPerRow; pixelFirstByteIndexWithinRow += bytesPerPixel {
			pixelIndex1 := rowFirstByteIndex1 + pixelFirstByteIndexWithinRow
			pixelIndex2 := rowFirstByteIndex2 + pixelFirstByteIndexWithinRow

			screenshot.Pix[pixelIndex1], screenshot.Pix[pixelIndex2] = screenshot.Pix[pixelIndex2], screenshot.Pix[pixelIndex1]
			screenshot.Pix[pixelIndex1+1], screenshot.Pix[pixelIndex2+1] = screenshot.Pix[pixelIndex2+1], screenshot.Pix[pixelIndex1+1]
			screenshot.Pix[pixelIndex1+2], screenshot.Pix[pixelIndex2+2] = screenshot.Pix[pixelIndex2+2], screenshot.Pix[pixelIndex1+2]
			screenshot.Pix[pixelIndex1+3], screenshot.Pix[pixelIndex2+3] = screenshot.Pix[pixelIndex2+3], screenshot.Pix[pixelIndex1+3]
		}
	}

	if err := jpeg.Encode(f, screenshot, nil); err != nil {
		panic(err)
	}
	f.Close()
}
