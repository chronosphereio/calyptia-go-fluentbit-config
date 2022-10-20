package fluentbitconf

import (
	"strings"
)

type Config struct {
	Env      Properties `json:"env" yaml:"env"`
	Service  Properties `json:"service" yaml:"service"`
	Customs  []ByName   `json:"customs" yaml:"customs"`
	Pipeline Pipeline   `json:"pipeline" yaml:"pipeline"`
}

type Pipeline struct {
	Inputs  []ByName `json:"inputs" yaml:"inputs"`
	Filters []ByName `json:"filters" yaml:"filters"`
	Outputs []ByName `json:"outputs" yaml:"outputs"`
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
