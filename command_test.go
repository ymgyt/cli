package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

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

func TestHelpFunc(t *testing.T) {
	want := `when in the go, do as gophers do

SubCommands
      get: get resources
  version: print version
`
	var b strings.Builder
	cmd := buildCmd(nil, nil, nil)
	cli.HelpFunc(&b, cmd)
	if diff := cmp.Diff(b.String(), want); diff != "" {
		t.Errorf("help message does not match. (-got +want)%s", diff)
	}

	podCmdWant := `get pod long desc...

Options
  -h, --help : print this
  -l, --level: level
      --nums : numbers
  -v         : verbose
`
	podCmd := cmd.Lookup("get").Lookup("pod")
	if podCmd == nil {
		t.Fatal("look up failed")
	}
	b = strings.Builder{}
	cli.HelpFunc(&b, podCmd)
	if diff := cmp.Diff(b.String(), podCmdWant); diff != "" {
		t.Errorf("pod command help message does not match. (-got +want)%s", diff)
	}
}

func TestRealCase(t *testing.T) {

	tests := map[string]struct {
		args    []string
		wantOut string
		wantErr string
	}{
		"get pod 1": {
			args:    []string{"get", "pod", "--level=2", "--nums=100,200"},
			wantOut: "get pod level=2 numbers=[100 200]\n",
		},
		"show version": {
			args:    []string{"version"},
			wantOut: "v9.9.9\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
			cmd := buildCmd(nil, stdout, stderr)
			cmd.ExecuteWithArgs(context.Background(), tc.args)
			gotOut, gotErr := stdout.String(), stderr.String()
			if diff := cmp.Diff(gotOut, tc.wantOut); diff != "" {
				t.Fatalf("command output does not mathc. (-got +want)%s", diff)
			}
			if diff := cmp.Diff(gotErr, tc.wantErr); diff != "" {
				t.Fatalf("command err output does not mathc. (-got +want)%s", diff)
			}
		})
	}
}

type pod struct {
	showHelp bool
	verbose  bool
	level    int
	numbers  []int
}

func (pod *pod) run(ctx context.Context, cmd *cli.Command, args []string) {
	if pod.showHelp {
		fmt.Fprintln(cmd.Stdout, "pod help message")
		return
	}

	fmt.Fprintf(cmd.Stdout, "get pod level=%d numbers=%v\n", pod.level, pod.numbers)
}

type rs struct {
	label string
}

func (rs *rs) run(ctx context.Context, cmd *cli.Command, args []string) {
	fmt.Fprintf(cmd.Stdout, "get replicaset label=%s\n", rs.label)
}

// nolint: gochecknoglobals
var buildCmd = func(stdin io.Reader, stdout, stderr io.Writer) *cli.Command {
	root := &cli.Command{
		Name:      "clictl",
		ShortDesc: "cli example command",
		LongDesc:  "when in the go, do as gophers do",
		Stdin:     stdin, Stdout: stdout, Stderr: stderr,
	}

	versionCmd := &cli.Command{
		Name:      "version",
		ShortDesc: "print version",
		Stdin:     stdin, Stdout: stdout, Stderr: stderr,
		Run: func(_ context.Context, cmd *cli.Command, args []string) {
			fmt.Fprintln(cmd.Stdout, "v9.9.9")
		},
	}

	getCmd := &cli.Command{
		Name:      "get",
		ShortDesc: "get resources",
		LongDesc:  "get long desc...",
		Stdin:     stdin, Stdout: stdout, Stderr: stderr,
	}

	pod := &pod{}
	podCmd := &cli.Command{
		Name:      "pod",
		ShortDesc: "get pod resources",
		LongDesc:  "get pod long desc...",
		Run:       pod.run,
		Stdin:     stdin, Stdout: stdout, Stderr: stderr,
	}
	err := podCmd.Options().
		Add(&cli.BoolOpt{Var: &pod.showHelp, Long: "help", Short: "h", Description: "print this"}).
		Add(&cli.BoolOpt{Var: &pod.verbose, Short: "v", Description: "verbose"}).
		Add(&cli.IntsOpt{Var: &pod.numbers, Long: "nums", Description: "numbers"}).
		Add(&cli.IntOpt{Var: &pod.level, Long: "level", Short: "l", Default: 1, Description: "level"}).Err
	if err != nil {
		panic(err)
	}

	rs := &rs{}
	rsCmd := &cli.Command{
		Name:    "replicaset",
		Aliases: []string{"rs"},
		Run:     rs.run,
		Stdin:   stdin, Stdout: stdout, Stderr: stderr,
	}

	return root.
		AddCommand(versionCmd).
		AddCommand(getCmd.
			AddCommand(podCmd).
			AddCommand(rsCmd))
}
