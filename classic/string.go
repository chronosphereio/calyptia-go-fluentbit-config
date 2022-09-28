package classic

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

func (c Config) String() string {
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 4, 1, ' ', 0)

	for _, entry := range c.Entries {
		switch entry.Kind {
		case EntryKindCommand:
			_, _ = fmt.Fprintf(&sb, "@%s %s", entry.AsCommand.Name, entry.AsCommand.Instruction)
		case EntryKindSection:
			_, _ = fmt.Fprintf(&sb, "[%s]\n", entry.AsSection.Name)
			for _, prop := range entry.AsSection.Properties {
				_, _ = fmt.Fprintf(tw, "    %s\t%s\n", prop.Name, prop.Value.String())
			}
			_ = tw.Flush()
		}
	}

	return sb.String()
}
