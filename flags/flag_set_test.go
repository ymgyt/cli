package flags_test

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"

	"github.com/ymgyt/cli/flags"
)

func TestFlagSet_Lookup(t *testing.T) {

	flagLabel := &flags.Flag{
		Long:    "label",
		Short:   "l",
		Aliases: []string{"tag"},
	}

	t.Run("not found", func(t *testing.T) {
		fs := &flags.FlagSet{}
		_, err := fs.Lookup("key")
		if err != flags.ErrFlagNotFound {
			t.Errorf("got %v, want ErrFlagNotFound", err)
		}
	})

	t.Run("found", func(t *testing.T) {
		fs := &flags.FlagSet{}
		addFlags(t, fs, flagLabel)
		got, err := fs.Lookup(flagLabel.Long)
		if err != nil {
			t.Fatalf("FlagSet.Loopkup(%v)", flagLabel.Long)
		}
		if diff := cmp.Diff(got, flagLabel); diff != "" {
			t.Errorf("got %v, want %v", got, flagLabel)
		}
	})
}

func TestFlagSet_Add(t *testing.T) {
	t.Run("no name", func(t *testing.T) {
		fs := &flags.FlagSet{}
		err := fs.Add(&flags.Flag{})
		if err != flags.ErrFlagNameRequired {
			t.Errorf("got %v, want ErrFlagNameRequired", err)
		}
	})

	t.Run("conflict", func(t *testing.T) {
		fs := &flags.FlagSet{}
		addFlags(t, fs, &flags.Flag{Long: "label"})
		err := fs.Add(&flags.Flag{Long: "label"})
		if err != flags.ErrFlagAlreadyExists {
			t.Errorf("got %v, want ErrFlagAlreadyExists", err)
		}
	})
}
func TestFlagSet_Parse(t *testing.T) {

	t.Run("root command", func(t *testing.T) {
		fs := &flags.FlagSet{}

		boolFlag := flags.BoolVar(false)
		addFlags(t, fs,
			&flags.Flag{Long: "label"},
			&flags.Flag{Long: "enable", Short: "e", Var: &boolFlag},
		)

		tests := []struct {
			ctx     flags.ParseContext
			want    flags.ParseContext
			wantErr bool
			err     error
		}{
			{
				ctx:  flags.ParseContext{Remain: []string{"cmd"}, State: flags.ParseStart},
				want: flags.ParseContext{Remain: []string{}, State: flags.ParseCommand, Value: "cmd"},
			},
			{
				ctx:  flags.ParseContext{Remain: []string{"--label", "app"}, State: flags.ParseStart},
				want: flags.ParseContext{Remain: []string{"app"}, State: flags.ParseFlag, FlagName: "label"},
			},
			{
				ctx:  flags.ParseContext{Remain: []string{"app"}, State: flags.ParseFlag, FlagName: "label"},
				want: flags.ParseContext{Remain: []string{}, State: flags.ParseFlagArg, FlagName: "label", Value: "app"},
			},
			{
				ctx:  flags.ParseContext{Remain: []string{"--enable"}, State: flags.ParseStart},
				want: flags.ParseContext{Remain: []string{}, State: flags.ParseBoolFlag, FlagName: "enable"},
			},
			{
				ctx:  flags.ParseContext{Remain: []string{"subcmd"}, State: flags.ParseBoolFlag},
				want: flags.ParseContext{Remain: []string{}, State: flags.ParseCommand, Value: "subcmd"},
			},
			{
				ctx:     flags.ParseContext{Remain: []string{"--enable"}, State: flags.ParseFlag, FlagName: "label"},
				wantErr: true,
				err:     &flags.ParseError{FlagName: "--label", Msg: "value required"},
			},
			{
				ctx:     flags.ParseContext{Remain: []string{"-e"}, State: flags.ParseFlag, FlagName: "label"},
				wantErr: true,
				err:     &flags.ParseError{FlagName: "--label", Msg: "value required"},
			},
		}

		for _, tc := range tests {
			t.Run(fmt.Sprintf("%v", tc.ctx.Remain), func(t *testing.T) {
				got, err := fs.Parse(&tc.ctx)
				if tc.wantErr {
					if diff := cmp.Diff(err, tc.err); diff != "" {
						t.Errorf("got error: %v, want error: %v", err, tc.err)
					}
					return
				}
				if err != nil {
					t.Fatalf("FlagSet.Parse(%v)", tc.ctx)
				}
				if diff := cmp.Diff(got, &tc.want); diff != "" {
					t.Errorf("(-got +want)\n%s", diff)
				}
			})
		}
	})
}

func TestFlagSet_ParseAll(t *testing.T) {
	t.Run("simple", func(t *testing.T) {

		var label flags.StringVar
		var enable flags.BoolVar

		newFlagSet := func(t *testing.T) *flags.FlagSet {
			fs := &flags.FlagSet{}
			label = flags.StringVar("")
			enable = flags.BoolVar(false)
			fss := []*flags.Flag{
				&flags.Flag{Long: "label", Var: &label},
				&flags.Flag{Long: "enable", Var: &enable},
			}
			addFlags(t, fs, fss...)
			return fs
		}

		type checkFn func(t *testing.T, fs *flags.FlagSet)

		tests := []struct {
			args  []string
			check checkFn
		}{
			{
				args: []string{"--label", "app", "--enable"},
				check: func(t *testing.T, fs *flags.FlagSet) {
					if string(label) != "app" {
						t.Errorf("--label does not match. got %v, want %v", label, "app")
					}
					if !bool(enable) {
						t.Errorf("--enable does not match. got %v, want %v", false, true)
					}
				},
			},
			{
				args: []string{"--label=app", "--enable"},
				check: func(t *testing.T, fs *flags.FlagSet) {
					if string(label) != "app" {
						t.Errorf("--label does not match. got %v, want %v", label, "app")
					}
					if !bool(enable) {
						t.Errorf("--enable does not match. got %v, want %v", false, true)
					}
				},
			},
		}

		for _, tc := range tests {
			t.Run(fmt.Sprintf("%v", tc.args), func(t *testing.T) {
				fs := newFlagSet(t)
				_, err := fs.ParseAll(tc.args)
				if err != nil {
					t.Fatal(err)
				}
				tc.check(t, fs)
			})
		}
	})
}

func addFlags(t *testing.T, fs *flags.FlagSet, flags ...*flags.Flag) {
	t.Helper()
	for _, flag := range flags {
		if err := fs.Add(flag); err != nil {
			t.Fatalf("failed to FlagSet.Add(%v)", flag)
		}
	}
}

var _ = spew.Dump
