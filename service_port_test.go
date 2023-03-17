package fluentbitconfig

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/calyptia/go-fluentbit-config/v2/networking"
)

func TestConfig_ServicePorts(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		config, err := ParseAs(`
			[SERVICE]
				http_server on
			[INPUT]
				name collectd
			[INPUT]
				name elasticsearch
			[INPUT]
				name forward
			[INPUT]
				name http
			[INPUT]
				name mqtt
			[INPUT]
				name opentelemetry
			[INPUT]
				name statsd
			[INPUT]
				name syslog
				mode udp
			[INPUT]
				name tcp
			[INPUT]
				name udp
			[OUTPUT]
				name prometheus_exporter
		`, FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, ServicePorts{
			{Port: 2020, Protocol: networking.ProtocolTCP, Kind: SectionKindService},
			{Port: 25826, Protocol: networking.ProtocolUDP, Kind: SectionKindInput, Plugin: &Plugin{ID: "collectd.0", Name: "collectd"}},
			{Port: 9200, Protocol: networking.ProtocolTCP, Kind: SectionKindInput, Plugin: &Plugin{ID: "elasticsearch.1", Name: "elasticsearch"}},
			{Port: 24224, Protocol: networking.ProtocolTCP, Kind: SectionKindInput, Plugin: &Plugin{ID: "forward.2", Name: "forward"}},
			{Port: 9880, Protocol: networking.ProtocolTCP, Kind: SectionKindInput, Plugin: &Plugin{ID: "http.3", Name: "http"}},
			{Port: 1883, Protocol: networking.ProtocolTCP, Kind: SectionKindInput, Plugin: &Plugin{ID: "mqtt.4", Name: "mqtt"}},
			{Port: 4318, Protocol: networking.ProtocolTCP, Kind: SectionKindInput, Plugin: &Plugin{ID: "opentelemetry.5", Name: "opentelemetry"}},
			{Port: 8125, Protocol: networking.ProtocolUDP, Kind: SectionKindInput, Plugin: &Plugin{ID: "statsd.6", Name: "statsd"}},
			{Port: 5140, Protocol: networking.ProtocolUDP, Kind: SectionKindInput, Plugin: &Plugin{ID: "syslog.7", Name: "syslog"}},
			{Port: 5170, Protocol: networking.ProtocolTCP, Kind: SectionKindInput, Plugin: &Plugin{ID: "tcp.8", Name: "tcp"}},
			{Port: 5170, Protocol: networking.ProtocolUDP, Kind: SectionKindInput, Plugin: &Plugin{ID: "udp.9", Name: "udp"}},
			{Port: 2021, Protocol: networking.ProtocolTCP, Kind: SectionKindOutput, Plugin: &Plugin{ID: "prometheus_exporter.0", Name: "prometheus_exporter"}},
		}, config.ServicePorts())
	})

	t.Run("explicit", func(t *testing.T) {
		config, err := ParseAs(`
			[SERVICE]
				http_server on
				http_port 1
			[INPUT]
				name collectd
				port 2
			[INPUT]
				name elasticsearch
				port 3
			[INPUT]
				name forward
				port 4
			[INPUT]
				name http
				port 5
			[INPUT]
				name mqtt
				port 6
			[INPUT]
				name opentelemetry
				port 7
			[INPUT]
				name statsd
				port 8
			[INPUT]
				name syslog
				mode udp
				port 9
			[INPUT]
				name tcp
				port 10
			[INPUT]
				name udp
				port 11
			[OUTPUT]
				name prometheus_exporter
				port 12
		`, FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, ServicePorts{
			{Port: 1, Protocol: networking.ProtocolTCP, Kind: SectionKindService},
			{Port: 2, Protocol: networking.ProtocolUDP, Kind: SectionKindInput, Plugin: &Plugin{ID: "collectd.0", Name: "collectd"}},
			{Port: 3, Protocol: networking.ProtocolTCP, Kind: SectionKindInput, Plugin: &Plugin{ID: "elasticsearch.1", Name: "elasticsearch"}},
			{Port: 4, Protocol: networking.ProtocolTCP, Kind: SectionKindInput, Plugin: &Plugin{ID: "forward.2", Name: "forward"}},
			{Port: 5, Protocol: networking.ProtocolTCP, Kind: SectionKindInput, Plugin: &Plugin{ID: "http.3", Name: "http"}},
			{Port: 6, Protocol: networking.ProtocolTCP, Kind: SectionKindInput, Plugin: &Plugin{ID: "mqtt.4", Name: "mqtt"}},
			{Port: 7, Protocol: networking.ProtocolTCP, Kind: SectionKindInput, Plugin: &Plugin{ID: "opentelemetry.5", Name: "opentelemetry"}},
			{Port: 8, Protocol: networking.ProtocolUDP, Kind: SectionKindInput, Plugin: &Plugin{ID: "statsd.6", Name: "statsd"}},
			{Port: 9, Protocol: networking.ProtocolUDP, Kind: SectionKindInput, Plugin: &Plugin{ID: "syslog.7", Name: "syslog"}},
			{Port: 10, Protocol: networking.ProtocolTCP, Kind: SectionKindInput, Plugin: &Plugin{ID: "tcp.8", Name: "tcp"}},
			{Port: 11, Protocol: networking.ProtocolUDP, Kind: SectionKindInput, Plugin: &Plugin{ID: "udp.9", Name: "udp"}},
			{Port: 12, Protocol: networking.ProtocolTCP, Kind: SectionKindOutput, Plugin: &Plugin{ID: "prometheus_exporter.0", Name: "prometheus_exporter"}},
		}, config.ServicePorts())
	})

	t.Run("skips", func(t *testing.T) {
		config, err := ParseAs(`
			[INPUT]
				name forward
				unix_path /tmp/forward.sock
				port 1
			[INPUT]
				name syslog
				port 2
			[INPUT]
				name syslog
				mode unix_udp
				port 3
			[INPUT]
				name syslog
				mode unix_tcp
				port 4
		`, FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, nil, config.ServicePorts())
	})

	t.Run("with_mode", func(t *testing.T) {
		config, err := ParseAs(`
			[INPUT]
				name syslog
				mode tcp
				port 1
		`, FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, ServicePorts{
			{Port: 1, Protocol: networking.ProtocolTCP, Kind: SectionKindInput, Plugin: &Plugin{ID: "syslog.0", Name: "syslog"}},
		}, config.ServicePorts())
	})
}
