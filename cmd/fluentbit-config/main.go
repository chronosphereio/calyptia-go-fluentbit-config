package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	fluent "github.com/calyptia/go-fluentbit-config/v2"
	"github.com/spf13/pflag"
)

func main() {
	var outputFormat string
	var outputFilename string
	var outFile *os.File
	var schema fluent.Schema = fluent.DefaultSchema
	var schemaVersion string
	var checkSchema bool
	var dryRun bool

	pflag.StringVarP(&outputFormat, "format", "f", "yaml",
		"output format, one of: json, yaml (yml), ini or conf")
	pflag.StringVarP(&outputFilename, "output", "o", "",
		"output file (optional, default is stdout)")
	pflag.BoolVarP(&checkSchema, "schema", "s", false,
		"validate the schema of the properties")
	pflag.StringVarP(&schemaVersion, "schema-version", "v", "",
		"schema version to valdiate against")
	pflag.BoolVarP(&dryRun, "dry-run", "n", false,
		"dry-run without writing to stdout (used for validation)")
	pflag.Parse()
	args := pflag.Args()

	if pflag.NArg() < 1 {
		usage(os.Args[0])
		os.Exit(1)
	}

	ext := filepath.Ext(args[0])

	config, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	inFormat, err := getFormatFromExt(ext[1:])
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	cfg, err := fluent.ParseAs(string(config), inFormat)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	if checkSchema {
		if schemaVersion != "" {
			schema, err = fluent.GetSchema(schemaVersion)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				os.Exit(2)
			}
		}
		if err := cfg.ValidateWithSchema(schema); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
	} else {
		if err := cfg.Validate(); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
	}

	if dryRun {
		if checkSchema {
			if schemaVersion == "" {
				fmt.Printf("file %s has valid fluent-bit config syntax\n", args[0])
			} else {
				fmt.Printf("file %s has valid fluent-bit config syntax for version %s\n",
					args[0], schemaVersion)
			}
		}
		return
	}

	outFormat, err := getFormatFromExt(outputFormat)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	out, err := cfg.DumpAs(outFormat)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	if outputFilename != "" {
		outFile, err = os.Create(outputFilename)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
		defer outFile.Close()
	} else {
		outFile = os.Stdout
	}

	switch outFormat {
	case fluent.FormatJSON:
		indented, err := indentJSON(out)
		if err != nil {
			fmt.Printf("ERROR: parse error: %s\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(outFile, indented)
	default:
		fmt.Fprintln(outFile, out)
	}
}

func usage(progname string) {
	fmt.Printf("%s <options> [input]\n", progname)
	pflag.CommandLine.PrintDefaults()
}

func getFormatFromExt(ext string) (fluent.Format, error) {
	switch strings.ToLower(ext) {
	case "json":
		return fluent.FormatJSON, nil
	case "ini", "conf":
		return fluent.FormatClassic, nil
	case "yml", "yaml":
		return fluent.FormatYAML, nil
	}
	return "", fmt.Errorf("unknown format extension: %s", ext)
}

func indentJSON(in string) (string, error) {
	var obj any
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
