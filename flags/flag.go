package flags

import (
	"strconv"
)

type Flag struct {
	Long        string
	Short       string
	Aliases     []string
	Var         Var
	Description string
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
