package flags_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ymgyt/cli/flags"
)

func TestStringVar_Set(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		var s string
		sv := flags.StringVar(s)
		psv := &sv
		want := "value"
		if err := psv.Set(want); err != nil {
			t.Fatalf("StringVar.Set(%s) %v", want, err)
		}
		got := string(sv)
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})
}

func TestIntVar_Set(t *testing.T) {
	t.Run("valid num", func(t *testing.T) {
		var i int
		iv := flags.IntVar(i)
		piv := &iv
		if err := piv.Set("100"); err != nil {
			t.Fatalf("IntVar.Set(100) %v", err)
		}
		got, want := int(iv), 100
		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})
}

func TestBoolVar(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		var b bool
		vb := flags.BoolVar(b)
		pvb := &vb
		if err := pvb.Set("true"); err != nil {
			t.Fatalf("BoolVar.Set(\"true\") %v", err)
		}
		got, want := bool(vb), true
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("SetBool", func(t *testing.T) {
		var b bool
		vb := flags.BoolVar(b)
		pvb := &vb
		if err := pvb.SetBool(true); err != nil {
			t.Fatalf("BoolVar.Set(true) %v", err)
		}
		got, want := bool(vb), true
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestStringSliceVar_Set(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		var ss []string
		ssv := flags.StringSliceVar(ss)
		pssv := &ssv
		wants := []string{"aaa", "bbb", "ccc"}
		for _, want := range wants {
			if err := pssv.Set(want); err != nil {
				t.Fatalf("StringSliceVar.Set(%s) %v", want, err)
			}
		}
		if diff := cmp.Diff([]string(ssv), wants); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}
	})
}

func TestDurationVar_Set(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		var d time.Duration
		dv := flags.DurationVar(d)
		pdv := &dv
		want := time.Second
		if err := pdv.Set("1s"); err != nil {
			t.Fatalf("DurationVar.Set(%s) %v", "1s", err)
		}
		if time.Duration(dv) != want {
			t.Errorf("got %v, want %v", time.Duration(dv), want)
		}
	})
}
