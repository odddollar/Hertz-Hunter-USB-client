package widgets

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Duration of history displayed
const waterfallWindow = 30 * time.Second
const waterfallWindowBuffer = 2 * time.Second

// Single row of history captured at given moment
type waterfallRow struct {
	values    []int
	timestamp time.Time
}

// Custom widget that displays rssi waterfall and shows historical strength when hovered
type WaterfallGraph struct {
	widget.BaseWidget

	// Rows of history, oldest first, used for drawing and tooltip lookup
	rows []waterfallRow

	// Ui elements
	graphCanvas *canvas.Image
	tooltipBg   *canvas.Rectangle
	tooltipText *canvas.Text

	// Constants
	graphWidth  int
	graphHeight int

	// Calculate tooltip data
	minCalibration int
	maxCalibration int
	minFrequency   int
	maxFrequency   int

	// Update tooltip position
	lastMousePos fyne.Position
	mouseIn      bool
}

// Creates new WaterfallGraph widget
func NewWaterfallGraph(graphWidth, graphHeight int) *WaterfallGraph {
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
	graph := &WaterfallGraph{
		graphCanvas:    graphCanvas,
		tooltipBg:      tooltipBg,
		tooltipText:    tooltipText,
		graphWidth:     graphWidth,
		graphHeight:    graphHeight,
		minCalibration: 0,
		maxCalibration: 4096,
	}

	// Extend base widget and return
	graph.ExtendBaseWidget(graph)
	return graph
}

// Updates tooltip when mouse enters widget
func (w *WaterfallGraph) MouseIn(event *desktop.MouseEvent) {
	w.mouseIn = true
	w.lastMousePos = event.Position
	w.updateTooltip(event.Position)
}

// Updates tooltip when mouse moves over widget
func (w *WaterfallGraph) MouseMoved(event *desktop.MouseEvent) {
	w.lastMousePos = event.Position
	w.updateTooltip(event.Position)
}

// Hides tooltip when mouse leaves widget
func (w *WaterfallGraph) MouseOut() {
	w.mouseIn = false
	w.tooltipBg.Hide()
	w.tooltipText.Hide()
	w.Refresh()
}

// Adds newest sample to waterfall and redraws image
func (w *WaterfallGraph) UpdateGraph(
	numbers []int,
	minCalibration, maxCalibration int,
	minFrequency, maxFrequency int,
) {
	if len(numbers) == 0 {
		return
	}

	// Used for calculating tooltip text
	// Updated every time data polled from device
	w.minCalibration = minCalibration
	w.maxCalibration = maxCalibration
	w.minFrequency = minFrequency
	w.maxFrequency = maxFrequency

	// Append newest row, drawn along bottom edge as time passes
	w.rows = append(w.rows, waterfallRow{
		values:    numbers,
		timestamp: time.Now(),
	})

	// Drop rows older than display window
	cutoff := time.Now().Add(-waterfallWindow).Add(-waterfallWindowBuffer)
	for len(w.rows) > 0 && w.rows[0].timestamp.Before(cutoff) {
		w.rows = w.rows[1:]
	}

	// Create blank image
	img := newEmptyImage(w.graphWidth, w.graphHeight, color.Black)

	now := time.Now()

	// Draw each row
	for i, row := range w.rows {
		// Top of row based on difference between current and row times
		yTop := w.elapsedToY(now.Sub(row.timestamp))

		// Bottom of row based on being last row or not
		yBottom := w.graphHeight
		if i+1 < len(w.rows) {
			yBottom = w.elapsedToY(now.Sub(w.rows[i+1].timestamp))
		}

		// Ignore too narrow rows
		if yBottom <= yTop {
			continue
		}

		// Draw each bar in row
		for x := range w.graphWidth {
			// Convert x position to value bin
			bin := x * len(row.values) / w.graphWidth
			value := row.values[bin]

			// Calculate rssi intensity
			intensity := mapClamped(value, minCalibration, maxCalibration, 0, 255)
			pixelColour := waterfallColour(intensity)

			// Fill in pixels
			for y := yTop; y < yBottom; y++ {
				img.Set(x, y, pixelColour)
			}
		}
	}

	w.graphCanvas.Image = img
	w.Refresh()

	// Update tooltip if mouse still inside
	if w.mouseIn {
		w.updateTooltip(w.lastMousePos)
	}
}

// Updates tooltip position and text
// TODO: Fix reliance on time.Now() as y position breaks
// when tooltip updating by graph not
func (w *WaterfallGraph) updateTooltip(localPos fyne.Position) {
	// Get drawn graph dimensions
	drawSize := w.graphCanvas.Size()
	if drawSize.Width <= 0 || drawSize.Height <= 0 {
		return
	}

	// Need at least one row of history to look up
	if len(w.rows) == 0 {
		return
	}

	// Scale mouse y position into graph height space used by elapsedToY
	displayHeight := int(drawSize.Height)
	mouseY := int(localPos.Y)
	if mouseY == displayHeight {
		mouseY--
	}
	scaledY := (mouseY * w.graphHeight) / displayHeight

	// Find row drawn at this vertical position
	// Search newest to oldest (end of array to beginning)
	now := time.Now()
	var hoveredRow *waterfallRow
	for i := len(w.rows) - 1; i >= 0; i-- {
		// Calculate bounds of current row
		yTop := w.elapsedToY(now.Sub(w.rows[i].timestamp))
		yBottom := w.graphHeight
		if i+1 < len(w.rows) {
			yBottom = w.elapsedToY(now.Sub(w.rows[i+1].timestamp))
		}

		// Check scaled mouse y inside current row bounds
		if scaledY >= yTop && scaledY < yBottom {
			hoveredRow = &w.rows[i]
			break
		}
	}
	if hoveredRow == nil {
		return
	}

	// Calculate number of bars over from 0
	// Use hovered row's own length to prevent out of bounds
	barCount := len(hoveredRow.values)
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
	freqRange := w.maxFrequency - w.minFrequency
	denom := barCount - 1
	num := barsOver * freqRange
	frequency := (num+denom/2)/denom + w.minFrequency

	// Calculate signal strength and time since row captured
	rssi := hoveredRow.values[barsOver]
	rssiStrength := mapClamped(rssi, w.minCalibration, w.maxCalibration, 0, 100)
	elapsedSeconds := int(now.Sub(hoveredRow.timestamp).Round(time.Second).Seconds())

	// Format tooltip text
	w.tooltipText.Text = fmt.Sprintf("%dMHz, %ds ago, %d%%", frequency, elapsedSeconds, rssiStrength)

	// Get proper tooltip sizing
	padding := float32(6)
	offset := float32(12)
	textSize := w.tooltipText.MinSize()
	bgSize := fyne.NewSize(textSize.Width+padding*2, textSize.Height+padding*2)

	// Put tooltip in bottom right corner of cursor
	tx := localPos.X + offset
	ty := localPos.Y + offset

	// Flip position horizontally
	if tx+bgSize.Width > w.Size().Width {
		tx = localPos.X - bgSize.Width
	}

	// Flip position vertically
	if ty+bgSize.Height > w.Size().Height {
		ty = localPos.Y - bgSize.Height
	}

	// Move tooltip
	w.tooltipBg.Move(fyne.NewPos(tx, ty))
	w.tooltipBg.Resize(bgSize)
	w.tooltipText.Move(fyne.NewPos(tx+padding, ty+padding))

	// Show tooltip
	w.tooltipBg.Show()
	w.tooltipText.Show()
	w.Refresh()
}

// Converts elapsed duration into vertical pixel position
// Bottow edge is now
func (w *WaterfallGraph) elapsedToY(elapsed time.Duration) int {
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed > waterfallWindow {
		elapsed = waterfallWindow
	}

	pixelsFromBottom := int(elapsed) * w.graphHeight / int(waterfallWindow)
	return w.graphHeight - pixelsFromBottom
}

// Returns new renderer for WaterfallGraph
func (w *WaterfallGraph) CreateRenderer() fyne.WidgetRenderer {
	return &waterfallGraphRenderer{waterfallGraph: w}
}

// Renderer for WaterfallGraph widget
type waterfallGraphRenderer struct {
	waterfallGraph *WaterfallGraph
}

// Returns minimum size of WaterfallGraph
func (r *waterfallGraphRenderer) MinSize() fyne.Size {
	return fyne.NewSize(500, 200)
}

// Lays out image to fill WaterfallGraph
func (r *waterfallGraphRenderer) Layout(size fyne.Size) {
	r.waterfallGraph.graphCanvas.Resize(size)
}

// Refreshes WaterfallGraph
func (r *waterfallGraphRenderer) Refresh() {
	r.waterfallGraph.graphCanvas.Refresh()
	r.waterfallGraph.tooltipBg.Refresh()
	r.waterfallGraph.tooltipText.Refresh()
}

// Returns child widgets of WaterfallGraph
func (r *waterfallGraphRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		r.waterfallGraph.graphCanvas,
		r.waterfallGraph.tooltipBg,
		r.waterfallGraph.tooltipText,
	}
}

// Does nothing as WaterfallGraph doesn't hold external resources
func (r *waterfallGraphRenderer) Destroy() {}
