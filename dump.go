package fluentbitconfig

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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
		return ParseAsClassic(raw)
	case "yml", "yaml":
		return ParseAsYAML(raw)
	case "json":
		return ParseAsJSON(raw)
	}
	return out, ErrFormatUnknown
}

func ParseAsClassic(raw string) (Config, error) {
	var out Config
	return out, out.UnmarshalClassic([]byte(raw))
}

func ParseAsYAML(raw string) (Config, error) {
	var out Config
	dec := yaml.NewDecoder(strings.NewReader(raw))
	dec.KnownFields(true)
	err := dec.Decode(&out)
	if errors.Is(err, io.EOF) {
		return out, nil
	}

	return out, err
}

func ParseAsJSON(raw string) (Config, error) {
	var out Config
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()
	err := dec.Decode(&out)
	if errors.Is(err, io.EOF) {
		return out, nil
	}

	return out, err
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
