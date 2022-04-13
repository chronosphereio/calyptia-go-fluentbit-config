package fluentbit_config

import (
	"testing"
)

func stringPtr(s string) *string {
	return &s
}

func numberPtr(n float64) *float64 {
	return &n
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
			expected: struct {
				Inputs      map[string][]Field
				Filters     map[string][]Field
				Customs     map[string][]Field
				Outputs     map[string][]Field
				Parsers     map[string][]Field
				InputIndex  int
				FilterIndex int
				OutputIndex int
				CustomIndex int
				ParserIndex int
			}{Inputs: map[string][]Field{}, Filters: map[string][]Field{}, Customs: map[string][]Field{}, Outputs: map[string][]Field{}, Parsers: map[string][]Field{}},
			expectedError: true,
		},
		{
			name:   "test invalid configuration",
			config: []byte(""),
			expected: struct {
				Inputs      map[string][]Field
				Filters     map[string][]Field
				Customs     map[string][]Field
				Outputs     map[string][]Field
				Parsers     map[string][]Field
				InputIndex  int
				FilterIndex int
				OutputIndex int
				CustomIndex int
				ParserIndex int
			}{Inputs: map[string][]Field{}, Filters: map[string][]Field{}, Customs: map[string][]Field{}, Outputs: map[string][]Field{}, Parsers: map[string][]Field{}},
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
			expected: struct {
				Inputs      map[string][]Field
				Filters     map[string][]Field
				Customs     map[string][]Field
				Outputs     map[string][]Field
				Parsers     map[string][]Field
				InputIndex  int
				FilterIndex int
				OutputIndex int
				CustomIndex int
				ParserIndex int
			}{
				Parsers: map[string][]Field{
					"apache2.0": []Field{
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
								String:   stringPtr("regexs"),
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
				Inputs: map[string][]Field{"tail.0": {
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
				}},
			}},
		{
			name: "test multiple values",
			config: []byte(`
					[FILTER]
						Name rewrite_tag
						Field 1 2 3
					`),
			// Rule  $topic ^tele\/[^\/]+\/SENSOR$ sonoff true
			expected: struct {
				Inputs      map[string][]Field
				Filters     map[string][]Field
				Customs     map[string][]Field
				Outputs     map[string][]Field
				Parsers     map[string][]Field
				InputIndex  int
				FilterIndex int
				OutputIndex int
				CustomIndex int
				ParserIndex int
			}{
				Filters: map[string][]Field{
					"rewrite_tag.0": []Field{
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
				},
				Inputs: map[string][]Field{},
			}},
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
			expected: struct {
				Inputs      map[string][]Field
				Filters     map[string][]Field
				Customs     map[string][]Field
				Outputs     map[string][]Field
				Parsers     map[string][]Field
				InputIndex  int
				FilterIndex int
				OutputIndex int
				CustomIndex int
				ParserIndex int
			}{
				Filters: map[string][]Field{
					"rewrite_tag.0": []Field{
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
				},
				Inputs: map[string][]Field{"tail.0": []Field{
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
				}},
			}},
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
			expected: struct {
				Inputs      map[string][]Field
				Filters     map[string][]Field
				Customs     map[string][]Field
				Outputs     map[string][]Field
				Parsers     map[string][]Field
				InputIndex  int
				FilterIndex int
				OutputIndex int
				CustomIndex int
				ParserIndex int
			}{
				Inputs: map[string][]Field{"tail.0": []Field{
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
				}},
				Filters: map[string][]Field{},
				Customs: map[string][]Field{},
				Outputs: map[string][]Field{"s3.0": []Field{
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
				}},
				Parsers: map[string][]Field{},
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
			expected: struct {
				Inputs      map[string][]Field
				Filters     map[string][]Field
				Customs     map[string][]Field
				Outputs     map[string][]Field
				Parsers     map[string][]Field
				InputIndex  int
				FilterIndex int
				OutputIndex int
				CustomIndex int
				ParserIndex int
			}{
				Inputs: map[string][]Field{
					"tcp.0": {
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
					"tcp.1": {
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
				},
				Filters: map[string][]Field{},
				Customs: map[string][]Field{},
				Outputs: map[string][]Field{"stdout.0": []Field{
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
				}},
				Parsers: map[string][]Field{},
			},
			expectedError: false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := NewFromBytes(tc.config)
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
			if want, got := len(tc.expected.Inputs), len(cfg.Inputs); want != got {
				t.Errorf("wants %v != got %v", want, got)
				return
			}

			for idx, input := range tc.expected.Inputs {
				if want, got := len(input), len(cfg.Inputs[idx]); want != got {
					t.Errorf("wants %v != got %v", want, got)
					return
				}
			}

			if want, got := len(tc.expected.Filters), len(cfg.Filters); want != got {
				t.Errorf("wants %v != got %v", want, got)
				return
			}

			for idx, filter := range tc.expected.Filters {
				if want, got := len(filter), len(cfg.Filters[idx]); want != got {
					t.Errorf("wants %v != got %v", want, got)
					return
				}
			}

			if want, got := len(tc.expected.Outputs), len(cfg.Outputs); want != got {
				t.Errorf("wants %v != got %v", want, got)
				return
			}

			for idx, output := range tc.expected.Outputs {
				if want, got := len(output), len(cfg.Outputs[idx]); want != got {
					t.Errorf("wants %v != got %v", want, got)
					return
				}
			}

			if want, got := len(tc.expected.Customs), len(cfg.Customs); want != got {
				t.Errorf("wants %v != got %v", want, got)
				return
			}

			for idx, custom := range tc.expected.Customs {
				if want, got := len(custom), len(cfg.Customs[idx]); want != got {
					t.Errorf("wants %v != got %v", want, got)
					return
				}
			}
		})
	}
}
