package fluentbitconfig

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestConfig_Ports(t *testing.T) {
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
		assert.Equal(t, Ports{
			{Number: 2020, Protocol: "http", Kind: SectionKindService},
			{Number: 25826, Protocol: "udp", Kind: SectionKindInput, Plugin: &PortPlugin{ID: "collectd.0", Name: "collectd"}},
			{Number: 9200, Protocol: "http", Kind: SectionKindInput, Plugin: &PortPlugin{ID: "elasticsearch.1", Name: "elasticsearch"}},
			{Number: 24224, Protocol: "http", Kind: SectionKindInput, Plugin: &PortPlugin{ID: "forward.2", Name: "forward"}},
			{Number: 9880, Protocol: "http", Kind: SectionKindInput, Plugin: &PortPlugin{ID: "http.3", Name: "http"}},
			{Number: 1883, Protocol: "http", Kind: SectionKindInput, Plugin: &PortPlugin{ID: "mqtt.4", Name: "mqtt"}},
			{Number: 4318, Protocol: "http", Kind: SectionKindInput, Plugin: &PortPlugin{ID: "opentelemetry.5", Name: "opentelemetry"}},
			{Number: 8125, Protocol: "udp", Kind: SectionKindInput, Plugin: &PortPlugin{ID: "statsd.6", Name: "statsd"}},
			{Number: 5140, Protocol: "udp", Kind: SectionKindInput, Plugin: &PortPlugin{ID: "syslog.7", Name: "syslog"}},
			{Number: 5170, Protocol: "tcp", Kind: SectionKindInput, Plugin: &PortPlugin{ID: "tcp.8", Name: "tcp"}},
			{Number: 5170, Protocol: "udp", Kind: SectionKindInput, Plugin: &PortPlugin{ID: "udp.9", Name: "udp"}},
			{Number: 2021, Protocol: "http", Kind: SectionKindOutput, Plugin: &PortPlugin{ID: "prometheus_exporter.0", Name: "prometheus_exporter"}},
		}, config.Ports())
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
		assert.Equal(t, []int{
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
			9,
			10,
			11,
			12,
		}, config.Ports().Numbers())
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
		assert.Equal(t, nil, config.Ports())
	})

	t.Run("with_mode", func(t *testing.T) {
		config, err := ParseAs(`
			[INPUT]
				name syslog
				mode tcp
				port 1
		`, FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, Ports{
			{Number: 1, Protocol: "tcp", Kind: SectionKindInput, Plugin: &PortPlugin{ID: "syslog.0", Name: "syslog"}},
		}, config.Ports())
	})
}
