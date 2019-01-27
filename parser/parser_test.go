package parser_test

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ymgyt/cli/parser"
)

type fakeCmd struct {
	name      string
	subs      []*fakeCmd
	boolFlags []string
}

func (f *fakeCmd) Name() string { return f.name }

func (f *fakeCmd) LookupSubCommand(name string) (parser.Commander, bool) {
	for _, sub := range f.subs {
		if sub.name == name {
			return sub, true
		}
	}
	return nil, false
}

func (f *fakeCmd) IsBoolFlag(name string) bool {
	for _, bf := range f.boolFlags {
		if name == bf {
			return true
		}
	}
	return false
}

func TestParser_Parse(t *testing.T) {

	type checkFn func(*testing.T, *parser.Result, error)
	check := func(fs ...checkFn) []checkFn { return fs }

	notNil := func(t *testing.T, r *parser.Result) {
		t.Helper()
		if r == nil {
			t.Fatalf("Result is nil")
		}
	}
	hasCmds := func(ss ...string) checkFn {
		return func(t *testing.T, r *parser.Result, err error) {
			notNil(t, r)
			if diff := cmp.Diff(r.Commands(), ss); diff != "" {
				t.Errorf("commands does not match. (-got +want)%s", diff)
			}
		}
	}
	hasArgs := func(ss ...string) checkFn {
		return func(t *testing.T, r *parser.Result, err error) {
			notNil(t, r)
			if diff := cmp.Diff(r.Args(), ss); diff != "" {
				t.Errorf("args does not match. (-got +want)%s", diff)
			}
		}
	}
	hasFlags := func(flags ...*parser.Flag) checkFn {
		return func(t *testing.T, r *parser.Result, err error) {
			notNil(t, r)
			if diff := cmp.Diff(r.AllFlags(), flags); diff != "" {
				t.Errorf("flags does not match. (-got +want)%s", diff)
			}
		}
	}
	hasCmdFlags := func(cmd string, flags ...*parser.Flag) checkFn {
		return func(t *testing.T, r *parser.Result, err error) {
			notNil(t, r)
			got := r.Flags(cmd)
			if diff := cmp.Diff(got, flags); diff != "" {
				t.Errorf("flags of %q does not match. (-got +want)%s", cmd, diff)
			}
		}
	}
	hasErr := func(wantErr error) checkFn {
		return func(t *testing.T, r *parser.Result, gotErr error) {
			got, want := reflect.TypeOf(gotErr), reflect.TypeOf(wantErr)
			if got != want {
				t.Errorf("error type does not match. got %v; want %v", got, want)
			}
		}
	}

	flag := func(name, v string) *parser.Flag { return &parser.Flag{Name: name, Value: v} }
	boolFlag := func(name string, b bool) *parser.Flag { return &parser.Flag{Name: name, IsBool: true, BoolValue: b} }

	tests := map[string]struct {
		args   []string
		checks []checkFn
	}{
		"single arg": {
			args:   []string{"arg"},
			checks: check(hasArgs("arg")),
		},
		"sub command": {
			args:   []string{"sub"},
			checks: check(hasCmds("sub")),
		},
		"subsub command": {
			args:   []string{"sub", "subsub"},
			checks: check(hasCmds("sub", "subsub")),
		},
		"subsub command and args": {
			args:   []string{"sub", "subsub", "arg1", "arg2"},
			checks: check(hasCmds("sub", "subsub"), hasArgs("arg1", "arg2")),
		},
		"single flag": {
			args:   []string{"--label", "app"},
			checks: check(hasFlags(flag("label", "app"))),
		},
		"single flag with value": {
			args:   []string{"--label=app"},
			checks: check(hasFlags(flag("label", "app"))),
		},
		"single bool flag with value": {
			args:   []string{"--verbose=true"},
			checks: check(hasFlags(boolFlag("verbose", true))),
		},
		"termination": {
			args:   []string{"sub", "--env=test", "exec", "--", "aaa", "-v", "--verbose", "--label", "bbb"},
			checks: check(hasCmds("sub"), hasFlags(flag("env", "test")), hasArgs("exec", "aaa", "-v", "--verbose", "--label", "bbb")),
		},
		"multi flag": {
			args:   []string{"-sSf"},
			checks: check(hasFlags(boolFlag("s", true), boolFlag("S", true), boolFlag("f", true))),
		},
		"command and flags": {
			args: []string{"--log", "warn", "sub", "--label=ops", "-v", "subsub", "--exclude", ".git", "arg"},
			checks: check(
				hasCmdFlags("root", flag("log", "warn")),
				hasCmdFlags("sub", flag("label", "ops"), boolFlag("v", true)),
				hasCmdFlags("subsub", flag("exclude", ".git")),
				hasArgs("arg"),
			),
		},
		"value not provided": {
			args:   []string{"--label"},
			checks: check(hasErr(&parser.Error{})),
		},
		"value not provided with bool": {
			args:   []string{"--label", "--verbose"},
			checks: check(hasErr(&parser.Error{})),
		},
		"invalid bool value": {
			args:   []string{"--verbose=XtrueX"},
			checks: check(hasErr(&parser.Error{})),
		},
	}

	boolFlags := []string{"v", "verbose", "s", "S", "f"}
	root := &fakeCmd{
		name: "root",
		subs: []*fakeCmd{
			{
				name: "sub",
				subs: []*fakeCmd{
					{
						name:      "subsub",
						boolFlags: boolFlags,
					},
				},
				boolFlags: boolFlags,
			},
		},
		boolFlags: boolFlags,
	}
	parser := &parser.Parser{Root: root}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := parser.Parse(tc.args)
			for _, check := range tc.checks {
				check(t, result, err)
			}
		})
	}
}
