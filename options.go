package cli

import (
	"time"

	"github.com/ymgyt/cli/flags"
)

func (c *Command) Options() *OptionConfigurator {
	if c.flagSet == nil {
		c.flagSet = &flags.FlagSet{}
	}
	return &OptionConfigurator{cmd: c}
}

type OptionConfigurator struct {
	Err error
	cmd *Command
}

type FlagProvider interface {
	Flag() *flags.Flag
}

type StringOpt struct {
	Var         *string
	Long        string
	Short       string
	Default     string
	Description string
	Aliases     []string
}

func (o *StringOpt) Flag() *flags.Flag {
	*o.Var = o.Default
	v := (*flags.StringVar)(o.Var)
	return &flags.Flag{Long: o.Long, Short: o.Short, Description: o.Description, Aliases: o.Aliases, Var: v}
}

type IntOpt struct {
	Var         *int
	Long        string
	Short       string
	Default     int
	Description string
	Aliases     []string
}

func (o *IntOpt) Flag() *flags.Flag {
	*o.Var = o.Default
	v := (*flags.IntVar)(o.Var)
	return &flags.Flag{Long: o.Long, Short: o.Short, Description: o.Description, Aliases: o.Aliases, Var: v}
}

type FloatOpt struct {
	Var         *float64
	Long        string
	Short       string
	Default     float64
	Description string
	Aliases     []string
}

func (o *FloatOpt) Flag() *flags.Flag {
	*o.Var = o.Default
	v := (*flags.FloatVar)(o.Var)
	return &flags.Flag{Long: o.Long, Short: o.Short, Description: o.Description, Aliases: o.Aliases, Var: v}
}

type BoolOpt struct {
	Var         *bool
	Long        string
	Short       string
	Default     bool
	Description string
	Aliases     []string
}

func (o *BoolOpt) Flag() *flags.Flag {
	*o.Var = o.Default
	v := (*flags.BoolVar)(o.Var)
	return &flags.Flag{Long: o.Long, Short: o.Short, Description: o.Description, Aliases: o.Aliases, Var: v}
}

type StringsOpt struct {
	Var         *[]string
	Long        string
	Short       string
	Default     []string
	Description string
	Aliases     []string
	Delimiter   string
}

func (o *StringsOpt) Flag() *flags.Flag {
	*o.Var = o.Default
	v := (*flags.StringsVar)(o.Var)
	return &flags.Flag{Long: o.Long, Short: o.Short, Description: o.Description, Aliases: o.Aliases, Var: v, AllowMultipleTimesSet: true, Delimiter: o.Delimiter}
}

type IntsOpt struct {
	Var         *[]int
	Long        string
	Short       string
	Default     []int
	Description string
	Aliases     []string
	Delimiter   string
}

func (o *IntsOpt) Flag() *flags.Flag {
	*o.Var = o.Default
	v := (*flags.IntsVar)(o.Var)
	return &flags.Flag{Long: o.Long, Short: o.Short, Description: o.Description, Aliases: o.Aliases, Var: v, AllowMultipleTimesSet: true, Delimiter: o.Delimiter}
}

type DurationOpt struct {
	Var         *time.Duration
	Long        string
	Short       string
	Default     time.Duration
	Description string
	Aliases     []string
}

func (o *DurationOpt) Flag() *flags.Flag {
	*o.Var = o.Default
	v := (*flags.DurationVar)(o.Var)
	return &flags.Flag{Long: o.Long, Short: o.Short, Description: o.Description, Aliases: o.Aliases, Var: v}
}

func (c *OptionConfigurator) Add(provider FlagProvider) *OptionConfigurator {
	fs := c.cmd.flagSet
	c.Err = fs.Add(provider.Flag())
	return c
}
