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

	t.Run("same name conflict", func(t *testing.T) {
		fs := &flags.FlagSet{}
		addFlags(t, fs, &flags.Flag{Long: "label"})
		err := fs.Add(&flags.Flag{Long: "label"})
		if err != flags.ErrFlagAlreadyExists {
			t.Errorf("adding same name flag should return FlagAlreadyExists error,but got %s", err)
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
	t.Run("empty name", func(t *testing.T) {
		if _, err := (&flags.FlagSet{}).Lookup(""); err != flags.ErrFlagNotFound {
			t.Errorf("want ErrFlagNotFound, got %v", err)
		}
	})
}

func TestFlagSet_Traverse(t *testing.T) {
	fs := &flags.FlagSet{}
	addFlags(t, fs,
		&flags.Flag{Short: "a"},
		&flags.Flag{Short: "b"},
		&flags.Flag{Short: "c"})
	var traversed []string
	fs.Traverse(func(f *flags.Flag) {
		traversed = append(traversed, f.Name())
	})
	got, want := traversed, []string{"a", "b", "c"}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("(-got +want)%s", diff)
	}
}

func TestParseError_Error(t *testing.T) {
	err := &flags.ParseError{FlagName: "label", Msg: "err message"}
	got, want := err.Error(), "flag label err message"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func addFlags(t *testing.T, fs *flags.FlagSet, flags ...*flags.Flag) {
	t.Helper()
	for _, flag := range flags {
		if err := fs.Add(flag); err != nil {
			t.Fatalf("failed to FlagSet.Add(%v)", flag)
		}
	}
}
