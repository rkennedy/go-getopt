package getopt

import (
	"iter"
)

// Iterate returns an iterator for options parsed from the given argument list. When iteration terminates, the slice
// pointer, if non-nil, will hold the remaining unparsed arguments.
func Iterate(args []string, opts string, remaining *[]string) iter.Seq2[*Opt, error] {
	g := New(args, opts)
	return func(yield func(*Opt, error) bool) {
		for opt, err := g.Getopt(); opt != nil || err != nil; opt, err = g.Getopt() {
			if !yield(opt, err) {
				break
			}
		}
		if remaining != nil {
			*remaining = g.Args[g.Optind():]
		}
	}
}

// IterateLong returns an iterator for options parsed from the given argument list and option definitions. When
// iteration terminates, the slice pointer, if non-nil, will hold the remaining unparsed arguments.
func IterateLong(args []string, opts string, longOptions []Option, remaining *[]string) iter.Seq2[*Opt, error] {
	g := NewLong(args, opts, longOptions)
	return func(yield func(*Opt, error) bool) {
		for opt, err := g.Getopt(); opt != nil || err != nil; opt, err = g.Getopt() {
			if !yield(opt, err) {
				break
			}
		}
		if remaining != nil {
			*remaining = g.Args[g.Optind():]
		}
	}
}

// IterateLongOnly returns an iterator for options parsed from the given argument list and option definitions. When
// iteration terminates, the slice pointer, if non-nil, will hold the remaining unparsed arguments.
func IterateLongOnly(args []string, opts string, longOptions []Option, remaining *[]string) iter.Seq2[*Opt, error] {
	g := NewLong(args, opts, longOptions)
	return func(yield func(*Opt, error) bool) {
		for opt, err := g.GetoptLongOnly(); opt != nil || err != nil; opt, err = g.GetoptLongOnly() {
			if !yield(opt, err) {
				break
			}
		}
		if remaining != nil {
			*remaining = g.Args[g.Optind():]
		}
	}
}
