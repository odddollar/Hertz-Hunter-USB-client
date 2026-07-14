package widgets

import (
	"image"
	"image/color"
)

// Generate empty image
func newEmptyImage(width, height int, c color.Color) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	// Fill background
	for y := range height {
		for x := range width {
			img.Set(x, y, c)
		}
	}

	return img
}

// Clamp and map value from one range to another
func mapClamped(value, inMin, inMax, outMin, outMax int) int {
	if value < inMin {
		value = inMin
	}
	if value > inMax {
		value = inMax
	}
	return outMin + (value-inMin)*(outMax-outMin)/(inMax-inMin)
}
