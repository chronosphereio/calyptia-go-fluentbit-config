package classic

import (
	"fmt"
	"strings"

	fluentbitconfig "github.com/calyptia/go-fluentbit-config"
	"github.com/calyptia/go-fluentbit-config/property"
)

func (c Classic) ToConfig() fluentbitconfig.Config {
	var out fluentbitconfig.Config

	addByName := func(props property.Properties, dest *[]fluentbitconfig.ByName) {
		if nameVal, ok := props.Get("Name"); ok {
			if name, ok := nameVal.(string); ok {
				config := property.Properties{}
				for _, p := range props {
					// Case when the property could be repeated. Example:
					// 	[FILTER]
					// 		Name record_modifier
					// 		Match *
					// 		Record hostname ${HOSTNAME}
					// 		Record product Awesome_Tool
					if v, ok := config.Get(p.Key); ok {
						if s, ok := v.([]any); ok {
							s = append(s, p.Value)
							config.Set(p.Key, s)
						} else {
							config.Set(p.Key, []any{v, p.Value})
						}
					} else {
						config.Set(p.Key, p.Value)
					}
				}
				*dest = append(*dest, fluentbitconfig.ByName{
					name: config,
				})
			}
		}
	}

	for _, entry := range c.Entries {
		switch entry.Kind {
		case EntryKindCommand:
			if strings.EqualFold(entry.AsCommand.Name, "SET") {
				if out.Env == nil {
					out.Env = property.Properties{}
				}
				out.Env.Set(splitCommand(entry.AsCommand.Instruction))
			}
		case EntryKindSection:
			switch {
			case strings.EqualFold(entry.AsSection.Name, "SERVICE"):
				if out.Service == nil {
					out.Service = property.Properties{}
				}
				for _, p := range entry.AsSection.Properties {
					out.Service.Set(p.Key, p.Value)
				}
			case strings.EqualFold(entry.AsSection.Name, "CUSTOM"):
				addByName(entry.AsSection.Properties, &out.Customs)
			case strings.EqualFold(entry.AsSection.Name, "INPUT"):
				addByName(entry.AsSection.Properties, &out.Pipeline.Inputs)
			case strings.EqualFold(entry.AsSection.Name, "FILTER"):
				addByName(entry.AsSection.Properties, &out.Pipeline.Filters)
			case strings.EqualFold(entry.AsSection.Name, "OUTPUT"):
				addByName(entry.AsSection.Properties, &out.Pipeline.Outputs)
			}
		}
	}

	return out
}

func FromConfig(conf fluentbitconfig.Config) Classic {
	var out Classic

	addCommands := func(name string, props property.Properties) {
		for _, p := range props {
			out.Entries = append(out.Entries, Entry{
				Kind: EntryKindCommand,
				AsCommand: &Command{
					Name:        name,
					Instruction: fmt.Sprintf("%s=%s", p.Key, stringFromAny(p.Value)),
				},
			})
		}
	}

	addSection := func(name string, properties property.Properties) {
		if len(properties) == 0 {
			return
		}

		out.Entries = append(out.Entries, Entry{
			Kind: EntryKindSection,
			AsSection: &Section{
				Name:       name,
				Properties: properties,
			},
		})
	}

	addSections := func(name string, byNames []fluentbitconfig.ByName) {
		for _, byName := range byNames {
			for _, props := range byName {
				addSection(name, props)
			}
		}
	}

	addCommands("SET", conf.Env)
	addSection("SERVICE", conf.Service)
	addSections("CUSTOM", conf.Customs)
	addSections("INPUT", conf.Pipeline.Inputs)
	addSections("FILTER", conf.Pipeline.Filters)
	addSections("OUTPUT", conf.Pipeline.Outputs)

	return out
}
