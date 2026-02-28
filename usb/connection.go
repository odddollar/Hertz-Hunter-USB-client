// DISCLAIMER
//
// The connection/communication code here, particularly the Communicate() and receive()
// functions, operates on a "works" rather than "is correct" basis. It's hacky, but it's
// the best I could come up with.

package usb

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.bug.st/serial"
)

var recentWait time.Duration

// Format of serial frame
type SerialFrame struct {
	Event    string         `json:"event"`
	Location string         `json:"location"`
	Payload  map[string]any `json:"payload"`
}

// Handles messaging and connection lifetime
type Connection struct {
	port serial.Port

	mu sync.Mutex
}

// Create new connection object
func NewConnection(portName string, baud int) (*Connection, error) {
	// Setup baud and don't reset
	mode := &serial.Mode{
		BaudRate: baud,
		InitialStatusBits: &serial.ModemOutputBits{
			DTR: false,
			RTS: false,
		},
	}

	// Connect to serial port
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}
	port.SetReadTimeout(50 * time.Millisecond)

	// Re-assert to not reset on connection
	port.SetDTR(false)
	port.SetRTS(false)

	// Create connection object with port
	connection := Connection{
		port: port,
	}

	// Check if connection succeeded
	if _, err := connection.ping(); err != nil {
		connection.Disconnect() // Close serial port to prevent future "Port busy" errors
		return nil, err
	}

	return &connection, nil
}

// Disconnect connection
func (c *Connection) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Purge and close connection stream
	c.port.ResetInputBuffer()
	c.port.ResetOutputBuffer()
	c.port.Close()
}

// Send message frame over serial and return response
func (c *Connection) Communicate(msg SerialFrame) (SerialFrame, error) {
	// Lock mutex so only one message in transit at time
	c.mu.Lock()
	defer c.mu.Unlock()

	var lastErr error
	maxAttempts := 2

	for i := range maxAttempts {
		// Drain serial to start buffer fresh
		c.port.ResetInputBuffer()
		c.port.ResetOutputBuffer()

		// Send message
		if err := c.send(msg); err != nil {
			lastErr = err
			fmt.Printf("Retrying send: %d\n", i)
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// Read response
		rec, err := c.receive()
		if err != nil {
			lastErr = err
			fmt.Printf("Retrying receive: %d, %s, %s\n", i, err, recentWait)
			time.Sleep(50 * time.Millisecond)
			continue
		}

		return rec, nil
	}

	return SerialFrame{}, lastErr
}

// Send message
func (c *Connection) send(msg SerialFrame) error {
	// Marshall struct
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Ensure ends with newline
	data = append(data, '\n')

	// Write data to port
	_, err = c.port.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// Receive message
func (c *Connection) receive() (SerialFrame, error) {
	// Set overall message deadline
	// Port timeout used for single read() calls
	deadline := time.Now().Add(500 * time.Millisecond)
	startTime := time.Now()
	defer func() {
		recentWait = time.Since(startTime)
	}()

	// Buffers for reading data from serial
	buffer := []byte{}
	tmp := make([]byte, 1024)

	readStarted := false

	// Retry receive message until deadline
	for time.Now().Before(deadline) {
		// Read bytes to tmp buffer
		nBytes, err := c.port.Read(tmp)
		if err != nil {
			return SerialFrame{}, err
		}
		if nBytes == 0 {
			continue
		}

		// Process each received byte
		for i := range nBytes {
			b := tmp[i]

			// Wait for start of frame
			if !readStarted {
				if b == '{' {
					readStarted = true
					buffer = append(buffer, b)
				}
				continue
			}

			// Process full frame when newline received
			if b == '\n' {
				var msg SerialFrame
				if err := json.Unmarshal(buffer, &msg); err != nil {
					fmt.Println(string(buffer))
					return SerialFrame{}, err
				}
				if msg.Event == "error" {
					// Safe type assertion
					if status, ok := msg.Payload["status"].(string); ok {
						return SerialFrame{}, errors.New(status)
					}
					return SerialFrame{}, errors.New("unknown device error")
				}

				return msg, nil
			}

			// Accumulate bytes
			buffer = append(buffer, b)
		}
	}

	return SerialFrame{}, errors.New("serial receive timeout")
}

// Checks if the serial port is connected with ping messages
func (c *Connection) ping() (bool, error) {
	// Send ping message
	if _, err := c.Communicate(SerialFrame{Event: "get", Location: "ping", Payload: map[string]any{}}); err != nil {
		return false, err
	}

	return true, nil
}
