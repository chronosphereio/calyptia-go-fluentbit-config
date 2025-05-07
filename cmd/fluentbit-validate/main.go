package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	fluent "github.com/calyptia/go-fluentbit-config/v2"
	"github.com/spf13/pflag"
)

func mustGetFormatFromExt(ext string) fluent.Format {
	switch strings.ToLower(ext[1:]) {
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

func usage(progname string) {
	fmt.Printf("%s <options> [input]\n", progname)
	pflag.CommandLine.PrintDefaults()
}

func main() {
	var schema fluent.Schema = fluent.DefaultSchema
	var schemaVersion string
	var checkSchema bool

	pflag.BoolVarP(&checkSchema, "schema", "s", false,
		"validate the schema of the properties")
	pflag.StringVarP(&schemaVersion, "schema-version", "v", "",
		"schema version to valdiate against")
	pflag.Parse()
	args := pflag.Args()

	if pflag.NArg() < 1 {
		usage(os.Args[0])
		os.Exit(0)
	}

	config, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(4)
	}

	cfg, err := fluent.ParseAs(string(config),
		mustGetFormatFromExt(filepath.Ext(args[0])))
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(3)
	}

	if schemaVersion != "" {
		schema, err = fluent.GetSchema(schemaVersion)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(2)
		}
	}

	if checkSchema {
		if err := cfg.ValidateWithSchema(schema); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("VALID SCHEMA")
	} else {
		if err := cfg.Validate(); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("VALID STRUCTURE")
	}
}
