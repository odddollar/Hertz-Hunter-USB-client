package ui

import (
	"Hertz-Hunter-USB-Client/widgets"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Use Fyne-X extensions to create about window
func (u *Ui) showAbout() {
	// Parse urls for documentation
	documentation, _ := url.Parse("https://github.com/odddollar/Hertz-hunter")
	clientDocumentation, _ := url.Parse("https://github.com/odddollar/Hertz-hunter-usb-client")

	links := []*widget.Hyperlink{
		widget.NewHyperlink("Hertz Hunter Documentation", documentation),
		widget.NewHyperlink("Hertz Hunter USB Client Documentation", clientDocumentation),
	}

	// Markdown program description
	content := "USB serial client for the **Hertz Hunter** spectrum analyser"

	// Use Fyne-X's about dialog
	d := widgets.NewAbout(content, links, u.batteryVoltage, u.a, u.w)
	d.Resize(fyne.NewSize(0, 424)) // 0 width as will always be at least content's min-width
	d.Show()
}

// Standard dialog to show error
func (u *Ui) showError(err error) {
	fyne.Do(func() { dialog.ShowError(err, u.w) })
}

// Standard dialog to show success
func (u *Ui) showSuccess(message string) {
	fyne.Do(func() { dialog.ShowInformation("Success", message, u.w) })
}
