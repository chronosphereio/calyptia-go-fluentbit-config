package fluentbitconfig

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//go:embed schemas/25.4.2.json
var rawSchema []byte

var DefaultSchema = func() Schema {
	var s Schema
	err := json.Unmarshal(rawSchema, &s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse fluent-bit schema: %v\n", err)
		os.Exit(1)
	}

	s.InjectLTSPlugins()

	return s
}()

type Schema struct {
	FluentBit SchemaFluentBit `json:"fluent-bit"`
	Customs   []SchemaSection `json:"customs"`
	Inputs    []SchemaSection `json:"inputs"`
	Filters   []SchemaSection `json:"filters"`
	Outputs   []SchemaSection `json:"outputs"`
}

type SchemaFluentBit struct {
	Version       string `json:"version"`
	SchemaVersion string `json:"schema_version"`
	OS            string `json:"os"`
}

type SchemaSection struct {
	Type        string           `json:"type"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Properties  SchemaProperties `json:"properties"`
}

type SchemaProperties struct {
	Options []SchemaOptions `json:"options"`

	// Networking is only present in outputs.
	Networking []SchemaOptions `json:"networking"`
	// NetworkTLS is only present in outputs.
	NetworkTLS []SchemaOptions `json:"network_tls"`
}

func (sec SchemaSection) findOptions(name string) (SchemaOptions, bool) {
	for _, opts := range sec.Properties.all() {
		if strings.EqualFold(opts.Name, name) {
			return opts, true
		}
	}

	return SchemaOptions{}, false
}

func (pp SchemaProperties) all() []SchemaOptions {
	var out []SchemaOptions
	out = append(out, pp.Options...)
	out = append(out, pp.Networking...)
	out = append(out, pp.NetworkTLS...)
	return out
}

type SchemaOptions struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     any    `json:"default,omitempty"`
	Type        string `json:"type"`
}

func (s Schema) findSection(kind SectionKind, name string) (SchemaSection, bool) {
	sections, ok := s.findSections(kind)
	if !ok {
		return SchemaSection{}, false
	}

	for _, section := range sections {
		if strings.EqualFold(section.Name, name) {
			return section, true
		}
	}

	return SchemaSection{}, false
}

func (s Schema) findSections(kind SectionKind) ([]SchemaSection, bool) {
	switch kind {
	case SectionKindCustom:
		return s.Customs, true
	case SectionKindInput:
		return s.Inputs, true
	case SectionKindFilter:
		return s.Filters, true
	case SectionKindOutput:
		return s.Outputs, true
	default:
		return nil, false
	}
}

func (s *Schema) InjectLTSPlugins() {
	s.Inputs = append(s.Inputs, SchemaSection{
		// See https://github.com/chronosphereio/calyptia-core-fluent-bit/tree/main/goplugins/go-s3-replay-plugin
		Type:        "input",
		Name:        "go-s3-replay-plugin",
		Description: "Calyptia LTS advanced plugin providing logs replay from s3",
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{
					Name: "aws_access_key",
					Type: "string",
				},
				{
					Name: "aws_secret_key",
					Type: "string",
				},
				{
					Name: "aws_bucket_name",
					Type: "string",
				},
				{
					Name:        "aws_s3_endpoint",
					Type:        "string",
					Description: "Either aws_s3_endpoint or aws_bucket_region has to be provided",
				},
				{
					Name:        "aws_bucket_region",
					Type:        "string",
					Description: "Either aws_s3_endpoint or aws_bucket_region has to be provided",
				},
				{
					Name:        "logs",
					Type:        "string",
					Description: "Log pattern",
				},
			},
		},
	}, SchemaSection{
		// See https://github.com/chronosphereio/calyptia-core-fluent-bit/tree/main/goplugins/dummy
		Type:        "input",
		Name:        "gdummy",
		Description: "dummy GO!",
	}, SchemaSection{
		// See https://github.com/chronosphereio/calyptia-core-fluent-bit/tree/main/goplugins/gsuite-reporter
		Type:        "input",
		Name:        "gsuite-reporter",
		Description: "A Calyptia LTS advanced plugin providing activity streams from Gsuite",
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{
					Name:        "access_token",
					Type:        "string",
					Description: "Set either access_token or creds_file",
				},
				{
					Name:        "creds_file",
					Type:        "string",
					Description: "Service account or refresh token. It must set subject if this is not empty",
				},
				{
					Name:        "subject",
					Type:        "string",
					Description: "Email of the user to impersonate",
				},
				{
					Name:        "pull_interval",
					Type:        "string",
					Description: "Must be >= 1s",
					Default:     "30s",
				},
				{
					Name:        "data_dir",
					Type:        "string",
					Description: "Directory to store data. It holds a 1MB cache",
					Default:     "/data/storage",
				},
				{
					Name:        "telemetry",
					Type:        "boolean",
					Description: "Enable telemetry on Gsuite API",
				},
				{
					Name:        "application_name",
					Type:        "string",
					Description: "See https://developers.google.com/admin-sdk/reports/reference/rest/v1/activities/list#ApplicationName",
				},
				{
					Name:        "user_key",
					Type:        "string",
					Description: "Profile ID or the user email for which the data should be filtered. Can be `all` for all information, or `userKey` for a user's unique Google Workspace profile ID or their primary email address",
					Default:     "all",
				},
			},
		},
	}, SchemaSection{
		// See https://github.com/chronosphereio/calyptia-core-fluent-bit/tree/main/goplugins/http_loader
		Type:        "input",
		Name:        "http_loader",
		Description: "HTTP Loader plugin provides a way to load/dump data from a paginated HTTP endpoint.",
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{
					Name:        "method",
					Type:        "string",
					Description: `Request method. Defaults to "GET", or "POST" if ` + "`body`" + ` is set. Supports templating.`,
					Default:     "GET",
				},
				{
					Name:        "url",
					Type:        "string",
					Description: `Request URL. Required. Supports templating.`,
				},
				{
					Name:        "header",
					Type:        "string",
					Description: "Request headers, string separated by new line character `\n`. Supports templating.",
					Default:     "User-Agent: Fluent-Bit HTTP Loader Plugin",
				},
				{
					Name:        "body",
					Type:        "string",
					Description: "Request body. Supports templating.",
				},
				{
					Name:        "oauth2_token_url",
					Type:        "string",
					Description: "OAuth2 token endpoint at where to exchange a token. Enables OAuth2 using the client-credentials flow.",
				},
				{
					Name:        "oauth2_client_id",
					Type:        "string",
					Description: "OAuth2 client ID.",
				},
				{
					Name:        "oauth2_client_secret",
					Type:        "string",
					Description: "OAuth2 client secret. Sensible field, prefer using pipeline secrets.",
				},
				{
					Name:        "oauth2_scopes",
					Type:        "string",
					Description: "OAuth2 scopes, string, each scope separated by space.",
				},
				{
					Name:        "oauth2_endpoint_params",
					Type:        "string",
					Description: "OAuth2 endpoint params, string in URL query string format.",
				},
				{
					Name:        "timeout",
					Type:        "string",
					Description: "Controls the request timeout, string duration. If set, it must be greater than 0s.",
					Default:     "0s",
				},
				{
					Name:        "pull_interval",
					Type:        "string",
					Description: "Controls the time between requests, string duration. If set, it must be greater than 0s. Supports templating.",
					Default:     "1s",
				},
				{
					Name:        "auth_cookie_url",
					Type:        "string",
					Description: "Cookie based authentication URL.",
				},
				{
					Name:        "auth_cookie_method",
					Type:        "string",
					Description: `Cookie based authentication request method. Defaults to "GET", or "POST" if ` + "`auth_cookie_body`" + ` is set.`,
					Default:     "GET",
				},
				{
					Name:        "auth_cookie_header",
					Type:        "string",
					Description: "Cookie based authentication request headers. String separated by new line character `\n`.",
				},
				{
					Name:        "auth_cookie_body",
					Type:        "string",
					Description: "Cookie based authentication request body.",
				},
				{
					Name:        "auth_digest_username",
					Type:        "string",
					Description: "Username for HTTP Digest authentication.",
				},
				{
					Name:        "auth_digest_password",
					Type:        "string",
					Description: "Password for HTTP Digest authentication.",
				},
				{
					Name:        "wait",
					Type:        "string",
					Description: "Controls the time to wait before starting to collect, string duration. If set, it must be greater or equal than 0s. Supports templating.",
					Default:     "0s",
				},
				{
					Name:        "retry",
					Type:        "string",
					Description: "Tells whether to retry a request, boolean. Supports templating.",
					Default:     "false",
				},
				{
					Name:        "max_retries",
					Type:        "string",
					Description: "Controls the maximum number of retries, integer. If set, it must be greater or equal than 0. Supports templating.",
					Default:     "1",
				},
				{
					Name:        "skip",
					Type:        "string",
					Description: "Controls when to skip sending records to fluent-bit, supports templating.\nDefaults to ignore error status codes, and empty response body.",
					Default:     "{{or (ge .Response.StatusCode 400) (empty .Response.Body)}}",
				},
				{
					Name:        "out",
					Type:        "string",
					Description: "Controls what to send to fluent-bit, supports templating. Defaults to send the response body.",
					Default:     "{{toJson .Response.Body}}",
				},
				{
					Name:        "stop",
					Type:        "string",
					Description: "Controls when to stop collecting, supports templating. Defaults to never stop.",
					Default:     "false",
				},
				{
					Name:        "proxy",
					Type:        "string",
					Description: "Proxy URL, allows comma separated list of URLs.",
				},
				{
					Name:        "no_proxy",
					Type:        "string",
					Description: "Exclude URLs from proxy, allows comma separated list of URLs.",
				},
				{
					Name:        "tls_cert_file",
					Type:        "string",
					Description: "TLS certificate file path.",
				},
				{
					Name:        "tls_key_file",
					Type:        "string",
					Description: "TLS key file path.",
				},
				{
					Name:        "tls_cert",
					Type:        "string",
					Description: "TLS certificate in PEM format.",
				},
				{
					Name:        "tls_key",
					Type:        "string",
					Description: "TLS key in PEM format.",
				},
				{
					Name:        "ca_cert_file",
					Type:        "string",
					Description: "CA certificate file path.",
				},
				{
					Name:        "ca_cert",
					Type:        "string",
					Description: "CA certificate in PEM format.",
				},
				{
					Name:        "data_dir",
					Type:        "string",
					Description: "Controls where to store data, data which is used to resume collecting.\nDefaults to `/data/storage` if exists, or a temporary directory if available, otherwise storage is disabled.",
					Default:     "/data/storage",
				},
				{
					Name:        "data_exp",
					Type:        "string",
					Description: "Controls for how much time data can be used after resume.",
					Default:     "0s",
				},
				{
					Name:        "store_response_body",
					Type:        "string",
					Description: "JSON value to store as response body, supports templating.",
					Default:     "{{toJson .Response.Body}}",
				},
			},
		},
	}, SchemaSection{
		// See https://github.com/chronosphereio/calyptia-core-fluent-bit/tree/main/goplugins/azeventgrid
		Type:        "input",
		Name:        "azeventgrid",
		Description: "A Calyptia Core fluent-bit plugin providing input from Azure Event Grid.",
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{
					Name:        "topicName",
					Type:        "string",
					Description: "The name of the topic to subscribe to.",
				},
				{
					Name:        "eventSubscriptionName",
					Type:        "string",
					Description: "The name of the event subscription to subscribe to.",
				},
				{
					Name:        "endpoint",
					Type:        "string",
					Description: "The endpoint domain to use for the subscription.",
				},
				{
					Name:        "key",
					Type:        "string",
					Description: "The key to use to authenticate.",
				},
			},
		},
	}, SchemaSection{
		// See https://github.com/chronosphereio/calyptia-core-fluent-bit/tree/main/goplugins/aws_kinesis_stream
		Type:        "input",
		Name:        "aws_kinesis_stream",
		Description: "AWS Kinesis stream input plugin.",
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{
					Name:        "aws_access_key_id",
					Type:        "string",
					Description: "AWS access key ID.",
				},
				{
					Name:        "aws_secret_access_key",
					Type:        "string",
					Description: "AWS secret access key.",
				},
				{
					Name:        "aws_region",
					Type:        "string",
					Description: "AWS region.",
				},
				{
					Name:        "stream_name",
					Type:        "string",
					Description: "AWS Kinesis stream name.",
				},
				{
					Name:        "empty_interval",
					Type:        "string",
					Description: "Interval to wait for new records when the stream is empty, string duration.",
					Default:     "10s",
				},
				{
					Name:        "limit",
					Type:        "integer",
					Description: "Maximum number of records to read per request, integer.",
				},
				{
					Name:        "data_dir",
					Type:        "string",
					Description: "Directory to store data. It holds a 1MB cache.",
					Default:     "/data/storage",
				},
			},
		},
	}, SchemaSection{
		// See https://github.com/chronosphereio/calyptia-core-fluent-bit-azure-blob-input
		Type:        "input",
		Name:        "flb_core_fluent_bit_azure_blob",
		Description: "Calyptia LTS Azure Blob Storage Input Plugin",
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{
					Name:        "account_name",
					Type:        "string",
					Description: "Azure Storage Account Name",
				},
				{
					Name:        "container",
					Type:        "string",
					Description: "If set, the plugin will only read from this container. Otherwise, it will read from all containers in the account.",
				},
				{
					Name:        "bucket",
					Type:        "string",
					Description: "-",
				},
			},
		},
	}, SchemaSection{
		// See https://github.com/chronosphereio/calyptia-core-fluent-bit/tree/main/goplugins/s3_sqs
		Type:        "input",
		Name:        "s3_sqs",
		Description: "Calyptia LTS advanced plugin providing logs replay from sqs events",
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{
					Name:        "delete_messages",
					Type:        "boolean",
					Description: "If true, messages will be deleted from the queue after being processed.",
					Default:     "true",
				},
				{
					Name:        "aws_s3_role_arn",
					Type:        "string",
					Description: "The role ARN to assume when reading from S3.",
				},
				{
					Name:        "aws_s3_role_session_name",
					Type:        "string",
					Description: "The session name to use when assuming the role.",
				},
				{
					Name:        "aws_s3_role_external_id",
					Type:        "string",
					Description: "The external ID to use when assuming the role.",
				},
				{
					Name:        "aws_s3_role_duration",
					Type:        "string",
					Description: "The duration to assume the role for, string duration.",
				},
				{
					Name:        "aws_sqs_role_arn",
					Type:        "string",
					Description: "The role ARN to assume when reading from SQS.",
				},
				{
					Name:        "aws_sqs_role_session_name",
					Type:        "string",
					Description: "The session name to use when assuming the role.",
				},
				{
					Name:        "aws_sqs_role_external_id",
					Type:        "string",
					Description: "The external ID to use when assuming the role.",
				},
				{
					Name:        "aws_sqs_role_duration",
					Type:        "string",
					Description: "The duration to assume the role for, string duration.",
				},
				{
					Name:        "aws_access_key",
					Type:        "string",
					Description: "AWS access key.",
				},
				{
					Name:        "aws_secret_key",
					Type:        "string",
					Description: "AWS secret key.",
				},
				{
					Name:        "aws_bucket_name",
					Type:        "string",
					Description: "AWS S3 bucket name.",
				},
				{
					Name:        "aws_bucket_region",
					Type:        "string",
					Description: "AWS S3 bucket region.",
				},
				{
					Name:        "match_regexp",
					Type:        "string",
					Description: "The regular expression to match against the SQS message body.",
					Default:     ".*",
				},
				{
					Name:        "aws_s3_enable_imds",
					Type:        "boolean",
					Description: "If true, the plugin will use the Instance Metadata Service to retrieve credentials.",
				},
				{
					Name:        "sqs_queue_name",
					Type:        "string",
					Description: "The name of the SQS queue to read from.",
				},
				{
					Name:        "sqs_queue_region",
					Type:        "string",
					Description: "The region of the SQS queue to read from.",
				},
				{
					Name:        "aws_sqs_endpoint",
					Type:        "string",
					Description: "The endpoint to use when reading from SQS.",
				},
				{
					Name:        "aws_sqs_enable_imds",
					Type:        "boolean",
					Description: "If true, the plugin will use the Instance Metadata Service to retrieve credentials.",
				},
				{
					Name:        "max_line_buffer_size",
					Type:        "size",
					Description: "The maximum size of the line buffer, size.",
					Default:     "10 MB",
				},
				{
					Name:        "s3_read_concurrency",
					Type:        "integer",
					Description: "The number of concurrent S3 reads, integer. Defaults to the number of CPUs.",
				},
			},
		},
	}, SchemaSection{
		// See https://github.com/chronosphereio/calyptia-core-fluent-bit/tree/main/goplugins/datagen
		Type:        "input",
		Name:        "datagen",
		Description: "Datagen input plugin generates fake logs at a given interval",
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{
					Name:        "template",
					Type:        "string",
					Description: "Golang template that evaluates into a JSON string.",
				},
				{
					Name:        "rate",
					Type:        "string",
					Description: "Duration rate at which records are produced.",
					Default:     "1s",
				},
			},
		},
	}, SchemaSection{
		// See https://github.com/chronosphereio/calyptia-core-fluent-bit/tree/main/goplugins/sqldb
		Type:        "input",
		Name:        "sqldb",
		Description: "SQL Database input",
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{
					Name:        "driver",
					Type:        "string",
					Description: "The SQL driver to use. Allowed values are: postgres, mysql, oracle, sqlserver and sqlite.",
					Default:     "postgres",
				},
				{
					Name:        "dsn",
					Type:        "string",
					Description: "Connection string to the database.",
				},
				{
					Name:        "query",
					Type:        "string",
					Description: `SQL query to perform. It supports "@named" arguments. See option columnsForArgs.`,
				},
				{
					Name:        "columnsForArgs",
					Type:        "string",
					Description: "Space separated list of columns. If you want to paginate over the data, you can convert the last row into arguments. Example id created_at will create two arguments that you can use in the query: @last_id and @last_created_at.",
				},
				{
					Name:        "timeFrom",
					Type:        "string",
					Description: "Column from which extract the log ingestion time. If not set, current time will be used.",
				},
				{
					Name:        "timeFormat",
					Type:        "string",
					Description: "Time format to use when timeFrom is set. Besides the standard RFCs time formats to parse strings, it can also parse integers by setting the format to: unix_sec, unix_ms, unix_us, or unix_ns (default for integers).",
					Default:     "2006-01-02T15:04:05.999999999Z07:00",
				},
				{
					Name:        "fetchInterval",
					Type:        "string",
					Description: "Duration between executing each query. Cannot be less than 0.",
					Default:     "1s",
				},
				{
					Name:        "storageKey",
					Type:        "string",
					Description: "Storage key used to store the query arguments to allow restart and resuming. Pass your own key to have more control, or to reset the arguments. The key should be unique and does not need to end with .gob.",
					Default:     "sqldb_{hash}.gob",
				},
				{
					Name:        "storageDir",
					Type:        "string",
					Description: "Storage path where to store data. If default /data/storage does not exists, a temporary directory will be used.",
					Default:     "/data/storage",
				},
			},
		},
	}, SchemaSection{
		// See https://github.com/chronosphereio/calyptia-core-fluent-bit/tree/main/goplugins/cloudflare
		Type:        "input",
		Name:        "cloudflare",
		Description: "HTTP server input for cloudflare with chunked transfer encoding support",
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{
					Name:        "addr",
					Type:        "string",
					Description: "Address to listen on.",
					Default:     ":9880",
				},
				{
					Name:        "resp_headers",
					Type:        "string",
					Description: "Response headers to set, separated by new line. Supports templating.",
					Default:     "Content-Type: application/json",
				},
				{
					Name:        "resp_status_code",
					Type:        "string",
					Description: "Response status code to set. Supports templating. Should evaluate to an integer.",
					Default:     "200",
				},
				{
					Name:        "resp_body",
					Type:        "string",
					Description: "Response body to set. Supports templating.",
					Default:     "{\"status\": \"ok\"}",
				},
				{
					Name:        "time_from",
					Type:        "string",
					Description: "Optional time to set. Supports templating with record access. Should evaluate to a RFC3339 formatted string. Defaults to current time.",
				},
				{
					Name:        "cert_file",
					Type:        "string",
					Description: "Path to the certificate file to enable TLS.",
				},
				{
					Name:        "key_file",
					Type:        "string",
					Description: "Path to the key file to enable TLS.",
				},
			},
		},
	})
}
