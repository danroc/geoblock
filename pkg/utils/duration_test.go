package utils_test

import (
	"testing"
	"time"

	"github.com/danroc/geoblock/pkg/utils"
)

func TestParseDuration(t *testing.T) {
	// Test valid durations
	validDurations := map[string]time.Duration{
		"0s":                  0,
		"1ms 1us 1ns":         1*time.Millisecond + 1*time.Microsecond + 1*time.Nanosecond,
		"1ns 1ms 1us":         1*time.Millisecond + 1*time.Microsecond + 1*time.Nanosecond,
		"10s":                 10 * time.Second,
		"1m30s":               90 * time.Second,
		"2m":                  2 * time.Minute,
		"03h":                 3 * time.Hour,
		"3h30m":               3*time.Hour + 30*time.Minute,
		"0h30m":               30 * time.Minute,
		" 1h 30m ":            90 * time.Minute,
		"4d":                  4 * utils.TimeDay,
		"1w":                  1 * utils.TimeWeek,
		"1M":                  1 * utils.TimeMonth,
		"1y":                  1 * utils.TimeYear,
		"1 minute":            1 * time.Minute,
		"2 hours":             2 * time.Hour,
		"1 day":               1 * utils.TimeDay,
		"3 days":              3 * utils.TimeDay,
		"1 minute 30 seconds": 90 * time.Second,
	}

	for input, expected := range validDurations {
		duration, err := utils.ParseDuration(input)
		if err != nil {
			t.Errorf("Unexpected error for input '%s': %v", input, err)
		}
		if duration != expected {
			t.Errorf(
				"Incorrect duration for input '%s'. Expected: %v, Got: %v",
				input,
				expected,
				duration,
			)
		}
	}

	// Test invalid durations
	invalidDurations := []string{
		"-1s",
		"1x",
		"1.5h",
		"1.5",
		"1",
		"x",
		"1m 30",
		"1m x",
		"1m s",
	}

	for _, input := range invalidDurations {
		_, err := utils.ParseDuration(input)
		if err == nil {
			t.Errorf("Expected error for invalid input '%s', but got no error", input)
		}
	}
}
