package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	fluent "github.com/calyptia/go-fluentbit-config/v2"
	"github.com/spf13/pflag"
)

func mustGetFormatFromExt(ext string) fluent.Format {
	switch strings.ToLower(ext) {
	case "json":
		return fluent.FormatJSON
	case "ini":
		fallthrough
	case "conf":
		return fluent.FormatClassic
	case "yml":
		fallthrough
	case "yaml":
		return fluent.FormatYAML
	}
	panic(fmt.Errorf("unknown format extension: %s", ext))
}

func indentJSON(in string) (string, error) {
	var obj interface{}
	if err := json.Unmarshal([]byte(in), &obj); err != nil {
		return "", err
	}

	indentedJSON, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		fmt.Println("Failed to indent JSON:", err)
		return "", err
	}

	return string(indentedJSON), nil
}

func main() {
	var outputFormat string

	pflag.StringVar(&outputFormat, "format", "json", "output format, one of: json, yaml, ini")
	pflag.Parse()

	parts := strings.SplitN(os.Args[1], ".", 2)

	config, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	cfg, err := fluent.ParseAs(string(config),
		mustGetFormatFromExt(parts[1]))
	if err != nil {
		panic(err)
	}

	out, err := cfg.DumpAs(mustGetFormatFromExt(outputFormat))
	if err != nil {
		panic(fmt.Errorf("parse error: %s", err))
	}

	switch mustGetFormatFromExt(outputFormat) {
	case fluent.FormatJSON:
		indented, err := indentJSON(out)
		if err != nil {
			panic(fmt.Errorf("parse error: %s", err))
		}
		fmt.Println(indented)
	default:
		fmt.Println(out)
	}
}
