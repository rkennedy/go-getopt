// Package getopt provides GNU-style argument parsing for both short and long arguments.
//
// Differences from Posix getopt:
//  1. There are no global variables. All operations are performed on a Getopt struct that maintains state between
//     successive calls.
//  2. The opterr setting is permanently false. Errors are never printed anywhere by this library. Instead, errors are
//     returned and the caller can choose what to do with them. The text of the errors corresponds to messages that
//     would be printed by GNU getopt. The leading ':' in the option spec that controls error-reporting is accepted for
//     compatibility, but it's ignored.
//  3. A struct is returned instead of just the matched option character. The struct includes the option character, any
//     value that would have been in optarg, as well as any value that would have been returned in the longindex
//     argument to getopt_long.
//  4. The optopt value is not used. Instead, the relevant unrecognized character or option name is available in the
//     Option field of whatever error gets returned.
//  5. The Flag and Val fields of Option have type rune, not int.
//  6. The list of options and arguments cannot be changed in the middle of parsing. The argument list and option
//     definition are set once at the start, and then you just call Getopt or GetoptLong with no parameters.
package getopt

import (
	"slices"
	"strings"
)

const (
	dash               = "-"
	argumentTerminator = "--"
)

// ArgumentDisposition is an enum specifying whether an option expects to be followed by an argument. Use it when
// defining the Option list for long options.
type ArgumentDisposition int

// These values are used for the HasArg field of Options.
const (
	NoArgument       ArgumentDisposition = iota // The option does not take an argument.
	RequiredArgument                            // The option requires an argument.
	OptionalArgument                            // The option takes an optional argument.
)

// Option describes the long-named options requested by the application. The longopts arguments to NewLong, ResetLong,
// and others are slices of these types.
//
// If Flag is not nil, then it points to a variable that will have its value set to Val when the option is found, but
// will be left unchanged if the option is not found.
//
// To have a long-named option do something other than set a rune to a compiled-in constant, such as set a value from
// Opt.Arg, set the option's Flag to nil and its Val to a nonzero value (such as the option's equivalent single-letter
// option character, if there is one). For long options that have a nil Flag, Getopt returns the Val field in Opt.C.
type Option struct {
	Name   string
	HasArg ArgumentDisposition
	Flag   *rune
	Val    rune
}

// ordering describes how to deal with options that follow non-option arguments.
//
// The special argument '--' forces an end of option-scanning regardless of the value of 'ordering'. In the case of
// ReturnInOrder, only '--' can cause Getopt to return -1 with Optind != len(Args).
type ordering int

const (
	// RequireOrder means don't recognize them as options; stop option processing when the first non-option is seen.
	// This is what POSIX specifies should happen.
	RequireOrder ordering = iota

	// Permute means permute the contents of Args as we scan, so that eventually all the non-options are at the end.
	// This allows options to be given in any order, even with programs that were not written to expect this.
	Permute

	// ReturnInOrder is an option available to programs that were written to expect options and other arguments in
	// any order and that care about the ordering of the two. We describe each non-option argument as if it were the
	// argument of an option with character code 1.
	ReturnInOrder
)

// Getopt is an option parser.
type Getopt struct {
	Args         []string // Args holds a copy of the argument list. It gets permuted during parsing.
	shortOptions optinfo
	longOptions  []Option

	optind   int  // Optind is the index into parent argv vector.
	optreset bool // Optreset resets getopt.

	// The rest of the argument to be scanned in the option-element in which the last option character we returned
	// was found. This allows us to pick up the scan where we left off.
	//
	// If this is empty, it means to resume the scan by advancing to the next argument.
	nextChar []rune

	firstNonopt int // Index in Args of the first non-option that has been skipped.
	lastNonopt  int // Index in Args after the last non-option that was skipped.
}

// Opt is a result from Getopt. If C is 0, then a long option was matched, Flag pointed at a variable and it has been
// assigned a value from Val, but Opt.Arg holds the argument for that option, if any, and LongInd holds the index of the
// long option that matched. If C is 1, then ordering is ReturnInOrder and Arg points to the current non-option
// argument. Otherwise, C holds the rune value of the matched short option or Val of the matched long option (in which
// case LongInd also holds the index of the matched long option).
type Opt struct {
	C       rune
	Arg     *string
	LongInd int
}

// Optind returns the argument index of the next argument to be scanned. When Getopt returns -1, Optind will be the
// index of the first non-option element in Args, which is where the caller should pick up scanning.
//
// To reset the scanner, call Optreset instead of setting optind=0.
func (g *Getopt) Optind() int {
	return g.optind
}

// Getopt scans elements of Args for option characters.
//
// If an element of Args starts with '-', and is not exactly "-" or "--", then it is an option element. The characters
// of this element (aside from the initial '-') are option characters. If Getopt is called repeatedly, it returns
// successively each of the option characters from each of the option elements.
//
// If Getopt finds another option character, it returns an Opt and updates 'optind' and 'nextchar' so that the next call
// to Getopt can resume the scan with the following option character or argument.
//
// If there are no more option characters, Getopt returns nil. Then 'optind' is the index in Args of the first argument
// that is not an option. (The arguments have been permuted so that those that are not options now come last.)
//
// If an option character is seen that was not listed in the opt string when calling New or Reset, then Getopt returns
// an UnrecognizedOptionError. It does not print messages to stderr; in that respect, it behaves as if the Posix value
// opterr is always false.
//
// If an option wants an argument, then the following text in the same Args element, or the text of the following Args
// element, is returned in Opt.Arg. If the option's argument is optional, then if there is text in the current Args
// element, it is returned in Opt.Arg. Otherwise, Opt.Arg will be nil.
//
// Long-named options begin with '--' instead of '-'. Their names may be abbreviated as long as the abbreviation is
// unique or is an exact match for some defined option. If they have an argument, it follows the option name in the same
// Args element, separated from the option name by a '=', or else the in next Args element. When Getopt finds a
// long-named option, it returns an Opt whose C field is 0 if that option's 'Flag' field is non-nil, or the value of the
// option's 'Val' field if the 'Flag' field is nil. The Opt.LongInd field is only valid when a long-named option has
// been found.
//
// The elements of Args aren't really const, because we permute them.
func (g *Getopt) Getopt() (*Opt, error) {
	return g.getoptInternal(false)
}

// GetoptLong is identical to Getopt.
func (g *Getopt) GetoptLong() (*Opt, error) {
	return g.Getopt()
}

// GetoptLongOnly is identical to Getopt and GetoptLong, except that '-' as well as '--' can introduce long-named
// options.
func (g *Getopt) GetoptLongOnly() (*Opt, error) {
	return g.getoptInternal(true)
}

// Reset initializes the Getopt for a new round of argument-parsing, using the argument list and short option
// specification passed in here. Unlike the Posix getopt, this library stores the argument list and the option
// definition in the Getopt struct so that each successive call in the parsing loop doesn't need to repeat the same list
// of parameters.
//
// The opts string is a list of characters that are recognized option letters, optionally followed by colons to specify
// that that letter takes an argument (returned via Opt.Arg). If a letter in opts is followed by two colons, its
// argument is optional. This behavior mimics the GNU extension.
//
// The argument '--' causes premature termination of argument scanning, explicitly telling Getopt that there are no more
// options.
//
// If opts begins with '-', then non-option arguments are treated as arguments to the option 1. This behavior mimics the
// GNU extension. If opts begins with '+', or if POSIXLY_CORRECT is set in the environment at the time Reset is called,
// then arguments will not be permuted during parsing.
//
// The argument list is assumed to include the program name at index 0; it is not returned or processed as a real
// argument.
func (g *Getopt) Reset(args []string, opts string) {
	g.Args = args
	g.shortOptions = parseShortOptionSpec(opts)

	g.longOptions = nil
	g.optind = 1
	g.optreset = true
	g.nextChar = nil
	g.firstNonopt = g.optind
	g.lastNonopt = g.optind
}

// ResetLong initializes the Getopt for a new round of argument-parsing using the argument list and short and long
// option specifications given.
//
// If opts includes 'W' followed by ';', then a GNU extension is enabled that allows long options to be specified as
// arguments to the short option '-W'. The argument sequence '-W foo=bar' will behave just as if it were '--foo=bar'.
func (g *Getopt) ResetLong(args []string, opts string, longOptions []Option) {
	g.Reset(args, opts)
	g.longOptions = longOptions
}

// New creates a new Getopt initialized as by Reset.
func New(args []string, opts string) *Getopt {
	var g Getopt
	g.Reset(args, opts)
	return &g
}

// NewLong creates a new Getopt initialized as by ResetLong.
func NewLong(args []string, opts string, longOptions []Option) *Getopt {
	var g Getopt
	g.ResetLong(args, opts, longOptions)
	return &g
}

// exchange swaps two adjacent subsequences of Args. One subsequence is elements [firstNonopt,lastNonopt) which
// contains all the non-options that have been skipped so far. The other is elements [lastNonopt,optind), which
// contains all the options processed since those non-options were skipped.
//
// 'firstNonopt' and 'lastNonopt' are relocated so that they describe the new indices of the non-options in Args after
// they are moved.
func (g *Getopt) exchange() {
	bottom := g.firstNonopt
	middle := g.lastNonopt
	top := g.optind

	// Exchange the shorter segment with the far end of the longer segment. That puts the shorter segment into the
	// right place. It leaves the longer segment in the right place overall, but it consists of two parts that need
	// to be swapped next.
	for top > middle && middle > bottom {
		if top-middle > middle-bottom {
			// Bottom segment is the short one.
			length := middle - bottom

			// Swap it with the top part of the top segment.
			for i := 0; i < length; i++ {
				tem := g.Args[bottom+i]
				g.Args[bottom+i] = g.Args[top-(middle-bottom)+i]
				g.Args[top-(middle-bottom)+i] = tem
			}
			// Exclude the moved bottom segment from further swapping.
			top -= length
		} else {
			// Top segment is the short one.
			length := top - middle

			// Swap it with the bottom part of the bottom segment.
			for i := 0; i < length; i++ {
				tem := g.Args[bottom+i]
				g.Args[bottom+i] = g.Args[middle+i]
				g.Args[middle+i] = tem
			}
			// Exclude the moved top segment from further swapping.
			bottom += length
		}
	}

	// Update records for the slots the non-options now occupy.
	g.firstNonopt += (g.optind - g.lastNonopt)
	g.lastNonopt = g.optind
}

// Process the argument starting with nextChar as a long option. optind should *not* have been advanced over this
// argument.
//
// If the value returned is -1, it was not actually a long option, the state is unchanged, and the argument should be
// processed as a set of short options (this can only happen when longOnly is true). Otherwise, the option (and its
// argument, if any) have been consumed and the return value is the value to return from getoptInternalR.
func (g *Getopt) processLongOption(longOnly bool, prefix string) (*Opt, error) {
	namelen := slices.Index(g.nextChar, '=')
	if namelen == -1 {
		namelen = len(g.nextChar)
	}
	nameend := g.nextChar[namelen:]

	// First look for an exact match, counting the options as a side effect.
	targetName := string(g.nextChar[:namelen])
	optionIndex := slices.IndexFunc(g.longOptions, func(p Option) bool {
		return targetName == p.Name
	})
	var pfound *Option
	if optionIndex != -1 {
		// Exact match found.
		pfound = &g.longOptions[optionIndex]
	}

	if pfound == nil {
		// Didn't find an exact match, so look for abbreviations.
		var ambig AmbiguousOptionError

		for i, p := range g.longOptions {
			if strings.HasPrefix(p.Name, string(g.nextChar[:namelen])) {
				if pfound == nil {
					// First nonexact match found.
					optionIndex = i
					pfound = &g.longOptions[optionIndex]
					ambig.Candidates = append(ambig.Candidates, p.Name)
				} else if longOnly || pfound.HasArg != p.HasArg || pfound.Flag != p.Flag || pfound.Val != p.Val {
					// Second or later nonexact match found.
					ambig.Candidates = append(ambig.Candidates, p.Name)
				}
			}
		}

		if len(ambig.Candidates) > 1 {
			ambig.Option = string(g.nextChar)
			ambig.prefix = prefix

			g.nextChar = nil
			g.optind++
			return nil, ambig
		}
	}

	if pfound == nil {
		// Can't find it as a long option. If this is not GetoptLongOnly, or the option starts with '--' or is
		// not a valid short option, then it's an error.
		if !longOnly || g.Args[g.optind][1] == '-' || !g.shortOptions.HasOpt(g.nextChar[0]) {
			unrecog := UnrecognizedOptionError{
				Option: string(g.nextChar),
				prefix: prefix,
			}
			g.nextChar = nil
			g.optind++
			return nil, unrecog
		}

		// Otherwise interpret it as a short option.
		return nil, nil
	}

	// We have found a matching long option. Consume it.
	g.optind++
	g.nextChar = nil
	var arg *string
	if len(nameend) != 0 {
		if pfound.HasArg == NoArgument {
			return nil, ArgumentNotAllowedError{
				Option: pfound.Name,
				prefix: prefix,
			}
		}
		s := string(nameend[1:])
		arg = &s
	} else if pfound.HasArg == RequiredArgument {
		if g.optind >= len(g.Args) {
			return nil, ArgumentRequiredError{
				Option: pfound.Name,
				prefix: prefix,
			}
		}
		arg = &g.Args[g.optind]
		g.optind++
	}

	if pfound.Flag != nil {
		*pfound.Flag = pfound.Val
		return &Opt{
			LongInd: optionIndex,
			Arg:     arg,
		}, nil
	}
	return &Opt{
		C:       pfound.Val,
		LongInd: optionIndex,
		Arg:     arg,
	}, nil
}

// nonoption tests whether ARGV[optind] holds a non-option argument.
func nonoption(s string) bool {
	return !strings.HasPrefix(s, dash) || len(s) == 1
}

func (g *Getopt) getoptInternalR(longOnly bool) (*Opt, error) {
	if len(g.Args) < 1 {
		return nil, nil
	}

	if len(g.nextChar) == 0 {
		// Advance to the next ARGV-element.

		// Give FIRST_NONOPT & LAST_NONOPT rational values if OPTIND has been moved back by the user (who may
		// also have changed the arguments).
		if g.lastNonopt > g.optind {
			g.lastNonopt = g.optind
		}
		if g.firstNonopt > g.optind {
			g.firstNonopt = g.optind
		}

		if g.shortOptions.Ordering == Permute {
			// If we have just processed some options following
			// some non-options, exchange them so that the options
			// come first.
			if g.firstNonopt != g.lastNonopt && g.lastNonopt != g.optind {
				g.exchange()
			} else if g.lastNonopt != g.optind {
				g.firstNonopt = g.optind
			}

			// Skip any additional non-options and extend the range of non-options previously skipped.
			for g.optind < len(g.Args) && nonoption(g.Args[g.optind]) {
				g.optind++
			}
			g.lastNonopt = g.optind
		}

		// The special ARGV-element '--' means premature end of options. Skip it like a null option, then
		// exchange with previous non-options as if it were an option, then skip everything else like a
		// non-option.
		if g.optind != len(g.Args) && g.Args[g.optind] == argumentTerminator {
			g.optind++

			if g.firstNonopt != g.lastNonopt && g.lastNonopt != g.optind {
				g.exchange()
			} else if g.firstNonopt == g.lastNonopt {
				g.firstNonopt = g.optind
			}
			g.lastNonopt = len(g.Args)

			g.optind = len(g.Args)
		}

		// If we have done all the ARGV-elements, stop the scan and back over any non-options that we skipped
		// and permuted.
		if g.optind == len(g.Args) {
			// Set the next-arg-index to point at the non-options that we previously skipped, so the caller
			// will digest them.
			if g.firstNonopt != g.lastNonopt {
				g.optind = g.firstNonopt
			}
			return nil, nil
		}

		// If we have come to a non-option and did not permute it, either stop the scan or describe it to the
		// caller and pass it by.
		if nonoption(g.Args[g.optind]) {
			if g.shortOptions.Ordering == RequireOrder {
				return nil, nil
			}
			arg := &g.Args[g.optind]
			g.optind++
			return &Opt{
				C:   1,
				Arg: arg,
			}, nil
		}

		// We have found another option-ARGV-element. Check whether it might be a long option.
		if len(g.longOptions) > 0 {
			if []rune(g.Args[g.optind])[1] == '-' {
				// "--foo" is always a long option. The
				// special option "--" was handled above.
				g.nextChar = []rune(g.Args[g.optind])[len(argumentTerminator):]
				return g.processLongOption(longOnly, argumentTerminator)
			}

			// If longOnly and the ARGV-element has the form "-f", where f is a valid short option, don't
			// consider it an abbreviated form of a long option that starts with f. Otherwise there would be
			// no way to give the -f short option.
			//
			// On the other hand, if there's a long option "fubar" and the ARGV-element is "-fu", do
			// consider that an abbreviation of the long option, just like "--fu", and not "-f" with arg
			// "u".
			//
			// This distinction seems to be the most useful approach.
			if longOnly && (len(g.Args[g.optind]) > 1 || !g.shortOptions.HasOpt([]rune(g.Args[g.optind])[1])) {
				g.nextChar = []rune(g.Args[g.optind])[1:]
				opt, err := g.processLongOption(longOnly, dash)
				if opt != nil {
					return opt, err
				}
			}
		}

		// It is not a long option. Skip the initial punctuation.
		g.nextChar = []rune(g.Args[g.optind])[1:]
	}

	// Look at and handle the next short option-character.

	c := g.nextChar[0]
	g.nextChar = g.nextChar[1:]

	// Increment 'optind' when we start to process its last character.
	if len(g.nextChar) == 0 {
		g.optind++
	}

	if !g.shortOptions.HasOpt(c) {
		return nil, UnrecognizedOptionError{
			Option: string(c),
			prefix: dash,
		}
	}

	// Convenience. Treat POSIX -W foo same as long option --foo
	if c == 'W' && g.shortOptions.W && len(g.longOptions) > 0 {
		// This is an option that requires an argument.
		if len(g.nextChar) == 0 {
			if g.optind == len(g.Args) {
				return nil, ArgumentRequiredError{
					Option: string(c),
					prefix: dash,
				}
			}
			g.nextChar = []rune(g.Args[g.optind])
		}

		return g.processLongOption(false /* longOnly */, "-W ")
	}

	var arg *string
	switch d, _ := g.shortOptions.Opts[c]; d {
	case OptionalArgument:
		if len(g.nextChar) != 0 {
			s := string(g.nextChar)
			arg = &s
			g.optind++
		}
		g.nextChar = nil
	case RequiredArgument:
		if len(g.nextChar) != 0 {
			s := string(g.nextChar)
			arg = &s
			// We've ended this ARGV-element by taking the rest as an arg. We must advance to the next
			// element now.
			g.optind++
		} else if g.optind == len(g.Args) {
			return nil, ArgumentRequiredError{
				Option: string(c),
				prefix: dash,
			}
		} else {
			// We already incremented 'optind' once; increment it again when taking next ARGV-elt as
			// argument.
			arg = &g.Args[g.optind]
			g.optind++
		}
		g.nextChar = nil
	}
	return &Opt{
		C:   c,
		Arg: arg,
	}, nil
}

func (g *Getopt) getoptInternal(longOnly bool) (*Opt, error) {
	return g.getoptInternalR(longOnly)
}
