package schema

import (
	"Hertz-Hunter-USB-Client/usb"
	"context"
	"time"
)

// Start polling for rssi values from device
func (s *Schema) StartPollValues(period time.Duration, pollBattery bool) (<-chan ValuesResult, <-chan error) {
	ctx, cancel := context.WithCancel(context.Background())
	s.pollingCancel = cancel

	valuesCh := make(chan ValuesResult)
	errCh := make(chan error, 1)

	go func() {
		defer close(valuesCh)
		defer close(errCh)

		// Start ticker
		ticker := time.NewTicker(period)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Get values data
				data, err := s.connection.Communicate(usb.SerialFrame{Event: "get", Location: "values", Payload: map[string]any{}})
				if err != nil {
					errCh <- err
					return
				}

				// Get battery voltage
				var batteryVoltage float64
				if pollBattery {
					batteryVoltage, err = s.GetBatteryVoltage()
					if err != nil {
						errCh <- err
						return
					}
				}

				// Convert data to proper type
				raw, _ := data.Payload["values"].([]any)
				values := make([]int, len(raw))
				for i, v := range raw {
					f, _ := v.(float64)
					values[i] = int(f)
				}

				// Get lowband state
				lowband, _ := data.Payload["lowband"].(bool)

				// Get min and max frequency
				minFrequency, _ := data.Payload["min_frequency"].(float64)
				maxFrequency, _ := data.Payload["max_frequency"].(float64)

				// Send data over channel
				valuesCh <- ValuesResult{
					Values:         values,
					Lowband:        lowband,
					MinFrequency:   int(minFrequency),
					MaxFrequency:   int(maxFrequency),
					BatteryVoltage: batteryVoltage,
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return valuesCh, errCh
}
