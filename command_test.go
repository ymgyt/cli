package cli_test

import (
	"context"
	"fmt"
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
			Name: "test",
			Run: func(_ context.Context, _ *cli.Command, args []string) {
				execute = true
			},
		}

		cmd.ExecuteWithArgs(context.Background(), nil)
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
	var files []string
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
		Add(&cli.StringsOpt{
			Var:         &files,
			Long:        "files",
			Short:       "f",
			Description: "strings flag",
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

	for _, key := range []string{"label", "enable", "num", "files", "interval"} {
		_, err := cmd.FlagSet.Lookup(key)
		if err != nil {
			t.Fatalf("failed to option add. %s not found", key)
		}
	}
}
