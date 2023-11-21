package getopt

import "fmt"

// AmbiguousOptionError is returned when there is no exact match for Option, but more than one abbreviated match, which
// are given in Candidates.
type AmbiguousOptionError struct {
	Option     string
	Candidates []string
	prefix     string
}

func (e AmbiguousOptionError) Error() string {
	result := fmt.Sprintf("option '%s%s' is ambiguous; possibilities:", e.prefix, e.Option)
	for _, opt := range e.Candidates {
		result = result + fmt.Sprintf(" '%s%s'", e.prefix, opt)
	}
	return result
}

// UnrecognizedOptionError is returned when Option on the command line is not a recogized option.
type UnrecognizedOptionError struct {
	Option string
	prefix string
}

func (e UnrecognizedOptionError) Error() string {
	return fmt.Sprintf("unrecognized option '%s%s'", e.prefix, e.Option)
}

// ArgumentNotAllowedError is returned when Option does not accept arguments but one is provided anyway.
type ArgumentNotAllowedError struct {
	Option string
	prefix string
}

func (e ArgumentNotAllowedError) Error() string {
	return fmt.Sprintf("option '%s%s' doesn't allow an argument", e.prefix, e.Option)
}

// ArgumentRequiredError is returned when Option expects an argument and none is given.
type ArgumentRequiredError struct {
	Option string
	prefix string
}

func (e ArgumentRequiredError) Error() string {
	return fmt.Sprintf("option '%s%s' requires an argument", e.prefix, e.Option)
}
