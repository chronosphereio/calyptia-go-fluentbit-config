package fluentbitconfig

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//go:embed schemas/23.8.5.json
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
		// See https://github.com/calyptia/lts-advanced-plugin-s3-replay
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
		// See https://github.com/calyptia/lts-advanced-plugin-dummy
		Type:        "input",
		Name:        "gdummy",
		Description: "dummy GO!",
	}, SchemaSection{
		// See https://github.com/calyptia/lts-advanced-plugin-gsuite-reporter
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
		// See https://github.com/calyptia/core-fluent-bit-plugin-http-loader
		Type:        "input",
		Name:        "http_loader",
		Description: "HTTP Loader plugin provides a way to load/dump data from a paginated HTTP endpoint.",
		Properties: SchemaProperties{
			Options: []SchemaOptions{
				{
					Name:        "method",
					Type:        "string",
					Description: `HTTP request method, default is "GET" or "POST" if body is set.`,
					Default:     "GET",
				},
				{
					Name:        "url",
					Type:        "string",
					Description: "Target URL to scrap. It must be an absolute URL starting with http or https.",
					Default:     "",
				},
				{
					Name:        "body",
					Type:        "string",
					Description: "Optional request body.",
				},
				{
					Name:        "headers_separator",
					Type:        "string",
					Description: "Option to override the headers separator.",
					Default:     "\r\n",
				},
				{
					Name:        "headers",
					Type:        "string",
					Description: `Optional request headers, each one separated by a new line character "\r\n". "User-Agent" will always be set to "Fluent-Bit HTTP Loader Plugin" unless overriden.`,
					Default:     "User-Agent: Fluent-Bit HTTP Loader Plugin",
				},
				{
					Name:        "next_url_template",
					Type:        "string",
					Description: "Optional Golang template you can use to modify the original request URL. Use it to advance and iterate over your paginated API. For an offset API do something like:\nhttps://example.org/pages?offset={{.index}}",
				},
				{
					Name:        "next_body_template",
					Type:        "string",
					Description: "Optional Golang template used to modify the original request body.",
				},
				{
					Name: "next_headers_template",
					Type: "string",
				},
				{
					Name:        "stop_template",
					Type:        "string",
					Description: "Required golang template that evaluates to a boolean value which tells the plugin when to stop.",
				},
				{
					Name:        "pull_interval",
					Type:        "string",
					Description: `Optional interval between each request. It supports Golang templating. Useful if you want to set different values depending on the response. For example: \n{{if .body.nextCursor}}1s{{else}}6h{{end}}\nWould use a short interval if paginating, or a long interval if not.`,
					Default:     "0s",
				},
				{
					Name:        "timeout",
					Type:        "string",
					Description: "Request timeout, default 10s.",
					Default:     "10s",
				},
				{
					Name:        "continue_on_error",
					Type:        "boolean",
					Description: "Whether to continue on >400 response status code.",
					Default:     false,
				},
				{
					Name:        "split_records",
					Type:        "boolean",
					Description: "Controls if we should send multiple records to fluent-bit when we find an array. In which case, each record will contain the array `index` and `value`.",
					Default:     false,
				},
				{
					Name:        "max_response_bytes",
					Type:        "integer",
					Description: "Option to limit the amount of bytes to read from the response body. Default is 15 MB.",
					Default:     15 << 20,
				},
				{
					Name:        "template",
					Type:        "string",
					Description: "Optional Golang template to apply over the record. You can find `statusCode` (int), `headers` (http.Header), `body` which can be any and `index` which a increasing number for the request being made. This is the data that will be sent to fluent-bit for processing.",
				},
				{
					Name:        "oauth2_client_id",
					Type:        "string",
					Description: "Optional OAuth2 Client ID. You need to at least pass the client ID, secret and token URL to enable OAUth2. It uses the client_credentials flow.",
				},
				{
					Name:        "oauth2_client_secret",
					Type:        "string",
					Description: "Optional OAuth2 client secret.",
				},
				{
					Name:        "oauth2_token_url",
					Type:        "string",
					Description: "Optional OAuth2 token URL.",
				},
				{
					Name:        "oauth2_scopes_separator",
					Type:        "string",
					Description: "Optional list of additional scopes for OAuth2 separated by space.",
				},
				{
					Name:        "oauth2_scopes",
					Type:        "string",
					Description: "Option to override the OAuth2 scopes separator.",
					Default:     " ",
				},
				{
					Name:        "oauth2_endpoint_params",
					Type:        "string",
					Description: "Optional additional params to pass during OAuth2. The format is a URL query string.",
				},
				{
					Name:        "data_dir",
					Type:        "string",
					Description: "Optional storage path to allow resuming after a restart.",
					Default:     "/data/storage",
				},
				{
					Name:        "data_exp",
					Type:        "string",
					Description: "Optional duration time to allow stored data to be valid. Set this if you know stuff like pagination tokens have an expiration time.",
					Default:     "0s",
				},
			},
		},
	})
}
