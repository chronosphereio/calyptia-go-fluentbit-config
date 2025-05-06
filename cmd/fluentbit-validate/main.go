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
	pflag.Parse()
	args := pflag.Args()

	if pflag.NArg() < 1 {
		usage(os.Args[0])
		os.Exit(0)
	}

	config, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(3)
	}

	cfg, err := fluent.ParseAs(string(config),
		mustGetFormatFromExt(filepath.Ext(args[0])))
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(2)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("VALID")
}
