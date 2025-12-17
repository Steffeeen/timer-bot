package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTime(t *testing.T) {
	t.Run("relative time strings should return future times", func(t *testing.T) {
		testCases := []struct {
			name        string
			input       string
			minDuration time.Duration
			maxDuration time.Duration
			description string
		}{
			{
				name:        "1 minute",
				input:       "1 min",
				minDuration: 59 * time.Second,
				maxDuration: 61 * time.Second,
				description: "should be approximately 1 minute in the future",
			},
			{
				name:        "5 minutes",
				input:       "5 min",
				minDuration: 4*time.Minute + 59*time.Second,
				maxDuration: 5*time.Minute + 1*time.Second,
				description: "should be approximately 5 minutes in the future",
			},
			{
				name:        "1 hour",
				input:       "1 hour",
				minDuration: 59 * time.Minute,
				maxDuration: 61 * time.Minute,
				description: "should be approximately 1 hour in the future",
			},
			{
				name:        "2 hours",
				input:       "2 hours",
				minDuration: 119 * time.Minute,
				maxDuration: 121 * time.Minute,
				description: "should be approximately 2 hours in the future",
			},
			{
				name:        "1 day",
				input:       "1 day",
				minDuration: 23*time.Hour + 59*time.Minute,
				maxDuration: 24*time.Hour + 1*time.Minute,
				description: "should be approximately 1 day in the future",
			},
			{
				name:        "7 days",
				input:       "7 days",
				minDuration: 7*24*time.Hour - 1*time.Minute,
				maxDuration: 7*24*time.Hour + 1*time.Minute,
				description: "should be approximately 7 days in the future",
			},
			{
				name:        "1 week",
				input:       "1 week",
				minDuration: 7*24*time.Hour - 1*time.Minute,
				maxDuration: 7*24*time.Hour + 1*time.Minute,
				description: "should be approximately 1 week in the future",
			},
			{
				name:        "30 days",
				input:       "30 days",
				minDuration: 30*24*time.Hour - 1*time.Minute,
				maxDuration: 30*24*time.Hour + 1*time.Minute,
				description: "should be approximately 30 days in the future",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				now := time.Now()
				result, err := parseTime(tc.input)

				require.NoError(t, err, "parseTime should not return an error for %q", tc.input)

				// Ensure the result is in the future
				assert.True(t, result.After(now),
					"parsed time should be in the future for %q, got %v which is %v relative to now",
					tc.input, result, result.Sub(now))

				// Check it's within the expected range
				duration := result.Sub(now)
				assert.GreaterOrEqual(t, duration, tc.minDuration,
					"%s: duration should be at least %v, got %v", tc.description, tc.minDuration, duration)
				assert.LessOrEqual(t, duration, tc.maxDuration,
					"%s: duration should be at most %v, got %v", tc.description, tc.maxDuration, duration)
			})
		}
	})

	t.Run("explicit future times with 'in' prefix", func(t *testing.T) {
		testCases := []struct {
			name        string
			input       string
			minDuration time.Duration
			maxDuration time.Duration
		}{
			{
				name:        "in 1 min",
				input:       "in 1 min",
				minDuration: 59 * time.Second,
				maxDuration: 61 * time.Second,
			},
			{
				name:        "in 5 minutes",
				input:       "in 5 minutes",
				minDuration: 4*time.Minute + 59*time.Second,
				maxDuration: 5*time.Minute + 1*time.Second,
			},
			{
				name:        "in 2 hours",
				input:       "in 2 hours",
				minDuration: 119 * time.Minute,
				maxDuration: 121 * time.Minute,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				now := time.Now()
				result, err := parseTime(tc.input)

				require.NoError(t, err, "parseTime should not return an error for %q", tc.input)
				assert.True(t, result.After(now), "parsed time should be in the future")

				duration := result.Sub(now)
				assert.GreaterOrEqual(t, duration, tc.minDuration)
				assert.LessOrEqual(t, duration, tc.maxDuration)
			})
		}
	})

	t.Run("absolute future dates", func(t *testing.T) {
		testCases := []struct {
			name  string
			input string
		}{
			{
				name:  "tomorrow",
				input: "tomorrow",
			},
			{
				name:  "next week",
				input: "next week",
			},
			{
				name:  "next month",
				input: "next month",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				now := time.Now()
				result, err := parseTime(tc.input)

				require.NoError(t, err, "parseTime should not return an error for %q", tc.input)
				assert.True(t, result.After(now),
					"parsed time should be in the future for %q, got %v", tc.input, result)
			})
		}
	})

	t.Run("specific future dates", func(t *testing.T) {
		now := time.Now()
		futureDate := now.AddDate(0, 0, 10)

		// Test with ISO format
		input := futureDate.Format("2006-01-02 15:04:05")
		result, err := parseTime(input)

		require.NoError(t, err)
		assert.True(t, result.After(now), "specific future date should be parsed as future")

		// Compare with a larger tolerance due to potential timezone differences
		// The parsed time might be in UTC while futureDate is in local time
		assert.WithinDuration(t, futureDate, result, 24*time.Hour)

		// Also verify the date components match
		assert.Equal(t, futureDate.Year(), result.Year(), "year should match")
		assert.Equal(t, futureDate.Month(), result.Month(), "month should match")
		assert.Equal(t, futureDate.Day(), result.Day(), "day should match")
	})

	t.Run("edge cases - ensure past times are handled", func(t *testing.T) {
		// These should be interpreted as future times due to the retry logic
		testCases := []struct {
			name  string
			input string
		}{
			{
				name:  "monday (could be past or future)",
				input: "monday",
			},
			{
				name:  "friday",
				input: "friday",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				now := time.Now()
				result, err := parseTime(tc.input)

				require.NoError(t, err, "parseTime should not return an error for %q", tc.input)
				// The retry logic with "on" prefix should ensure this is in the future
				assert.True(t, result.After(now) || result.Equal(now),
					"parsed time should be in the future or now for %q, got %v which is %v relative to now",
					tc.input, result, result.Sub(now))
			})
		}
	})

	t.Run("various time formats", func(t *testing.T) {
		testCases := []struct {
			name  string
			input string
		}{
			{name: "10 seconds", input: "10 seconds"},
			{name: "30 sec", input: "30 sec"},
			{name: "15 minutes", input: "15 minutes"},
			{name: "3 hours", input: "3 hours"},
			{name: "2 weeks", input: "2 weeks"},
			{name: "1 month", input: "1 month"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				now := time.Now()
				result, err := parseTime(tc.input)

				require.NoError(t, err, "parseTime should not return an error for %q", tc.input)
				assert.True(t, result.After(now),
					"parsed time should be in the future for %q, got %v which is %v after now",
					tc.input, result, result.Sub(now))
			})
		}
	})

	t.Run("invalid input should return error", func(t *testing.T) {
		testCases := []string{
			"not a date",
			"invalid time string",
			"xyz123",
			"",
		}

		for _, input := range testCases {
			t.Run(input, func(t *testing.T) {
				_, err := parseTime(input)
				assert.Error(t, err, "parseTime should return an error for invalid input %q", input)
			})
		}
	})
}

func TestParseTimeAlwaysReturnsFuture(t *testing.T) {
	// This test specifically ensures that parseTime ALWAYS returns a future time
	// for common relative time strings
	relativeTimeStrings := []string{
		"1 min",
		"2 min",
		"5 min",
		"10 min",
		"30 min",
		"1 hour",
		"2 hours",
		"1 day",
		"2 days",
		"7 days",
		"1 week",
		"2 weeks",
		"1 month",
	}

	for _, input := range relativeTimeStrings {
		t.Run(input, func(t *testing.T) {
			now := time.Now()
			result, err := parseTime(input)

			require.NoError(t, err, "parseTime(%q) should not return an error", input)
			require.True(t, result.After(now),
				"parseTime(%q) MUST return a future time. Got %v (now: %v, diff: %v)",
				input, result, now, result.Sub(now))

			// Additional check: the difference should be positive
			diff := result.Sub(now)
			assert.Positive(t, diff.Seconds(),
				"time difference should be positive for %q, got %v seconds", input, diff.Seconds())
		})
	}
}

func TestParseTimeReturnsLocalTimezone(t *testing.T) {
	// This test ensures that parseTime ALWAYS returns times in the local timezone
	testCases := []struct {
		name  string
		input string
	}{
		{name: "relative time - 1 min", input: "1 min"},
		{name: "relative time - 1 hour", input: "1 hour"},
		{name: "relative time - 1 day", input: "1 day"},
		{name: "absolute time - tomorrow", input: "tomorrow"},
		{name: "weekday - monday", input: "monday"},
		{name: "explicit time - in 2 hours", input: "in 2 hours"},
		{name: "ISO format", input: "2026-01-01 12:00:00"},
	}

	localZone := time.Now().Location()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseTime(tc.input)

			require.NoError(t, err, "parseTime(%q) should not return an error", tc.input)

			// Check that the timezone matches the local timezone
			resultZone, resultOffset := result.Zone()
			localZoneName, localOffset := time.Now().Zone()

			assert.Equal(t, localZoneName, resultZone,
				"parseTime(%q) should return time in local timezone. Got %s, expected %s",
				tc.input, resultZone, localZoneName)

			assert.Equal(t, localOffset, resultOffset,
				"parseTime(%q) should return time with local timezone offset. Got %d, expected %d",
				tc.input, resultOffset, localOffset)

			// Verify the location is local
			assert.Equal(t, localZone.String(), result.Location().String(),
				"parseTime(%q) should return time in local location. Got %s, expected %s",
				tc.input, result.Location().String(), localZone.String())
		})
	}
}
