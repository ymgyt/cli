package flags_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ymgyt/cli/flags"
)

func TestFlag_HasName(t *testing.T) {
	t.Run("aliases work", func(t *testing.T) {
		f := &flags.Flag{Long: "label", Aliases: []string{"alias"}}
		if !f.HasName("alias") {
			t.Errorf("Flag.HasName() does not work by aliases")
		}
	})

	t.Run("does not has name", func(t *testing.T) {
		f := &flags.Flag{Long: "label", Aliases: []string{"alias"}}
		if f.HasName("noone") {
			t.Error("Flag.HasName() return true but, want false")
		}
	})
}

func TestFlag_Name(t *testing.T) {
	t.Run("long name take precedence over short name", func(t *testing.T) {
		f := &flags.Flag{Long: "label", Short: "l"}
		got, want := f.Name(), "label"
		if got != want {
			t.Errorf("flag.Name() does not match. got %s; want %s", got, want)
		}
		f = &flags.Flag{Short: "l"}
		got, want = f.Name(), "l"
		if got != want {
			t.Errorf("short name does not match. got %s; want %s", got, want)
		}
	})
}

func TestFlag_Set(t *testing.T) {
	t.Run("Set correctly recored", func(t *testing.T) {
		var label string
		v := (*flags.StringVar)(&label)
		f := &flags.Flag{Long: "label", Var: v}
		if err := f.Set("app"); err != nil {
			t.Fatalf("Flag.Set() %v", err)
		}
		if !f.IsSet {
			t.Errorf("Flag.Set() should be recoreded")
		}
		if f.Raw != "app" {
			t.Errorf("Flag.Set() parameter should be recorded")
		}
	})

	t.Run("multiple set return err", func(t *testing.T) {
		var label string
		v := (*flags.StringVar)(&label)
		f := &flags.Flag{Long: "label", Var: v, AllowMultipleTimesSet: false}
		if err := f.Set("app"); err != nil {
			t.Errorf("Flag.Set() %v", err)
		}
		err := f.Set("app")
		if err != flags.ErrMulitipleTimesSet {
			t.Errorf("multiple set should return ErrMulitipleTimesSet if now allowd")
		}
	})
}

func TestFlag_Validate(t *testing.T) {
	tests := map[string]struct {
		flag *flags.Flag
		err  error
	}{
		"invalid short flag": {
			flag: &flags.Flag{Short: "xy"},
			err:  flags.ErrInvalidShortFlag,
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			got, want := tc.flag.Validate(), tc.err
			if got != want {
				t.Errorf("Flag.Validate(). got %v, want %v", got, want)
			}
		})
	}
}

func TestFlag_IsBool(t *testing.T) {
	var b bool
	// bv := (*flags.BoolVar)(&b)
	if !(&flags.Flag{Var: (*flags.BoolVar)(&b)}).IsBool() {
		t.Error("bool Flag.IsBool() does not work")
	}
	var s string
	if (&flags.Flag{Var: (*flags.StringVar)(&s)}).IsBool() {
		t.Error("string Flag.IsBool() return true incorrectly")
	}
}

func TestStringVar_Set(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		var s string
		sv := (*flags.StringVar)(&s)
		if err := sv.Set("value"); err != nil {
			t.Fatalf("StringVar.Set(%s) %v", "value", err)
		}
		got, want := s, "value"
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})
}

func TestIntVar_Set(t *testing.T) {
	intVar := func() (*int, *flags.IntVar) {
		var i int
		iv := (*flags.IntVar)(&i)
		return &i, iv
	}
	t.Run("valid num", func(t *testing.T) {
		pi, iv := intVar()
		if err := iv.Set("100"); err != nil {
			t.Fatalf("IntVar.Set(100) %v", err)
		}
		got, want := *pi, 100
		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})
	t.Run("invalid num", func(t *testing.T) {
		_, iv := intVar()
		if err := iv.Set("invalid"); err == nil {
			t.Error("want error, but no error")
		}
	})
}

func TestFloatVar_set(t *testing.T) {
	floatVar := func() (*float64, *flags.FloatVar) {
		var f float64
		fv := (*flags.FloatVar)(&f)
		return &f, fv
	}
	t.Run("simple", func(t *testing.T) {
		pf, fv := floatVar()
		if err := fv.Set("123.456"); err != nil {
			t.Fatalf("FloatVar.Set(123.456) %v", err)
		}
		got, want := *pf, float64(123.456)
		if got != want {
			t.Errorf("got %f, want %f", got, want)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		_, fv := floatVar()
		if err := fv.Set("invalid"); err == nil {
			t.Error("want error, but no error")
		}
	})
}

func TestBoolVar(t *testing.T) {
	boolVar := func() (*bool, *flags.BoolVar) {
		var b bool
		vb := (*flags.BoolVar)(&b)
		return &b, vb
	}
	t.Run("Set", func(t *testing.T) {
		pv, bv := boolVar()
		if err := bv.Set("true"); err != nil {
			t.Fatalf("BoolVar.Set(\"true\") %v", err)
		}
		got, want := *pv, true
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("SetBool", func(t *testing.T) {
		pb, bv := boolVar()
		if err := bv.SetBool(true); err != nil {
			t.Fatalf("BoolVar.Set(true) %v", err)
		}
		got, want := *pb, true
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		_, bv := boolVar()
		if err := bv.Set("invalid"); err == nil {
			t.Error("want error, but no error")
		}
	})
}

func TestStringsVar_Set(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		var ss []string
		ssv := (*flags.StringsVar)(&ss)
		wants := []string{"aaa", "bbb", "ccc"}
		for _, want := range wants {
			if err := ssv.Set(want); err != nil {
				t.Fatalf("StringsVar.Set(%s) %v", want, err)
			}
		}
		if diff := cmp.Diff(ss, wants); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}
	})
	t.Run("multi var", func(t *testing.T) {
		var ss []string
		ssv := (*flags.StringsVar)(&ss)
		if err := ssv.SetMulti("aaa, bbb", ","); err != nil {
			t.Fatalf("StringsVar.SetMulti() %v", err)
		}
		got, want := ss, []string{"aaa", "bbb"}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}
	})
	t.Run("multi var multi times", func(t *testing.T) {
		var ss []string
		ssv := (*flags.StringsVar)(&ss)
		if err := ssv.SetMulti("aaa, bbb", ","); err != nil {
			t.Fatalf("StringsVar.SetMulti() %v", err)
		}
		if err := ssv.SetMulti(",ccc, ddd  ,", ","); err != nil {
			t.Fatalf("StringsVar.SetMulti() %v", err)
		}
		got, want := ss, []string{"aaa", "bbb", "ccc", "ddd"}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}
	})
}

func TestIntsVar_Set(t *testing.T) {
	intsVar := func() (*[]int, *flags.IntsVar) {
		var ns []int
		nsv := (*flags.IntsVar)(&ns)
		return &ns, nsv
	}
	t.Run("simple", func(t *testing.T) {
		pns, nsv := intsVar()
		for _, want := range []string{"10", "100", "1000"} {
			if err := nsv.Set(want); err != nil {
				t.Fatalf("IntsVar.Set(%s) %v", want, err)
			}
		}
		got, want := *pns, []int{10, 100, 1000}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		_, nsv := intsVar()
		if err := nsv.Set("invalid"); err == nil {
			t.Error("want error, gut no error")
		}
	})
	t.Run("multi var", func(t *testing.T) {
		var ns []int
		nsv := (*flags.IntsVar)(&ns)
		if err := nsv.SetMulti("100, 200", ","); err != nil {
			t.Fatalf("IntsVar.SetMulti() %v", err)
		}
		got, want := ns, []int{100, 200}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}
	})
	t.Run("multi var muti times", func(t *testing.T) {
		var ns []int
		nsv := (*flags.IntsVar)(&ns)
		if err := nsv.SetMulti("100, 200", ","); err != nil {
			t.Fatalf("IntsVar.SetMulti() %v", err)
		}
		if err := nsv.SetMulti(",300, 400,", ","); err != nil {
			t.Fatalf("IntsVar.SetMulti() %v", err)
		}
		got, want := ns, []int{100, 200, 300, 400}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}
	})
}

func TestDurationVar_Set(t *testing.T) {
	durationVar := func() (*time.Duration, *flags.DurationVar) {
		var d time.Duration
		dv := (*flags.DurationVar)(&d)
		return &d, dv
	}
	t.Run("simple", func(t *testing.T) {
		pd, dv := durationVar()
		if err := dv.Set("1s"); err != nil {
			t.Fatalf("DurationVar.Set(%s) %v", "1s", err)
		}
		got, want := *pd, time.Second
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		_, dv := durationVar()
		if err := dv.Set("invalid"); err == nil {
			t.Error("want error, gut no error")
		}
	})
}
