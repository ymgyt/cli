package parser

import (
	"fmt"
	"strconv"
	"strings"
)

type Result struct {
	commands []string
	args     []string
	flagMap  map[string][]*Flag

	rootCmdName string
}

func (r *Result) Commands() []string { return r.commands }
func (r *Result) Args() []string     { return r.args }
func (r *Result) Flags(cmd string) []*Flag {
	return r.flagMap[cmd]
}
func (r *Result) AllFlags() []*Flag {
	var fs []*Flag
	// root commandはargumentには現れないので自分で補ってやる.
	// real arg: root -v sub xxx
	// os.Exit(cmd.Execute(os.Args[1:]))
	for _, cmd := range append([]string{r.rootCmdName}, r.commands...) {
		fs = append(fs, r.Flags(cmd)...)
	}
	return fs
}

type Flag struct {
	Name      string
	Value     string
	IsBool    bool
	BoolValue bool
}

type Commander interface {
	Name() string
	LookupSubCommand(name string) (Commander, bool)
	IsBoolFlag(name string) bool
}

func New(root Commander) *Parser {
	return &Parser{Root: root}
}

type Parser struct {
	Root Commander
}

func (p *Parser) Parse(args []string) (*Result, error) {
	lexer := newLexer(args)

	ctx := newContext(p.Root.Name())
	parse(lexer, p.Root, ctx)
	if ctx.err != nil {
		return nil, ctx.err
	}
	return ctx.Result, nil
}

func parse(lexer *lexer, cmd Commander, ctx *context) {
	if ctx.err != nil {
		return
	}
	tk := lexer.read()
	switch tk.kind {
	case tkEnd:
		return
	case tkArgument:
		parseArgument(tk, lexer, &cmd, ctx)
	case tkFlag:
		if cmd.IsBoolFlag(tk.flagName) {
			parseBoolFlag(tk, lexer, cmd, ctx)
		} else {
			parseFlag(tk, lexer, cmd, ctx)
		}
	case tkFlagWithValue:
		if cmd.IsBoolFlag(tk.flagName) {
			parseBoolFlagWithValue(tk, lexer, cmd, ctx)
		} else {
			parseFlagWithValue(tk, lexer, cmd, ctx)
		}
	case tkMultiFlag:
		parseMultiFlag(tk, lexer, cmd, ctx)
	case tkTermination:
		parseTermination(lexer, cmd, ctx)
	}

	parse(lexer, cmd, ctx)
}

func parseArgument(tk token, _ *lexer, cmd *Commander, ctx *context) {
	sub, found := (*cmd).LookupSubCommand(tk.raw)
	if found {
		ctx.addCmd(tk.raw)
		*cmd = sub
	} else {
		ctx.addArg(tk.raw)
	}
}

func parseBoolFlag(tk token, _ *lexer, cmd Commander, ctx *context) {
	ctx.addBoolFlag(cmd.Name(), tk)
}

func parseFlag(tk token, lexer *lexer, cmd Commander, ctx *context) {
	next := lexer.read()
	if next.kind != tkArgument {
		ctx.err = &Error{Flag: tk.raw, Msg: "value not provided"}
		return
	}
	ctx.addFlag(cmd.Name(), &Flag{Name: tk.flagName, Value: next.raw})
}

func parseBoolFlagWithValue(tk token, _ *lexer, cmd Commander, ctx *context) {
	b, err := strconv.ParseBool(tk.flagValue)
	if err != nil {
		ctx.err = &Error{Flag: tk.raw, Msg: fmt.Sprintf("invalid bool value %q", tk.flagValue)}
		return
	}
	ctx.addFlag(cmd.Name(), &Flag{Name: tk.flagName, IsBool: true, BoolValue: b})
}

func parseFlagWithValue(tk token, _ *lexer, cmd Commander, ctx *context) {
	ctx.addFlag(cmd.Name(), &Flag{Name: tk.flagName, Value: tk.flagValue})
}

func parseMultiFlag(tk token, _ *lexer, cmd Commander, ctx *context) {
	for _, r := range tk.flagName {
		flag := string(r)
		if !cmd.IsBoolFlag(flag) {
			ctx.err = &Error{
				Flag: flag,
				Msg:  fmt.Sprintf("%s (%s) only bool flag is allowed as multi short flag.", tk.raw, flag),
			}
			return
		}
		ctx.addFlag(cmd.Name(), &Flag{Name: flag, IsBool: true, BoolValue: true})
	}
}

func parseTermination(lexer *lexer, _ Commander, ctx *context) {
	remain := lexer.readAllAsLiteral()
	for _, arg := range remain {
		ctx.addArg(arg)
	}
}

type Error struct {
	Flag string
	Msg  string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s %s", e.Flag, e.Msg)
}

type context struct {
	*Result
	err error
}

func newContext(rootName string) *context {
	return &context{
		Result: &Result{
			flagMap:     make(map[string][]*Flag),
			rootCmdName: rootName,
		},
	}
}

func (ctx *context) addArg(s string) { ctx.args = append(ctx.args, s) }
func (ctx *context) addCmd(s string) { ctx.commands = append(ctx.commands, s) }
func (ctx *context) addFlag(cmd string, f *Flag) {
	ctx.flagMap[cmd] = append(ctx.flagMap[cmd], f)
}
func (ctx *context) addBoolFlag(cmd string, t token) {
	ctx.addFlag(cmd, &Flag{Name: t.flagName, IsBool: true, BoolValue: true})
}

type lexer struct {
	args    []string
	current int
}

func newLexer(args []string) *lexer {
	stripped := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "" {
			continue
		}
		stripped = append(stripped, arg)
	}
	return &lexer{args: stripped}
}

type tokenKind int

const (
	tkFlag tokenKind = iota
	tkFlagWithValue
	tkArgument
	tkTermination
	tkMultiFlag
	tkInvalid
	tkEnd
)

type token struct {
	kind      tokenKind
	raw       string
	flagName  string
	flagValue string
}

func (l *lexer) read() token {
	if l.isEnd() {
		return token{kind: tkEnd}
	}
	v := l.args[l.current]
	defer func() { l.current++ }()
	return l.toToken(v)
}

func (l *lexer) isEnd() bool { return l.current >= len(l.args) }

func (l *lexer) toToken(v string) token {
	if strings.HasPrefix(v, "--") {
		if len(v) == 2 {
			return token{kind: tkTermination, raw: "--"}
		}

		fName := v[2:]
		// --label=app
		if strings.Contains(v, "=") {
			nameValue := strings.SplitN(fName, "=", 2)
			return token{kind: tkFlagWithValue, raw: v, flagName: nameValue[0], flagValue: nameValue[1]}
		}
		return token{kind: tkFlag, raw: v, flagName: fName}
	}

	if strings.HasPrefix(v, "-") {
		// "-" arg like indicating stdin
		if len(v) == 1 {
			return token{kind: tkArgument, raw: v}
		}
		fName := v[1:]
		// -v
		if len(fName) == 1 {
			return token{kind: tkFlag, raw: v, flagName: fName}
		}
		// -n=10
		if fName[1] == '=' {
			nameValue := strings.SplitN(fName, "=", 2)
			return token{kind: tkFlagWithValue, raw: v, flagName: nameValue[0], flagValue: nameValue[1]}
		}
		// -sSL=bbb
		if strings.Contains(fName, "=") {
			return token{kind: tkInvalid, raw: v}
		}
		// -sSL
		return token{kind: tkMultiFlag, raw: v, flagName: fName}
	}

	return token{kind: tkArgument, raw: v}
}

func (l *lexer) readAllAsLiteral() []string {
	remain := l.args[l.current:]
	l.current = len(l.args)
	return remain
}

func (tk tokenKind) String() string {
	switch tk {
	case tkFlag:
		return "flag"
	case tkFlagWithValue:
		return "flagWithValue"
	case tkArgument:
		return "argment"
	case tkTermination:
		return "termination"
	case tkMultiFlag:
		return "multiFlag"
	case tkInvalid:
		return "invalid"
	case tkEnd:
		return "end"
	}
	return ""
}
