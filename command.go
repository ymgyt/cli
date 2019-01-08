package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ymgyt/cli/flags"
)

type Command struct {
	Name    string
	Aliases []string
	Help    func(io.Writer)
	Run     func(context.Context, *Command, []string)
	FlagSet *flags.FlagSet

	parent      *Command
	subCommands []*Command
	onceInit    sync.Once
}

func (c *Command) Execute(ctx context.Context) {
	c.ExecuteWithArgs(ctx, os.Args[1:])
}

func (c *Command) ExecuteWithArgs(ctx context.Context, args []string) {
	remain, err := c.Parse(args)
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

// AddCommand add subcommand. if same name sub command already added, it panic.
func (c *Command) AddCommand(sub *Command) *Command {
	c.lasyInit()
	if sub := c.Lookup(sub.Name); sub != nil {
		panic(fmt.Sprintf("%s already exists", sub.Name))
	}
	if err := c.FlagSet.Merge(sub.FlagSet); err != nil {
		panic(fmt.Sprintf("merge flag set: %s", err))
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

func (c *Command) lasyInit() {
	c.onceInit.Do(func() {
		if c.FlagSet == nil {
			c.FlagSet = &flags.FlagSet{}
		}
	})
}

func (c *Command) Parse(args []string) (remain []string, err error) {
	return c.parse(args)
}

type ParseError struct {
	Message string
}

func (e *ParseError) Error() string { return e.Message }

func (c *Command) parse(args []string) (remain []string, err error) {
	var onParsing *token

	for {
		if len(args) == 0 {
			break
		}

		tk := toToken(args[0])
		switch tk.kind {
		case tkFlag:
			// --foo --bar
			if onParsing != nil {
				return nil, &ParseError{fmt.Sprintf("flag %s <-- %s value not provided", onParsing.value, tk.value)}
			}
			onParsing = &tk
		case tkArgument:
			if c.isBoolFlag(onParsing) {
				if err = c.handleFlag(onParsing.flagName, "true", true); err != nil {
					break
				}
				if tk.value == "true" {
					// verbose true
				} else {
					// verbose arg
					remain = append(remain, tk.value)
				}
				onParsing = nil
				break
			}
			if onParsing != nil {
				if onParsing.kind != tkFlag {
					panic("logic error")
				}
				err = c.handleFlag(onParsing.flagName, tk.value, false)
				onParsing = nil
				break
			}
			remain = append(remain, tk.value)
		case tkFlagWithValue:
			err = c.handleFlag(tk.flagName, tk.flagValue, c.isBoolFlag(&tk))
		case tkMultiFlag:
			if onParsing != nil {
				return nil, &ParseError{fmt.Sprintf("flag %s <-- %s value not provided", onParsing.value, tk.value)}
			}
			for _, f := range tk.flagName {
				if err = c.handleFlag(string(f), "true", true); err != nil {
					break
				}
			}
		case tkTermination:
			return append(remain, args[1:]...), nil
		default:
			err = &ParseError{fmt.Sprintf("unsupported flag: %v", tk.value)}
		}
		if err != nil {
			return nil, err
		}
		args = args[1:]
	}

	if onParsing == nil {
		return remain, nil
	}

	// last flag is bool (cmd get -v)
	if c.isBoolFlag(onParsing) {
		return remain, c.handleFlag(onParsing.flagName, "true", true)
	}

	// last flag no arument (cmd get --label)
	return nil, &ParseError{fmt.Sprintf("flag %s <-- value not provided", onParsing.value)}
}

func (c *Command) handleFlag(name, value string, boolFlag bool) error {
	fs, err := c.FlagSet.LookupAll(name)
	if err != nil || len(fs) == 0 {
		return &ParseError{fmt.Sprintf("flag %s undefined", name)}
	}

	for _, f := range fs {
		if boolFlag {
			if f.IsBool() {
				if err := f.Set(value); err != nil {
					return &ParseError{fmt.Sprintf("error while set bool flag value %s -> %s: %s ", name, value, err)}
				}
			} else {
				return &ParseError{fmt.Sprintf("flag %s is not bool flag", name)}
			}
			continue
		}
		if err := f.Set(value); err != nil {
			return &ParseError{fmt.Sprintf("error while set flag value %s -> %s: %s", name, value, err)}
		}
	}
	return nil
}

func (c *Command) isBoolFlag(tk *token) bool {
	if tk == nil {
		return false
	}
	flag, err := c.lookupFlag(tk.flagName)
	if err != nil {
		return false
	}
	// we does not assume bool flag and non-bool flag are set with same name
	return flag.IsBool()
}

func (c *Command) lookupFlag(name string) (*flags.Flag, error) {
	flag, err := c.FlagSet.Lookup(name)
	if err != nil {
		return nil, &ParseError{fmt.Sprintf("flag %s undefined", name)}
	}
	return flag, nil
}

type tokenKind int

const (
	tkFlag tokenKind = iota
	tkFlagWithValue
	tkArgument
	tkTermination
	tkMultiFlag
	tkEmpty
	tkInvalid
)

type token struct {
	kind      tokenKind
	value     string
	flagName  string
	flagValue string
}

func toToken(s string) token {
	if s == "" {
		return token{kind: tkEmpty}
	}
	if strings.HasPrefix(s, "--") {
		if len(s) == 2 {
			return token{kind: tkTermination, value: "--"}
		}

		fName := s[2:]
		// --label=app
		if strings.Contains(s, "=") {
			nameValue := strings.SplitN(fName, "=", 2)
			return token{kind: tkFlagWithValue, value: s, flagName: nameValue[0], flagValue: nameValue[1]}
		}
		return token{kind: tkFlag, value: s, flagName: fName}
	}
	if strings.HasPrefix(s, "-") {
		// "-" arg like indicating stdin
		if len(s) == 1 {
			return token{kind: tkArgument, value: s}
		}
		fName := s[1:]
		// -v
		if len(fName) == 1 {
			return token{kind: tkFlag, value: s, flagName: fName}
		}
		// -n=10
		if fName[1] == '=' {
			nameValue := strings.SplitN(fName, "=", 2)
			return token{kind: tkFlagWithValue, value: s, flagName: nameValue[0], flagValue: nameValue[1]}
		}
		// -sSL=bbb
		if strings.Contains(fName, "=") {
			return token{kind: tkInvalid, value: s}
		}
		// -sSL
		return token{kind: tkMultiFlag, value: s, flagName: fName}
	}

	// arg
	return token{kind: tkArgument, value: s}
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
	optKindStrings
	optKindDuration
)

type optionProvider interface {
	kind() optionKind
}

func (*StringOpt) kind() optionKind   { return optKindString }
func (*IntOpt) kind() optionKind      { return optKindInt }
func (*BoolOpt) kind() optionKind     { return optKindBool }
func (*StringsOpt) kind() optionKind  { return optKindStrings }
func (*DurationOpt) kind() optionKind { return optKindDuration }

type StringOpt struct {
	Var         *string
	Long        string
	Short       string
	Default     string
	Description string
	Aliases     []string
}

type IntOpt struct {
	Var         *int
	Long        string
	Short       string
	Default     int
	Description string
	Aliases     []string
}

type BoolOpt struct {
	Var         *bool
	Long        string
	Short       string
	Default     bool
	Description string
	Aliases     []string
}

type StringsOpt struct {
	Var         *[]string
	Long        string
	Short       string
	Default     []string
	Description string
	Aliases     []string
}

type DurationOpt struct {
	Var         *time.Duration
	Long        string
	Short       string
	Default     time.Duration
	Description string
	Aliases     []string
}

func (c *OptionConfigurator) Add(provider optionProvider) *OptionConfigurator {
	if c.Err != nil {
		return c
	}
	current := c.cmd
	for {
		fs := current.FlagSet
		switch provider.kind() {
		case optKindString:
			opt := provider.(*StringOpt)
			*opt.Var = opt.Default
			v := (*flags.StringVar)(opt.Var)
			c.Err = fs.Add(&flags.Flag{Long: opt.Long, Short: opt.Short, Description: opt.Description, Aliases: opt.Aliases, Var: v})
		case optKindInt:
			opt := provider.(*IntOpt)
			*opt.Var = opt.Default
			v := (*flags.IntVar)(opt.Var)
			c.Err = fs.Add(&flags.Flag{Long: opt.Long, Short: opt.Short, Description: opt.Description, Aliases: opt.Aliases, Var: v})
		case optKindBool:
			opt := provider.(*BoolOpt)
			*opt.Var = opt.Default
			v := (*flags.BoolVar)(opt.Var)
			c.Err = fs.Add(&flags.Flag{Long: opt.Long, Short: opt.Short, Description: opt.Description, Aliases: opt.Aliases, Var: v})
		case optKindStrings:
			opt := provider.(*StringsOpt)
			*opt.Var = opt.Default
			v := (*flags.StringSliceVar)(opt.Var)
			c.Err = fs.Add(&flags.Flag{Long: opt.Long, Short: opt.Short, Description: opt.Description, Aliases: opt.Aliases, Var: v, AllowMultipleTimesSet: true})
		case optKindDuration:
			opt := provider.(*DurationOpt)
			*opt.Var = opt.Default
			v := (*flags.DurationVar)(opt.Var)
			c.Err = fs.Add(&flags.Flag{Long: opt.Long, Short: opt.Short, Description: opt.Description, Aliases: opt.Aliases, Var: v})
		}
		if c.Err != nil {
			return c
		}
		current = current.parent
		if current == nil {
			break
		}
	}

	return c
}
