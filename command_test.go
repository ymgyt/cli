package cli_test

import (
	"testing"

	"github.com/ymgyt/cli"
)

func TestCommand_Parse(t *testing.T) {
	type checkFn func(t *testing.T, fs *cli.FlagSet)

	checkExist := func(flags []string) checkFn {
		return func(t *testing.T, fs *cli.FlagSet) {
			for _, f := range flags {
				flag, err := fs.Lookup(f)
				if err != nil {
					t.Fatalf("flag %s not found.", f)
				}
				if flag.Name != f {
					t.Errorf("FlagSet.Lookup(%s) = %s. want %s", f, flag.Name, f)
				}
			}
		}
	}

	tests := map[string]struct {
		args  []string
		check checkFn
	}{
		"simple": {
			args:  []string{"--vervose"},
			check: checkExist([]string{"vervose"}),
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			cmd := &cli.Command{}
			err := cmd.Parse(tc.args)
			if err != nil {
				t.Fatalf("Command.Parse(%v) failed: %s", tc.args, err)
			}
			tc.check(t, cmd.FlagSet)
		})
	}
}
