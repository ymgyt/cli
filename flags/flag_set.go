package flags

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrFlagNotFound                    = errors.New("flag not found")
	ErrFlagAlreadyExists               = errors.New("flag already exists")
	ErrFlagNameRequired                = errors.New("flag name required")
	ErrBoolAndNonBoolFlagNotCompatible = errors.New("bool and non-bool flags are not compatible")
)

type ParseError struct {
	FlagName string
	Msg      string
}

func (pe *ParseError) Error() string {
	return fmt.Sprintf("flag %s %s", pe.FlagName, pe.Msg)
}

type FlagSet struct {
	Flags []*Flag
	*sync.RWMutex
	once sync.Once
}

// Add add given flag.
// if flag Name(Long, Short, Aliases) conflict already exists flags, returns ErrFlagAlreadyExists.
func (fs *FlagSet) Add(f *Flag) error {
	fs.lasyInit()
	name := f.Long
	if name == "" {
		name = f.Short
	}
	if name == "" {
		return ErrFlagNameRequired
	}
	found, err := fs.Lookup(name)
	if err == ErrFlagNotFound {
		// ok
	} else if err != nil {
		return err
	}

	// bool and non-bool flags are not compatible.
	if found != nil {
		if found.IsBool() != f.IsBool() {
			return ErrBoolAndNonBoolFlagNotCompatible
		}
	}

	fs.Lock()
	fs.Flags = append(fs.Flags, f)
	fs.Unlock()
	return nil
}

// Lookup lookup flag by given flag name. if not found, returns ErrFlagNotFound.
// Long, Short, Aliases are checked.
func (fs *FlagSet) Lookup(name string) (*Flag, error) {
	fs.lasyInit()
	if name == "" {
		return nil, ErrFlagNotFound
	}
	fs.RLock()
	defer fs.RUnlock()
	for _, f := range fs.Flags {
		if f.HasName(name) {
			return f, nil
		}
	}
	return nil, ErrFlagNotFound
}

func (fs *FlagSet) LookupAll(name string) ([]*Flag, error) {
	fs.lasyInit()
	var ret []*Flag
	fs.RLock()
	defer fs.RUnlock()
	for _, f := range fs.Flags {
		if f.HasName(name) {
			ret = append(ret, f)
		}
	}
	return ret, nil
}

func (fs *FlagSet) Merge(o *FlagSet) error {
	if o == nil {
		return nil
	}
	for _, f := range o.Flags {
		if err := fs.Add(f); err != nil {
			return err
		}
	}
	return nil
}

func (fs *FlagSet) lasyInit() {
	if fs == nil {
		fs = &FlagSet{}
	}
	fs.once.Do(func() {
		fs.RWMutex = &sync.RWMutex{}
	})
}
