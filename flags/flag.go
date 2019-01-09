package flags

import (
	"errors"
	"strconv"
	"time"
)

var (
	ErrMulitipleTimesSet = errors.New("multiple times set")
)

type Flag struct {
	Long                  string
	Short                 string
	Aliases               []string
	Var                   Var
	Description           string
	IsSet                 bool
	Raw                   string
	AllowMultipleTimesSet bool
}

func (f Flag) HasName(name string) bool {
	if f.Long == name || f.Short == name {
		return true
	}
	for _, alias := range f.Aliases {
		if alias == name {
			return true
		}
	}
	return false
}

func (f *Flag) Set(s string) error {
	if f.IsSet && !f.AllowMultipleTimesSet {
		return ErrMulitipleTimesSet
	}
	f.IsSet = true
	f.Raw = s
	return f.Var.Set(s)
}

func (f *Flag) IsBool() bool {
	_, isBool := f.Var.(BooleanVar)
	return isBool
}

type Var interface {
	Set(string) error
}

type StringVar string

func (sv *StringVar) Set(s string) error {
	*sv = StringVar(s)
	return nil
}

type IntVar int

func (iv *IntVar) Set(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*iv = IntVar(n)
	return nil
}

type FloatVar float64

func (fv *FloatVar) Set(s string) error {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*fv = FloatVar(f)
	return nil
}

type BooleanVar interface {
	Var
	SetBool(bool) error
}

type BoolVar bool

func (bv *BoolVar) Set(s string) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	*bv = BoolVar(b)
	return nil
}

func (bv *BoolVar) SetBool(b bool) error {
	*bv = BoolVar(b)
	return nil
}

type StringsVar []string

func (sv *StringsVar) Set(s string) error {
	*sv = append(*sv, s)
	return nil
}

type IntsVar []int

func (iv *IntsVar) Set(s string) error {
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*iv = append(*iv, i)
	return nil
}

type DurationVar time.Duration

func (dv *DurationVar) Set(s string) error {
	d, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*dv = DurationVar(d)
	return nil
}
