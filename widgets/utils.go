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
	// Get primary and error theme colours
	pr, pg, pb, _ := theme.Color(theme.ColorNamePrimary).RGBA()
	er, eg, eb, _ := theme.Color(theme.ColorNameError).RGBA()

	// Interpolate from black to primary
	if intensity <= 128 {
		t := uint32(intensity)
		return color.RGBA{
			R: uint8((pr * t) / (128 * 257)),
			G: uint8((pg * t) / (128 * 257)),
			B: uint8((pb * t) / (128 * 257)),
			A: 255,
		}
	}

	// Interpolate from primary to error
	t := uint32(intensity - 128)
	return color.RGBA{
		R: uint8(pr + (er-pr)*t/(127*257)),
		G: uint8(pg + (eg-pg)*t/(127*257)),
		B: uint8(pb + (eb-pb)*t/(127*257)),
		A: 255,
	}
}
