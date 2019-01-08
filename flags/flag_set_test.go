package flags_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ymgyt/cli/flags"
)

func TestFlagSet_Add(t *testing.T) {
	t.Run("no name", func(t *testing.T) {
		fs := &flags.FlagSet{}
		err := fs.Add(&flags.Flag{})
		if err != flags.ErrFlagNameRequired {
			t.Errorf("got %v, want ErrFlagNameRequired", err)
		}
	})

	t.Run("same name not conflict", func(t *testing.T) {
		fs := &flags.FlagSet{}
		addFlags(t, fs, &flags.Flag{Long: "label"})
		err := fs.Add(&flags.Flag{Long: "label"})
		if err != nil {
			t.Errorf("same name flag should be compatible, but got %s", err)
		}
	})

	t.Run("bool and non-bool are not compatible", func(t *testing.T) {
		fs := &flags.FlagSet{}
		boolean := flags.BoolVar(false)
		addFlags(t, fs, &flags.Flag{Long: "label"})
		err := fs.Add(&flags.Flag{Long: "label", Var: &boolean})
		if err != flags.ErrBoolAndNonBoolFlagNotCompatible {
			t.Errorf("bool and non-bool are not compatible, but got %s", err)
		}
	})
}

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

func addFlags(t *testing.T, fs *flags.FlagSet, flags ...*flags.Flag) {
	t.Helper()
	for _, flag := range flags {
		if err := fs.Add(flag); err != nil {
			t.Fatalf("failed to FlagSet.Add(%v)", flag)
		}
	}
}

func TestFlagSet_Merge(t *testing.T) {
	type checkFn func(t *testing.T, fs *flags.FlagSet)

	stringFlag := func(long string) *flags.Flag { return &flags.Flag{Long: long} }

	hasFlag := func(long string) checkFn {
		return func(t *testing.T, fs *flags.FlagSet) {
			_, err := fs.Lookup(long)
			if err != nil {
				t.Errorf("flag %s not found", long)
			}
		}
	}

	check := func(fs ...checkFn) []checkFn { return fs }

	tests := map[string]struct {
		receiver *flags.FlagSet
		other    *flags.FlagSet
		checks   []checkFn
	}{
		"just one flag": {
			receiver: &flags.FlagSet{},
			other: &flags.FlagSet{
				Flags: []*flags.Flag{
					stringFlag("label"),
				},
			},
			checks: check(
				hasFlag("label"),
			),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := tc.receiver.Merge(tc.other)
			if err != nil {
				t.Fatalf("FlagSet.Merge() %v", err)
			}
			for _, check := range tc.checks {
				check(t, tc.receiver)
			}
		})
	}
}
