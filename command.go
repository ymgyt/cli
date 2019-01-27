package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/ymgyt/cli/flags"
	"github.com/ymgyt/cli/parser"
)

type Command struct {
	Name      string
	Aliases   []string
	ShortDesc string
	LongDesc  string
	Help      func(io.Writer, *Command)
	Run       func(context.Context, *Command, []string)

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	flagSet     *flags.FlagSet
	parent      *Command
	subCommands []*Command
	onceInit    sync.Once
}

func (c *Command) Execute(ctx context.Context) {
	c.ExecuteWithArgs(ctx, os.Args[1:])
}

func (c *Command) ExecuteWithArgs(ctx context.Context, args []string) {
	c.lasyInit()
	pr, err := c.Parse(args)
	if err != nil {
		c.handleParseErr(err)
		return
	}

	var runCmd = c
	for _, sub := range pr.Commands() {
		if runCmd == nil {
			panic("runCmd == nil, something went wrong")
		}
		runCmd = runCmd.Lookup(sub)
	}
	if err := runCmd.ConsumeFlags(pr.AllFlags()); err != nil {
		c.handleParseErr(err)
		return
	}
	runCmd.Run(ctx, runCmd, pr.Args())
}

// AddCommand add subcommand. if same name sub command already added, it panic.
func (c *Command) AddCommand(sub *Command) *Command {
	c.lasyInit()
	if sub := c.Lookup(sub.Name); sub != nil {
		panic(fmt.Sprintf("%s already exists", sub.Name))
	}
	sub.parent = c
	c.subCommands = append(c.subCommands, sub)
	return c
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

func (c *Command) ConsumeFlags(pfs []*parser.Flag) error {
	flagSet := c.flagSet
	for _, pf := range pfs {
		f, err := flagSet.Lookup(pf.Name)
		if err != nil {
			return &ParseError{FlagName: pf.Name, Message: fmt.Sprintf("flag %s not found", pf.Name)}
		}
		value := pf.Value
		if pf.IsBool {
			value = "false"
			if pf.BoolValue {
				value = "true"
			}
		}
		if err := f.Set(value); err != nil {
			return &ParseError{FlagName: pf.Name, Message: err.Error()}
		}
	}
	return nil
}

func (c *Command) lasyInit() {
	c.onceInit.Do(func() {
		if c.flagSet == nil {
			c.flagSet = &flags.FlagSet{}
		}
		if c.Stdin == nil {
			c.Stdin = os.Stdin
		}
		if c.Stdout == nil {
			c.Stdout = os.Stdout
		}
		if c.Stderr == nil {
			c.Stderr = os.Stderr
		}
		if c.Help == nil {
			c.Help = c.DefaultHelp()
		}
		if c.Run == nil {
			c.Run = func(_ context.Context, _ *Command, _ []string) {
				c.Help(c.Stderr, c)
			}
		}
	})
}

func (c *Command) DefaultHelp() func(io.Writer, *Command) {
	return HelpFunc()
}

func (c *Command) Parse(args []string) (*parser.Result, error) {
	return parser.New(&commander{c}).Parse(args)
}

type ParseError struct {
	Message  string
	FlagName string
}

func (e *ParseError) Error() string { return e.Message }

func (c *Command) handleParseErr(err error) {
	fmt.Fprintf(c.Stderr, "parse error: %s\n", err)
}

type commander struct {
	c *Command
}

func (c *commander) Name() string { return c.c.Name }
func (c *commander) LookupSubCommand(name string) (parser.Commander, bool) {
	sub := c.c.Lookup(name)
	if sub == nil {
		return nil, false
	}
	return &commander{c: sub}, true
}
func (c *commander) IsBoolFlag(name string) bool {
	// まず自分のflagsetをみにいく
	f, err := c.c.flagSet.Lookup(name)
	if err == nil {
		return f.IsBool()
	}
	for _, sub := range c.c.subCommands {
		if (&commander{c: sub}).IsBoolFlag(name) {
			return true
		}
	}
	return false
}
