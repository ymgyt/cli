package cli_test

import (
	"testing"
	"time"

	"github.com/ymgyt/cli"
)

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
}
