package cli_test

import (
	"context"
	"testing"

	"github.com/ymgyt/cli"
)

func TestOptionConfigurator_Add(t *testing.T) {
	cmd := &cli.Command{}

	var label string
	var enable bool
	var num int

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
		}).Err

	if err != nil {
		t.Fatalf("Command.Options().Add() %v", err)
	}

	for _, key := range []string{"label", "enable", "num"} {
		_, err := cmd.FlagSet.Lookup(key)
		if err != nil {
			t.Fatalf("failed to option add. %s not found", key)
		}
	}
}

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
