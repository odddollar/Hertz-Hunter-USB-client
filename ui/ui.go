package ui

import (
	"Hertz-Hunter-USB-client/schema"
	"Hertz-Hunter-USB-client/widgets"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Ui struct {
	// Main app elements
	a fyne.App
	w fyne.Window

	// Main ui components
	titleLabel                      *canvas.Text
	aboutButton                     *widget.Button
	switchBandButton                *widget.Button
	spectrumGraph                   *widgets.RssiGraph
	spectrumLeftRssiLabels          *fyne.Container
	spectrumRightRssiLabels         *fyne.Container
	spectrumHighbandFrequencyLabels *fyne.Container
	spectrumLowbandFrequencyLabels  *fyne.Container

	// Connection ui components
	portsSelect                *widget.Select
	portsRefreshButton         *widget.Button
	baudRateSelect             *widget.Select
	maxComRetriesSelect        *widget.Select
	maxComRetriesHelp          *widget.Button
	graphRefreshIntervalSelect *widget.Select
	graphRefreshIntervalHelp   *widget.Button
	connectButton              *widget.Button
	disconnectButton           *widget.Button

	// Settings ui components
	scanIntervalSelect *widget.Select
	buzzerSelect       *widget.Select
	batteryAlarmSelect *widget.Select
	settingsSetButton  *widget.Button

	// Calibration ui components
	highRssiCalibrationEntry *widget.Entry
	lowRssiCalibrationEntry  *widget.Entry
	calibrationSetButton     *widget.Button

	// Stores config items
	configAccordion *widget.Accordion

	// Store band state
	lowband bool

	// Store current calibration values
	highRssiCalibration int
	lowRssiCalibration  int

	// Stores battery state
	batteryEnabled bool
	batteryVoltage binding.Float

	// Global schema object store
	schema *schema.Schema
}

// Create new ui layout
func (u *Ui) NewUI() {
	// Create window
	u.a = app.New()
	u.w = u.a.NewWindow("Hertz Hunter USB Client")

	// Create title widget
	u.titleLabel = canvas.NewText("Hertz Hunter USB Client", theme.Color(theme.ColorNameForeground))
	u.titleLabel.Alignment = fyne.TextAlignCenter
	u.titleLabel.TextStyle.Bold = true
	u.titleLabel.TextSize = 20

	// Create about button
	u.aboutButton = widget.NewButtonWithIcon("", theme.InfoIcon(), u.showAbout)

	// Create switch band button
	u.switchBandButton = widget.NewButton("Switch Band", u.switchBand)

	// Create spectrum graph display area
	u.spectrumGraph = widgets.NewRssiGraph(GRAPH_WIDTH, GRAPH_HEIGHT)

	// Create rssi labels
	u.spectrumLeftRssiLabels = newRssiScale(fyne.TextAlignTrailing)
	u.spectrumRightRssiLabels = newRssiScale(fyne.TextAlignLeading)

	// Create highband labels
	u.spectrumHighbandFrequencyLabels = newFrequencyScale("5645MHz", "5795MHz", "5945MHz")

	// Create lowband labels
	u.spectrumLowbandFrequencyLabels = newFrequencyScale("5345MHz", "5495MHz", "5645MHz")
	u.spectrumLowbandFrequencyLabels.Hide()

	// Create port selection dropdown with serial ports
	u.portsSelect = widget.NewSelect([]string{}, func(s string) {})

	// Create refresh ports button
	u.portsRefreshButton = widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), u.refreshPortsDisplay)

	// Create baudrate dropdown
	u.baudRateSelect = widget.NewSelect(intsToStrings(BAUD_RATES), func(s string) {})
	u.baudRateSelect.SetSelected(fmt.Sprint(DEFAULT_BAUD_RATE))

	// Create communication retries dropdown
	u.maxComRetriesSelect = widget.NewSelect(intsToStrings(COMMUNICATION_RETRIES), func(s string) {})
	u.maxComRetriesSelect.SetSelected(fmt.Sprint(DEFAULT_COMMUNICATION_RETRIES))
	u.maxComRetriesHelp = widget.NewButtonWithIcon("", theme.InfoIcon(), u.showMaxComRetriesHelp)

	// Create refresh graph dropdown
	u.graphRefreshIntervalSelect = widget.NewSelect(durationsToStrings(REFRESH_INTERVALS), func(s string) {})
	u.graphRefreshIntervalSelect.SetSelected(fmt.Sprintf("%.2gs", DEFAULT_REFRESH_INTERVAL.Seconds()))
	u.graphRefreshIntervalHelp = widget.NewButtonWithIcon("", theme.InfoIcon(), u.showGraphRefreshIntervalHelp)

	// Create connect button
	u.connectButton = widget.NewButton("Connect", func() { go u.connectUSBSerial() })
	u.connectButton.Importance = widget.HighImportance

	// Create disconnect button
	u.disconnectButton = widget.NewButton("Disconnect", func() { go u.disconnectUSBSerial() })
	u.disconnectButton.Hide()

	// Create container for connection items
	connectionContainer := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Serial Port", container.NewBorder(
				nil, nil, nil,
				u.portsRefreshButton,
				u.portsSelect,
			)),
			widget.NewFormItem("Baud Rate", u.baudRateSelect),
			widget.NewFormItem("Max Communication Retries", container.NewBorder(
				nil, nil, nil,
				u.maxComRetriesHelp,
				u.maxComRetriesSelect,
			)),
			widget.NewFormItem("Graph Refresh Interval", container.NewBorder(
				nil, nil, nil,
				u.graphRefreshIntervalHelp,
				u.graphRefreshIntervalSelect,
			)),
		),
		u.connectButton,
		u.disconnectButton,
	)

	// Create selects for settings
	u.scanIntervalSelect = widget.NewSelect(SCAN_INTERVALS, func(s string) {})
	u.buzzerSelect = widget.NewSelect(BUZZER_STATES, func(s string) {})
	u.batteryAlarmSelect = widget.NewSelect(BATTERY_ALARM_THRESHOLDS, func(s string) {})

	// Create settings set button
	u.settingsSetButton = widget.NewButton("Set", func() { go u.setSettingsIndices() })
	u.settingsSetButton.Importance = widget.HighImportance

	// Create container for settings items
	settingsContainer := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Scan Interval", u.scanIntervalSelect),
			widget.NewFormItem("Buzzer", u.buzzerSelect),
			widget.NewFormItem("Battery Alarm Threshold", u.batteryAlarmSelect),
		),
		u.settingsSetButton,
	)

	// Create entries for calibration rssi
	u.highRssiCalibrationEntry = widget.NewEntry()
	u.highRssiCalibrationEntry.Validator = validation.NewRegexp(
		`^(?:[0-9]|[1-9][0-9]{1,2}|[1-3][0-9]{3}|40[0-8][0-9]|409[0-5])$`,
		"Must be integer between 0 and 4095 inclusive",
	)
	u.highRssiCalibrationEntry.OnChanged = func(s string) {
		u.validateCalibrationEntries()
		u.configAccordion.Refresh()
	}
	u.lowRssiCalibrationEntry = widget.NewEntry()
	u.lowRssiCalibrationEntry.Validator = validation.NewRegexp(
		`^(?:[0-9]|[1-9][0-9]{1,2}|[1-3][0-9]{3}|40[0-8][0-9]|409[0-5])$`,
		"Must be integer between 0 and 4095 inclusive",
	)
	u.lowRssiCalibrationEntry.OnChanged = func(s string) {
		u.validateCalibrationEntries()
		u.configAccordion.Refresh()
	}

	// Create set button for calibration
	u.calibrationSetButton = widget.NewButton("Set", func() { go u.setCalibrationValues() })
	u.calibrationSetButton.Importance = widget.HighImportance

	// Create container for calibration items
	calibrationContainer := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("High Value", u.highRssiCalibrationEntry),
			widget.NewFormItem("Low Value", u.lowRssiCalibrationEntry),
		),
		u.calibrationSetButton,
	)

	// Create accordion for configuration items
	u.configAccordion = widget.NewAccordion(
		widget.NewAccordionItem("Connection", connectionContainer),
		widget.NewAccordionItem("Settings", settingsContainer),
		widget.NewAccordionItem("Calibration", calibrationContainer),
	)
	u.configAccordion.MultiOpen = true
	u.configAccordion.Open(0)

	// Create top elements group
	topElementsGroup := container.NewVBox(
		container.NewBorder(
			nil,
			nil,
			widgets.NewSpacer(widget.NewButtonWithIcon("", theme.InfoIcon(), func() {}).MinSize()), // Keeps title centred
			u.aboutButton,
			u.titleLabel,
		),
		u.configAccordion,
		u.switchBandButton,
	)

	// Create spectrum graph group
	spectrumGraphGroup := container.NewBorder(
		widgets.NewSpacer(fyne.NewSize(0, 0)), // Funky spacing hack
		container.NewVBox(
			u.spectrumHighbandFrequencyLabels,
			u.spectrumLowbandFrequencyLabels,
		),
		u.spectrumLeftRssiLabels,
		u.spectrumRightRssiLabels,
		u.spectrumGraph,
	)

	// Create waterfall graph group
	waterfallGraphGroup := container.NewBorder(
		nil, nil, nil, nil,
		widget.NewLabel("Placeholder"),
	)

	// Create window layout and set content
	u.w.SetContent(container.NewBorder(
		topElementsGroup,
		nil, nil, nil,
		container.NewAppTabs(
			container.NewTabItemWithIcon("Spectrum", theme.StorageIcon(), spectrumGraphGroup),
			container.NewTabItemWithIcon("Waterfall", theme.ColorChromaticIcon(), waterfallGraphGroup),
		),
	))

	// Initial refresh of available ports
	u.refreshPortsDisplay()

	// Disable settings elements as not connected
	u.disableSettingsUi()

	// Create battery voltage binding
	u.batteryVoltage = binding.NewFloat()
}

// Show and run app
func (u *Ui) Run() {
	u.w.Resize(fyne.NewSize(800, 0))
	u.w.Show()
	u.a.Run()
}

// Disable ui elements related to connection
func (u *Ui) disableConnectionUi() {
	fyne.Do(func() {
		u.portsSelect.Disable()
		u.portsRefreshButton.Disable()
		u.baudRateSelect.Disable()
		u.maxComRetriesSelect.Disable()
		u.graphRefreshIntervalSelect.Disable()
		u.connectButton.Disable()
	})
}

// Enable ui elements related to connection
func (u *Ui) enableConnectionUi() {
	fyne.Do(func() {
		u.portsSelect.Enable()
		u.portsRefreshButton.Enable()
		u.baudRateSelect.Enable()
		u.maxComRetriesSelect.Enable()
		u.graphRefreshIntervalSelect.Enable()
		u.connectButton.Enable()
	})
}

// Disable ui elements related to settings
func (u *Ui) disableSettingsUi() {
	fyne.Do(func() {
		u.scanIntervalSelect.Disable()
		u.buzzerSelect.Disable()
		u.batteryAlarmSelect.Disable()
		u.settingsSetButton.Disable()
		u.highRssiCalibrationEntry.Disable()
		u.lowRssiCalibrationEntry.Disable()
		u.calibrationSetButton.Disable()
		u.switchBandButton.Disable()
	})
}

// Enable ui elements related to settings
func (u *Ui) enableSettingsUi() {
	fyne.Do(func() {
		u.scanIntervalSelect.Enable()
		u.buzzerSelect.Enable()
		if u.batteryEnabled {
			u.batteryAlarmSelect.Enable()
		}
		u.settingsSetButton.Enable()
		u.highRssiCalibrationEntry.Enable()
		u.lowRssiCalibrationEntry.Enable()
		u.calibrationSetButton.Enable()
		u.switchBandButton.Enable()
	})
}

// Show connect button
func (u *Ui) showConnectButton() {
	fyne.Do(func() {
		u.connectButton.Show()
		u.disconnectButton.Hide()
	})
}

// Show disconnect button
func (u *Ui) showDisconnectButton() {
	fyne.Do(func() {
		u.connectButton.Hide()
		u.disconnectButton.Show()
	})
}

// Switch which band labels are visible
func (u *Ui) switchBandLabels(lowband bool) {
	if lowband {
		fyne.Do(func() {
			u.spectrumHighbandFrequencyLabels.Hide()
			u.spectrumLowbandFrequencyLabels.Show()
		})
	} else {
		fyne.Do(func() {
			u.spectrumHighbandFrequencyLabels.Show()
			u.spectrumLowbandFrequencyLabels.Hide()
		})
	}
}

// Disables calibration set button if entries are invalid
func (u *Ui) validateCalibrationEntries() {
	highErr := u.highRssiCalibrationEntry.Validate()
	lowErr := u.lowRssiCalibrationEntry.Validate()

	if highErr == nil && lowErr == nil {
		fyne.Do(func() { u.calibrationSetButton.Enable() })
	} else {
		fyne.Do(func() { u.calibrationSetButton.Disable() })
	}
}

// Update values in calibration entries
func (u *Ui) updateCalibrationEntries() {
	fyne.Do(func() {
		// Fill calibration entries with values
		u.highRssiCalibrationEntry.SetText(fmt.Sprint(u.highRssiCalibration))
		u.lowRssiCalibrationEntry.SetText(fmt.Sprint(u.lowRssiCalibration))
	})
}

// Updates selected entries for settings
func (u *Ui) updateSettingsIndices(scan_interval_index, buzzer_index, battery_alarm_index int) {
	fyne.Do(func() {
		u.scanIntervalSelect.SetSelectedIndex(scan_interval_index)
		u.buzzerSelect.SetSelectedIndex(buzzer_index)
		u.batteryAlarmSelect.SetSelectedIndex(battery_alarm_index)
	})
}
