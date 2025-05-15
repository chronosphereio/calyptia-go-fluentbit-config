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

	pflag.StringVarP(&outputFormat, "format", "f", "yaml",
		"output format, one of: json, yaml (yml), ini or conf")
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
