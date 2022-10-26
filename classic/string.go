package classic

import (
	"fmt"
	"strings"
	"text/tabwriter"

	fluentbitconf "github.com/calyptia/go-fluentbit-conf"
)

func String(conf fluentbitconf.Config) string {
	var sb strings.Builder

	writeByNames := func(byNames []fluentbitconf.ByName, section string) {
		for _, byName := range byNames {
			for _, in := range byName {
				_, _ = fmt.Fprintf(&sb, "[%s]\n", section)
				tw := tabwriter.NewWriter(&sb, 0, 4, 1, ' ', 0)
				for k, v := range in {
					// Case when the property could be repeated. Example:
					// 	[FILTER]
					// 		Name record_modifier
					// 		Match *
					// 		Record hostname ${HOSTNAME}
					// 		Record product Awesome_Tool
					if s, ok := v.([]any); ok {
						for _, v := range s {
							_, _ = fmt.Fprintf(tw, "    %s\t%s\n", k, stringFromAny(v))
						}
					} else {
						_, _ = fmt.Fprintf(tw, "    %s\t%s\n", k, stringFromAny(v))
					}
				}
				_ = tw.Flush()
			}
		}
	}

	for k, v := range conf.Env {
		_, _ = fmt.Fprintf(&sb, "@SET %s=%s\n", k, stringFromAny(v))
	}

	if len(conf.Service) != 0 {
		sb.WriteString("[SERVICE]\n")
		tw := tabwriter.NewWriter(&sb, 0, 4, 1, ' ', 0)
		for k, v := range conf.Service {
			_, _ = fmt.Fprintf(tw, "    %s\t%s\n", k, stringFromAny(v))
		}
		_ = tw.Flush()
	}

	writeByNames(conf.Customs, "CUSTOM")
	writeByNames(conf.Pipeline.Inputs, "INPUT")
	writeByNames(conf.Pipeline.Filters, "FILTER")
	writeByNames(conf.Pipeline.Outputs, "OUTPUT")

	return sb.String()
}
