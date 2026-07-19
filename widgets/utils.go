package widgets

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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

// Show tooltip with given text at mouse position inside parent widget
func showTooltip(
	text string,
	localPos fyne.Position,
	textWidget *canvas.Text,
	bg *canvas.Rectangle,
	parentSize fyne.Size,
) {
	// Set tooltip text and calculate sizing
	textWidget.Text = text

	// Get proper tooltip sizing
	padding := float32(6)
	offset := float32(12)
	textSize := textWidget.MinSize()
	bgSize := fyne.NewSize(textSize.Width+padding*2, textSize.Height+padding*2)

	// Put tooltip in bottom right corner of cursor
	tx := localPos.X + offset
	ty := localPos.Y + offset

	// Flip position horizontally
	if tx+bgSize.Width > parentSize.Width {
		tx = localPos.X - bgSize.Width
	}

	// Flip position vertically
	if ty+bgSize.Height > parentSize.Height {
		ty = localPos.Y - bgSize.Height
	}

	// Move tooltip
	bg.Move(fyne.NewPos(tx, ty))
	bg.Resize(bgSize)
	textWidget.Move(fyne.NewPos(tx+padding, ty+padding))

	// Show tooltip
	bg.Show()
	textWidget.Show()
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
