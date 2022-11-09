package fluentbitconfig

import "github.com/calyptia/go-fluentbit-config/property"

// TODO: add includes.
type Config struct {
	Env      property.Properties `json:"env,omitempty" yaml:"env,omitempty"`
	Service  property.Properties `json:"service,omitempty" yaml:"service,omitempty"`
	Customs  []ByName            `json:"customs,omitempty" yaml:"customs,omitempty"`
	Pipeline Pipeline            `json:"pipeline,omitempty" yaml:"pipeline,omitempty"`
}

// TODO: add parsers.
type Pipeline struct {
	Inputs  []ByName `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Parsers []ByName `json:"parsers,omitempty" yaml:"parsers,omitempty"`
	Filters []ByName `json:"filters,omitempty" yaml:"filters,omitempty"`
	Outputs []ByName `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

type ByName map[string]property.Properties
