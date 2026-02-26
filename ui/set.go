package ui

import (
	"strconv"
)

// Callback for set settings button
func (u *Ui) setSettingsIndices() {
	// Get values from dropdowns
	scan_interval_index := u.scanIntervalSelect.SelectedIndex()
	buzzer_index := u.buzzerSelect.SelectedIndex()
	battery_alarm_index := u.batteryAlarmSelect.SelectedIndex()

	// Set indices
	err := u.schema.SetSettingsIndices(scan_interval_index, buzzer_index, battery_alarm_index, u.batteryEnabled)
	if err != nil {
		u.showError(err)
		return
	}

	u.showSuccess("Settings set")
}

// Callback for set calibration button
func (u *Ui) setCalibrationValues() {
	// Get values from entries
	high, _ := strconv.Atoi(u.highRssiCalibrationEntry.Text)
	low, _ := strconv.Atoi(u.lowRssiCalibrationEntry.Text)

	// Set values
	err := u.schema.SetCalibratedValues(low, high)
	if err != nil {
		u.showError(err)
		return
	}

	// Update store of calibration values for graph scaling only after values are known-good
	u.highRssiCalibration = high
	u.lowRssiCalibration = low

	u.showSuccess("Calibration set")
}

// Callback for switch band button
func (u *Ui) switchBand() {
	// Toggle band
	u.lowband = !u.lowband

	// Send to device
	err := u.schema.SetBand(u.lowband)
	if err != nil {
		u.showError(err)
		return
	}
}
