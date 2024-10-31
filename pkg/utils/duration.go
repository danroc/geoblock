package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Some common time durations. Note that the durations for months and years are
// approximate and do not take into account leap years or months with different
// numbers of days.
const (
	TimeDay   time.Duration = 24 * time.Hour
	TimeWeek                = 7 * TimeDay
	TimeMonth               = 30 * TimeDay
	TimeYear                = 365 * TimeDay
)

var (
	reDurationSegment = regexp.MustCompile(`(\d+)([^\d\s]+)`)
	reDurationString  = regexp.MustCompile(`^(\s*(\d+)\s*([^\d\s]+)\s*)+$`)
)

// ParseDuration parses a string representation of a duration and returns the
// corresponding time.Duration value.
//
// The string should be in the format of a number followed by a unit, such as
// "5s" for 5 seconds or "2h" for 2 hours. It can also contain multiple
// durations that will be added together, such as "1h30m" for 1 hour and 30
// minutes.
//
// Valid units are: "ns" (nanoseconds), "us" (microseconds), "ms"
// (milliseconds), "s" (seconds), "m" (minutes), "h" (hours), "d" (days), "w"
// (weeks), "M" (months), and "y" (years). If the string is not in a valid
// format, an error is returned.
func ParseDuration(s string) (time.Duration, error) {
	// This check rejects negative values or invalid durations combined with
	// valid ones. E.g. "-1s", "1m 30" or "1m s".
	if !reDurationString.MatchString(s) {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}

	matches := reDurationSegment.FindAllStringSubmatch(
		strings.ReplaceAll(s, " ", ""), -1,
	)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}

	totalDuration := time.Duration(0)
	for _, match := range matches {
		duration, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return 0, err
		}

		unit := match[2]
		switch unit {
		case "ns", "nanosecond", "nanoseconds":
			totalDuration += time.Duration(duration) * time.Nanosecond
		case "us", "microsecond", "microseconds":
			totalDuration += time.Duration(duration) * time.Microsecond
		case "ms", "millisecond", "milliseconds":
			totalDuration += time.Duration(duration) * time.Millisecond
		case "s", "second", "seconds":
			totalDuration += time.Duration(duration) * time.Second
		case "m", "minute", "minutes":
			totalDuration += time.Duration(duration) * time.Minute
		case "h", "hour", "hours":
			totalDuration += time.Duration(duration) * time.Hour
		case "d", "day", "days":
			totalDuration += time.Duration(duration) * TimeDay
		case "w", "week", "weeks":
			totalDuration += time.Duration(duration) * TimeWeek
		case "M", "month", "months":
			totalDuration += time.Duration(duration) * TimeMonth
		case "y", "year", "years":
			totalDuration += time.Duration(duration) * TimeYear
		default:
			return 0, fmt.Errorf("invalid duration unit: %s", unit)
		}
	}

	return totalDuration, nil
}

// IsDuration checks if the given string represents a valid duration. It
// returns true if the string matches the duration format, otherwise false.
func IsDuration(s string) bool {
	_, err := ParseDuration(s)
	return err == nil
}
