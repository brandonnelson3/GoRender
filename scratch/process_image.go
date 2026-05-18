package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func main() {
	// Open the image file
	file, err := os.Open("slideframes/1.png")
	if err != nil {
		fmt.Printf("Error opening image: %v\n", err)
		return
	}
	defer file.Close()

	// Decode the image
	img, err := png.Decode(file)
	if err != nil {
		fmt.Printf("Error decoding image: %v\n", err)
		return
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	fmt.Printf("Image dimensions: %d x %d\n", width, height)

	// Exact coordinates of the 16:9 slide canvas:
	// X: 633 to 3429 (Width: 2797)
	// Y: 227 to 1799 (Height: 1573)
	// Ratio: 2797 / 1573 = 1.778 (~16:9)
	minX, maxX := 633, 3429
	minY, maxY := 227, 1799

	fmt.Printf("Using exact white rectangle bounds: X[%d to %d], Y[%d to %d] (Width: %d, Height: %d)\n",
		minX, maxX, minY, maxY, maxX-minX+1, maxY-minY+1)

	// Create a new RGBA image
	outImg := image.NewRGBA(bounds)
	draw.Draw(outImg, bounds, img, bounds.Min, draw.Src)

	// Make the detected rectangle transparent
	transparentColor := color.RGBA{0, 0, 0, 0}
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			outImg.Set(x, y, transparentColor)
		}
	}

	// Save the output image
	outFile, err := os.Create("slideframes/hud.png")
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outFile.Close()

	err = png.Encode(outFile, outImg)
	if err != nil {
		fmt.Printf("Error encoding PNG: %v\n", err)
		return
	}

	fmt.Println("Successfully saved transparent HUD image to slideframes/hud.png")
}
