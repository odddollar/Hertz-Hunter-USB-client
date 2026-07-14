package widgets

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2/theme"
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

// Convert rssi intensity to colour from gradient
func waterfallColour(intensity int) color.Color {
	// Get primary colour
	r, g, b, _ := theme.Color(theme.ColorNamePrimary).RGBA()

	// Calculation intensity position on colour gradient
	return color.RGBA{
		R: uint8((r * uint32(intensity)) / (255 * 257)),
		G: uint8((g * uint32(intensity)) / (255 * 257)),
		B: uint8((b * uint32(intensity)) / (255 * 257)),
		A: 255,
	}
}
