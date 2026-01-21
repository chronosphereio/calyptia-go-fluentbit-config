package fluentbitconfig

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
	yaml "github.com/goccy/go-yaml"
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
		{
			name: "input_aws_kinesis_stream_correct",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: aws_kinesis_stream
					  aws_access_key_id: test
					  aws_secret_access_key: test
					  aws_region: test
					  stream_name: test
					  empty_interval: 1s
					  limit: 1
					  data_dir: /test/test
			`),
			want: "",
		},
		{
			name: "input_aws_kinesis_stream_incorrect_key",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: aws_kinesis_stream
					  incorrect_key: foo
			`),
			want: "input: aws_kinesis_stream: unknown property \"incorrect_key\"",
		},
		{
			name: "input_azeventgrid_correct",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: azeventgrid
					  topicName: test
					  eventSubscriptionName: test
					  endpoint: test
					  key: test
			`),
			want: "",
		},
		{
			name: "input_azeventgrid_incorrect_key",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: azeventgrid
					  incorrect_key: foo
			`),
			want: "input: azeventgrid: unknown property \"incorrect_key\"",
		},
		{
			name: "input_azure_blob_input_correct",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: azure-blob-input
					  tag: blob
					  account_name: test_account
					  connection_string: "DefaultEndpointsProtocol=http;AccountName=test_account;AccountKey=foo;BlobEndpoint=http://127.0.0.1:10000/test_account;"
					  container: test_container
					  service_url: test_service_url
					  tenant_id: test_tenant_id
			`),
			want: "",
		},
		{
			name: "input_azure_blob_input_incorrect_key",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: azure-blob-input
					  incorrect_key: foo
			`),
			want: "input: azure-blob-input: unknown property \"incorrect_key\"",
		},
		{
			name: "input_cloudflare_correct",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: cloudflare
					  addr: test
					  resp_headers: "User Agent: test"
					  resp_status_code: 200
					  resp_body: test
					  time_from: test
					  cert_file: /test/cert_file
					  key_file: /test/key_file
					  http_user: test
					  http_passwd: test
					  cloudflareApiKey: test
					  cloudflareEmail: test
					  destination: test
					  cloudflareAccountId: test
					  cloudflareZoneId: test
					  skipOwnershipChallenge: false
					  baseUrl: http://test
			`),
			want: "",
		},
		{
			name: "input_cloudflare_incorrect_key",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: cloudflare
					  incorrect_key: foo
			`),
			want: "input: cloudflare: unknown property \"incorrect_key\"",
		},
		{
			name: "input_datagen_correct",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: datagen
					  template: test
					  rate: 1s
			`),
			want: "",
		},
		{
			name: "input_datagen_incorrect_key",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: datagen
					  incorrect_key: foo
			`),
			want: "input: datagen: unknown property \"incorrect_key\"",
		},
		{
			name: "input_gdummy_correct",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: gdummy
			`),
			want: "",
		},
		{
			name: "input_gdummy_incorrect_key",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: gdummy
					  incorrect_key: foo
			`),
			want: "input: gdummy: unknown property \"incorrect_key\"",
		},
		{
			name: "input_go-s3-replay-plugin_correct",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: go-s3-replay-plugin
					  aws_access_key: test
					  aws_secret_key: test
					  aws_bucket_name: test
					  aws_bucket_region: test
					  aws_s3_endpoint: test
					  aws_s3_role_arn: test
					  aws_s3_role_session_name: test
					  aws_s3_role_external_id: test
					  aws_s3_role_duration: 1s
					  logs: test
					  s3_read_concurrency: 1
					  max_line_buffer_size: 10 MiB
			`),
			want: "",
		},
		{
			name: "input_go-s3-replay-plugin_incorrect_key",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: go-s3-replay-plugin
					  incorrect_key: foo
			`),
			want: "input: go-s3-replay-plugin: unknown property \"incorrect_key\"",
		},
		{
			name: "input_gsuite-reporter_correct",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: gsuite-reporter
					  access_token: test
					  creds_file: /test/file
					  subject: test
					  pull_interval: 1s
					  data_dir: /test/dir
					  telemetry: false
					  application_name: test
					  user_key: test
			`),
			want: "",
		},
		{
			name: "input_gsuite-reporter_incorrect_key",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: gsuite-reporter
					  incorrect_key: foo
			`),
			want: "input: gsuite-reporter: unknown property \"incorrect_key\"",
		},
		{
			name: "input_http_loader_correct",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: http_loader
					  method: POST
					  body: test
					  url: http://test
					  wait: test
					  stop: test
					  header: "User-Agent: test"
					  proxy: test
					  tls_cert_file: /test/file
					  tls_key_file: /test/file
					  tls_cert: test
					  tls_key: test
					  ca_cert_file: /test/file
					  ca_cert: test
					  oauth2_token_url: http://test
					  oauth2_client_id: test
					  oauth2_client_secret: test
					  oauth2_scopes: "test1 test2"
					  oauth2_endpoint_params: test
					  timeout: 1s
					  pull_interval: 10s
					  auth_cookie_url: http://test
					  auth_cookie_method: POST
					  auth_cookie_header: "User-Agent: test"
					  auth_cookie_body: test
					  auth_cookie_exp: 100s
					  auth_digest_username: test
					  auth_digest_password: test
					  skip: test
					  out: test
					  data_dir: /test/dir
					  data_exp: 10m
					  retry: test
					  max_retries: test
					  store_response_body: test
			`),
			want: "",
		},
		{
			name: "input_http_loader_incorrect_key",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: http_loader
					  incorrect_key: foo
			`),
			want: "input: http_loader: unknown property \"incorrect_key\"",
		},
		{
			name: "input_http_scraper_correct",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: http_scraper
					  method: POST
					  url: http://test
					  body: test
					  headers: "User-Agent: test"
					  headers_separator: ","
					  pull_interval: 1s
					  timeout: 1s
					  continue_on_error: true
					  max_response_bytes: 1024
					  go_template: test
					  oauth2_client_id: test
					  oauth2_client_secret: test
					  oauth2_token_url: http://test
					  oauth2_scopes_separator: ","
					  oauth2_scopes: "test1,test2"
					  oauth2_endpoint_params: test
			`),
			want: "",
		},
		{
			name: "input_http_scraper_incorrect_key",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: http_scraper
					  incorrect_key: foo
			`),
			want: "input: http_scraper: unknown property \"incorrect_key\"",
		},
		{
			name: "input_s3_sqs_correct",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: s3_sqs
					  delete_messages: true
					  aws_s3_role_arn: test
					  aws_s3_role_session_name: test
					  aws_s3_role_external_id: test
					  aws_s3_role_duration: 1m
					  aws_sqs_role_arn: test
					  aws_sqs_role_session_name: test
					  aws_sqs_role_external_id: test
					  aws_sqs_role_duration: 1m
					  aws_access_key: test
					  aws_secret_key: test
					  aws_bucket_name: test
					  aws_bucket_region: test
					  aws_s3_endpoint: test
					  match_regexp: test
					  aws_s3_enable_imds: true
					  sqs_queue_name: test
					  sqs_queue_region: test
					  aws_sqs_endpoint: test
					  aws_sqs_enable_imds: true
					  max_line_buffer_size: 10 MB
					  s3_read_concurrency: 10
			`),
			want: "",
		},
		{
			name: "input_s3_sqs_incorrect_key",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: s3_sqs
					  incorrect_key: foo
			`),
			want: "input: s3_sqs: unknown property \"incorrect_key\"",
		},

		{
			name: "input_sqldb_correct",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: sqldb
					  driver: mysql
					  dsn: test
					  query: SELECT \"test\" AS test
					  columnsForArgs: test
					  timeFrom: test
					  fetchInterval: 100s
					  storageKey: test
					  dataDir: /test/data
			`),
			want: "",
		},
		{
			name: "input_sqldb_incorrect_key",
			yaml: configLiteral(`
			pipeline:
				inputs:
					- name: sqldb
					  incorrect_key: foo
			`),
			want: "input: sqldb: unknown property \"incorrect_key\"",
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
