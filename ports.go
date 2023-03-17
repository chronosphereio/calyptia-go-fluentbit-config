package fluentbitconfig

import (
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

var fluentBitNetworkingPorts = map[string]portDefault{
	// Inputs.
	"collectd":      {Number: 25826, Protocol: "udp"},
	"elasticsearch": {Number: 9200, Protocol: "http"},
	"forward":       {Number: 24224, Protocol: "http"}, // only if `unix_path` is not set.
	"http":          {Number: 9880, Protocol: "http"},
	"mqtt":          {Number: 1883, Protocol: "http"},
	"opentelemetry": {Number: 4318, Protocol: "http"},
	"statsd":        {Number: 8125, Protocol: "udp"},
	"syslog":        {Number: 5140, Protocol: "udp"}, // only if `mode` is not `unix_udp` (default) or `unix_tcp`
	"tcp":           {Number: 5170, Protocol: "tcp"},
	"udp":           {Number: 5170, Protocol: "udp"},

	// Outputs.
	"prometheus_exporter": {Number: 2021, Protocol: "http"},
}

type portDefault struct {
	Number   int
	Protocol string
}

type Port struct {
	Number   int
	Protocol string
	Kind     SectionKind
	Plugin   *PortPlugin
}

type PortPlugin struct {
	ID   string
	Name string
}

type Ports []Port

func (p Ports) Numbers() []int {
	var out []int
	for _, port := range p {
		out = append(out, port.Number)
	}

	slices.Sort(out)
	out = slices.Compact(out)

	return out
}

func (c *Config) Ports() Ports {
	var out Ports

	enabledVal, ok := c.Service.Get("http_server")
	if ok && (enabledVal == true || strings.ToLower(stringFromAny(enabledVal)) == "on") {
		portVal, ok := c.Service.Get("http_port")
		if !ok {
			out = append(out, Port{
				Number:   2020,
				Protocol: "http",
				Kind:     SectionKindService,
			})
		} else if i, ok := intFromAny(portVal); ok {
			out = append(out, Port{
				Number:   i,
				Protocol: "http",
				Kind:     SectionKindService,
			})
		}
	}

	lookup := func(kind SectionKind, plugins Plugins) {
		for _, plugin := range plugins {
			if plugin.Name == "forward" && plugin.Properties.Has("unix_path") {
				continue
			}

			if plugin.Name == "syslog" {
				modeVal, ok := plugin.Properties.Get("mode")
				if !ok {
					continue
				}

				if ok {
					mode := strings.ToLower(stringFromAny(modeVal))
					if mode == "unix_udp" || mode == "unix_tcp" {
						continue
					}
				}
			}

			portVal, ok := plugin.Properties.Get("port")
			if !ok {
				defaultPort, ok := fluentBitNetworkingPorts[plugin.Name]
				if ok {
					out = append(out, Port{
						Number:   defaultPort.Number,
						Protocol: defaultPort.Protocol,
						Kind:     kind,
						Plugin: &PortPlugin{
							ID:   plugin.ID,
							Name: plugin.Name,
						},
					})
				}
			}

			if ok {
				port, ok := intFromAny(portVal)
				if ok {
					var protocol string
					if plugin.Name == "syslog" {
						modeVal, ok := plugin.Properties.Get("mode")
						if ok {
							protocol = strings.ToLower(stringFromAny(modeVal))
						}
					}

					if protocol == "" {
						defaultPort, ok := fluentBitNetworkingPorts[plugin.Name]
						if ok {
							protocol = defaultPort.Protocol
						}
					}

					if protocol == "" {
						protocol = "http"
					}

					out = append(out, Port{
						Number:   port,
						Protocol: protocol,
						Kind:     kind,
						Plugin: &PortPlugin{
							ID:   plugin.ID,
							Name: plugin.Name,
						},
					})
				}
			}
		}
	}

	lookup(SectionKindInput, c.Pipeline.Inputs)
	lookup(SectionKindOutput, c.Pipeline.Outputs)

	return out
}

func intFromAny(v any) (int, bool) {
	if v == nil {
		return 0, false
	}

	switch v := v.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		if int64(int(v)) == v {
			return int(v), true
		}
	case float32:
		return int(v), true
	case float64:
		if float64(int(v)) == v {
			return int(v), true
		}
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	}
	return 0, false
}
