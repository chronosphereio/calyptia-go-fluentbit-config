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
	var outfile *os.File

	pflag.StringVarP(&outputFormat, "format", "f", "yaml",
		"output format, one of: json, yaml, ini")
	pflag.StringVarP(&outputFilename, "output", "o", "",
		"output file (optional, default is stdout)")
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

	cfg, err := fluent.ParseAs(string(config),
		mustGetFormatFromExt(ext[1:]))
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	out, err := cfg.DumpAs(mustGetFormatFromExt(outputFormat))
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	if outputFilename != "" {
		outfile, err = os.Create(outputFilename)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
		defer outfile.Close()
	} else {
		outfile = os.Stdout
	}

	switch mustGetFormatFromExt(outputFormat) {
	case fluent.FormatJSON:
		indented, err := indentJSON(out)
		if err != nil {
			panic(fmt.Errorf("parse error: %s", err))
		}
		fmt.Fprintln(outfile, indented)
	default:
		fmt.Fprintln(outfile, out)
	}
}

func usage(progname string) {
	fmt.Printf("%s <options> [input]\n", progname)
	pflag.CommandLine.PrintDefaults()
}

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
