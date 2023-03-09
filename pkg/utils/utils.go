package utils

import (
	"fmt"
	"strings"
	"time"
)

const (
	defaultFormat  = "{{data}}"
	dateTimeFormat = "{{datetime}}"
	dateFormat     = "{{date}}"
	timeFormat     = "{{time}}"
	countFormat    = "{{count}}"
)

// FormatMessage formats the message according to the format string
func FormatMessage(msg, format string, counter int) string {
	ts := time.Now()

	format = dateTimeHelper(msg, format, ts)
	format = dateHelper(msg, format, ts)
	format = timeHelper(msg, format, ts)
	format = countHelper(msg, format, counter)

	return strings.ReplaceAll(format, defaultFormat, msg)
}

// SelectFormat returns the format string in the following order of precedence:
// 1. cliFormat
// 2. configFormat
// 3. defaultFormat
func SelectFormat(cliFormat, configFormat string) string {
	if cliFormat != "" {
		return cliFormat
	} else if configFormat != "" {
		return configFormat
	}
	return defaultFormat
}

func dateTimeHelper(msg, format string, now time.Time) string {
	if strings.Contains(format, dateTimeFormat) {
		format = strings.ReplaceAll(format, dateTimeFormat, now.Format("01-02-2006 15:04:05-0700"))
	}

	return format
}

func timeHelper(msg, format string, now time.Time) string {
	if strings.Contains(format, timeFormat) {
		format = strings.ReplaceAll(format, timeFormat, now.Format("15:04:05-0700"))
	}

	return format
}

func dateHelper(msg, format string, now time.Time) string {
	if strings.Contains(format, dateFormat) {
		format = strings.ReplaceAll(format, dateFormat, now.Format("01-02-2006"))
	}

	return format
}

func countHelper(msg, format string, counter int) string {
	if strings.Contains(format, countFormat) {
		format = strings.ReplaceAll(format, countFormat, fmt.Sprint(counter))
	}

	return format
}
