package main

import (
	"fmt"
	"os"

	"github.com/calyptia/go-fluentbit-conf/classic"
	"gopkg.in/yaml.v3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	conf, err := classic.Parse(`
		[FILTER]
			Name   record_modifier
			Match  *
			Record hostname ${HOSTNAME}
			Record product Awesome_Tool
	`)
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", string(b))
	fmt.Println(classic.String(conf))
	return nil
}
