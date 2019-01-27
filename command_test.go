package cli_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/ymgyt/cli"
	"github.com/ymgyt/cli/parser"
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

	t.Run("parse error not panic", func(t *testing.T) {
		root := &cli.Command{Name: "root"}
		var label string
		root.Options().Add(&cli.StringOpt{Long: "label", Var: &label})
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("parse error should not panic")
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

func TestCommand_Parse_return_ParseError(t *testing.T) {
	tests := map[string]struct {
		input []string
		err   error
		setup func(*testing.T, *cli.Command)
	}{
		"no value provided": {
			input: []string{"--label"},
			err:   &parser.Error{},
			setup: func(t *testing.T, root *cli.Command) {
				var label string
				root.Options().Add(&cli.StringOpt{Var: &label, Long: "label"})
			},
		},
		"no value with flagvalue": {
			input: []string{"--label", "--user=admin"},
			err:   &parser.Error{},
			setup: func(t *testing.T, root *cli.Command) {
				var label string
				var user string
				root.Options().Add(&cli.StringOpt{Var: &label, Long: "label"})
				root.Options().Add(&cli.StringOpt{Var: &user, Long: "user"})
			},
		},
		"one of flags no value provided": {
			input: []string{"--label", "app", "--interval"},
			err:   &parser.Error{},
			setup: func(t *testing.T, root *cli.Command) {
				var label string
				root.Options().Add(&cli.StringOpt{Var: &label, Long: "label"})
			},
		},
		"undefined flag in multi short flags": {
			input: []string{"-sxS"},
			err:   &parser.Error{},
			setup: func(t *testing.T, root *cli.Command) {
				var s, S bool
				root.Options().Add(&cli.BoolOpt{Var: &s, Short: "s"}).Add(&cli.BoolOpt{Var: &S, Short: "S"})
			},
		},
		"invalid flag value(bool)": {
			input: []string{"--verbose=ok"},
			err:   &parser.Error{},
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

func TestParseError_Error(t *testing.T) {
	t.Run("simple return msg", func(t *testing.T) {
		err := &cli.ParseError{Message: "parse error"}
		if err.Error() != "parse error" {
			t.Errorf("make sure ParseError.Error() just return message")
		}
	})
}
