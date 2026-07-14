package widgets

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

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
