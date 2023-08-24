package fluentbitconfig

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//go:embed schemas/23.8.7.json
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
			},
		},
	})
}
