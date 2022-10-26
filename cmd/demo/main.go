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
	var classicConf classic.Classic
	err := classicConf.UnmarshalText([]byte(`
		@SET foo=bar
		[SERVICE]
			Log_Level error
		[CUSTOM]
			Name calyptia
			foo  bar
		[INPUT]
			Name  dummy
			dummy {"foo": "bar"}
		[INPUT]
			Name random
		[FILTER]
			Name   record_modifier
			Match  *
			Record hostname ${HOSTNAME}
			Record product Awesome_Tool
		[OUTPUT]
			Name stdout
			Match *
	`))
	if err != nil {
		return err
	}

	conf := classicConf.ToConfig()

	b, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	fmt.Printf("as yaml:\n%s\n", string(b))

	classicConf = classic.FromConfig(conf)
	b, err = classicConf.MarshalText()
	if err != nil {
		return err
	}

	fmt.Printf("back to classic:\n%s\n", string(b))
	return nil
}
