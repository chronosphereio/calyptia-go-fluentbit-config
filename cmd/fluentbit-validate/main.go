package main

import (
	"fmt"
	"os"
	"strings"
	"path/filepath"

	fluent "github.com/calyptia/go-fluentbit-config/v2"
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

func main() {
	config, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	cfg, err := fluent.ParseAs(string(config),
		mustGetFormatFromExt(filepath.Ext(os.Args[1])))
	if err != nil {
		panic(err)
	}

	if err := cfg.Validate(); err != nil {
		panic(err)
	}
}
