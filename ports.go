package fluentbitconfig

import (
	"strconv"
	"strings"
)

type Protocol string

const (
	ProtocolTCP  Protocol = "TCP"
	ProtocolUDP  Protocol = "UDP"
	ProtocolSCTP Protocol = "SCTP"
)

func (p Protocol) OK() bool {
	return p == ProtocolTCP || p == ProtocolUDP || p == ProtocolSCTP
}

var fluentBitNetworkingDefaults = map[string]portDefault{
	// Inputs.
	"collectd":      {Port: 25826, Protocol: ProtocolUDP},
	"elasticsearch": {Port: 9200, Protocol: ProtocolTCP},
	"forward":       {Port: 24224, Protocol: ProtocolTCP}, // only if `unix_path` is not set.
	"http":          {Port: 9880, Protocol: ProtocolTCP},
	"mqtt":          {Port: 1883, Protocol: ProtocolTCP},
	"opentelemetry": {Port: 4318, Protocol: ProtocolTCP},
	"statsd":        {Port: 8125, Protocol: ProtocolUDP},
	"syslog":        {Port: 5140, Protocol: ProtocolUDP}, // only if `mode` is not `unix_udp` (default) or `unix_tcp`
	"tcp":           {Port: 5170, Protocol: ProtocolTCP},
	"udp":           {Port: 5170, Protocol: ProtocolUDP},

	// Outputs.
	"prometheus_exporter": {Port: 2021, Protocol: ProtocolTCP},
}

type portDefault struct {
	Port     int
	Protocol Protocol
}

type ServicePort struct {
	Port     int
	Protocol Protocol
	Kind     SectionKind
	Plugin   *ServicePortPlugin
}

type ServicePortPlugin struct {
	ID   string
	Name string
}

type ServicePorts []ServicePort

func (c *Config) ServicePorts() ServicePorts {
	var out ServicePorts

	enabledVal, ok := c.Service.Get("http_server")
	if ok && (enabledVal == true || strings.ToLower(stringFromAny(enabledVal)) == "on") {
		portVal, ok := c.Service.Get("http_port")
		if !ok {
			out = append(out, ServicePort{
				Port:     2020,
				Protocol: ProtocolTCP,
				Kind:     SectionKindService,
			})
		} else if i, ok := intFromAny(portVal); ok {
			out = append(out, ServicePort{
				Port:     i,
				Protocol: ProtocolTCP,
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
				defaults, ok := fluentBitNetworkingDefaults[plugin.Name]
				if ok {
					out = append(out, ServicePort{
						Port:     defaults.Port,
						Protocol: defaults.Protocol,
						Kind:     kind,
						Plugin: &ServicePortPlugin{
							ID:   plugin.ID,
							Name: plugin.Name,
						},
					})
				}
			}

			if ok {
				port, ok := intFromAny(portVal)
				if ok {
					var protocol Protocol
					if plugin.Name == "syslog" {
						modeVal, ok := plugin.Properties.Get("mode")
						if ok {
							if v := Protocol(strings.ToUpper(stringFromAny(modeVal))); v.OK() {
								protocol = v
							}
						}
					}

					if protocol == "" {
						defaultPort, ok := fluentBitNetworkingDefaults[plugin.Name]
						if ok {
							protocol = defaultPort.Protocol
						}
					}

					if protocol == "" {
						protocol = ProtocolTCP
					}

					out = append(out, ServicePort{
						Port:     port,
						Protocol: protocol,
						Kind:     kind,
						Plugin: &ServicePortPlugin{
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
