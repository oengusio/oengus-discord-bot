package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func ParseAndMakeDurationPretty(duration string) string {
	dur, _ := ParseDuration(duration)

	// Remnants of failed attempts
	//return fmt.Sprintf("%f hours, %f minutes, %f seconds", dur.Hours(), dur.Minutes(), dur.Seconds())
	//return fmt.Sprintf("%02d:%02d:%02d", int64(dur.Hours()), int64(dur.Minutes()), int64(dur.Seconds()))
	//return dur.String()

	dur = dur.Round(time.Second)

	hours := dur / time.Hour
	dur -= hours * time.Hour

	minutes := dur / time.Minute
	dur -= minutes * time.Minute

	seconds := dur / time.Second

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// ParseDuration parses an ISO 8601 string representing a duration,
// and returns the resultant golang time.Duration instance.
//
// Source: https://gist.github.com/spatialtime/2a54a6dbf80121997b2459b2d3b9b380
func ParseDuration(isoDuration string) (time.Duration, error) {
	re := regexp.MustCompile(`^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:.\d+)?)S)?$`)
	matches := re.FindStringSubmatch(isoDuration)
	if matches == nil {
		return 0, errors.New("input string is of incorrect format")
	}

	seconds := 0.0

	//skipping years and months

	//days
	if matches[3] != "" {
		f, err := strconv.ParseFloat(matches[3], 32)
		if err != nil {
			return 0, err
		}

		seconds += (f * 24 * 60 * 60)
	}
	//hours
	if matches[4] != "" {
		f, err := strconv.ParseFloat(matches[4], 32)
		if err != nil {
			return 0, err
		}

		seconds += (f * 60 * 60)
	}
	//minutes
	if matches[5] != "" {
		f, err := strconv.ParseFloat(matches[5], 32)
		if err != nil {
			return 0, err
		}

		seconds += (f * 60)
	}
	//seconds & milliseconds
	if matches[6] != "" {
		f, err := strconv.ParseFloat(matches[6], 32)
		if err != nil {
			return 0, err
		}

		seconds += f
	}

	goDuration := strconv.FormatFloat(seconds, 'f', -1, 32) + "s"
	return time.ParseDuration(goDuration)

}

// FormatDuration returns an ISO 8601 duration string.
func FormatDuration(dur time.Duration) string {
	return "PT" + strings.ToUpper(dur.Truncate(time.Millisecond).String())
}
