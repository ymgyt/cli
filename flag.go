package cli

import (
	"errors"
	"flag"
)

var (
	ErrFlagNotFound = errors.New("flag not found")
)

// FlagSet is collection of Flag.
type FlagSet struct{}

// Lookup lookup flag by given name.
// if not found, Lookup returns ErrFlagNotFound.
func (fs *FlagSet) Lookup(name string) (*Flag, error) {
	if fs == nil {
		return nil, ErrFlagNotFound
	}
	return nil, nil
}

// Flag represents command line flag.
type Flag struct {
	*flag.Flag
}
