package fluentbit_config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	diff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

func TestParseINI(t *testing.T) {
	tt := []struct {
		name          string
		config        []byte
		expected      *Config
		expectedError bool
	}{
		{
			name:          "empty ini",
			config:        []byte(""),
			expected:      nil,
			expectedError: true,
		},
		{
			name:          "bad short ini",
			config:        []byte("new-configuration"),
			expected:      nil,
			expectedError: true,
		},
		{
			name:          "single field ini",
			config:        []byte("interval_flush 1"),
			expected:      nil,
			expectedError: true,
		},
		{
			name:          "just an include",
			config:        []byte("@include foobar.conf"),
			expected:      nil,
			expectedError: false,
		},
		{
			name:          "just set a variable",
			config:        []byte("@set foo=bar\n"),
			expected:      nil,
			expectedError: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := ParseINI(tc.config)
			if tc.expectedError && err == nil {
				t.Errorf("%s expected error", tc.name)
				return
			}
			if tc.expectedError && err != nil {
				return
			}
			if !tc.expectedError && err != nil {
				t.Error(err)
				return
			}

			if tc.expected != nil {
				if want, got := len(tc.expected.Sections), len(cfg.Sections); want != got {
					fmt.Printf("%+v\n", cfg.Sections)
					t.Errorf("wants %v != got %v", want, got)
					return
				}

				for idx, section := range tc.expected.Sections {
					if want, got := len(section.Fields), len(cfg.Sections[idx].Fields); want != got {
						t.Errorf("input[%d] wants %v != got %v", idx, want, got)
						fmt.Printf("GOT=%+v\n", cfg.Sections[idx].Fields)
						return
					}
				}
			}
		})
	}
}

func TestParseYAML(t *testing.T) {
	tt := []struct {
		name          string
		config        []byte
		expected      Config
		expectedError bool
	}{
		{
			name: "basic configuration",
			config: []byte(`service:
  flush_interval: 1
  log_level: error
pipeline:
  inputs:
  - dummy:
      name: dummy
      rate: 5
  - dummy:
      dummy: '{"FOO": "BAR"}'
      name: dummy
      rate: 1
  filters:
  - record_modifier:
      match: '*'
      name: record_modifier
      record: powered_by calyptia
  outputs:
  - stdout:
      match: '*'
      name: stdout
  - exit:
      flush_count: 10
      match: '*'
      name: exit
`),
			expected: Config{
				Sections: []ConfigSection{
					{
						Type: ServiceSection,
						Fields: []Field{
							{Key: "flush_interval", Values: []Value{{
								Number: ptr(int64(1)),
							}}},
							{Key: "log_level", Values: []Value{{
								String: ptr("error"),
							}}},
						},
					},
					{
						Type: InputSection,
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: ptr("dummy"),
							}}},
							{Key: "rate", Values: []Value{{
								Number: ptr(int64(5)),
							}}},
						},
					},
					{
						Type: InputSection,
						ID:   "dummy.1",
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: ptr("dummy"),
							}}},
							{Key: "dummy", Values: []Value{{
								String: ptr(`{"FOO": "BAR"}`),
							}}},
							{Key: "rate", Values: []Value{{
								Number: ptr(int64(1)),
							}}},
						},
					},
					{
						Type: FilterSection,
						ID:   "record_modifier.0",
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: ptr("record_modifier"),
							}}},
							{Key: "match", Values: []Value{{
								String: ptr("*"),
							}}},
							{Key: "record", Values: []Value{{
								String: ptr("powered_by calyptia"),
							}}},
						},
					},
					{
						Type: OutputSection,
						ID:   "stdout.0",
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: ptr("stdout"),
							}}},
							{Key: "match", Values: []Value{{
								String: ptr("*"),
							}}},
						},
					},
					{
						Type: OutputSection,
						ID:   "exit.1",
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: ptr("exit"),
							}}},
							{Key: "match", Values: []Value{{
								String: ptr("*"),
							}}},
							{Key: "flush_count", Values: []Value{{
								Number: ptr(int64(10)),
							}}},
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := ParseYAML(tc.config)
			if tc.expectedError && err == nil {
				t.Errorf("%s expected error", tc.name)
				return
			}
			if tc.expectedError && err != nil {
				return
			}
			if !tc.expectedError && err != nil {
				t.Error(err)
				return
			}

			if want, got := len(tc.expected.Sections), len(cfg.Sections); want != got {
				fmt.Printf("%+v\n", cfg.Sections)
				t.Errorf("wants %v != got %v", want, got)
				return
			}

			for idx, section := range tc.expected.Sections {
				if want, got := len(section.Fields), len(cfg.Sections[idx].Fields); want != got {
					t.Errorf("input[%d] wants %v != got %v", idx, want, got)
					fmt.Printf("GOT=%+v\n", cfg.Sections[idx].Fields)
					return
				}
			}

			y, err := DumpYAML(cfg)
			if err != nil {
				t.Error(err)
				return
			}

			if !bytes.Equal(y, tc.config) {
				fmt.Printf("\n'%s'\n!=\n'%s'\n", string(y), string(tc.config))
				t.Errorf("yaml is not reversible")
				return
			}
		})
	}
}

func TestParseJSON(t *testing.T) {
	tt := []struct {
		name          string
		config        []byte
		expected      Config
		expectedError bool
	}{
		{
			name: "basic configuration",
			config: []byte(`{
  "service": {
    "flush_interval": 1,
    "log_level": "error"
  },
  "pipeline": {
    "inputs": [
      {
        "dummy": {
          "name": "dummy",
          "rate": 5
        }
      },
      {
        "dummy": {
          "dummy": "{\"FOO\": \"BAR\"}",
          "name": "dummy",
          "rate": 1
        }
      }
    ],
    "filters": [
      {
        "record_modifier": {
          "match": "*",
          "name": "record_modifier",
          "record": "powered_by calyptia"
        }
      }
    ],
    "outputs": [
      {
        "stdout": {
          "match": "*",
          "name": "stdout"
        }
      },
      {
        "exit": {
          "flush_count": 10,
          "match": "*",
          "name": "exit"
        }
      }
    ]
  }
}`),
			expected: Config{
				Sections: []ConfigSection{
					{
						Type: ServiceSection,
						Fields: []Field{
							{Key: "flush_interval", Values: []Value{{
								String:   nil,
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   ptr(int64(1)),
								Float:    nil,
								List:     nil,
							}}},
							{Key: "log_level", Values: []Value{{
								String:   ptr("error"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}}},
						},
					},
					{
						Type: InputSection,
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String:   ptr("dummy"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}}},
							{Key: "rate", Values: []Value{{
								Number: ptr(int64(5)),
							}}},
						},
					},
					{
						Type: InputSection,
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: ptr("dummy"),
							}}},
							{Key: "dummy", Values: []Value{{
								String: ptr(`{"FOO": "BAR"}`),
							}}},
							{Key: "rate", Values: []Value{{
								Number: ptr(int64(1)),
							}}},
						},
					},
					{
						Type: FilterSection,
						ID:   "record_modifier.0",
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: ptr("record_modifier"),
							}}},
							{Key: "match", Values: []Value{{
								String: ptr("*"),
							}}},
							{Key: "record", Values: []Value{{
								String: ptr("powered_by calyptia"),
							}}},
						},
					},
					{
						Type: OutputSection,
						ID:   "stdout.0",
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: ptr("stdout"),
							}}},
							{Key: "match", Values: []Value{{
								String: ptr("*"),
							}}},
						},
					},
					{
						Type: OutputSection,
						ID:   "exit.1",
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: ptr("exit"),
							}}},
							{Key: "match", Values: []Value{{
								String: ptr("*"),
							}}},
							{Key: "flush_count", Values: []Value{{
								Number: ptr(int64(10)),
							}}},
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := ParseYAML(tc.config)
			if tc.expectedError && err == nil {
				t.Errorf("%s expected error", tc.name)
				return
			}
			if tc.expectedError && err != nil {
				return
			}
			if !tc.expectedError && err != nil {
				t.Error(err)
				return
			}

			if want, got := len(tc.expected.Sections), len(cfg.Sections); want != got {
				fmt.Printf("%+v\n", cfg.Sections)
				t.Errorf("wants %v != got %v", want, got)
				return
			}

			for idx, section := range tc.expected.Sections {
				if want, got := len(section.Fields), len(cfg.Sections[idx].Fields); want != got {
					t.Errorf("input[%d] wants %v != got %v", idx, want, got)
					fmt.Printf("GOT=%+v\n", cfg.Sections[idx].Fields)
					return
				}
			}
		})
	}
}

func TestNewConfigFromBytes(t *testing.T) {
	tt := []struct {
		name          string
		config        []byte
		expected      Config
		expectedError bool
	}{
		{
			name: "test invalid configuration",
			config: []byte(`
				asdasdasdasdasd
				"@@"
			`),
			expected:      Config{},
			expectedError: true,
		},
		{
			name:          "test invalid configuration",
			config:        []byte(""),
			expected:      Config{},
			expectedError: true,
		},
		{
			name: "test valid simple configuration",
			config: []byte(
				`[INPUT]
	Name tail
	tag kube.*
	mem_buf_limit 4.8

[PARSER]
	Name   apache2
	Format regex
	Regex  ^(?<host>[^ ]*) [^ ]* (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^ ]*) +\S*)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>.*)")?$
	Time_Key time
	Time_Format %d/%b/%Y:%H:%M:%S %z
`),
			expected: Config{
				Sections: []ConfigSection{{
					Type: InputSection,
					Fields: []Field{
						{Key: "name", Values: []Value{{
							String: ptr("tail"),
						}}},
						{Key: "tag", Values: []Value{{
							String: ptr("kube.*"),
						}}},
						{Key: "mem_buf_limit", Values: []Value{{
							String: ptr("4.8M"),
						}}},
					},
				}, {
					Type: ParserSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: ptr("apache2"),
						}}},
						{Key: "Format", Values: []Value{{
							String: ptr("regex"),
						}}},
						{Key: "Regex", Values: []Value{{
							Regex: ptr(`^(?<host>[^ ]*) [^ ]* (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^ ]*) +\S*)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>.*)")?$`),
						}}},
						{Key: "Time_Key", Values: []Value{{
							String: ptr("time"),
						}}},
						{Key: "Time_Format", Values: []Value{{
							TimeFormat: ptr("%d/%b/%Y:%H:%M:%S %z"),
						}}},
					},
				}},
			},
		},
		{
			name: "test multiple values",
			config: []byte(`[FILTER]
						Name rewrite_tag
						Field 1 2 3
					`),
			// Rule  $topic ^tele\/[^\/]+\/SENSOR$ sonoff true
			expected: Config{
				Sections: []ConfigSection{{
					Type: FilterSection,
					Fields: []Field{
						{
							Key: "Name",
							Values: []Value{{
								String:   ptr("rewrite_tag"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}}},
						{
							Key: "Field",
							Values: []Value{{
								String:   ptr("1 2 3 4"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}}},
					},
				}},
			},
		},

		{
			name: "test rewrite_tag",
			config: []byte(`[INPUT]
						name tail
						tag kube.*
						mem_buf_limit 4.8M
					[FILTER]
						Name rewrite_tag
						Match mqtt
						Rule topic tele sonoff true
			`),
			// Rule  $topic ^tele\/[^\/]+\/SENSOR$ sonoff true
			expected: Config{
				Sections: []ConfigSection{{
					Type: InputSection,
					Fields: []Field{
						{Key: "name", Values: []Value{{
							String: ptr("tail"),
						}}},
						{Key: "tag", Values: []Value{{
							String: ptr("kube.*"),
						}}},
						{Key: "mem_buf_limit", Values: []Value{{
							String: ptr("4.8M"),
						}}},
					},
				}, {
					Type: InputSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: ptr("rewrite_tag"),
						}}},
						{Key: "Match", Values: []Value{{
							String: ptr("mqtt"),
						}}},
						{Key: "Rule", Values: []Value{{
							String: ptr("topic tele sonoff true"),
						}}},
					},
				}},
			},
		},
		{
			name: "test valid large configuration bug 157",
			config: []byte(`[INPUT]
			Name          forward
			Host          0.0.0.0
			Port          24284

		[OUTPUT]
			Name http
			Match dummy.*
			Host api.coralogix.com
			Port 443
			URI /logs/rest/singles
			Header private_key 0886464e-ccba-ff8b-83ff-4fa9c3cc0f72
			Format json_lines
			Compress On
		`),
			expected: Config{
				Sections: []ConfigSection{{
					Type: InputSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: ptr("forward"),
						}}},
						{Key: "Host", Values: []Value{{
							String: ptr("0.0.0.0"),
						}}},
						{Key: "Port", Values: []Value{{
							String: ptr("24284"),
						}}},
					},
				}, {
					Type: OutputSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: ptr("http"),
						}}},
						{Key: "Match", Values: []Value{{
							String: ptr("dummy.*"),
						}}},
						{Key: "Host", Values: []Value{{
							String: ptr("api.coralogix.com"),
						}}},
						{Key: "Port", Values: []Value{{
							String: ptr("443"),
						}}},
						{Key: "URI", Values: []Value{{
							String: ptr("/logs/rest/singles"),
						}}},
						{Key: "Header", Values: []Value{{
							String: ptr("private_key Private_Key-Super-Complex-testing-123"),
						}}},
						{Key: "Format", Values: []Value{{
							String: ptr("json_lines"),
						}}},
						{Key: "Compress", Values: []Value{{
							String: ptr("On"),
						}}},
					},
				}},
			},
			expectedError: false,
		},
		{
			name: "test valid large configuration bug 157 - 2nd case",
			config: []byte(`[INPUT]
			Name          forward
			Host          0.0.0.0
			Port          24284

		[OUTPUT]
			Name http
			Match dummy.*
			Host api.coralogix.com
			Port 443
			URI /logs/rest/singles
			Header private_key EXPLICITLY-LONG-KEY-12345
			Format json_lines
			Compress On
		`),
			expected: Config{
				Sections: []ConfigSection{{
					Type: InputSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: ptr("forward"),
						}}},
						{Key: "Host", Values: []Value{{
							String: ptr("0.0.0.0"),
						}}},
						{Key: "Port", Values: []Value{{
							String: ptr("24284"),
						}}},
					},
				}, {
					Type: OutputSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: ptr("http"),
						}}},
						{Key: "Match", Values: []Value{{
							String: ptr("dummy.*"),
						}}},
						{Key: "Host", Values: []Value{{
							String: ptr("api.coralogix.com"),
						}}},
						{Key: "Port", Values: []Value{{
							String: ptr("443"),
						}}},
						{Key: "URI", Values: []Value{{
							String: ptr("/logs/rest/singles"),
						}}},
						{Key: "Header", Values: []Value{{
							String: ptr("private_key EXPLICITLY-LONG-KEY-12345"),
						}}},
						{Key: "Format", Values: []Value{{
							String: ptr("json_lines"),
						}}},
						{Key: "Compress", Values: []Value{{
							String: ptr("On"),
						}}},
					},
				}},
			},
			expectedError: false,
		},

		{
			name: "test valid larger configuration",
			config: []byte(`[INPUT]
			Name        tail
			Tag         tail.01
			Path        /var/log/system.log
		[OUTPUT]
			Name s3
			Match *
			bucket your-bucket
	`),
			expected: Config{
				Sections: []ConfigSection{{
					Type: InputSection,
					Fields: []Field{
						{Key: "name", Values: []Value{{
							String: ptr("tail"),
						}}},
						{Key: "tag", Values: []Value{{
							String: ptr("tail.01"),
						}}},
						{Key: "Path", Values: []Value{{
							String: ptr("/var/log/system.log"),
						}}},
					},
				}, {
					Type: OutputSection,
					Fields: []Field{
						{Key: "name", Values: []Value{{
							String: ptr("s3"),
						}}},
						{Key: "Match", Values: []Value{{
							String: ptr("*"),
						}}},
						{Key: "bucket", Values: []Value{{
							String: ptr("your-bucket"),
						}}},
					},
				}},
			},
			expectedError: false,
		},
		{
			name: "test valid multiple inputs",
			config: []byte(`[INPUT]
					Name tcp
					Port 5556
					Tag  foobar

				[INPUT]
					Name tcp
					Port 5557
					Tag  foobar

				[OUTPUT]
					Name  stdout
					Match *
			`),
			expected: Config{
				Sections: []ConfigSection{{
					Type: InputSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: ptr("tcp"),
						}}},
						{Key: "Port", Values: []Value{{
							String: ptr("5556"),
							List:   nil},
						}},
						{Key: "Tag", Values: []Value{{
							String: ptr("foobar"),
						}}},
					},
				}, {
					Type: InputSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: ptr("tcp"),
						}}},
						{Key: "Port", Values: []Value{{
							String: ptr("5557"),
						}}},
						{Key: "Tag", Values: []Value{{
							String: ptr("foobar"),
						}}},
					},
				}, {
					Type: OutputSection,
					Fields: []Field{
						{Key: "name", Values: []Value{{
							String: ptr("stdout"),
						}}},
						{Key: "Match", Values: []Value{{
							String: ptr("*"),
						}}},
					},
				}},
			},
			expectedError: false,
		},
		{
			name: "test valid multiple outputs with hosts",
			config: []byte(`[OUTPUT]
					Name tcp
					Host 172.168.0.1
					Port 5556
					Tag  foobar

				[OUTPUT]
					Name opensearch
					Host 127.0.0.1
					Port 5550
					Tag  foobar

				[OUTPUT]
					Name gelf
					Host api.gelf.test
					Match *
			`),
			expected: Config{
				Sections: []ConfigSection{{
					Type: OutputSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: ptr("tcp"),
						}}},
						{Key: "Host", Values: []Value{{
							Address: ptr("172.168.0.1"),
						}}},
						{Key: "Port", Values: []Value{{
							String: ptr("5556"),
							List:   nil},
						}},
						{Key: "Tag", Values: []Value{{
							String: ptr("foobar"),
						}}},
					},
				}, {
					Type: OutputSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: ptr("opensearch"),
						}}},
						{Key: "Host", Values: []Value{{
							Address: ptr("127.0.0.1"),
						}}},
						{Key: "Port", Values: []Value{{
							String: ptr("5557"),
						}}},
						{Key: "Tag", Values: []Value{{
							String: ptr("foobar"),
						}}},
					},
				}, {
					Type: OutputSection,
					Fields: []Field{
						{Key: "name", Values: []Value{{
							String: ptr("gelf"),
						}}},
						{Key: "Host", Values: []Value{{
							Address: ptr("api.gelf.test"),
						}}},
						{Key: "Match", Values: []Value{{
							String: ptr("*"),
						}}},
					},
				}},
			},
			expectedError: false,
		},

		{
			name: "test time formats",
			config: []byte(`
				[PARSER]
					Name        syslog-rfc3164-local
					Format      regex
					Regex       ^\<(?<pri>[0-9]+)\>(?<time>[^ ]* {1,2}[^ ]* [^ ]*) (?<ident>[a-zA-Z0-9_\/\.\-]*)(?:\[(?<pid>[0-9]+)\])?(?:[^\:]*\:)? *(?<message>.*)$
					Time_Key    time
					Time_Format %b %d %H:%M:%S
					Time_Keep   On
				[PARSER]
					Name        mongodb
					Format      regex
					Regex       ^(?<time>[^ ]*)\s+(?<severity>\w)\s+(?<component>[^ ]+)\s+\[(?<context>[^\]]+)]\s+(?<message>.*?) *(?<ms>(\d+))?(:?ms)?$
					Time_Format %Y-%m-%dT%H:%M:%S.%L
					Time_Keep   On
					Time_Key    time
			`),
			expected: Config{
				Sections: []ConfigSection{{
					Type: ParserSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: ptr("syslog-rfc3164-local"),
						}}},
						{Key: "Format", Values: []Value{{
							String: ptr("regex"),
						}}},
						{Key: "Regex", Values: []Value{{
							Regex: ptr(`^\<(?<pri>[0-9]+)\>(?<time>[^ ]* {1,2}[^ ]* [^ ]*) (?<ident>[a-zA-Z0-9_\/\.\-]*)(?:\[(?<pid>[0-9]+)\])?(?:[^\:]*\:)? *(?<message>.*)$`),
						}}},
						{Key: "Time_Key", Values: []Value{{
							String: ptr("time"),
						}}},
						{Key: "Time_Format", Values: []Value{{
							TimeFormat: ptr("%b %d %H:%M:%S"),
						}}},
						{Key: "Time_Keep", Values: []Value{{
							String: ptr("On"),
						}}},
					},
				}, {
					Type: ParserSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: ptr("mongodb"),
						}}},
						{Key: "Format", Values: []Value{{
							String: ptr("regex"),
						}}},
						{Key: "Regex", Values: []Value{{
							Regex: ptr(`^(?<time>[^ ]*)\s+(?<severity>\w)\s+(?<component>[^ ]+)\s+\[(?<context>[^\]]+)]\s+(?<message>.*?) *(?<ms>(\d+))?(:?ms)?$`),
						}}},
						{Key: "Time_Format", Values: []Value{{
							TimeFormat: ptr("%Y-%m-%dT%H:%M:%S.%L"),
						}}},
						{Key: "Time_Keep", Values: []Value{{
							String: ptr("On"),
						}}},
						{Key: "Time_Key", Values: []Value{{
							String: ptr("time"),
						}}},
					},
				}},
			},
			expectedError: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := ParseINI(tc.config)
			if tc.expectedError && err == nil {
				t.Errorf("%s expected error", tc.name)
				return
			}
			if tc.expectedError && err != nil {
				return
			}
			if !tc.expectedError && err != nil {
				t.Error(err)
				return
			}

			if want, got := len(tc.expected.Sections), len(cfg.Sections); want != got {
				t.Errorf("wants %v != got %v", want, got)
				return
			}

			for idx, section := range tc.expected.Sections {
				if want, got := len(section.Fields), len(cfg.Sections[idx].Fields); want != got {
					t.Errorf("section[%d] wants %v != got %v", idx, want, got)
					return
				}
			}
		})
	}
}

func TestINItoJSON(t *testing.T) {
	inipaths, err := filepath.Glob("tests/ini-to-json/*.ini")
	if err != nil {
		t.Error(err)
		return
	}

	for _, inipath := range inipaths {
		inifile := filepath.Base(inipath)
		ininame := inifile[:len(inifile)-len(filepath.Ext(inifile))]

		t.Run(fmt.Sprintf("ini_%s", inifile), func(t *testing.T) {
			config, err := os.ReadFile(inipath)
			if err != nil {
				t.Error(err)
				return
			}

			cfg, err := ParseINI(config)
			if err != nil {
				t.Error(err)
				return
			}

			jsonconfig, err := DumpJSON(cfg)
			if err != nil {
				t.Error(err)
				return
			}

			expected, err := os.ReadFile(fmt.Sprintf("tests/ini-to-json/%s.json", ininame))
			if err != nil {
				t.Error(err)
				return
			}

			differ := diff.New()
			d, err := differ.Compare(expected, jsonconfig)
			if err != nil {
				t.Error(err)
				return
			}

			if d.Modified() {
				var eJson map[string]interface{}
				if err := json.Unmarshal(expected, &eJson); err != nil {
					t.Error(err)
					return
				}

				config := formatter.AsciiFormatterConfig{
					ShowArrayIndex: true,
					Coloring:       true,
				}

				formatter := formatter.NewAsciiFormatter(eJson, config)
				diff, err := formatter.Format(d)
				if err != nil {
					t.Error(err)
					return
				}

				fmt.Print(diff)

				t.Error("JSON does not match")
				return
			}
		})
	}
}

func TestJSONtoINI(t *testing.T) {
	jsonpaths, err := filepath.Glob("tests/ini-to-json/*.json")
	if err != nil {
		t.Error(err)
		return
	}

	for _, jsonpath := range jsonpaths {
		jsonfile := filepath.Base(jsonpath)
		jsonname := jsonfile[:len(jsonfile)-len(filepath.Ext(jsonfile))]

		t.Run(fmt.Sprintf("ini_%s", jsonfile), func(t *testing.T) {
			config, err := os.ReadFile(jsonpath)
			if err != nil {
				t.Error(err)
				return
			}

			cfg, err := ParseJSON(config)
			if err != nil {
				t.Error(err)
				return
			}

			ini, err := cfg.DumpINI()
			if err != nil {
				t.Error(err)
				return
			}

			expected, err := os.ReadFile(fmt.Sprintf("tests/ini-to-json/%s.ini", jsonname))
			if err != nil {
				t.Error(err)
				return
			}

			if expected[len(expected)-1] != '\n' {
				expected = append(expected, '\n')
			}

			if !bytes.Equal(ini, expected) {
				t.Errorf("INI does not match: \n'%s' => \n'%s'\n",
					string(ini), string(expected))
				return
			}
		})
	}
}
