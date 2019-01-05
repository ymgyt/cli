package flags

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	ErrFlagNotFound      = errors.New("flag not found")
	ErrFlagAlreadyExists = errors.New("flag already exists")
	ErrFlagNameRequired  = errors.New("flag name required")
)

type ParseError struct {
	FlagName string
	Msg      string
}

func (pe *ParseError) Error() string {
	return fmt.Sprintf("flag %s %s", pe.FlagName, pe.Msg)
}

type FlagSet struct {
	Flags []*Flag
	*sync.RWMutex
	once sync.Once
}

// Add add given flag.
// if flag Name(Long, Short, Aliases) conflict already exists flags, returns ErrFlagAlreadyExists.
func (fs *FlagSet) Add(f *Flag) error {
	fs.init()
	name := f.Long
	if name == "" {
		name = f.Short
	}
	if name == "" {
		return ErrFlagNameRequired
	}
	found, err := fs.Lookup(name)
	canAdd := err == ErrFlagNotFound
	if !canAdd {
		if found != nil {
			return ErrFlagAlreadyExists
		}
		return err
	}
	fs.Lock()
	fs.Flags = append(fs.Flags, f)
	fs.Unlock()
	return nil
}

// Lookup lookup flag by given flag name. if not found, returns ErrFlagNotFound.
// Long, Short, Aliases are checked.
func (fs *FlagSet) Lookup(name string) (*Flag, error) {
	fs.init()
	if name == "" {
		return nil, ErrFlagNotFound
	}
	fs.RLock()
	defer fs.RUnlock()
	for _, f := range fs.Flags {
		if f.Long == name || f.Short == name {
			return f, nil
		}
		for _, alias := range f.Aliases {
			if alias == name {
				return f, nil
			}
		}
	}
	return nil, ErrFlagNotFound
}

func (fs *FlagSet) init() {
	fs.once.Do(func() {
		fs.RWMutex = &sync.RWMutex{}
	})
}

type ParseState int

const (
	ParseStart ParseState = iota
	ParseCommand
	ParseFlag
	ParseFlagArg
	ParseBoolFlag
	ParseEnd
)

type ParseContext struct {
	Remain   []string
	State    ParseState
	FlagName string
	Value    string
	IsShort  bool
}

func (pc *ParseContext) clone() *ParseContext {
	c := *pc
	return &c
}

func (pc *ParseContext) state(s ParseState) *ParseContext {
	pc.State = s
	return pc
}

func (pc *ParseContext) remain(args []string) *ParseContext {
	pc.Remain = args
	return pc
}

func (pc *ParseContext) name(name string) *ParseContext {
	pc.FlagName = name
	return pc
}

func (pc *ParseContext) value(v string) *ParseContext {
	pc.Value = v
	return pc
}

func (pc *ParseContext) isShort(b bool) *ParseContext {
	pc.IsShort = b
	return pc
}

func (fs *FlagSet) ParseAll(args []string) (remain []string, err error) {
	ctx := &ParseContext{
		Remain: args,
		State:  ParseStart,
	}

	for {
		ctx, err = fs.Parse(ctx)
		if err != nil {
			return nil, err
		}
		switch ctx.State {
		case ParseCommand:
			return append([]string{ctx.Value}, ctx.Remain...), nil
		case ParseFlag:
			// do nothing
		case ParseFlagArg:
			flag, err := fs.Lookup(ctx.FlagName)
			if err != nil {
				return nil, err
			}
			if err := flag.Var.Set(ctx.Value); err != nil {
				return nil, err
			}
			ctx = ctx.name("").value("")
		case ParseBoolFlag:
			flag, err := fs.Lookup(ctx.FlagName)
			if err != nil {
				return nil, err
			}
			if err := flag.Var.(BooleanVar).SetBool(true); err != nil {
				return nil, err
			}
			ctx = ctx.name("")
		case ParseEnd:
			return nil, nil
		}

	}
}

func (fs *FlagSet) Parse(ctx *ParseContext) (*ParseContext, error) {
	ctx = ctx.clone()
	remain := ctx.Remain
	if len(remain) == 0 {
		return ctx.state(ParseEnd), nil
	}
	remain, token := fs.read(ctx.Remain)
	ctx = ctx.remain(remain)

	switch token.kind {
	case tkLongOption, tkShortOption:
		// root --label --enable <-
		if ctx.State == ParseFlag {
			var prefix = "--"
			if ctx.IsShort {
				prefix = "-"
			}
			return nil, &ParseError{FlagName: prefix + ctx.FlagName, Msg: "value required"}
		}
		ctx = ctx.state(ParseFlag).name(token.value)
		flag, err := fs.Lookup(ctx.FlagName)
		if err != nil {
			return nil, err
		}
		if isBoolFlag(flag) {
			ctx = ctx.state(ParseBoolFlag)
		}
		ctx = ctx.isShort(token.kind == tkShortOption)
	case tkArgument:
		// root --label app <-
		if ctx.State == ParseFlag {
			ctx = ctx.state(ParseFlagArg).value(token.value)
		} else {
			// root cmd <-
			ctx = ctx.state(ParseCommand).value(token.value)
		}
	}

	return ctx, nil
}

type tokenKind string

const (
	tkLongOption  tokenKind = "long"
	tkShortOption tokenKind = "short"
	tkArgument    tokenKind = "argument"
	tkOptionEnd   tokenKind = "optend"
	tkInvalid     tokenKind = "invalid"
)

type token struct {
	kind  tokenKind
	value string
}

func (fs *FlagSet) read(args []string) (remain []string, t token) {
	if len(args) == 0 {
		return nil, token{kind: tkInvalid}
	}
	v := args[0]
	// --label=app
	if strings.Contains(v, "=") {
		kv := strings.SplitN(v, "=", 2)
		v = kv[0]
		flagArg := kv[1]
		if len(flagArg) == 0 {
			return nil, token{kind: tkInvalid}
		}
		// --label=app --second -> app --second
		remain = append([]string{flagArg}, args[1:]...)
	} else {
		remain = args[1:]
	}

	return remain, toToken(v)
}

func toToken(s string) token {
	if s == "" {
		return token{kind: tkInvalid}
	}
	if strings.HasPrefix(s, "--") {
		// --
		if len(s) == 2 {
			return token{kind: tkOptionEnd, value: s}
		}
		return token{kind: tkLongOption, value: s[2:]}
	}
	if strings.HasPrefix(s, "-") {
		// currently we do not support "-" option(indacting stdinput option)
		if len(s) == 1 {
			return token{kind: tkInvalid}
		}
		return token{kind: tkShortOption, value: s[1:]}
	}
	return token{kind: tkArgument, value: s}
}

func isBoolFlag(flag *Flag) bool {
	_, ok := flag.Var.(BooleanVar)
	return ok
}
