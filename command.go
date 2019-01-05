package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ymgyt/cli/flags"
)

type Command struct {
	Name    string
	Aliases []string
	Help    func(io.Writer)
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Run     func(context.Context, *Command, []string)
	FlagSet *flags.FlagSet

	parent      *Command
	subCommands []*Command
}

func (c *Command) Execute(ctx context.Context) {
	c.ExecuteWithArgs(ctx, os.Args)
}

func (c *Command) ExecuteWithArgs(ctx context.Context, args []string) {
	remain, err := c.FlagSet.ParseAll(args)
	if err != nil {
		panic(err)
	}
	var sub *Command
	if len(remain) > 0 {
		sub = c.Lookup(remain[0])
	}

	if sub != nil {
		sub.ExecuteWithArgs(ctx, remain[1:])
	} else {
		c.Run(ctx, c, remain)
	}
}

func (c *Command) AddCommand(sub *Command) {
	if sub := c.Lookup(sub.Name); sub != nil {
		panic(fmt.Sprintf("%s already exists", sub.Name))
	}
	sub.parent = c
	c.subCommands = append(c.subCommands, sub)
}

func (c *Command) Lookup(name string) *Command {
	for _, sub := range c.subCommands {
		if sub.Name == name {
			return sub
		}
		for _, alias := range sub.Aliases {
			if alias == name {
				return sub
			}
		}
	}
	return nil
}

func (c *Command) Options() *OptionConfigurator {
	if c.FlagSet == nil {
		c.FlagSet = &flags.FlagSet{}
	}
	return &OptionConfigurator{cmd: c}
}

type OptionConfigurator struct {
	Err error
	cmd *Command
}

type optionKind int

const (
	optKindString optionKind = iota
	optKindInt
	optKindBool
)

type optionProvider interface {
	kind() optionKind
}

type StringOpt struct {
	Var         *string
	Long        string
	Short       string
	Default     string
	Description string
	Aliases     []string
}

func (*StringOpt) kind() optionKind { return optKindString }

type IntOpt struct {
	Var         *int
	Long        string
	Short       string
	Default     int
	Description string
	Aliases     []string
}

func (*IntOpt) kind() optionKind { return optKindInt }

type BoolOpt struct {
	Var         *bool
	Long        string
	Short       string
	Default     bool
	Description string
	Aliases     []string
}

func (*BoolOpt) kind() optionKind { return optKindBool }

func (c *OptionConfigurator) Add(provider optionProvider) *OptionConfigurator {
	if c.Err != nil {
		return c
	}
	fs := c.cmd.FlagSet
	switch provider.kind() {
	case optKindString:
		opt := provider.(*StringOpt)
		v := (*flags.StringVar)(opt.Var)
		c.Err = fs.Add(&flags.Flag{Long: opt.Long, Short: opt.Short, Description: opt.Description, Aliases: opt.Aliases, Var: v})
	case optKindInt:
		opt := provider.(*IntOpt)
		v := (*flags.IntVar)(opt.Var)
		c.Err = fs.Add(&flags.Flag{Long: opt.Long, Short: opt.Short, Description: opt.Description, Aliases: opt.Aliases, Var: v})
	case optKindBool:
		opt := provider.(*BoolOpt)
		v := (*flags.BoolVar)(opt.Var)
		c.Err = fs.Add(&flags.Flag{Long: opt.Long, Short: opt.Short, Description: opt.Description, Aliases: opt.Aliases, Var: v})
	}

	return c
}
