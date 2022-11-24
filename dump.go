package fluentbitconfig

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

var ErrFormatUnknown = errors.New("format unknown")

type Format string

const (
	FormatClassic = "classic"
	FormatYAML    = "yaml"
	FormatJSON    = "json"
)

func ParseAs(raw string, format Format) (Config, error) {
	var out Config
	switch strings.ToLower(string(format)) {
	case "", "ini", "conf", "classic":
		return out, out.UnmarshalClassic([]byte(raw))
	case "yml", "yaml":
		return out, yaml.Unmarshal([]byte(raw), &out)
	case "json":
		return out, json.Unmarshal([]byte(raw), &out)
	}
	return out, ErrFormatUnknown
}

func (c Config) DumpAs(format Format) (string, error) {
	switch strings.ToLower(string(format)) {
	case "", "ini", "conf", "classic":
		return c.DumpAsClassic()
	case "yml", "yaml":
		return c.DumpAsYAML()
	case "json":
		return c.DumpAsJSON()
	}
	return "", ErrFormatUnknown
}

func (c Config) DumpAsClassic() (string, error) {
	b, err := c.MarshalClassic()
	if err != nil {
		return "", err
	}

	return string(b), err
}

func (c Config) DumpAsYAML() (string, error) {
	b, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (c Config) DumpAsJSON() (string, error) {
	var buff bytes.Buffer
	enc := json.NewEncoder(&buff)
	enc.SetEscapeHTML(false)
	err := enc.Encode(c)
	if err != nil {
		return "", err
	}

	return buff.String(), nil
}
