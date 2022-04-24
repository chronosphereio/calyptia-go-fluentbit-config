package fluentbit_config

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
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
							{Key: "flush_interval", Values: []*Value{{
								String:   nil,
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   numberPtr(1),
								Float:    nil,
								List:     nil,
							}}},
							{Key: "log_level", Values: []*Value{{
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
						ID:   "dummy.0",
						Fields: []Field{{
							Key: "name", Values: []*Value{{
								String:   stringPtr("dummy"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}},
						},
							{
								Key: "rate", Values: []*Value{{
									String:   nil,
									DateTime: nil,
									Date:     nil,
									Time:     nil,
									Bool:     nil,
									Number:   numberPtr(5),
									Float:    nil,
									List:     nil,
								}},
							},
						},
					},
					{
						Type: InputSection,
						ID:   "dummy.1",
						Fields: []Field{{
							Key: "name", Values: []*Value{{
								String:   stringPtr("dummy"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}},
						},
							{
								Key: "dummy", Values: []*Value{{
									String:   stringPtr(`{"FOO": "BAR"}`),
									DateTime: nil,
									Date:     nil,
									Time:     nil,
									Bool:     nil,
									Number:   nil,
									Float:    nil,
									List:     nil,
								}},
							},
							{
								Key: "rate", Values: []*Value{{
									String:   nil,
									DateTime: nil,
									Date:     nil,
									Time:     nil,
									Bool:     nil,
									Number:   numberPtr(1),
									Float:    nil,
									List:     nil,
								}},
							},
						},
					},
					{
						Type: FilterSection,
						ID:   "record_modifier.0",
						Fields: []Field{{
							Key: "name", Values: []*Value{{
								String:   stringPtr("record_modifier"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}},
						},
							{
								Key: "match", Values: []*Value{{
									String:   stringPtr("*"),
									DateTime: nil,
									Date:     nil,
									Time:     nil,
									Bool:     nil,
									Number:   nil,
									Float:    nil,
									List:     nil,
								}},
							},
							{
								Key: "record", Values: []*Value{{
									String:   stringPtr("powered_by calyptia"),
									DateTime: nil,
									Date:     nil,
									Time:     nil,
									Bool:     nil,
									Number:   nil,
									Float:    nil,
									List:     nil,
								}},
							},
						},
					},
					{
						Type: OutputSection,
						ID:   "stdout.0",
						Fields: []Field{{
							Key: "name", Values: []*Value{{
								String:   stringPtr("stdout"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}},
						},
							{
								Key: "match", Values: []*Value{{
									String:   stringPtr("*"),
									DateTime: nil,
									Date:     nil,
									Time:     nil,
									Bool:     nil,
									Number:   nil,
									Float:    nil,
									List:     nil,
								}},
							},
						},
					},
					{
						Type: OutputSection,
						ID:   "exit.1",
						Fields: []Field{
							{
								Key: "name", Values: []*Value{{
									String:   stringPtr("exit"),
									DateTime: nil,
									Date:     nil,
									Time:     nil,
									Bool:     nil,
									Number:   nil,
									Float:    nil,
									List:     nil,
								}},
							},
							{
								Key: "match", Values: []*Value{{
									String:   stringPtr("*"),
									DateTime: nil,
									Date:     nil,
									Time:     nil,
									Bool:     nil,
									Number:   nil,
									Float:    nil,
									List:     nil,
								}},
							},
							{
								Key: "flush_count", Values: []*Value{{
									String:   nil,
									DateTime: nil,
									Date:     nil,
									Time:     nil,
									Bool:     nil,
									Number:   numberPtr(10),
									Float:    nil,
									List:     nil,
								}},
							},
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
			config: []byte(`
				[INPUT]
					name tail
					tag kube.*
					mem_buf_limit 4.8M
				[PARSER]
					Name   apache2
					Format regex
					Regex  ^(?<host>[^ ]*) [^ ]* (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^ ]*) +\S*)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>.*)")?$
					Time_Key time
					Time_Format %d/%b/%Y:%H:%M:%S %z`),
			expected: Config{
				Sections: []ConfigSection{{
					Type: InputSection,
					Fields: []Field{
						{Key: "name", Values: []*Value{{
							String:   stringPtr("tail"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
						{Key: "tag", Values: []*Value{{
							String:   stringPtr("kube.*"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
						{Key: "mem_buf_limit", Values: []*Value{{
							String:   stringPtr("4.8M"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
					},
				}, {
					Type: ParserSection,
					Fields: []Field{
						{
							Key: "Name",
							Values: []*Value{{
								String:   stringPtr("apache2"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}},
						},
						{
							Key: "Format", Values: []*Value{{
								String:   stringPtr("regex"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}},
						},
						{
							Key: "Regex", Values: []*Value{{
								String:   stringPtr(`^(?<host>[^ ]*) [^ ]* (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^ ]*) +\S*)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>.*)")?$`),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}},
						},
						{
							Key: "Time_Key", Values: []*Value{{
								String:   stringPtr("time"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}},
						},
						{
							Key: "Time_Format", Values: []*Value{{
								String:   stringPtr("%d/%b/%Y:%H:%M:%S %z"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}},
						},
					},
				}},
			},
		},
		{
			name: "test multiple values",
			config: []byte(`
					[FILTER]
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
							Values: []*Value{{
								String:   stringPtr("rewrite_tag"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}},
						},
						{
							Key: "Field",
							Values: []*Value{{
								String:   stringPtr("1"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}, {
								String:   stringPtr("2"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}, {
								String:   stringPtr("3"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}},
						},
					},
				}},
			},
		},

		{
			name: "test rewrite_tag",
			config: []byte(`
					[INPUT]
						name tail
						tag kube.*
						mem_buf_limit 4.8M
					[FILTER]
						name rewrite_tag
						Rule topic tele sonoff true
						match mqtt
			`),
			// Rule  $topic ^tele\/[^\/]+\/SENSOR$ sonoff true
			expected: Config{
				Sections: []ConfigSection{{
					Type: InputSection,
					Fields: []Field{
						{Key: "name", Values: []*Value{{
							String:   stringPtr("tail"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
						{Key: "tag", Values: []*Value{{
							String:   stringPtr("kube.*"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
						{Key: "mem_buf_limit", Values: []*Value{{
							String:   stringPtr("4.8M"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
					},
				}, {
					Type: InputSection,
					Fields: []Field{
						{
							Key: "Name",
							Values: []*Value{{
								String:   stringPtr("rewrite_tag"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}},
						},
						{
							Key: "Match", Values: []*Value{{
								String:   stringPtr("mqtt"),
								DateTime: nil,
								Date:     nil,
								Time:     nil,
								Bool:     nil,
								Number:   nil,
								Float:    nil,
								List:     nil,
							}},
						},
						{
							Key: "Rule", Values: []*Value{
								{
									String:   stringPtr("mqtt"),
									DateTime: nil,
									Date:     nil,
									Time:     nil,
									Bool:     nil,
									Number:   nil,
									Float:    nil,
									List:     nil,
								},
							},
						},
					},
				}},
			},
		},

		{
			name: "test valid larger configuration",
			config: []byte(`
		[INPUT]
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
						{Key: "name", Values: []*Value{{
							String:   stringPtr("tail"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
						{Key: "tag", Values: []*Value{{
							String:   stringPtr("tail.01"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
						{Key: "Path", Values: []*Value{{
							String:   stringPtr("/var/log/system.log"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
					},
				}, {
					Type: OutputSection,
					Fields: []Field{
						{Key: "name", Values: []*Value{{
							String:   stringPtr("s3"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
						{Key: "Match", Values: []*Value{{
							String:   stringPtr("*"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
						{Key: "bucket", Values: []*Value{{
							String:   stringPtr("your-bucket"),
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
			expectedError: false,
		},
		{
			name: "test valid larger configuration",
			config: []byte(`
				[INPUT]
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
						{Key: "Name", Values: []*Value{{
							String:   stringPtr("tcp"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
						{Key: "Port", Values: []*Value{{
							String:   nil,
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   numberPtr(5556),
							Float:    nil,
							List:     nil,
						}}},
						{Key: "Tag", Values: []*Value{{
							String:   stringPtr("foobar"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
					},
				}, {
					Type: InputSection,
					Fields: []Field{
						{Key: "Name", Values: []*Value{{
							String:   stringPtr("tcp"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
						{Key: "Port", Values: []*Value{{
							String:   nil,
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   numberPtr(5557),
							Float:    nil,
							List:     nil,
						}}},
						{Key: "Tag", Values: []*Value{{
							String:   stringPtr("foobat"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
					},
				}, {
					Type: OutputSection,
					Fields: []Field{
						{Key: "name", Values: []*Value{{
							String:   stringPtr("stdout"),
							DateTime: nil,
							Date:     nil,
							Time:     nil,
							Bool:     nil,
							Number:   nil,
							Float:    nil,
							List:     nil,
						}}},
						{Key: "Match", Values: []*Value{{
							String:   stringPtr("*"),
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
			expectedError: false,
		},
	}

	stripIndentation := regexp.MustCompile("\n[\t]+")
	stripMultipleSpaces := regexp.MustCompile(`[\ ]{2,}`)

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
					t.Errorf("input[%d] wants %v != got %v", idx, want, got)
					fmt.Printf("GOT=%+v\n", cfg.Sections[idx].Fields)
					return
				}
			}

			config := stripIndentation.ReplaceAll(tc.config, []byte("\n"))
			config = stripMultipleSpaces.ReplaceAll(config, []byte(" "))

			if ini, err := cfg.DumpINI(); err != nil {
				t.Error(err)
				return
			} else if !bytes.Equal(ini, config) {
				fmt.Printf("\n'%s'\n != \n'%s'\n", string(config), string(ini))
				t.Errorf("unable to reconstitute INI")
				return
			}
		})
	}
}
