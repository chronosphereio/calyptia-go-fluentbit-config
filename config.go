package fluentbitconf

import "github.com/calyptia/go-fluentbit-conf/property"

type Config struct {
	Env      property.Properties `json:"env,omitempty" yaml:"env,omitempty"`
	Service  property.Properties `json:"service,omitempty" yaml:"service,omitempty"`
	Customs  []ByName            `json:"customs,omitempty" yaml:"customs,omitempty"`
	Pipeline Pipeline            `json:"pipeline,omitempty" yaml:"pipeline,omitempty"`
}

type Pipeline struct {
	Inputs  []ByName `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Filters []ByName `json:"filters,omitempty" yaml:"filters,omitempty"`
	Outputs []ByName `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

type ByName map[string]property.Properties
