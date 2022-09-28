package fluentbitconf

import (
	"fmt"
	"strings"
)

type Config struct {
	Env      Properties `json:"env" yaml:"env"`
	Service  Properties `json:"service" yaml:"service"`
	Pipeline Pipeline   `json:"pipeline" yaml:"pipeline"`
}

type Pipeline struct {
	Inputs  []Properties `json:"inputs" yaml:"inputs"`
	Filters []Properties `json:"filters" yaml:"filters"`
	Outputs []Properties `json:"outputs" yaml:"outputs"`
}

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

func (p Properties) Get(key string) any {
	if p == nil {
		return nil
	}

	for k, v := range p {
		if strings.EqualFold(k, key) {
			return v
		}
	}

	return ""
}

func (p Properties) GetString(key string) string {
	v := p.Get(key)
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func (p Properties) Set(key string, value any) {
	if p == nil {
		return
	}

	for k := range p {
		if strings.EqualFold(k, key) {
			p[key] = value
			return
		}
	}

	p[key] = value
}
