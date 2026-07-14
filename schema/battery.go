package schema

import (
	"Hertz-Hunter-USB-client/usb"
)

// Get battery voltage
func (s *Schema) GetBatteryVoltage() (float64, error) {
	frame := usb.SerialFrame{
		Event:    "get",
		Location: "battery",
		Payload:  map[string]any{},
	}

	raw, err := s.connection.Communicate(frame)
	if err != nil {
		return 0.0, err
	}

	return raw.Payload["voltage"].(float64) / 10, nil
}
