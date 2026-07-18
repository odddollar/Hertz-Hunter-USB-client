package widgets

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

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
	rssiValues     []int
	minCalibration int
	maxCalibration int
	minFrequency   int
	maxFrequency   int

	// Update tooltip position
	lastMousePos fyne.Position
	mouseIn      bool
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
	r.mouseIn = true
	r.lastMousePos = event.Position
	r.updateTooltip(event.Position)
}

// Updates tooltip when mouse moves over widget
func (r *RssiGraph) MouseMoved(event *desktop.MouseEvent) {
	r.lastMousePos = event.Position
	r.updateTooltip(event.Position)
}

// Hides tooltip when mouse leaves widget
func (r *RssiGraph) MouseOut() {
	r.mouseIn = false
	r.tooltipBg.Hide()
	r.tooltipText.Hide()
	r.Refresh()
}

// Reset clears the graph display and any stored data
func (r *RssiGraph) Reset() {
	// Clear stored data
	r.rssiValues = nil

	// Hide tooltip
	r.tooltipBg.Hide()
	r.tooltipText.Hide()

	// Draw blank black canvas
	img := newEmptyImage(r.graphWidth, r.graphHeight, color.Black)
	r.graphCanvas.Image = img

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
	r.minCalibration = minCalibration
	r.maxCalibration = maxCalibration
	r.minFrequency = minFrequency
	r.maxFrequency = maxFrequency

	// Calculation values
	barCount := len(numbers)

	// Create blank image
	img := newEmptyImage(r.graphWidth, r.graphHeight, color.Black)

	for i, value := range numbers {
		// Map value to bar height
		barHeight := mapClamped(value, minCalibration, maxCalibration, 0, r.graphHeight)

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

	// Update tooltip if mouse still inside
	if r.mouseIn {
		r.updateTooltip(r.lastMousePos)
	}
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
	if barCount < 2 {
		return
	}
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

	// Calculate signal strength
	rssi := r.rssiValues[barsOver]
	rssiStrength := mapClamped(rssi, r.minCalibration, r.maxCalibration, 0, 100)

	// Format tooltip text
	r.tooltipText.Text = fmt.Sprintf("%dMHz, %d%%", frequency, rssiStrength)

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
	r.Refresh()
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
	return fyne.NewSize(500, 200)
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
