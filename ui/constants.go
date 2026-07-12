package ui

import (
	"fmt"
	"time"
)

var (
	// Baud rates used in ui dropdown
	BAUD_RATES        = []int{9600, 19200, 38400, 57600, 115200}
	DEFAULT_BAUD_RATE = 115200

	// Communication retries used in ui dropdown
	COMMUNICATION_RETRIES         = []int{1, 2, 3, 4, 5}
	DEFAULT_COMMUNICATION_RETRIES = 3

	// Graph refresh intervals used in ui dropdown
	REFRESH_INTERVALS        = []time.Duration{100 * time.Millisecond, 250 * time.Millisecond, 500 * time.Millisecond, 1 * time.Second}
	DEFAULT_REFRESH_INTERVAL = 250 * time.Millisecond

	// Scan intervals used in ui dropdown
	SCAN_INTERVALS = []string{"2.5MHz", "5MHz", "10MHz"}

	// Buzzer states used in ui dropdown
	BUZZER_STATES = []string{"On", "Off"}

	// Battery alarm thresholds used in ui dropdown
	BATTERY_ALARM_THRESHOLDS = []string{"3.6v", "3.3v", "3.0v"}
)

// Dimensions for graph images
const (
	GRAPH_WIDTH  = 998
	GRAPH_HEIGHT = 600
)

// Convert array of ints to array of strings
func intsToStrings(a []int) []string {
	out := make([]string, len(a))
	for i, v := range a {
		out[i] = fmt.Sprint(v)
	}
	return out
}

// Convert array of durations to array of formatted strings
func durationsToStrings(a []time.Duration) []string {
	out := make([]string, len(a))
	for i, d := range a {
		out[i] = fmt.Sprintf("%.2gs", d.Seconds())
	}
	return out
}
