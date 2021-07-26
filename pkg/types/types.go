package types

import "github.com/projectdiscovery/goflags"

type Options struct {
	Verbose        bool                `yaml:"verbose,omitempty"`
	NoColor        bool                `yaml:"no_color,omitempty"`
	Silent         bool                `yaml:"silent,omitempty"`
	Version        bool                `yaml:"version,omitempty"`
	ProviderConfig string              `yaml:"provider_config,omitempty"`
	Providers      goflags.StringSlice `yaml:"providers,omitempty"`
	Profiles       goflags.StringSlice `yaml:"profiles,omitempty"`

	Stdin     bool
	Bulk      bool   `yaml:"bulk,omitempty"`
	CharLimit int    `yaml:"char_limit,omitempty"`
	Data      string `yaml:"data,omitempty"`
}
