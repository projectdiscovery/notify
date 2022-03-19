package utils

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/projectdiscovery/gologger"
)

const defaultFormat = "{{.Data}}"

func Contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok

}

func FormatMessage(msg, format string) string {
	buf := new(bytes.Buffer)
	tpl, err := template.New("msg").Funcs(sprig.TxtFuncMap()).Parse(format)
	if err != nil {
		gologger.Error().Msgf("failed to parse message template %s:%v", format, err)
		return msg
	}

	err = tpl.Execute(buf, struct{ Data string }{Data: msg})
	if err != nil {
		gologger.Error().Msgf("failed to execute message template: %v", err)
		return msg
	}
	return buf.String()
}

func SelectFormat(cliFormat, configFormat string) string {
	if cliFormat != "" && cliFormat != defaultFormat {
		return cliFormat
	} else if configFormat != "" && configFormat != defaultFormat {
		return configFormat
	}
	return defaultFormat
}
