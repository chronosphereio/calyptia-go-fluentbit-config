package fluentbitconfig

import (
	"errors"
	"strings"

	"github.com/chronosphereio/calyptia-go-fluentbit-config/v2/networking"
)

var servicePortDefaultsByPlugin = map[string]servicePortDefaults{
	// Inputs.
	"collectd":      {Port: 25826, Protocol: networking.ProtocolUDP},
	"elasticsearch": {Port: 9200, Protocol: networking.ProtocolTCP},
	"forward":       {Port: 24224, Protocol: networking.ProtocolTCP}, // only if `unix_path` is not set.
	"http":          {Port: 9880, Protocol: networking.ProtocolTCP},
	"mqtt":          {Port: 1883, Protocol: networking.ProtocolTCP},
	"opentelemetry": {Port: 4318, Protocol: networking.ProtocolTCP},
	"statsd":        {Port: 8125, Protocol: networking.ProtocolUDP},
	"syslog":        {Port: 5140, Protocol: networking.ProtocolUDP}, // only if `mode` is not `unix_udp` (default) or `unix_tcp`
	"tcp":           {Port: 5170, Protocol: networking.ProtocolTCP},
	"udp":           {Port: 5170, Protocol: networking.ProtocolUDP},

	// Outputs.
	"prometheus_exporter": {Port: 2021, Protocol: networking.ProtocolTCP},
}

type servicePortDefaults struct {
	Port     int
	Protocol networking.Protocol
}

type ServicePort struct {
	Port     int
	Protocol networking.Protocol
	Kind     SectionKind
	// Plugin is not available for `service`` section kind.
	Plugin *Plugin
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
				Protocol: networking.ProtocolTCP,
				Kind:     SectionKindService,
			})
		} else if i, ok := intFromAny(portVal); ok {
			out = append(out, ServicePort{
				Port:     i,
				Protocol: networking.ProtocolTCP,
				Kind:     SectionKindService,
			})
		}
	}

	lookup := func(kind SectionKind, plugins Plugins) {
		for _, plugin := range plugins {
			err := ValidateSection(kind, plugin.Properties)
			if errors.Is(err, ErrMissingName) {
				continue
			}

			var e *UnknownPluginError
			if errors.As(err, &e) {
				continue
			}

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
				defaults, ok := servicePortDefaultsByPlugin[plugin.Name]
				if ok {
					plugin := plugin
					plugin.Properties = nil
					out = append(out, ServicePort{
						Port:     defaults.Port,
						Protocol: defaults.Protocol,
						Kind:     kind,
						Plugin:   &plugin,
					})
				}
			}

			if ok {
				port, ok := intFromAny(portVal)
				if ok {
					var protocol networking.Protocol
					if plugin.Name == "syslog" {
						modeVal, ok := plugin.Properties.Get("mode")
						if ok {
							if v := networking.Protocol(strings.ToUpper(stringFromAny(modeVal))); v.OK() {
								protocol = v
							}
						}
					}

					if protocol == "" {
						defaultPort, ok := servicePortDefaultsByPlugin[plugin.Name]
						if ok {
							protocol = defaultPort.Protocol
						}
					}

					if protocol == "" {
						protocol = networking.ProtocolTCP
					}

					plugin := plugin
					plugin.Properties = nil
					out = append(out, ServicePort{
						Port:     port,
						Protocol: protocol,
						Kind:     kind,
						Plugin:   &plugin,
					})
				}
			}
		}
	}

	lookup(SectionKindInput, c.Pipeline.Inputs)
	lookup(SectionKindOutput, c.Pipeline.Outputs)

	return out
}
