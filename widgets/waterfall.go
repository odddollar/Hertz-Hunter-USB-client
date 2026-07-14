package widgets

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// Duration of history displayed
const waterfallWindow = 30 * time.Second

// Single row of history captured at given moment
type waterfallRow struct {
	values    []int
	timestamp time.Time
}

// Custom widget that displays rssi waterfall and shows historical strength when hovered
type WaterfallGraph struct {
	widget.BaseWidget

	// Ui elements
	graphCanvas *canvas.Image

	// Constants
	graphWidth  int
	graphHeight int

	// Rows of history, oldest first, used for drawing and tooltip lookup
	rows []waterfallRow
}

// Creates new WaterfallGraph widget
func NewWaterfallGraph(graphWidth, graphHeight int) *WaterfallGraph {
	// Create graph canvas from given image
	graphCanvas := canvas.NewImageFromImage(newEmptyImage(graphWidth, graphHeight, color.Black))
	graphCanvas.FillMode = canvas.ImageFillStretch
	graphCanvas.ScaleMode = canvas.ImageScalePixels

	// Create new object
	graph := &WaterfallGraph{
		graphCanvas: graphCanvas,
		graphWidth:  graphWidth,
		graphHeight: graphHeight,
	}

	// Extend base widget and return
	graph.ExtendBaseWidget(graph)
	return graph
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

	// Append newest row, drawn along bottom edge as time passes
	w.rows = append(w.rows, waterfallRow{
		values:    numbers,
		timestamp: time.Now(),
	})

	// Drop rows older than display window
	cutoff := time.Now().Add(-waterfallWindow)
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
}

// Returns child widgets of WaterfallGraph
func (r *waterfallGraphRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		r.waterfallGraph.graphCanvas,
	}
}

// Does nothing as WaterfallGraph doesn't hold external resources
func (r *waterfallGraphRenderer) Destroy() {}
