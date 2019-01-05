package examples_test

import (
	"context"
	"fmt"

	"github.com/ymgyt/cli"
)

func ExampleCommand_Execute() {
	var label string
	var enable bool
	var num int

	root := &cli.Command{Name: "root"}
	sub := &cli.Command{Name: "sub"}
	subsub := &cli.Command{
		Name: "subsub",
		Run: func(ctx context.Context, cmd *cli.Command, args []string) {
			fmt.Printf("label: %s, enable: %v, num: %d", label, enable, num)
		},
	}

	root.AddCommand(sub)
	sub.AddCommand(subsub)

	err := subsub.Options().
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
		panic(err)
	}

	root.ExecuteWithArgs(context.Background(), []string{"sub", "subsub", "--label", "gopher", "--enable", "-n=5"})

	// Output: label: gopher, enable: true, num: 5
}
