# go-fluentbit-config

This library parses fluent-bit configuration in three different formats:

  * INI: the standard INI like format (legacy).
  * YAML: the new YAML format
  * JSON: a JSON format derived from the YAML formats structure.

It consists of 2 functions for each format as well as the NewFromBytes
function which is supported for applications which already use it which
invokes the ParseINI function.

  * INI
    * ParseINI([]byte) (*Config, error)
    * DumpINI(*Config) ([]byte, error)
  * YAML
    * ParseYAML([]byte) (*Config, error)
    * DumpYAML(*Config) ([]byte, error)
  * JSON
    * ParseJSON([]byte) (*Config, error)
    * DumpJSON(*Config) ([]byte, error)

The Parse functions return a pointer to a Config structure which holds
a series of sections.

```go
type Config struct {
	Sections    []ConfigSection
}
```

These sections describe a single section of the configuration file:

```go
const (
	NoneSection ConfigSectionType = iota
	ServiceSection
	InputSection
	FilterSection
	OutputSection
	CustomSection
	ParserSection
)

type ConfigSection struct {
	Type   ConfigSectionType
	ID     string
	Fields []Field
}
```

Each Field has a Key and a set of Values. Most fields contain a single,
but some exceptions, like the `Rewrite_Tag` property `Rule`, have
multiple Values.

```go
type Field struct {
	Key    string
	Values []*Value
}

type Value struct {
	String     *string
	DateTime   *string
	Date       *string
	Time       *string
	TimeFormat *string
	Topic      *string
	Regex      *string
	Bool       *bool
	Number     *float64
	Float      *float64
	List       []*Value
}
```

The Value type has a two receiver methods that which make using values
easier.

```go
func (a *Value) Equals(b *Value) bool
func (v *Value) Value() interface{}
```

The `Equals` receiver can be used to compare two `Value`s, and the `Value`
receiver can be used to return an interface{} which can be type switched.

To more easily manage multiple values the Values type, `type Values []Value`, has a
receiver method that turns all the values into a single string.

```
func (vs Values) ToString() string
```

## example usage
```go
import (
	fbitconf "github.com/calyptia/go-fluentbit-config"
)

cfg, err := fbitconf.ParseINI([]byte(`[INPUT] ....`))
if err != nil {
	panic(err)
}

for _, section := range cfg.Sections {
	if section.Type == fbitconf.FilterSection {
		for _, field := section.Fields {
			if field.Key == "Name" {
				fmt.Printf("filter: %s\n", field.Values.ToString())
			}
		}
	}
}
```

References:

  * Fluentbit schema: https://docs.fluentbit.io/manual/administration/configuring-fluent-bit/format-schema
