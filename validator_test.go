package fluentbitconfig

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConfig_Validate(t *testing.T) {
	tt := []struct {
		name string
		ini  string
		want string
	}{
		{
			name: "service_with_cloud_variable",
			ini: `
				[SERVICE]
					Parsers_File {{ files.parszerz }}
				[INPUT]
					Name            forward
					Buffer_Max_Size {{ secrets.size }}
			`,
		},
		{
			name: "input_cpu_pid_float",
			ini: `
				[INPUT]
					Name cpu
					pid  3.4
			`,
			want: `input: cpu: expected "pid" to be a valid integer, got 3.4`,
		},
		{
			name: "input_cpu_pid_bool",
			ini: `
				[INPUT]
					Name cpu
					pid  true
			`,
			want: `input: cpu: expected "pid" to be a valid integer, got true`,
		},
		{
			name: "input_cpu_pid_bytes",
			ini: `
				[INPUT]
					Name cpu
					pid  1024k
			`,
			want: `input: cpu: expected "pid" to be a valid integer, got 1024k`,
		},
		{
			name: "input_cpu_pid_string",
			ini: `
				[INPUT]
					Name cpu
					pid  foo
			`,
			want: `input: cpu: expected "pid" to be a valid integer, got foo`,
		},
		{
			name: "input_cpu_pid_integer",
			ini: `
				[INPUT]
					Name cpu
					pid  5
			`,
		},
		{
			name: "input_tail_docker_mode_string",
			ini: `
				[INPUT]
					Name        tail
					docker_mode foo
			`,
			want: `input: tail: expected "docker_mode" to be a valid boolean, got foo`,
		},
		{
			name: "input_tail_docker_mode_integer",
			ini: `
				[INPUT]
					Name        tail
					docker_mode 5
			`,
			want: `input: tail: expected "docker_mode" to be a valid boolean, got 5`,
		},
		{
			name: "input_tail_docker_mode_boolean",
			ini: `
				[INPUT]
					Name        tail
					docker_mode true
			`,
		},
		{
			name: "filter_throttle_rate_integer",
			ini: `
				[FILTER]
					Name throttle
					rate 5
			`,
			want: `filter: throttle: expected "rate" to be a valid double, got 5`,
		},
		{
			name: "filter_throttle_rate_string",
			ini: `
				[FILTER]
					Name throttle
					rate foo
			`,
			want: `filter: throttle: expected "rate" to be a valid double, got foo`,
		},
		{
			name: "filter_throttle_rate_double",
			ini: `
				[FILTER]
					Name throttle
					rate 3.5
			`,
		},
		{
			name: "output_syslog_syslog_maxsize_string",
			ini: `
				[OUTPUT]
					Name syslog
					syslog_maxsize foo
			`,
			want: `output: syslog: expected "syslog_maxsize" to be a valid size, got foo`,
		},
		{
			name: "output_syslog_syslog_maxsize_bool",
			ini: `
				[OUTPUT]
					Name syslog
					syslog_maxsize true
			`,
			want: `output: syslog: expected "syslog_maxsize" to be a valid size, got true`,
		},
		{
			name: "output_syslog_syslog_maxsize_no_unit",
			ini: `
				[OUTPUT]
					Name syslog
					syslog_maxsize 1024
			`,
		},
		{
			name: "output_syslog_syslog_maxsize_size_no_spaces",
			ini: `
				[OUTPUT]
					Name syslog
					syslog_maxsize 5MB
			`,
		},
		{
			name: "output_syslog_syslog_maxsize_size_with_space",
			ini: `
				[OUTPUT]
					Name syslog
					syslog_maxsize 5 MB
			`,
		},
		{
			name: "output_syslog_syslog_maxsize_bytefmt",
			ini: `
				[OUTPUT]
					Name syslog
					syslog_maxsize 5 bytes
			`,
			want: `output: syslog: expected "syslog_maxsize" to be a valid size, got 5 bytes`,
		},
		{
			name: "filter_record_modifier_record_space_delimited_strings_2",
			ini: `
				[FILTER]
					Name record_modifier
					record foo
			`,
			want: `filter: record_modifier: expected "record" to be a valid space delimited strings (minimum 2), got foo`,
		},
		{
			name: "filter_record_modifier_record_space_delimited_strings_2",
			ini: `
				[FILTER]
					Name record_modifier
					record foo bar
			`,
		},
		{
			name: "filter_record_modifier_record_space_delimited_strings_2",
			ini: `
				[FILTER]
					Name record_modifier
					record 1 true
			`,
		},
		{
			name: "input_tail_path_comma_delimited_strings",
			ini: `
				[INPUT]
					Name tail
					path foo
			`,
		},
		{
			name: "input_tail_path_comma_delimited_strings",
			ini: `
				[INPUT]
					Name tail
					path foo,bar
			`,
		},
		{
			name: "input_systemd_path_bool",
			ini: `
				[INPUT]
					Name systemd
					path true
			`,
			// `true` is a valid string too.
			// want: `input: systemd: expected "path" to be a valid string, got true`,
		},
		{
			name: "input_systemd_path_string",
			ini: `
				[INPUT]
					Name systemd
					path foo
			`,
		},
		// TODO: fix parser not uncluding unknown sections.
		// {
		// 	name: "unknown_section",
		// 	ini: `
		// 		[NOPE]
		// 			Name test
		// 	`,
		// 	want: "unknown section \"NOPE\"",
		// },
		{
			name: "unknown_plugin",
			ini: `
				[INPUT]
					Name nope
			`,
			want: `input: unknown plugin "nope"`,
		},
		{
			name: "lts_gsuite_reporter_unknown_property",
			ini: `
				[INPUT]
					Name gsuite-reporter
					nope test
			`,
			want: `input: gsuite-reporter: unknown property "nope"`,
		},
		{
			name: "lts_gsuite_reporter_pull_interval",
			ini: `
				[INPUT]
					Name 		  gsuite-reporter
					pull_interval 10s
			`,
		},
		{
			name: "input_http_with_host_and_port",
			ini: `
				[INPUT]
					Name http
					Host api.example.org
					Port 443
			`,
		},
		{
			name: "output_http_with_host_and_port",
			ini: `
				[OUTPUT]
					Name http
					Host api.example.org
					Port 443
			`,
		},
		{
			name: "output_http_header",
			ini: `
				[OUTPUT]
					Name   http
					Header private_key 008(test)
			`,
		},
		{
			name: "input_tail_refresh_interval",
			ini: `
				[INPUT]
					Name   tail
					refresh_interval 30
			`,
		},
		{
			name: "http_loader_ok",
			ini: `
				[INPUT]
					Name        http_loader
					method      POST
					url         https://example.org
					header      Authorization Bearer 123
					body        {"foo": "bar"}
					out         {{toJson .body}}
					skip        false
					stop        true
					retry       true
					max_retries 3
					timeout     5s
			`,
		},
		{
			name: "out_splunk_event_field",
			// out_splunk event_field is: space delimited strings (minimum 2)
			ini: `
				[OUTPUT]
					Name        splunk
					event_field $source foo
					event_field $vendor bar
					event_field $product baz
			`,
		},
		{
			name: "out_stackdriver_labels",
			// out_stackdriver labels is: multiple comma delimited strings
			ini: `
				[OUTPUT]
					Name   stackdriver
					labels foo
					labels bar
			`,
		},
		{
			name: "in_sqldb_unknown_property",
			ini: `
				[INPUT]
					Name sqldb
					nope test
			`,
			want: `input: sqldb: unknown property "nope"`,
		},
		{
			name: "in_sqldb_ok",
			ini: `
				[INPUT]
					Name   sqldb
					driver slite
					dsn   file::memory:?cache=shared
					query SELECT 'hello from sqldb' AS message
			`,
		},
		{
			name: "custom_core_property",
			ini: `
				[INPUT]
					Name dummy
				[OUTPUT]
					Name s3
					core.metadata name=foo
			`,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var classicConf Config
			err := classicConf.UnmarshalClassic([]byte(tc.ini))
			require.NoError(t, err)

			err = classicConf.Validate()
			if tc.want == "" && err == nil {
				return
			}

			if tc.want == "" && err != nil {
				t.Error(err)
				return
			}

			if tc.want != "" && err == nil {
				t.Errorf("expected error:\n%s\ngot: nil\n", tc.want)
				return
			}

			if tc.want != err.Error() {
				t.Errorf("expected error:\n%s\ngot:\n%s\n", tc.want, err.Error())
				return
			}
		})
	}
}

func TestConfig_Validate_Schema_YAML(t *testing.T) {
	tt := []struct {
		name string
		yaml string
		want string
	}{
		// fluent-bit does not support using arrays under
		// the service section under any circumstance.
		{
			name: "service_with_arrays",
			yaml: configLiteral(`
				service:
					parsers:
						- one
						- two
			`),
			want: "invalid type ([]interface {}) for service setting: parsers",
		},
		// fluent-bit does not support using mappings under
		// the service section under any circumstance.
		{
			name: "service_with_mapping",
			yaml: configLiteral(`
				service:
					headers:
						one: two
			`),
			want: "invalid type (map[string]interface {}) for service setting: headers",
		},
		// it does support booleans
		{
			name: "service_with_http_server",
			yaml: configLiteral(`
				service:
					http_server: true
			`),
			want: "",
		},
		// integers
		{
			name: "service_with_http_port",
			yaml: configLiteral(`
				service:
					http_port: 2020
			`),
			want: "",
		},
		// and strings
		{
			name: "service_with_string_setting",
			yaml: configLiteral(`
				service:
					label: FOOBAR
			`),
			want: "",
		},
		// and floating point
		{
			name: "service_with_flush_microseconds",
			yaml: configLiteral(`
				service:
					flush: 0.2
			`),
			want: "",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var yamlConf Config
			err := yaml.Unmarshal([]byte(tc.yaml), &yamlConf)
			require.NoError(t, err)

			err = yamlConf.Validate()
			if tc.want == "" && err == nil {
				return
			}

			if tc.want == "" && err != nil {
				t.Error(err)
				return
			}

			if tc.want != "" && err == nil {
				t.Errorf("expected error:\n%s\ngot: nil\n", tc.want)
				return
			}

			if tc.want != err.Error() {
				t.Errorf("expected error:\n%s\ngot:\n%s\n", tc.want, err.Error())
				return
			}
		})
	}
}

func TestConfig_Validate_SchemaWithVersion(t *testing.T) {
	// Default Schema
	t.Run("default_schema", func(t *testing.T) {
		ini := `
			[INPUT]
				Name dummy
			[OUTPUT]
				Name s3
				core.metadata name=foo
		`
		var conf Config
		err := conf.UnmarshalClassic([]byte(ini))
		require.NoError(t, err)

		err = conf.ValidateWithSchema(DefaultSchema)
		require.NoError(t, err)
	})

	// Load oldest schema
	t.Run("oldest_schema", func(t *testing.T) {
		ini := `
			[INPUT]
				Name dummy
			[OUTPUT]
				Name s3
				core.metadata name=foo
		`
		var conf Config
		err := conf.UnmarshalClassic([]byte(ini))
		require.NoError(t, err)

		schema, err := GetSchema("22.7.2")
		require.NoError(t, err)

		err = conf.ValidateWithSchema(schema)
		require.NoError(t, err)
	})

	t.Run("some_random_version", func(t *testing.T) {
		ini := `
			[INPUT]
				Name dummy
			[OUTPUT]
				Name s3
				core.metadata name=foo
		`
		var conf Config
		err := conf.UnmarshalClassic([]byte(ini))
		require.NoError(t, err)

		schema, err := GetSchema("25.1.1")
		require.NoError(t, err)

		err = conf.ValidateWithSchema(schema)
		require.NoError(t, err)
	})
}
