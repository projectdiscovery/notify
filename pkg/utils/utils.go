package utils

import (
	"strings"
)

const defaultFormat = "{{data}}"

func Contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok

}

func FormatMessage(msg, format string) string {
	return strings.Replace(format, defaultFormat, msg, -1)
}

func SelectFormat(cliFormat, configFormat string) string {
	if cliFormat != "" {
		return cliFormat
	} else if configFormat != "" {
		return configFormat
	}
	return defaultFormat
}
