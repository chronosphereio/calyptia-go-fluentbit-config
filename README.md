# go-fluentbit-config

Library for parsing a fluentbit configuration file according to the specs presented here [0]

[0] https://docs.fluentbit.io/manual/administration/configuring-fluent-bit/format-schema

## example usage
```go
    import (
        "github.com/calyptia/go-fluent-bit-config"
    )

    cfg, err := fluentbit_config.NewFromBytes([]byte(`[INPUT] ....`))
    if err != nil {
    }

    for name, filters := range cfg.Filters {
	    for name, property := range filters {
            }
    }   	
```
