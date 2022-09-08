package fluentbit_config

import (
	_ "embed"
	"testing"
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
			// TODO: fix float not conversing its original value.
			want: `cpu: expected "pid" to be a valid integer; got "3.400000"`,
		},
		{
			name: "input_cpu_pid_bool",
			ini: `
				[INPUT]
					Name cpu
					pid  true
			`,
			want: `cpu: expected "pid" to be a valid integer; got "true"`,
		},
		{
			name: "input_cpu_pid_bytes",
			ini: `
				[INPUT]
					Name cpu
					pid  1024k
			`,
			// TODO: fix byte size not being parsed.
			want: `cpu: expected "pid" to be a valid integer; got ""`,
		},
		{
			name: "input_cpu_pid_string",
			ini: `
				[INPUT]
					Name cpu
					pid  foo
			`,
			want: `cpu: expected "pid" to be a valid integer; got "foo"`,
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
			want: `tail: expected "docker_mode" to be a valid boolean; got "foo"`,
		},
		{
			name: "input_tail_docker_mode_integer",
			ini: `
				[INPUT]
					Name        tail
					docker_mode 5
			`,
			want: `tail: expected "docker_mode" to be a valid boolean; got "5"`,
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
			want: `throttle: expected "rate" to be a valid double; got "5"`,
		},
		{
			name: "filter_throttle_rate_string",
			ini: `
				[FILTER]
					Name throttle
					rate foo
			`,
			want: `throttle: expected "rate" to be a valid double; got "foo"`,
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
			want: `syslog: expected "syslog_maxsize" to be a valid size; got "foo"`,
		},
		{
			name: "output_syslog_syslog_maxsize_bool",
			ini: `
				[OUTPUT]
					Name syslog
					syslog_maxsize true
			`,
			want: `syslog: expected "syslog_maxsize" to be a valid size; got "true"`,
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
			// TODO: fix parser not parsing unit.
			want: `syslog: expected "syslog_maxsize" to be a valid size; got " B"`,
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
		},
		{
			name: "output_syslog_syslog_maxsize_bytefmt",
			ini: `
				[OUTPUT]
					Name syslog
					syslog_maxsize 5 bytes
			`,
		},
		{
			name: "filter_record_modifier_record_space_delimited_strings_2",
			ini: `
				[FILTER]
					Name record_modifier
					record foo
			`,
			want: `record_modifier: expected "record" to be a valid space delimited strings (minimum 2); got "foo"`,
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
		// TODO: fix parser not handling comma separated strings.
		// {
		// 	name: "input_tail_path_comma_delimited_strings",
		// 	ini: `
		// 		[INPUT]
		// 			Name tail
		// 			path foo,bar
		// 	`,
		// },
		{
			name: "input_systemd_path_no_string",
			ini: `
				[INPUT]
					Name systemd
					path
			`,
			want: `systemd: expected "path" to be a valid string; got ""`,
		},
		{
			name: "input_systemd_path_bool",
			ini: `
				[INPUT]
					Name systemd
					path true
			`,
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
			want: "unknown plugin \"nope\"",
		},
		{
			name: "lts_gsuite_reporter_unknown_property",
			ini: `
				[INPUT]
					Name gsuite-reporter
					nope test
			`,
			want: "gsuite-reporter: unknown property \"nope\"",
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
			name: "output_http_with_host_and_port",
			ini: `
				[OUTPUT]
					Name http
					Host api.example.org
					Port 443
			`,
		},
		// TODO: fix parser
		// {
		// 	name: "output_http_header",
		// 	ini: `
		// 		[OUTPUT]
		// 			Name   http
		// 			Header private_key 008(test)
		// 	`,
		// },
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := ParseINI([]byte(tc.ini))
			if err != nil {
				t.Errorf("parsing ini: %v\n", err)
				return
			}

			err = cfg.Validate()
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
