package widgets

import (
	"fmt"
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
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

// Custom widget that displays rssi graph and shows bar data when hovered
type RssiGraph struct {
	widget.BaseWidget

	// Ui elements
	graphCanvas *canvas.Image
	tooltipBg   *canvas.Rectangle
	tooltipText *canvas.Text

	// Constants
	graphWidth  int
	graphHeight int

	// Calculate tooltip data
	rssiValues   []int
	minFrequency int
	maxFrequency int
}

// Creates new RssiGraph widget
func NewRssiGraph(graphWidth, graphHeight int) *RssiGraph {
	// Create graph canvas from given image
	graphCanvas := canvas.NewImageFromImage(newEmptyImage(graphWidth, graphHeight, color.Black))
	graphCanvas.FillMode = canvas.ImageFillStretch
	graphCanvas.ScaleMode = canvas.ImageScalePixels

	// Create background
	tooltipBg := canvas.NewRectangle(color.RGBA{R: 32, G: 32, B: 36, A: 235})

	// Create text
	tooltipText := canvas.NewText("", color.White)
	tooltipText.TextSize = 13

	// Hide tooltip by default
	tooltipBg.Hide()
	tooltipText.Hide()

	// Create new object
	graph := &RssiGraph{
		graphCanvas: graphCanvas,
		tooltipBg:   tooltipBg,
		tooltipText: tooltipText,
		graphWidth:  graphWidth,
		graphHeight: graphHeight,
	}

	// Extend base widget and return
	graph.ExtendBaseWidget(graph)
	return graph
}

// Updates tooltip when mouse enters widget
func (r *RssiGraph) MouseIn(event *desktop.MouseEvent) {
	r.updateTooltip(event.Position)
}

// Updates tooltip when mouse moves over widget
func (r *RssiGraph) MouseMoved(event *desktop.MouseEvent) {
	r.updateTooltip(event.Position)
}

// Hides tooltip when mouse leaves widget
func (r *RssiGraph) MouseOut() {
	r.tooltipBg.Hide()
	r.tooltipText.Hide()
	r.Refresh()
}

// Updates graph image
func (r *RssiGraph) UpdateGraph(
	numbers []int,
	minCalibration, maxCalibration int,
	minFrequency, maxFrequency int,
) {
	if len(numbers) == 0 {
		return
	}

	// Used for calculating tooltip text
	// Updated every time data polled from device
	r.rssiValues = numbers
	r.minFrequency = minFrequency
	r.maxFrequency = maxFrequency

	// Calculation values
	barCount := len(numbers)
	calibrationRange := maxCalibration - minCalibration

	// Create blank image
	img := newEmptyImage(r.graphWidth, r.graphHeight, color.Black)

	for i, value := range numbers {
		// Clamp value to range
		if value < minCalibration {
			value = minCalibration
		}
		if value > maxCalibration {
			value = maxCalibration
		}

		// Height as percentage of calibration range
		barHeight := (value - minCalibration) * r.graphHeight / calibrationRange

		// Dimensions and position of bar
		x1 := i * r.graphWidth / barCount
		x2 := (i + 1) * r.graphWidth / barCount
		y1 := r.graphHeight - barHeight

		// Draw bar
		for y := y1; y < r.graphHeight; y++ {
			for x := x1; x < x2 && x < r.graphWidth; x++ {
				img.Set(x, y, theme.Color(theme.ColorNamePrimary))
			}
		}
	}

	r.graphCanvas.Image = img
	r.Refresh()
}

// Updates tooltip position and text
func (r *RssiGraph) updateTooltip(localPos fyne.Position) {
	// Get drawn graph dimensions
	drawSize := r.graphCanvas.Size()
	if drawSize.Width <= 0 || drawSize.Height <= 0 {
		return
	}

	// Calculate number of bars over from 0
	barCount := len(r.rssiValues)
	displayWidth := int(drawSize.Width)
	mouseX := int(localPos.X)
	if mouseX == displayWidth {
		mouseX--
	}
	barsOver := (mouseX * barCount) / displayWidth

	// Calculate frequency with integer rounding
	freqRange := r.maxFrequency - r.minFrequency
	denom := barCount - 1
	num := barsOver * freqRange
	frequency := (num+denom/2)/denom + r.minFrequency

	// Format tooltip text
	textChanged := false
	tooltipText := fmt.Sprintf("%dMHz", frequency)
	if r.tooltipText.Text != tooltipText {
		r.tooltipText.Text = tooltipText
		textChanged = true
	}

	// Get proper tooltip sizing
	padding := float32(6)
	offset := float32(12)
	textSize := r.tooltipText.MinSize()
	bgSize := fyne.NewSize(textSize.Width+padding*2, textSize.Height+padding*2)

	// Put tooltip in bottom right corner of cursor
	tx := localPos.X + offset
	ty := localPos.Y + offset

	// Flip position horizontally
	if tx+bgSize.Width > r.Size().Width {
		tx = localPos.X - bgSize.Width
	}

	// Flip position vertically
	if ty+bgSize.Height > r.Size().Height {
		ty = localPos.Y - bgSize.Height
	}

	// Move tooltip
	r.tooltipBg.Move(fyne.NewPos(tx, ty))
	r.tooltipBg.Resize(bgSize)
	r.tooltipText.Move(fyne.NewPos(tx+padding, ty+padding))

	// Show tooltip
	r.tooltipBg.Show()
	r.tooltipText.Show()
	if textChanged {
		r.Refresh()
	}
}

// Returns new renderer for RssiGraph
func (r *RssiGraph) CreateRenderer() fyne.WidgetRenderer {
	return &rssiGraphRenderer{rssiGraph: r}
}

// Renderer for RssiGraph widget
type rssiGraphRenderer struct {
	rssiGraph *RssiGraph
}

// Returns minimum size of RssiGraph
func (r *rssiGraphRenderer) MinSize() fyne.Size {
	return fyne.NewSize(160, 120)
}

// Lays out image to fill RssiGraph
func (r *rssiGraphRenderer) Layout(size fyne.Size) {
	r.rssiGraph.graphCanvas.Resize(size)
}

// Refreshes RssiGraph
func (r *rssiGraphRenderer) Refresh() {
	r.rssiGraph.graphCanvas.Refresh()
	r.rssiGraph.tooltipBg.Refresh()
	r.rssiGraph.tooltipText.Refresh()
}

// Returns child widgets of RssiGraph
func (r *rssiGraphRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		r.rssiGraph.graphCanvas,
		r.rssiGraph.tooltipBg,
		r.rssiGraph.tooltipText,
	}
}

// Does nothing as RssiGraph doesn't hold external resources
func (r *rssiGraphRenderer) Destroy() {}
