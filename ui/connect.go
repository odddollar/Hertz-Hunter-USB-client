package ui

import (
	"Hertz-Hunter-USB-Client/schema"
	"errors"

	"fyne.io/fyne/v2"
)

// Attempt to connect to usb serial
func (u *Ui) connectUSBSerial() {
	// Error if no port selected
	if u.portsSelect.SelectedIndex() == -1 {
		u.showError(errors.New("port must be selected"))
		return
	}

	// Disable ui elements whilst attempting connection
	u.disableConnectionUi()

	// Get port
	portName := u.portsSelect.Selected

	// Get baud rate
	baudRate := BAUDRATES[u.baudrateSelect.SelectedIndex()]

	// Get poll rate
	pollRate := REFRESH_INTERVALS[u.graphRefreshIntervalSelect.SelectedIndex()]

	// Create new schema
	var err error
	u.schema, err = schema.NewSchema(portName, baudRate)
	if err != nil {
		u.enableConnectionUi()
		u.showError(err)
		return
	}

	// Switch which connection button is visible
	u.showDisconnectButton()

	// Get battery enabled state and voltage
	u.batteryEnabled = true
	voltage, err := u.schema.GetBatteryVoltage()
	if err != nil {
		if err.Error() == "'location' must be 'values', 'settings', 'calibration', or 'ping'" {
			u.batteryEnabled = false
		} else {
			u.connectionError(err)
			return
		}
	}
	u.batteryVoltage.Set(voltage)

	// Enable settings ui
	// Battery alarm setting is disabled if necessary
	u.enableSettingsUi()

	// Set to high band
	err = u.schema.SetBand(false)
	if err != nil {
		u.connectionError(err)
		return
	}

	// Get settings values
	scan_interval_index, buzzer_index, battery_alarm_index, err := u.schema.GetSettingsIndices(u.batteryEnabled)
	if err != nil {
		u.connectionError(err)
		return
	}
	u.updateSettingsIndices(scan_interval_index, buzzer_index, battery_alarm_index)

	// Get calibration values
	u.lowRssiCalibration, u.highRssiCalibration, err = u.schema.GetCalibratedValues()
	if err != nil {
		u.connectionError(err)
		return
	}

	// Update entries with calibration values
	u.updateCalibrationEntries()

	// Start polling for values
	valuesCh, errCh := u.schema.StartPollValues(pollRate, u.batteryEnabled)

	go func() {
		for {
			select {
			case values, ok := <-valuesCh: // Update graph
				if !ok {
					return
				}

				fyne.Do(func() {
					// Update image with new values
					u.graphImage.UpdateGraph(
						values.Values,
						u.lowRssiCalibration,
						u.highRssiCalibration,
						values.MinFrequency,
						values.MaxFrequency,
					)

					// Automatically switch band labels
					u.switchBandLabels(values.Lowband)

					// Update battery voltage
					u.batteryVoltage.Set(values.BatteryVoltage)
				})

			case err, ok := <-errCh: // Handle errors
				if !ok {
					return
				}

				u.connectionError(err)
				return
			}
		}
	}()
}

// Callback for disconnect button
func (u *Ui) disconnectUSBSerial() {
	// Cancel connection
	u.schema.Stop()

	// Switch ui elements
	u.enableConnectionUi()
	u.showConnectButton()
	u.disableSettingsUi()
}
