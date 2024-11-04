package fluentbitconfig

import (
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/calyptia/go-fluentbit-config/v2/networking"
	"github.com/calyptia/go-fluentbit-config/v2/property"
)

func TestConfig_ServicePorts(t *testing.T) {
	expected := func(port int32, protocol networking.Protocol, kind SectionKind, name string, index int, props ...property.Property) ServicePort {
		pp := property.Properties{
			{Key: "name", Value: name},
		}
		pp = append(pp, props...)
		return ServicePort{
			Port:     port,
			Protocol: protocol,
			Kind:     kind,
			Plugin: &Plugin{
				ID:         name + "." + strconv.Itoa(index),
				Name:       name,
				Properties: pp,
			},
		}
	}

	t.Run("defaults", func(t *testing.T) {
		config, err := ParseAs(`
			[SERVICE]
				http_server on
			[INPUT]
				name cloudflare
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
				name prometheus_remote_write
			[INPUT]
				name splunk
			[INPUT]
				name statsd
			[INPUT]
				name syslog
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
			expected(9880, networking.ProtocolTCP, SectionKindInput, "cloudflare", 0),
			expected(25826, networking.ProtocolUDP, SectionKindInput, "collectd", 1),
			expected(9200, networking.ProtocolTCP, SectionKindInput, "elasticsearch", 2),
			expected(24224, networking.ProtocolTCP, SectionKindInput, "forward", 3),
			expected(9880, networking.ProtocolTCP, SectionKindInput, "http", 4),
			expected(1883, networking.ProtocolTCP, SectionKindInput, "mqtt", 5),
			expected(4318, networking.ProtocolTCP, SectionKindInput, "opentelemetry", 6),
			expected(8080, networking.ProtocolTCP, SectionKindInput, "prometheus_remote_write", 7),
			expected(8088, networking.ProtocolTCP, SectionKindInput, "splunk", 8),
			expected(8125, networking.ProtocolUDP, SectionKindInput, "statsd", 9),
			// expected(5140, networking.ProtocolTCP, SectionKindInput, "syslog", 10), // default syslog without explicit mode tcp or udp is skipped
			expected(5170, networking.ProtocolTCP, SectionKindInput, "tcp", 11),
			expected(5170, networking.ProtocolUDP, SectionKindInput, "udp", 12),
			expected(2021, networking.ProtocolTCP, SectionKindOutput, "prometheus_exporter", 0),
		}, config.ServicePorts())
	})

	t.Run("explicit", func(t *testing.T) {
		config, err := ParseAs(`
			[SERVICE]
				http_server on
				http_port 1
			[INPUT]
				name cloudflare
				addr :2
			[INPUT]
				name collectd
				port 3
			[INPUT]
				name elasticsearch
				port 4
			[INPUT]
				name forward
				port 5
			[INPUT]
				name http
				port 6
			[INPUT]
				name mqtt
				port 7
			[INPUT]
				name opentelemetry
				port 8
			[INPUT]
				name prometheus_remote_write
				port 9
			[INPUT]
				name splunk
				port 10
			[INPUT]
				name statsd
				port 11
			[INPUT]
				name syslog
				mode tcp
				port 12
			[INPUT]
				name syslog
				mode udp
				port 13
			[INPUT]
				name tcp
				port 14
			[INPUT]
				name udp
				port 15
			[OUTPUT]
				name prometheus_exporter
				port 16
		`, FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, ServicePorts{
			{Port: 1, Protocol: networking.ProtocolTCP, Kind: SectionKindService},
			expected(2, networking.ProtocolTCP, SectionKindInput, "cloudflare", 0, property.Property{Key: "addr", Value: ":2"}),
			expected(3, networking.ProtocolUDP, SectionKindInput, "collectd", 1, property.Property{Key: "port", Value: int64(3)}),
			expected(4, networking.ProtocolTCP, SectionKindInput, "elasticsearch", 2, property.Property{Key: "port", Value: int64(4)}),
			expected(5, networking.ProtocolTCP, SectionKindInput, "forward", 3, property.Property{Key: "port", Value: int64(5)}),
			expected(6, networking.ProtocolTCP, SectionKindInput, "http", 4, property.Property{Key: "port", Value: int64(6)}),
			expected(7, networking.ProtocolTCP, SectionKindInput, "mqtt", 5, property.Property{Key: "port", Value: int64(7)}),
			expected(8, networking.ProtocolTCP, SectionKindInput, "opentelemetry", 6, property.Property{Key: "port", Value: int64(8)}),
			expected(9, networking.ProtocolTCP, SectionKindInput, "prometheus_remote_write", 7, property.Property{Key: "port", Value: int64(9)}),
			expected(10, networking.ProtocolTCP, SectionKindInput, "splunk", 8, property.Property{Key: "port", Value: int64(10)}),
			expected(11, networking.ProtocolUDP, SectionKindInput, "statsd", 9, property.Property{Key: "port", Value: int64(11)}),
			expected(12, networking.ProtocolTCP, SectionKindInput, "syslog", 10, property.Property{Key: "mode", Value: "tcp"}, property.Property{Key: "port", Value: int64(12)}),
			expected(13, networking.ProtocolUDP, SectionKindInput, "syslog", 11, property.Property{Key: "mode", Value: "udp"}, property.Property{Key: "port", Value: int64(13)}),
			expected(14, networking.ProtocolTCP, SectionKindInput, "tcp", 12, property.Property{Key: "port", Value: int64(14)}),
			expected(15, networking.ProtocolUDP, SectionKindInput, "udp", 13, property.Property{Key: "port", Value: int64(15)}),
			expected(16, networking.ProtocolTCP, SectionKindOutput, "prometheus_exporter", 0, property.Property{Key: "port", Value: int64(16)}),
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
			[INPUT]
				name does_not_exists
				port 5
		`, FormatClassic)
		assert.NoError(t, err)
		assert.Equal(t, nil, config.ServicePorts())
	})
}
