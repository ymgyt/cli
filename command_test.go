package cli_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/ymgyt/cli"
)

func TestCommand_Execute(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		execute := false
		cmd := &cli.Command{
			Name: "root",
			Run: func(_ context.Context, _ *cli.Command, args []string) {
				execute = true
			},
		}

		org := os.Args
		defer func() { os.Args = org }()
		os.Args = []string{"root"}
		cmd.Execute(context.Background())
		if !execute {
			t.Error("Command.Execute() does not work")
		}
	})

	t.Run("sub", func(t *testing.T) {
		execute := false
		root := &cli.Command{Name: "root"}
		sub := &cli.Command{
			Name: "sub",
			Run: func(_ context.Context, _ *cli.Command, args []string) {
				execute = true
			},
		}
		root.AddCommand(sub)

		root.ExecuteWithArgs(context.Background(), []string{"sub"})
		if !execute {
			t.Error("Command.Execute() does not work")
		}
	})

	t.Run("sub sub", func(t *testing.T) {
		execute := false
		root := &cli.Command{Name: "root"}
		sub := &cli.Command{Name: "sub"}
		subsub := &cli.Command{
			Name: "subsub",
			Run: func(_ context.Context, _ *cli.Command, args []string) {
				execute = true
			},
		}
		sub.AddCommand(subsub)
		root.AddCommand(sub)

		root.ExecuteWithArgs(context.Background(), []string{"sub", "subsub"})
		if !execute {
			t.Error("Command.Execute() does not work")
		}
	})

	t.Run("parse error panic", func(t *testing.T) {
		root := &cli.Command{Name: "root"}
		var label string
		root.Options().Add(&cli.StringOpt{Long: "label", Var: &label})
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("parse error should panic")
			}
		}()
		root.ExecuteWithArgs(context.Background(), []string{"--label"})
	})
}

func TestCommand_AddCommand(t *testing.T) {
	t.Run("dupulicate add panic", func(t *testing.T) {
		root := &cli.Command{Name: "root"}
		sub1 := &cli.Command{Name: "sub"}
		sub2 := &cli.Command{Name: "sub"}
		root.AddCommand(sub1)
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("adding same name command should panic")
			}
		}()
		root.AddCommand(sub2)
	})
	t.Run("flag set conflict panic", func(t *testing.T) {
		var opt1 string
		var opt2 bool
		root := &cli.Command{Name: "root"}
		root.Options().Add(&cli.StringOpt{Var: &opt1, Long: "label"})
		sub := &cli.Command{Name: "sub"}
		sub.Options().Add(&cli.BoolOpt{Var: &opt2, Long: "label"})
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("flag set conflict should panic")
			}
		}()
		root.AddCommand(sub)
	})
}

func TestCommand_Lookup(t *testing.T) {
	t.Run("by aliaes", func(t *testing.T) {
		root := &cli.Command{Name: "root"}
		found := root.AddCommand(&cli.Command{Name: "sub", Aliases: []string{"alias"}}).Lookup("alias")
		if found == nil || found.Name != "sub" {
			t.Errorf("Command.Lookup does not work")
		}
	})
}

func TestCommand_Parse(t *testing.T) {

	type checkFn func()

	checkString := func(t *testing.T, sp *string, name, want string) {
		if *sp != want {
			t.Errorf("flag %s = %q; want %q", name, *sp, want)
		}
	}
	checkBool := func(t *testing.T, bp *bool, name string, want bool) {
		if *bp != want {
			t.Errorf("flag %s = %v; want %v", name, *bp, want)
		}
	}
	checkStrings := func(t *testing.T, ssp *[]string, name string, want []string) {
		if diff := cmp.Diff(*ssp, want); diff != "" {
			t.Errorf("flag %s (-got +want)\n%s", name, diff)
		}
	}

	t.Run("root command", func(t *testing.T) {
		tests := []struct {
			input  []string
			remain []string
			check  func(*testing.T, *cli.Command) checkFn
		}{
			{
				input: []string{"-v"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					var verbose bool
					root.Options().Add(&cli.BoolOpt{Var: &verbose, Short: "v"})
					return func() {
						checkBool(t, &verbose, "-v", true)
					}
				},
			},
			{
				input: []string{"-v=true"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					var verbose bool
					root.Options().Add(&cli.BoolOpt{Var: &verbose, Short: "v"})
					return func() {
						checkBool(t, &verbose, "-v", true)
					}
				},
			},
			{
				input: []string{"-v", "true"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					var verbose bool
					root.Options().Add(&cli.BoolOpt{Var: &verbose, Short: "v"})
					return func() {
						checkBool(t, &verbose, "-v", true)
					}
				},
			},
			{
				input:  []string{"-v=true", "true"},
				remain: []string{"true"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					var verbose bool
					root.Options().Add(&cli.BoolOpt{Var: &verbose, Short: "v"})
					return func() {
						checkBool(t, &verbose, "-v", true)
					}
				},
			},
			{
				input:  []string{"-v", "cmd"},
				remain: []string{"cmd"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					var verbose bool
					root.Options().Add(&cli.BoolOpt{Var: &verbose, Short: "v"})
					return func() {
						checkBool(t, &verbose, "-v", true)
					}
				},
			},
			{
				input: []string{"--label", "app"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					var label string
					root.Options().Add(&cli.StringOpt{Var: &label, Long: "label"})
					return func() {
						checkString(t, &label, "--label", "app")
					}
				},
			},
			{
				input: []string{"--label=app"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					var label string
					root.Options().Add(&cli.StringOpt{Var: &label, Long: "label"})
					return func() {
						checkString(t, &label, "--label", "app")
					}
				},
			},
			{
				input:  []string{"get", "--label", "app"},
				remain: []string{"get"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					var label string
					root.Options().Add(&cli.StringOpt{Var: &label, Long: "label"})
					return func() {
						checkString(t, &label, "--label", "app")
					}
				},
			},
			{
				input:  []string{"--label", "app", "get"},
				remain: []string{"get"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					var label string
					root.Options().Add(&cli.StringOpt{Var: &label, Long: "label"})
					return func() {
						checkString(t, &label, "--label", "app")
					}
				},
			},
			{
				input:  []string{"--label", "app=go,env=dev", "get"},
				remain: []string{"get"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					var label string
					root.Options().Add(&cli.StringOpt{Var: &label, Long: "label"})
					return func() {
						checkString(t, &label, "--label", "app=go,env=dev")
					}
				},
			},
			{
				input:  []string{"get", "--", "--label", "app", "--user=ops", "--", "arg"},
				remain: []string{"get", "--label", "app", "--user=ops", "--", "arg"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					var label string
					var user string
					root.Options().Add(&cli.StringOpt{Var: &label, Long: "label"})
					root.Options().Add(&cli.StringOpt{Var: &label, Long: "user"})
					return func() {
						checkString(t, &label, "--label", "")
						checkString(t, &user, "--user", "")
					}
				},
			},
			{
				input:  []string{"-"},
				remain: []string{"-"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					return func() {
					}
				},
			},
			{
				input:  []string{"--files", "aaa.yml", "get", "--files=bbb.yml", "arg", "--files", "ccc.yml"},
				remain: []string{"get", "arg"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					var files []string
					root.Options().Add(&cli.StringsOpt{Var: &files, Long: "files"})
					return func() {
						checkStrings(t, &files, "--files", []string{"aaa.yml", "bbb.yml", "ccc.yml"})
					}
				},
			},
			{
				input: []string{"-sSL"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					var silent, show, location bool
					root.Options().Add(&cli.BoolOpt{Var: &silent, Short: "s"})
					root.Options().Add(&cli.BoolOpt{Var: &show, Short: "S"})
					root.Options().Add(&cli.BoolOpt{Var: &location, Short: "L"})
					return func() {
						checkBool(t, &silent, "-s", true)
						checkBool(t, &show, "-S", true)
						checkBool(t, &location, "-L", true)
					}
				},
			},
			{
				input:  []string{"-sSLo", "value", "arg"},
				remain: []string{"arg"},
				check: func(t *testing.T, root *cli.Command) checkFn {
					var silent, show, location bool
					var out string
					root.Options().Add(&cli.BoolOpt{Var: &silent, Short: "s"}).
						Add(&cli.BoolOpt{Var: &show, Short: "S"}).
						Add(&cli.BoolOpt{Var: &location, Short: "L"}).
						Add(&cli.StringOpt{Var: &out, Short: "o"})
					return func() {
						checkBool(t, &silent, "-s", true)
						checkBool(t, &show, "-S", true)
						checkBool(t, &location, "-L", true)
						checkString(t, &out, "-o", "value")
					}
				},
			},
		}

		for _, tc := range tests {
			t.Run(fmt.Sprintf("%v", tc.input), func(t *testing.T) {
				root := &cli.Command{Name: "root"}
				check := tc.check(t, root)
				remain, err := root.Parse(tc.input)
				if err != nil {
					t.Fatalf("Command.Parse() %v", err)
				}

				if diff := cmp.Diff(remain, tc.remain); diff != "" {
					t.Errorf("remain (-got +want)\n%s", diff)
				}

				check()
			})
		}
	})

	t.Run("sub command", func(t *testing.T) {
		buildCmd := func() []*cli.Command {
			root := &cli.Command{Name: "root"}
			sub := &cli.Command{Name: "sub"}
			return []*cli.Command{root, sub}
		}
		tests := []struct {
			input  []string
			remain []string
			check  func(*testing.T, []*cli.Command) checkFn
		}{
			{
				input:  []string{"sub", "-v"},
				remain: []string{"sub"},
				check: func(t *testing.T, cmds []*cli.Command) checkFn {
					root, sub := cmds[0], cmds[1]
					var verbose bool
					root.AddCommand(sub)
					sub.Options().Add(&cli.BoolOpt{Var: &verbose, Short: "v"})
					return func() {
						checkBool(t, &verbose, "-v", true)
					}
				},
			},
			{
				input:  []string{"--label=app", "sub", "--user=sys", "arg", "--verbose"},
				remain: []string{"sub", "arg"},
				check: func(t *testing.T, cmds []*cli.Command) checkFn {
					root, sub := cmds[0], cmds[1]
					var label, user string
					var verbose bool
					sub.Options().Add(&cli.BoolOpt{Var: &verbose, Long: "verbose"})
					sub.Options().Add(&cli.StringOpt{Var: &label, Long: "label"})
					sub.Options().Add(&cli.StringOpt{Var: &user, Long: "user"})
					root.AddCommand(sub)
					return func() {
						checkBool(t, &verbose, "-v", true)
						checkString(t, &label, "--label", "app")
						checkString(t, &user, "--user", "sys")
					}
				},
			},
		}

		for _, tc := range tests {
			t.Run(fmt.Sprintf("%v", tc.input), func(t *testing.T) {
				cmds := buildCmd()
				check := tc.check(t, cmds)
				remain, err := cmds[0].Parse(tc.input)
				if err != nil {
					t.Fatalf("Command.Parse() %v", err)
				}

				if diff := cmp.Diff(remain, tc.remain); diff != "" {
					t.Errorf("remain (-got +want)\n%s", diff)
				}

				check()
			})
		}
	})
}

func TestCommand_Parse_return_ParseError(t *testing.T) {
	tests := map[string]struct {
		input []string
		err   error
		setup func(*testing.T, *cli.Command)
	}{
		"no value provided": {
			input: []string{"--label"},
			err:   &cli.ParseError{},
			setup: func(t *testing.T, root *cli.Command) {
				var label string
				root.Options().Add(&cli.StringOpt{Var: &label, Long: "label"})
			},
		},
		"no value with flagvalue": {
			input: []string{"--label", "--user=admin"},
			err:   &cli.ParseError{},
			setup: func(t *testing.T, root *cli.Command) {
				var label string
				var user string
				root.Options().Add(&cli.StringOpt{Var: &label, Long: "label"})
				root.Options().Add(&cli.StringOpt{Var: &user, Long: "user"})
			},
		},
		"one of flags no value provided": {
			input: []string{"--label", "app", "--interval"},
			err:   &cli.ParseError{},
			setup: func(t *testing.T, root *cli.Command) {
				var label string
				root.Options().Add(&cli.StringOpt{Var: &label, Long: "label"})
			},
		},
		"-sSo=v type flag not supported": {
			input: []string{"-sSo=value"},
			err:   &cli.ParseError{},
			setup: func(t *testing.T, root *cli.Command) {
				var s, S bool
				var o string
				root.Options().Add(&cli.BoolOpt{Var: &s, Short: "s"}).Add(&cli.BoolOpt{Var: &S, Short: "S"}).Add(&cli.StringOpt{Var: &o, Short: "o"})
			},
		},
		"undefined flag in multi short flags": {
			input: []string{"-sxS"},
			err:   &cli.ParseError{},
			setup: func(t *testing.T, root *cli.Command) {
				var s, S bool
				root.Options().Add(&cli.BoolOpt{Var: &s, Short: "s"}).Add(&cli.BoolOpt{Var: &S, Short: "S"})
			},
		},
		"invalid flag value(int)": {
			input: []string{"--max", "three"},
			err:   &cli.ParseError{},
			setup: func(t *testing.T, root *cli.Command) {
				var max int
				root.Options().Add(&cli.IntOpt{Var: &max, Long: "max"})
			},
		},
		"invalid flag value(bool)": {
			input: []string{"--verbose=ok"},
			err:   &cli.ParseError{},
			setup: func(t *testing.T, root *cli.Command) {
				var v bool
				root.Options().Add(&cli.BoolOpt{Var: &v, Long: "verbose"})
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			root := &cli.Command{Name: "root"}
			tc.setup(t, root)
			_, err := root.Parse(tc.input)
			gotT, wantT := reflect.TypeOf(err), reflect.TypeOf(tc.err)
			if gotT != wantT {
				t.Errorf("got %v; want %v", gotT, wantT)
			}
		})
	}
}

func TestOptionConfigurator_Add(t *testing.T) {
	cmd := &cli.Command{}

	var label string
	var enable bool
	var num int
	var rate float64
	var files []string
	var backoffs []int
	var interval time.Duration

	err := cmd.Options().
		Add(&cli.StringOpt{
			Var:         &label,
			Long:        "label",
			Short:       "l",
			Default:     "app",
			Description: "target label",
			Aliases:     []string{"tag"},
		}).
		Add(&cli.BoolOpt{
			Var:         &enable,
			Long:        "enable",
			Short:       "e",
			Description: "bool flag",
		}).
		Add(&cli.IntOpt{
			Var:         &num,
			Long:        "num",
			Short:       "n",
			Description: "int flag",
		}).
		Add(&cli.FloatOpt{
			Var:         &rate,
			Long:        "rate",
			Short:       "r",
			Description: "float flag",
		}).
		Add(&cli.StringsOpt{
			Var:         &files,
			Long:        "files",
			Short:       "f",
			Description: "strings flag",
		}).
		Add(&cli.IntsOpt{
			Var:         &backoffs,
			Long:        "backoffs",
			Description: "ints flag",
		}).
		Add(&cli.DurationOpt{
			Var:         &interval,
			Long:        "interval",
			Short:       "d",
			Description: "duration flag",
		}).Err

	if err != nil {
		t.Fatalf("Command.Options().Add() %v", err)
	}

	for _, key := range []string{"label", "enable", "num", "rate", "files", "backoffs", "interval"} {
		_, err := cmd.FlagSet.Lookup(key)
		if err != nil {
			t.Fatalf("failed to option add. %s not found", key)
		}
	}
}

func TestParseError_Error(t *testing.T) {
	t.Run("simple return msg", func(t *testing.T) {
		err := &cli.ParseError{Message: "parse error"}
		if err.Error() != "parse error" {
			t.Errorf("make sure ParseError.Error() just return message")
		}
	})
}
