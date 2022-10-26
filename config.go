package fluentbitconf

import (
	"strings"
)

type Config struct {
	Env      Properties `json:"env,omitempty" yaml:"env,omitempty"`
	Service  Properties `json:"service,omitempty" yaml:"service,omitempty"`
	Customs  []ByName   `json:"customs,omitempty" yaml:"customs,omitempty"`
	Pipeline Pipeline   `json:"pipeline,omitempty" yaml:"pipeline,omitempty"`
}

type Pipeline struct {
	Inputs  []ByName `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Filters []ByName `json:"filters,omitempty" yaml:"filters,omitempty"`
	Outputs []ByName `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

type ByName map[string]Properties

type Properties map[string]any

func (p Properties) Has(key string) bool {
	if p == nil {
		return false
	}

	for k := range p {
		if strings.EqualFold(k, key) {
			return true
		}
	}

	return false
}

func (p Properties) Get(key string) (any, bool) {
	if p == nil {
		return nil, false
	}

	for k, v := range p {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}

	return nil, false
}

func (p *Properties) Set(key string, value any) {
	if p == nil {
		return
	}

	for k := range *p {
		if strings.EqualFold(k, key) {
			(*p)[key] = value
			return
		}
	}

	(*p)[key] = value
}
