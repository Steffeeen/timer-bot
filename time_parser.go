package main

import (
	"time"

	"github.com/markusmobius/go-dateparser"
)

func parseTime(timeStr string) (time.Time, error) {
	// Try standard Go time formats first for well-formed ISO dates
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		time.RFC1123,
		time.RFC1123Z,
	}

	for _, format := range formats {
		parsedTime, parseErr := time.Parse(format, timeStr)
		if parseErr == nil {
			// Convert to local timezone
			return parsedTime.Local(), nil
		}
	}

	// If standard formats fail, use dateparser for natural language
	configuration := &dateparser.Configuration{
		CurrentTime:         time.Now(),
		DateOrder:           dateparser.DMY,
		PreferredDateSource: dateparser.Future,
	}

	date, err := dateparser.Parse(configuration, timeStr)

	if err != nil {
		return time.Time{}, err
	}

	// If the parsed time is in the past, try adding "in" prefix
	if time.Now().After(date.Time) {
		dateWithIn, errIn := dateparser.Parse(configuration, "in "+timeStr)
		if errIn == nil && time.Now().Before(dateWithIn.Time) {
			// Convert to local timezone
			return dateWithIn.Time.Local(), nil
		}
	}

	// If still in the past, try adding "on" prefix for weekday names
	if time.Now().After(date.Time) {
		dateWithOn, errOn := dateparser.Parse(configuration, "on "+timeStr)
		if errOn == nil && time.Now().Before(dateWithOn.Time) {
			// Convert to local timezone
			return dateWithOn.Time.Local(), nil
		}
	}

	// If still in the past, try "next" prefix for weekday names
	if time.Now().After(date.Time) {
		dateWithNext, errNext := dateparser.Parse(configuration, "next "+timeStr)
		if errNext == nil && time.Now().Before(dateWithNext.Time) {
			// Convert to local timezone
			return dateWithNext.Time.Local(), nil
		}
	}

	// Convert to local timezone before returning
	return date.Time.Local(), err
}
