package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ymgyt/cli/flags"
)

// HlepFunc print help message to given writer
func HelpFunc(w io.Writer, c *Command) {

	var longestSubcmd string
	sortCmds := func(subs []*Command) []string {
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

	var b strings.Builder
	b.WriteString(c.LongDesc + "\n")

	sorted := sortCmds(c.subCommands)
	indent := "  "
	if len(sorted) > 0 {
		indent := "  "
		b.WriteString("\nSubCommands")
		for _, subName := range sorted {
			sub := c.Lookup(subName)
			b.WriteString("\n" + indent + fmt.Sprintf("%*s: %s", len(longestSubcmd), sub.Name, sub.ShortDesc))
		}
		b.WriteString("\n")
	} else {
		var longestFlag string
		var fs []*flags.Flag
		c.flagSet.Traverse(func(f *flags.Flag) {
			if len(f.Long) > len(longestFlag) {
				longestFlag = f.Long
			}
			fs = append(fs, f)
		})
		sort.Slice(fs, func(i, j int) bool {
			return fs[i].Name() < fs[j].Name()
		})

		for i, f := range fs {
			if i == 0 {
				b.WriteString("\nOptions")
			}
			long := f.Long
			if long == "" {
				long = strings.Repeat(" ", len(longestFlag)+2) // for minus minus
			} else {
				long = fmt.Sprintf("--%-*s", len(longestFlag), f.Long)
			}
			short := f.Short
			if short == "" {
				short = "   " // space for minus char ,
			} else {
				delimiter := ","
				if f.Long == "" {
					delimiter = " "
				}
				short = "-" + short + delimiter
			}
			b.WriteString("\n" + indent + fmt.Sprintf("%s %s: %s", short, long, f.Description))
		}
		b.WriteString("\n")
	}

	fmt.Fprint(w, b.String())
}
