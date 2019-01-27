package examples_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ymgyt/cli"
)

type pod struct {
	showHelp bool
	level    int
}

func (pod *pod) run(ctx context.Context, cmd *cli.Command, args []string) {
	if pod.showHelp {
		fmt.Fprintln(cmd.Stdout, "pod help message")
		return
	}

	fmt.Fprintf(cmd.Stdout, "get pod level=%d\n", pod.level)
}

type rs struct {
	label string
}

func (rs *rs) run(ctx context.Context, cmd *cli.Command, args []string) {
	fmt.Fprintf(cmd.Stdout, "get replicaset label=%s\n", rs.label)
}

func TestRealCase(t *testing.T) {
	buildCmd := func(stdin io.Reader, stdout, stderr io.Writer) *cli.Command {
		root := &cli.Command{
			Name:      "clictl",
			ShortDesc: "cli example command",
			LongDesc:  "when in the go, do as gophers do",
			Stdin:     stdin, Stdout: stdout, Stderr: stderr,
		}

		versionCmd := &cli.Command{
			Name:  "version",
			Stdin: stdin, Stdout: stdout, Stderr: stderr,
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
			Add(&cli.IntOpt{Var: &pod.level, Long: "level", Short: "l", Default: 1}).Err
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

	tests := map[string]struct {
		args    []string
		wantOut string
		wantErr string
	}{
		"get pod 1": {
			args:    []string{"get", "pod", "--level=2"},
			wantOut: "get pod level=2\n",
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
