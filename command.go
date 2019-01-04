package cli

// Command store command info.
type Command struct {
	FlagSet *FlagSet
}

// Parse parse arguments and populate FlagSet.
func (c *Command) Parse(args []string) error {
	return nil
}
