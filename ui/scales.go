package ui

import (
	"Hertz-Hunter-USB-client/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

// Create scale with given text alignment
func newSideScale(lowText, midText, highText string, alignment fyne.TextAlign) *fyne.Container {
	full := canvas.NewText(highText, theme.Color(theme.ColorNameForeground))
	full.Alignment = alignment
	full.TextStyle.Bold = true

	mid := canvas.NewText(midText, theme.Color(theme.ColorNameForeground))
	mid.Alignment = alignment
	mid.TextStyle.Bold = true

	none := canvas.NewText(lowText, theme.Color(theme.ColorNameForeground))
	none.Alignment = alignment
	none.TextStyle.Bold = true

	return container.NewBorder(
		full,
		none,
		nil,
		nil,
		mid,
	)
}

// Create frequency scale with given text
func newFrequencyScale(lowText, midText, highText, spacerText string) *fyne.Container {
	// Alignment spacer
	t := canvas.NewText(spacerText, theme.Color(theme.ColorNameForeground))
	t.TextStyle.Bold = true
	spacer := widgets.NewSpacer(t.MinSize())

	left := canvas.NewText(lowText, theme.Color(theme.ColorNameForeground))
	left.Alignment = fyne.TextAlignLeading
	left.TextStyle.Bold = true

	middle := canvas.NewText(midText, theme.Color(theme.ColorNameForeground))
	middle.Alignment = fyne.TextAlignCenter
	middle.TextStyle.Bold = true

	right := canvas.NewText(highText, theme.Color(theme.ColorNameForeground))
	right.Alignment = fyne.TextAlignTrailing
	right.TextStyle.Bold = true

	return container.NewGridWithColumns(3,
		container.NewBorder(nil, nil, spacer, nil, left),
		middle,
		container.NewBorder(nil, nil, nil, spacer, right),
	)
}
