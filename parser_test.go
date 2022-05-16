package fluentbit_config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	diff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

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
								Number: numberPtr(1),
							}}},
							{Key: "log_level", Values: []Value{{
								String: stringPtr("error"),
							}}},
						},
					},
					{
						Type: InputSection,
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: stringPtr("dummy"),
							}}},
							{Key: "rate", Values: []Value{{
								Number: numberPtr(5),
							}}},
						},
					},
					{
						Type: InputSection,
						ID:   "dummy.1",
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: stringPtr("dummy"),
							}}},
							{Key: "dummy", Values: []Value{{
								String: stringPtr(`{"FOO": "BAR"}`),
							}}},
							{Key: "rate", Values: []Value{{
								Number: numberPtr(1),
							}}},
						},
					},
					{
						Type: FilterSection,
						ID:   "record_modifier.0",
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: stringPtr("record_modifier"),
							}}},
							{Key: "match", Values: []Value{{
								String: stringPtr("*"),
							}}},
							{Key: "record", Values: []Value{{
								String: stringPtr("powered_by calyptia"),
							}}},
						},
					},
					{
						Type: OutputSection,
						ID:   "stdout.0",
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: stringPtr("stdout"),
							}}},
							{Key: "match", Values: []Value{{
								String: stringPtr("*"),
							}}},
						},
					},
					{
						Type: OutputSection,
						ID:   "exit.1",
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: stringPtr("exit"),
							}}},
							{Key: "match", Values: []Value{{
								String: stringPtr("*"),
							}}},
							{Key: "flush_count", Values: []Value{{
								Number: numberPtr(10),
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
								Number:   numberPtr(1),
								Float:    nil,
								List:     nil,
							}}},
							{Key: "log_level", Values: []Value{{
								String:   stringPtr("error"),
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
								String:   stringPtr("dummy"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}}},
							{Key: "rate", Values: []Value{{
								Number: numberPtr(5),
							}}},
						},
					},
					{
						Type: InputSection,
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: stringPtr("dummy"),
							}}},
							{Key: "dummy", Values: []Value{{
								String: stringPtr(`{"FOO": "BAR"}`),
							}}},
							{Key: "rate", Values: []Value{{
								Number: numberPtr(1),
							}}},
						},
					},
					{
						Type: FilterSection,
						ID:   "record_modifier.0",
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: stringPtr("record_modifier"),
							}}},
							{Key: "match", Values: []Value{{
								String: stringPtr("*"),
							}}},
							{Key: "record", Values: []Value{{
								String: stringPtr("powered_by calyptia"),
							}}},
						},
					},
					{
						Type: OutputSection,
						ID:   "stdout.0",
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: stringPtr("stdout"),
							}}},
							{Key: "match", Values: []Value{{
								String: stringPtr("*"),
							}}},
						},
					},
					{
						Type: OutputSection,
						ID:   "exit.1",
						Fields: []Field{
							{Key: "name", Values: []Value{{
								String: stringPtr("exit"),
							}}},
							{Key: "match", Values: []Value{{
								String: stringPtr("*"),
							}}},
							{Key: "flush_count", Values: []Value{{
								Number: numberPtr(10),
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
							String: stringPtr("tail"),
						}}},
						{Key: "tag", Values: []Value{{
							String: stringPtr("kube.*"),
						}}},
						{Key: "mem_buf_limit", Values: []Value{{
							String: stringPtr("4.8M"),
						}}},
					},
				}, {
					Type: ParserSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: stringPtr("apache2"),
						}}},
						{Key: "Format", Values: []Value{{
							String: stringPtr("regex"),
						}}},
						{Key: "Regex", Values: []Value{{
							Regex: stringPtr(`^(?<host>[^ ]*) [^ ]* (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^ ]*) +\S*)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>.*)")?$`),
						}}},
						{Key: "Time_Key", Values: []Value{{
							String: stringPtr("time"),
						}}},
						{Key: "Time_Format", Values: []Value{{
							TimeFormat: stringPtr("%d/%b/%Y:%H:%M:%S %z"),
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
								String:   stringPtr("rewrite_tag"),
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
								String:   stringPtr("1 2 3 4"),
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
							String: stringPtr("tail"),
						}}},
						{Key: "tag", Values: []Value{{
							String: stringPtr("kube.*"),
						}}},
						{Key: "mem_buf_limit", Values: []Value{{
							String: stringPtr("4.8M"),
						}}},
					},
				}, {
					Type: InputSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: stringPtr("rewrite_tag"),
						}}},
						{Key: "Match", Values: []Value{{
							String: stringPtr("mqtt"),
						}}},
						{Key: "Rule", Values: []Value{{
							String: stringPtr("topic tele sonoff true"),
						}}},
					},
				}},
			},
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
							String: stringPtr("tail"),
						}}},
						{Key: "tag", Values: []Value{{
							String: stringPtr("tail.01"),
						}}},
						{Key: "Path", Values: []Value{{
							String: stringPtr("/var/log/system.log"),
						}}},
					},
				}, {
					Type: OutputSection,
					Fields: []Field{
						{Key: "name", Values: []Value{{
							String: stringPtr("s3"),
						}}},
						{Key: "Match", Values: []Value{{
							String: stringPtr("*"),
						}}},
						{Key: "bucket", Values: []Value{{
							String: stringPtr("your-bucket"),
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
					Tag  foobat
		
				[OUTPUT]
					Name  stdout
					Match *
			`),
			expected: Config{
				Sections: []ConfigSection{{
					Type: InputSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: stringPtr("tcp"),
						}}},
						{Key: "Port", Values: []Value{{
							String: stringPtr("5556"),
							List:   nil},
						}},
						{Key: "Tag", Values: []Value{{
							String: stringPtr("foobar"),
						}}},
					},
				}, {
					Type: InputSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: stringPtr("tcp"),
						}}},
						{Key: "Port", Values: []Value{{
							String: stringPtr("5557"),
						}}},
						{Key: "Tag", Values: []Value{{
							String: stringPtr("foobat"),
						}}},
					},
				}, {
					Type: OutputSection,
					Fields: []Field{
						{Key: "name", Values: []Value{{
							String: stringPtr("stdout"),
						}}},
						{Key: "Match", Values: []Value{{
							String: stringPtr("*"),
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
							String: stringPtr("syslog-rfc3164-local"),
						}}},
						{Key: "Format", Values: []Value{{
							String: stringPtr("regex"),
						}}},
						{Key: "Regex", Values: []Value{{
							Regex: stringPtr(`^\<(?<pri>[0-9]+)\>(?<time>[^ ]* {1,2}[^ ]* [^ ]*) (?<ident>[a-zA-Z0-9_\/\.\-]*)(?:\[(?<pid>[0-9]+)\])?(?:[^\:]*\:)? *(?<message>.*)$`),
						}}},
						{Key: "Time_Key", Values: []Value{{
							String: stringPtr("time"),
						}}},
						{Key: "Time_Format", Values: []Value{{
							TimeFormat: stringPtr("%b %d %H:%M:%S"),
						}}},
						{Key: "Time_Keep", Values: []Value{{
							String: stringPtr("On"),
						}}},
					},
				}, {
					Type: ParserSection,
					Fields: []Field{
						{Key: "Name", Values: []Value{{
							String: stringPtr("mongodb"),
						}}},
						{Key: "Format", Values: []Value{{
							String: stringPtr("regex"),
						}}},
						{Key: "Regex", Values: []Value{{
							Regex: stringPtr(`^(?<time>[^ ]*)\s+(?<severity>\w)\s+(?<component>[^ ]+)\s+\[(?<context>[^\]]+)]\s+(?<message>.*?) *(?<ms>(\d+))?(:?ms)?$`),
						}}},
						{Key: "Time_Format", Values: []Value{{
							TimeFormat: stringPtr("%Y-%m-%dT%H:%M:%S.%L"),
						}}},
						{Key: "Time_Keep", Values: []Value{{
							String: stringPtr("On"),
						}}},
						{Key: "Time_Key", Values: []Value{{
							String: stringPtr("time"),
						}}},
					},
				}},
			},
			expectedError: false,
		},
	}

	//stripIndentation := regexp.MustCompile("\n[\t]+")
	//stripMultipleSpaces := regexp.MustCompile(`[\ ]{2,}`)

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

				/*
					for fidx, field := range section.Fields {
						if want, got := len(field.Values), len(cfg.Sections[idx].Fields[fidx].Values); want != got {
							t.Errorf("section[%d].fields[%s:%d] %d != got %d", idx, field.Key, fidx, want, got)
							return
						}
						for vidx, value := range cfg.Sections[idx].Fields[fidx].Values {
							if !value.Equals(field.Values[vidx]) {
								t.Errorf("unexpected value section[%d].fields[%s]", idx, field.Key)
								return
							}
						}
					}
				*/
			}

			/*
				config := stripIndentation.ReplaceAll(tc.config, []byte("\n"))
				config = stripMultipleSpaces.ReplaceAll(config, []byte(" "))

				if ini, err := cfg.DumpINI(); err != nil {
					t.Error(err)
					return
				} else {
					ini = stripIndentation.ReplaceAll(ini, []byte("\n"))
					ini = stripMultipleSpaces.ReplaceAll(ini, []byte(""))
					if !bytes.Equal(ini, config) {
						fmt.Printf("\n'%s'\n != \n'%s'\n", string(config), string(ini))
						t.Errorf("unable to reconstitute INI")
					}
					return
				}
			*/
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
			config, err := ioutil.ReadFile(inipath)
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

			expected, err := ioutil.ReadFile(fmt.Sprintf("tests/ini-to-json/%s.json", ininame))
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
			config, err := ioutil.ReadFile(jsonpath)
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

			expected, err := ioutil.ReadFile(fmt.Sprintf("tests/ini-to-json/%s.ini", jsonname))
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
