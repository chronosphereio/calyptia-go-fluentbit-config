package fluentbitconfig

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

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
				name csphere_http
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
		require.NoError(t, err)

		require.Equal(t, ServicePorts{
			{Port: 2020, Protocol: networking.ProtocolTCP, Kind: SectionKindService},
			expected(9880, networking.ProtocolTCP, SectionKindInput, "cloudflare", 0),
			expected(25826, networking.ProtocolUDP, SectionKindInput, "collectd", 1),
			expected(443, networking.ProtocolTCP, SectionKindInput, "csphere_http", 2),
			expected(9200, networking.ProtocolTCP, SectionKindInput, "elasticsearch", 3),
			expected(24224, networking.ProtocolTCP, SectionKindInput, "forward", 4),
			expected(9880, networking.ProtocolTCP, SectionKindInput, "http", 5),
			expected(1883, networking.ProtocolTCP, SectionKindInput, "mqtt", 6),
			expected(4318, networking.ProtocolTCP, SectionKindInput, "opentelemetry", 7),
			expected(8080, networking.ProtocolTCP, SectionKindInput, "prometheus_remote_write", 8),
			expected(8088, networking.ProtocolTCP, SectionKindInput, "splunk", 9),
			expected(8125, networking.ProtocolUDP, SectionKindInput, "statsd", 10),
			// expected(5140, networking.ProtocolTCP, SectionKindInput, "syslog", 11), // default syslog without explicit mode tcp or udp is skipped
			expected(5170, networking.ProtocolTCP, SectionKindInput, "tcp", 12),
			expected(5170, networking.ProtocolUDP, SectionKindInput, "udp", 13),
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
				name csphere_http
				port 4
			[INPUT]
				name elasticsearch
				port 5
			[INPUT]
				name forward
				port 6
			[INPUT]
				name http
				port 7
			[INPUT]
				name mqtt
				port 8
			[INPUT]
				name opentelemetry
				port 9
			[INPUT]
				name prometheus_remote_write
				port 10
			[INPUT]
				name splunk
				port 11
			[INPUT]
				name statsd
				port 12
			[INPUT]
				name syslog
				mode tcp
				port 13
			[INPUT]
				name syslog
				mode udp
				port 14
			[INPUT]
				name tcp
				port 15
			[INPUT]
				name udp
				port 16
			[OUTPUT]
				name prometheus_exporter
				port 17
		`, FormatClassic)
		require.NoError(t, err)
		require.Equal(t, ServicePorts{
			{Port: 1, Protocol: networking.ProtocolTCP, Kind: SectionKindService},
			expected(2, networking.ProtocolTCP, SectionKindInput, "cloudflare", 0, property.Property{Key: "addr", Value: ":2"}),
			expected(3, networking.ProtocolUDP, SectionKindInput, "collectd", 1, property.Property{Key: "port", Value: int64(3)}),
			expected(4, networking.ProtocolTCP, SectionKindInput, "csphere_http", 2, property.Property{Key: "port", Value: int64(4)}),
			expected(5, networking.ProtocolTCP, SectionKindInput, "elasticsearch", 3, property.Property{Key: "port", Value: int64(5)}),
			expected(6, networking.ProtocolTCP, SectionKindInput, "forward", 4, property.Property{Key: "port", Value: int64(6)}),
			expected(7, networking.ProtocolTCP, SectionKindInput, "http", 5, property.Property{Key: "port", Value: int64(7)}),
			expected(8, networking.ProtocolTCP, SectionKindInput, "mqtt", 6, property.Property{Key: "port", Value: int64(8)}),
			expected(9, networking.ProtocolTCP, SectionKindInput, "opentelemetry", 7, property.Property{Key: "port", Value: int64(9)}),
			expected(10, networking.ProtocolTCP, SectionKindInput, "prometheus_remote_write", 8, property.Property{Key: "port", Value: int64(10)}),
			expected(11, networking.ProtocolTCP, SectionKindInput, "splunk", 9, property.Property{Key: "port", Value: int64(11)}),
			expected(12, networking.ProtocolUDP, SectionKindInput, "statsd", 10, property.Property{Key: "port", Value: int64(12)}),
			expected(13, networking.ProtocolTCP, SectionKindInput, "syslog", 11, property.Property{Key: "mode", Value: "tcp"}, property.Property{Key: "port", Value: int64(13)}),
			expected(14, networking.ProtocolUDP, SectionKindInput, "syslog", 12, property.Property{Key: "mode", Value: "udp"}, property.Property{Key: "port", Value: int64(14)}),
			expected(15, networking.ProtocolTCP, SectionKindInput, "tcp", 13, property.Property{Key: "port", Value: int64(15)}),
			expected(16, networking.ProtocolUDP, SectionKindInput, "udp", 14, property.Property{Key: "port", Value: int64(16)}),
			expected(17, networking.ProtocolTCP, SectionKindOutput, "prometheus_exporter", 0, property.Property{Key: "port", Value: int64(17)}),
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
		require.NoError(t, err)
		require.Nil(t, config.ServicePorts())
	})
}
