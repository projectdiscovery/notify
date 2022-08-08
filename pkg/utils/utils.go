package utils

import (
	"fmt"
	"strings"
)

const defaultFormat = "{{data}}"

// FormatMessage formats the message according to the format string
func FormatMessage(msg, format string) string {
	return strings.Replace(format, defaultFormat, msg, -1)
}

// SelectFormat returns the format string in the following order of precedence:
// 1. cliFormat
// 2. configFormat
// 3. defaulFormat
func SelectFormat(cliFormat, configFormat string) string {
	fmt.Println("cliFormat: ", cliFormat)
	fmt.Println("configFormat: ", configFormat)
	if cliFormat != "" {
		return cliFormat
	} else if configFormat != "" {
		return configFormat
	}
	return defaultFormat
}
