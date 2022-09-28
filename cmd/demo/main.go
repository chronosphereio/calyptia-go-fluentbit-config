package main

import (
	"fmt"
	"os"

	"github.com/calyptia/go-fluentbit-conf/classic"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	conf, err := classic.Parse(`
		[INPUT]
			Name dummy
			rate 10.4
	`)
	if err != nil {
		return err
	}

	// fmt.Printf("%#v\n", conf)
	// litter.Dump(conf)
	fmt.Println(conf.String())
	return nil
}
