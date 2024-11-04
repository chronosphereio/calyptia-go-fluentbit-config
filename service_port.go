package fluentbitconfig

import (
	"net"
	"strconv"
	"strings"

	"github.com/calyptia/go-fluentbit-config/v2/networking"
)

type ServicePortGetter struct {
	EnabledFunc     func(p Plugin) bool
	PortFunc        func(p Plugin) (int32, bool)
	ProtocolFunc    func(p Plugin) (networking.Protocol, bool)
	DefaultPort     int32
	DefaultProtocol networking.Protocol
}

func fromPort(p Plugin) (int32, bool) {
	portVal, ok := p.Properties.Get("port")
	if !ok {
		return 0, false
	}

	i, ok := int32FromAny(portVal)
	return i, ok
}

var inputServicePorts = map[string]ServicePortGetter{
	"cloudflare": {
		PortFunc: func(p Plugin) (int32, bool) {
			addrVal, ok := p.Properties.Get("addr")
			if !ok {
				return 0, false
			}

			_, port, err := net.SplitHostPort(stringFromAny(addrVal))
			if err != nil {
				return 0, false
			}

			i, err := strconv.ParseInt(port, 10, 32)
			if err != nil {
				return 0, false
			}

			return int32(i), true
		},
		DefaultPort:     9880,
		DefaultProtocol: networking.ProtocolTCP,
	},
	"collectd": {
		PortFunc:        fromPort,
		DefaultPort:     25826,
		DefaultProtocol: networking.ProtocolUDP,
	},
	"elasticsearch": {
		PortFunc:        fromPort,
		DefaultPort:     9200,
		DefaultProtocol: networking.ProtocolTCP,
	},
	"forward": {
		EnabledFunc: func(p Plugin) bool {
			return !p.Properties.Has("unix_path")
		},
		PortFunc:        fromPort,
		DefaultPort:     24224,
		DefaultProtocol: networking.ProtocolTCP,
	},
	"http": {
		PortFunc:        fromPort,
		DefaultPort:     9880,
		DefaultProtocol: networking.ProtocolTCP,
	},
	"mqtt": {
		PortFunc:        fromPort,
		DefaultPort:     1883,
		DefaultProtocol: networking.ProtocolTCP,
	},
	"opentelemetry": {
		PortFunc:        fromPort,
		DefaultPort:     4318,
		DefaultProtocol: networking.ProtocolTCP,
	},
	"prometheus_remote_write": {
		PortFunc:        fromPort,
		DefaultPort:     8080,
		DefaultProtocol: networking.ProtocolTCP,
	},
	"splunk": {
		PortFunc:        fromPort,
		DefaultPort:     8088,
		DefaultProtocol: networking.ProtocolTCP,
	},
	"statsd": {
		PortFunc:        fromPort,
		DefaultPort:     8125,
		DefaultProtocol: networking.ProtocolUDP,
	},
	"syslog": {
		EnabledFunc: func(p Plugin) bool {
			modeVal, ok := p.Properties.Get("mode")
			if !ok {
				return false // default mode is unix_udp
			}

			mode := strings.ToLower(stringFromAny(modeVal))
			return mode == "tcp" || mode == "udp"
		},
		PortFunc: fromPort,
		ProtocolFunc: func(p Plugin) (networking.Protocol, bool) {
			modeVal, ok := p.Properties.Get("mode")
			if !ok {
				return networking.Protocol("unknown"), false
			}

			mode := strings.ToLower(stringFromAny(modeVal))
			switch mode {
			case "tcp":
				return networking.ProtocolTCP, true
			case "udp":
				return networking.ProtocolUDP, true
			}

			return networking.Protocol("unknown"), false
		},
		DefaultPort:     5140,
		DefaultProtocol: networking.ProtocolTCP,
	},
	"tcp": {
		PortFunc:        fromPort,
		DefaultPort:     5170,
		DefaultProtocol: networking.ProtocolTCP,
	},
	"udp": {
		PortFunc:        fromPort,
		DefaultPort:     5170,
		DefaultProtocol: networking.ProtocolUDP,
	},
}

var outputServicePorts = map[string]ServicePortGetter{
	"prometheus_exporter": {
		PortFunc:        fromPort,
		DefaultPort:     2021,
		DefaultProtocol: networking.ProtocolTCP,
	},
}

type ServicePort struct {
	Port     int32
	Protocol networking.Protocol
	Kind     SectionKind
	// Plugin is not available for `service`` section kind.
	Plugin *Plugin
}

type ServicePorts []ServicePort

func (c *Config) ServicePorts() ServicePorts {
	var out ServicePorts

	httpServerEnabled, ok := c.Service.Get("http_server")
	if ok && (httpServerEnabled == true || strings.ToLower(stringFromAny(httpServerEnabled)) == "on") {
		portVal, ok := c.Service.Get("http_port")
		if !ok {
			out = append(out, ServicePort{
				Port:     2020,
				Protocol: networking.ProtocolTCP,
				Kind:     SectionKindService,
			})
		} else if i, ok := int32FromAny(portVal); ok {
			out = append(out, ServicePort{
				Port:     i,
				Protocol: networking.ProtocolTCP,
				Kind:     SectionKindService,
			})
		}
	}

	process := func(getters map[string]ServicePortGetter, kind SectionKind, plugin Plugin) {
		getter, ok := getters[plugin.Name]
		if !ok {
			return
		}

		if getter.EnabledFunc != nil && !getter.EnabledFunc(plugin) {
			return
		}

		port := getter.DefaultPort
		if getter.PortFunc != nil {
			var ok bool
			port, ok = getter.PortFunc(plugin)
			if !ok {
				port = getter.DefaultPort
			}
		}

		protocol := getter.DefaultProtocol
		if getter.ProtocolFunc != nil {
			var ok bool
			protocol, ok = getter.ProtocolFunc(plugin)
			if !ok {
				protocol = getter.DefaultProtocol
			}
		}

		out = append(out, ServicePort{
			Port:     port,
			Protocol: protocol,
			Kind:     kind,
			Plugin:   &plugin,
		})
	}

	for _, input := range c.Pipeline.Inputs {
		process(inputServicePorts, SectionKindInput, input)
	}

	for _, output := range c.Pipeline.Outputs {
		process(outputServicePorts, SectionKindOutput, output)
	}

	return out
}
