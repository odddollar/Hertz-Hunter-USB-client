package schema

import (
	"Hertz-Hunter-USB-Client/usb"
	"context"
)

// Handles messaging device with required schema
type Schema struct {
	connection    *usb.Connection
	pollingCancel context.CancelFunc
}

// Used to return lowband state with values on polling
type ValuesResult struct {
	Values         []int
	Lowband        bool
	MinFrequency   int
	MaxFrequency   int
	BatteryVoltage float64
}

// Create new schema object
func NewSchema(portName string, baud int) (*Schema, error) {
	con, err := usb.NewConnection(portName, baud)
	if err != nil {
		return nil, err
	}

	return &Schema{
		connection: con,
	}, nil
}

// Cancels rssi polling and disconnects device
func (s *Schema) Stop() {
	if s.pollingCancel != nil {
		s.pollingCancel()
		s.pollingCancel = nil
	}

	s.connection.Disconnect()
}
