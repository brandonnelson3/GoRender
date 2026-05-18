package main

import (
	"fmt"
	"image/png"
	"os"
)

func main() {
	file, err := os.Open("slideframes/1.png")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	centerY := height / 2
	centerX := width / 2

	fmt.Printf("Analyzing image %d x %d...\n", width, height)

	// Let's scan horizontally across centerY and see the color transitions.
	// We want to see where the pure white area starts and ends in the middle.
	fmt.Println("\nHorizontal profile along centerY:")
	currentStart := -1
	var lastR, lastG, lastB, lastA uint32
	for x := 0; x < width; x++ {
		r, g, b, a := img.At(x, centerY).RGBA()
		if x == 0 {
			lastR, lastG, lastB, lastA = r, g, b, a
			currentStart = 0
		} else if r != lastR || g != lastG || b != lastB || a != lastA {
			fmt.Printf("X: %4d to %4d | Color: (%d, %d, %d, %d)\n", currentStart, x-1, lastR>>8, lastG>>8, lastB>>8, lastA>>8)
			lastR, lastG, lastB, lastA = r, g, b, a
			currentStart = x
		}
	}
	fmt.Printf("X: %4d to %4d | Color: (%d, %d, %d, %d)\n", currentStart, width-1, lastR>>8, lastG>>8, lastB>>8, lastA>>8)

	// Let's scan vertically across centerX and see the color transitions.
	fmt.Println("\nVertical profile along centerX:")
	currentStart = -1
	for y := 0; y < height; y++ {
		r, g, b, a := img.At(centerX, y).RGBA()
		if y == 0 {
			lastR, lastG, lastB, lastA = r, g, b, a
			currentStart = 0
		} else if r != lastR || g != lastG || b != lastB || a != lastA {
			fmt.Printf("Y: %4d to %4d | Color: (%d, %d, %d, %d)\n", currentStart, y-1, lastR>>8, lastG>>8, lastB>>8, lastA>>8)
			lastR, lastG, lastB, lastA = r, g, b, a
			currentStart = y
		}
	}
	fmt.Printf("Y: %4d to %4d | Color: (%d, %d, %d, %d)\n", currentStart, height-1, lastR>>8, lastG>>8, lastB>>8, lastA>>8)
}
