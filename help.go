package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ymgyt/cli/flags"
)

func HelpFunc() func(io.Writer, *Command) {

	var longestSubcmd string
	sort := func(subs []*Command) []string {
		var names = make([]string, 0, len(subs))
		for _, sub := range subs {
			if len(sub.Name) > len(longestSubcmd) {
				longestSubcmd = sub.Name
			}
			names = append(names, sub.Name)
		}
		sort.Strings(names)
		return names
	}

	return func(w io.Writer, c *Command) {
		var b strings.Builder
		b.WriteString(c.LongDesc + "\n")

		sorted := sort(c.subCommands)
		indent := "  "
		if len(sorted) > 0 {
			indent := "  "
			b.WriteString("\nSubCommands")
			for _, subName := range sorted {
				sub := c.Lookup(subName)
				if sub == nil {
					continue
				}
				b.WriteString("\n" + indent + fmt.Sprintf("%*s: %s", len(longestSubcmd), sub.Name, sub.ShortDesc))
			}
			b.WriteString("\n")
		} else {
			var longestFlag string
			var fs []*flags.Flag
			c.FlagSet.Traverse(func(f *flags.Flag) {
				if len(f.Long) > len(longestFlag) {
					longestFlag = f.Long
				}
				fs = append(fs, f)
			})

			for i, f := range fs {
				if i == 0 {
					b.WriteString("\nOptions")
				}
				short := f.Short
				if short == "" {
					short = "   " // space for minus char ,
				} else {
					short = "-" + short + ","
				}
				long := f.Long
				if long == "" {
					long = strings.Repeat("", len(longestFlag)+2) // for minus minus
				} else {
					long = fmt.Sprintf("--%-*s", len(longestFlag), f.Long)
				}

				b.WriteString("\n" + indent + fmt.Sprintf("%s %s: %s", short, long, f.Description))
			}
			b.WriteString("\n")
		}

		fmt.Fprint(w, b.String())
	}
}
