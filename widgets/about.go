// THIS CODE WAS TAKEN AND ADAPTED FROM
// https://github.com/fyne-io/fyne-x/blob/master/dialog/about.go
// ORIGINAL COPYRIGHT AND ATRRIBUTION GOES TO THE FYNE-X PROJECT
// MODIFICATIONS TO DISPLAY BATTERY VOLTAGE ARE MY OWN

package widgets

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewAbout creates a parallax about dialog using the app metadata along with the
// markdown content and links passed into this method.
// You should call Show on the returned dialog to display it.
func NewAbout(
	content string,
	links []*widget.Hyperlink,
	battery binding.Float,
	a fyne.App,
	w fyne.Window,
) dialog.Dialog {
	d := dialog.NewCustom(
		"About",
		"OK",
		aboutContent(content, links, battery, a),
		w,
	)
	d.Resize(fyne.NewSize(400, 360))

	return d
}

func aboutContent(content string, links []*widget.Hyperlink, battery binding.Float, a fyne.App) fyne.CanvasObject {
	rich := widget.NewRichTextFromMarkdown(content)
	footer := aboutFooter(links)

	logo := canvas.NewImageFromResource(a.Metadata().Icon)
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(128, 128))

	header := widget.NewRichText()

	updateHeader := func() {
		val, err := battery.Get()
		if err != nil {
			return
		}

		header.ParseMarkdown(fmt.Sprintf(
			"## %s\n\n**Version:** %s\n\n**Battery:** %.1fv",
			a.Metadata().Name,
			a.Metadata().Version,
			val,
		))

		centerText(header)
		header.Refresh()
	}

	updateHeader()
	battery.AddListener(binding.NewDataListener(updateHeader))

	space := canvas.NewRectangle(color.Transparent)
	space.SetMinSize(fyne.NewSquareSize(theme.Padding() * 4))

	body := container.NewVBox(
		space,
		logo,
		container.NewCenter(header),
		container.NewCenter(rich),
	)

	scroll := container.NewScroll(body)

	bgColor := withAlpha(theme.Color(theme.ColorNameBackground), 0xe0)
	shadowColor := withAlpha(theme.Color(theme.ColorNameBackground), 0x33)

	underlay := canvas.NewImageFromResource(a.Metadata().Icon)
	underlay.Resize(fyne.NewSize(512, 512))

	bg := canvas.NewRectangle(bgColor)
	footerBG := canvas.NewRectangle(shadowColor)

	underlayer := underLayout{}
	slideBG := container.New(underlayer, underlay)

	scroll.OnScrolled = func(p fyne.Position) {
		underlayer.offset = -p.Y / 3
		underlayer.Layout(slideBG.Objects, slideBG.Size())
	}

	watchTheme(bg, footerBG)

	bgClip := container.NewScroll(slideBG)
	bgClip.Direction = container.ScrollNone

	return container.NewStack(
		container.New(unpad{top: true}, bgClip, bg),
		container.NewBorder(
			nil,
			container.NewStack(footerBG, footer),
			nil,
			nil,
			container.New(unpad{top: true, bottom: true}, scroll),
		),
	)
}

func aboutFooter(links []*widget.Hyperlink) fyne.CanvasObject {
	footer := container.NewHBox(layout.NewSpacer())
	for i, link := range links {
		footer.Add(link)
		if i < len(links)-1 {
			footer.Add(widget.NewLabel("-"))
		}
	}
	footer.Add(layout.NewSpacer())
	return footer
}

func centerText(rich *widget.RichText) {
	for _, s := range rich.Segments {
		if t, ok := s.(*widget.TextSegment); ok {
			t.Style.Alignment = fyne.TextAlignCenter
		}
	}
}

func watchTheme(bg, footer *canvas.Rectangle) {
	fyne.CurrentApp().Settings().AddListener(func(_ fyne.Settings) {
		bg.FillColor = withAlpha(theme.Color(theme.ColorNameBackground), 0xe0)
		bg.Refresh()

		footer.FillColor = withAlpha(theme.Color(theme.ColorNameBackground), 0x33)
		footer.Refresh()
	})
}

func withAlpha(c color.Color, alpha uint8) color.Color {
	r, g, b, _ := c.RGBA()
	return color.NRGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: alpha,
	}
}

type underLayout struct {
	offset float32
}

func (u underLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	under := objs[0]
	left := size.Width/2 - under.Size().Width/2
	under.Move(fyne.NewPos(left, u.offset-50))
}

func (u underLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.Size{}
}

type unpad struct {
	top, bottom bool
}

func (u unpad) Layout(objs []fyne.CanvasObject, s fyne.Size) {
	pad := theme.Padding()
	var pos fyne.Position
	if u.top {
		pos.Y = -pad
	}
	size := s
	if u.top {
		size.Height += pad
	}
	if u.bottom {
		size.Height += pad
	}
	for _, o := range objs {
		o.Move(pos)
		o.Resize(size)
	}
}

func (u unpad) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(100, 100)
}
