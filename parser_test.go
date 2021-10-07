package fluentbit_config

import (
	"testing"
)

func stringPtr(s string) *string {
	return &s
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
				Inputs  map[string][]Field
				Filters map[string][]Field
				Customs map[string][]Field
				Outputs map[string][]Field
			}{Inputs: map[string][]Field{}, Filters: map[string][]Field{}, Customs: map[string][]Field{}, Outputs: map[string][]Field{}},
			expectedError: true,
		},
		{
			name:   "test invalid configuration",
			config: []byte(""),
			expected: struct {
				Inputs  map[string][]Field
				Filters map[string][]Field
				Customs map[string][]Field
				Outputs map[string][]Field
			}{Inputs: map[string][]Field{}, Filters: map[string][]Field{}, Customs: map[string][]Field{}, Outputs: map[string][]Field{}},
			expectedError: true,
		},
		{
			name: "test valid simple configuration",
			config: []byte(`
				[INPUT]
					name tail
					tag kube.*
					mem_buf_limit 4.8M
			`),
			expected: struct {
				Inputs  map[string][]Field
				Filters map[string][]Field
				Customs map[string][]Field
				Outputs map[string][]Field
			}{Inputs: map[string][]Field{"tail": []Field{
				{Key: "name", Value: &Value{
					String:   stringPtr("tail"),
					DateTime: nil,
					Date:     nil,
					Time:     nil,
					Bool:     nil,
					Number:   nil,
					Float:    nil,
					List:     nil,
				}},
				{Key: "tag", Value: &Value{
					String:   stringPtr("kube.*"),
					DateTime: nil,
					Date:     nil,
					Time:     nil,
					Bool:     nil,
					Number:   nil,
					Float:    nil,
					List:     nil,
				}},
				{Key: "mem_buf_limit", Value: &Value{
					String:   stringPtr("4.8M"),
					DateTime: nil,
					Date:     nil,
					Time:     nil,
					Bool:     nil,
					Number:   nil,
					Float:    nil,
					List:     nil,
				}},
			}}, Filters: map[string][]Field{}, Customs: map[string][]Field{}, Outputs: map[string][]Field{}},
			expectedError: false,
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
			expected: struct {
				Inputs  map[string][]Field
				Filters map[string][]Field
				Customs map[string][]Field
				Outputs map[string][]Field
			}{
				Inputs: map[string][]Field{"tail": []Field{
					{Key: "name", Value: &Value{
						String:   stringPtr("tail"),
						DateTime: nil,
						Date:     nil,
						Time:     nil,
						Bool:     nil,
						Number:   nil,
						Float:    nil,
						List:     nil,
					}},
					{Key: "tag", Value: &Value{
						String:   stringPtr("tail.01"),
						DateTime: nil,
						Date:     nil,
						Time:     nil,
						Bool:     nil,
						Number:   nil,
						Float:    nil,
						List:     nil,
					}},
					{Key: "Path", Value: &Value{
						String:   stringPtr("/var/log/system.log"),
						DateTime: nil,
						Date:     nil,
						Time:     nil,
						Bool:     nil,
						Number:   nil,
						Float:    nil,
						List:     nil,
					}},
				}},
				Filters: map[string][]Field{},
				Customs: map[string][]Field{},
				Outputs: map[string][]Field{"s3": []Field{
					{Key: "name", Value: &Value{
						String:   stringPtr("s3"),
						DateTime: nil,
						Date:     nil,
						Time:     nil,
						Bool:     nil,
						Number:   nil,
						Float:    nil,
						List:     nil,
					}},
					{Key: "Match", Value: &Value{
						String:   stringPtr("*"),
						DateTime: nil,
						Date:     nil,
						Time:     nil,
						Bool:     nil,
						Number:   nil,
						Float:    nil,
						List:     nil,
					}},
					{Key: "bucket", Value: &Value{
						String:   stringPtr("your-bucket"),
						DateTime: nil,
						Date:     nil,
						Time:     nil,
						Bool:     nil,
						Number:   nil,
						Float:    nil,
						List:     nil,
					}},
				}},
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
			if err != nil {
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
