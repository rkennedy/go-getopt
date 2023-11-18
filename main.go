// Package getopt provides GNU-style argument parsing for both short and long
// arguments.
package getopt

import (
	"fmt"
	"os"
	"strings"
)

// ArgumentDisposition is an enum specifying whether an option expects to be
// followed by an argument.
type ArgumentDisposition int

// These control which options require arguments.
const (
	NoArgument       ArgumentDisposition = iota // The option accepts no argument.
	RequiredArgument                            // The option requires an argument.
	OptionalArgument                            // The option's argument is optional.
)

// Option defines long options recognized by GetoptLong.
type Option struct {
	// name of long option
	Name string
	// one of NoArgument, RequiredArgument, and OptionalArgument:
	// whether option takes an argument
	HasArg ArgumentDisposition
	// if not nil, set *Flag to Val when option found
	Flag *int
	// if Flag not nil, value to set *Flag to; else return value
	Val int
}

type optinfo struct {
	scanningMode           scanningMode
	w                      bool
	longOnly               bool
	opts                   map[rune]ArgumentDisposition
}

// Getopt is an option parser.
type Getopt struct {
	Args []string // Args holds a copy of the argument list. It gets permuted during parsing.
	optinfo
	longOptions []Option

	Optarg   *string // Optarg points at the argument associated with the most recently found option.
	Opterr   bool    // Opterr is ignored. Handle errors returned in the Go way.
	Optind   int     // Optind is the index into parent argv vector.
	Optopt   rune    // Optopt is a character checked for validity.
	Optreset bool    // Optreset resets getopt.

	place string // option letter processing

	// XXX: set Optreset to true rather than these two
	nonoptStart int // first non option argument (for permute)
	nonoptEnd   int // first option after non options (for permute)
}

// New returns an initialized Getopt parser with no long options. The parser
// always works in Posixly correct mode.
func New(args []string, options string) *Getopt {
	result := &Getopt{
		Args:        args,
		optinfo:     parseShortOptionSpec(options),
		Optind:      1,
		Optopt:      '?',
		nonoptStart: -1,
		nonoptEnd:   -1,
	}
	if !strings.HasPrefix(options, posixPrefix) && !strings.HasPrefix(options, inorderPrefix) {
		// We don't pass FLAG_PERMUTE to getoptInternal() since
		// the BSD getopt(3) (unlike GNU) has never done this.

		// Furthermore, since many privileged programs call getopt()
		// before dropping privileges it makes sense to keep things
		// as simple (and bug-free) as possible.
		result.optinfo.scanningMode = posixlyCorrect
	}
	return result
}

// NewLong returns an initialized Getopt parser with long options.
func NewLong(args []string, options string, longOptions []Option) *Getopt {
	return &Getopt{
		Args:        args,
		optinfo:     parseShortOptionSpec(options),
		longOptions: longOptions,
		Optind:      1,
		Optopt:      '?',
		nonoptStart: -1,
		nonoptEnd:   -1,
	}
}

// NewLongOnly returns an initialized Getopt parser with long options that may
// be recognized with only a '-' prefix instead of '--'. If there are no
// candidate long options, then short options will be considered as well.
func NewLongOnly(args []string, options string, longOptions []Option) *Getopt {
	result := &Getopt{
		Args:        args,
		optinfo:     parseShortOptionSpec(options),
		longOptions: longOptions,
		Optind:      1,
		Optopt:      '?',
		nonoptStart: -1,
		nonoptEnd:   -1,
	}
	result.optinfo.longOnly = true
	return result
}

type scanningMode int

const (
	defaultPermute scanningMode = iota // permute non-options to the end of argv.
	posixlyCorrect                     // stop scanning at the first non-option.
	argsInOrder                        // treat non-options as args to option 1
)

// return values
const (
	BADCH   rune = '?'
	INORDER rune = 1
	BADARG  rune = ':'
)

// RecArgChar is the error returned when a short option lacks a required argument.
type RecArgChar rune

func (e RecArgChar) Error() string {
	return fmt.Sprintf("option requires an argument -- %c", e)
}

// RecArgString is the error returned when a long option lacks a required argument.
type RecArgString string

func (e RecArgString) Error() string {
	return fmt.Sprintf("option requires an argument -- %s", string(e))
}

// Ambig is the error returned when an ambiguous long option is detected.
type Ambig string

func (e Ambig) Error() string {
	return fmt.Sprintf("ambiguous option -- %s", string(e))
}

// Noarg is the error returned when an argument is given for a long option that doesn't accept one.
type Noarg string

func (e Noarg) Error() string {
	return fmt.Sprintf("option doesn't take an argument -- %s", string(e))
}

// IllOptChar is the error returned for an unknown short option.
type IllOptChar rune

func (e IllOptChar) Error() string {
	return fmt.Sprintf("unknown option -- %c", e)
}

// IllOptString is the error returned for an unknown long option.
type IllOptString string

func (e IllOptString) Error() string {
	return fmt.Sprintf("unknown option -- %s", string(e))
}

// Exchange the block from nonoptStart to nonoptEnd with the block
// from nonoptEnd to optEnd (keeping the same order of arguments
// in each block).
func permuteArgs(paNonoptStart int, paNonoptEnd int, optEnd int, nargv []string) {
	newArgs := nargv[0:paNonoptStart]
	newArgs = append(newArgs, nargv[paNonoptEnd:optEnd]...)
	newArgs = append(newArgs, nargv[paNonoptStart:paNonoptEnd]...)
	newArgs = append(newArgs, nargv[optEnd:]...)
	copy(nargv, newArgs)
}

const (
	posixPrefix   = "+"
	inorderPrefix = "-"
	dash          = "-"
)

func parseShortOptionSpec(options string) optinfo {
	var result optinfo
	_, ok := os.LookupEnv("POSIXLY_CORRECT")
	if ok || strings.HasPrefix(options, posixPrefix) {
		result.scanningMode = posixlyCorrect
	} else if strings.HasPrefix(options, inorderPrefix) {
		result.scanningMode = argsInOrder
	}
	if strings.HasPrefix(options, posixPrefix) || strings.HasPrefix(options, inorderPrefix) {
		options = options[1:]
	}
	if strings.HasPrefix(options, ":") {
		// Here we would mark to suppress printing errors, but we just always do that.
		options = options[1:]
	}
	result.opts = map[rune]ArgumentDisposition{}
	optrunes := []rune(options)
	for i := 0; i < len(optrunes); {
		c := optrunes[i]
		result.opts[c] = NoArgument
		i++
		if c == 'W' && i < len(optrunes) && optrunes[i] == ';' {
			result.w = true
			i++
		} else {
			if i < len(optrunes) && optrunes[i] == ':' {
				result.opts[c] = RequiredArgument
				i++
			}
			if i < len(optrunes) && optrunes[i] == ':' {
				result.opts[c] = OptionalArgument
				i++
			}
		}
	}
	return result
}

func (inf *optinfo) HasOpt(c rune) bool {
	_, ok := inf.opts[c]
	return ok
}

/*
 * parseLongOptions --
 *	Parse long options in argc/argv argument vector.
 * Returns -1 if shortToo is set and the option does not match longOptions.
 */
func (g *Getopt) parseLongOptions(shortToo bool) (
	ch rune, longIndex int, err error) {
	var currentArgvLen int
	currentArgv := g.place
	match := -1
	g.Optind++
	hasEqual := strings.IndexRune(currentArgv, '=')
	if hasEqual >= 0 {
		// argument found (--option=arg)
		currentArgvLen = hasEqual
		hasEqual++
	} else {
		currentArgvLen = len(currentArgv)
	}
	for i := range g.longOptions {
		// find matching long option
		if !strings.HasPrefix(g.longOptions[i].Name, currentArgv[:currentArgvLen]) {
			// option i is definitely not an abbreviation of currentArgv
			continue
		}
		if len(g.longOptions[i].Name) == currentArgvLen {
			// exact match
			match = i
			break
		}
		// If this is a known short option, don't allow
		// a partial match of a single character.
		if shortToo && currentArgvLen == 1 {
			continue
		}
		if match != -1 {
			// ambiguous abbreviation
			err := Ambig(currentArgv[:currentArgvLen])
			g.Optopt = 0
			return BADCH, 0, err
		}
		// partial match
		match = i
	}
	if match == -1 {
		// unknown option
		if shortToo {
			g.Optind--
			return -1, 0, nil
		}
		err := IllOptString(currentArgv)
		g.Optopt = 0
		return BADCH, 0, err
	}
	// option found
	if g.longOptions[match].HasArg == NoArgument && hasEqual >= 0 {
		err := Noarg(currentArgv[:currentArgvLen])
		// XXX: GNU sets Optopt to Val regardless of Flag
		if g.longOptions[match].Flag == nil {
			g.Optopt = rune(g.longOptions[match].Val)
		} else {
			g.Optopt = 0
		}
		return BADARG, 0, err
	}
	if g.longOptions[match].HasArg == RequiredArgument || g.longOptions[match].HasArg == OptionalArgument {
		if hasEqual >= 0 {
			argpart := currentArgv[hasEqual:]
			g.Optarg = &argpart
		} else if g.longOptions[match].HasArg == RequiredArgument {
			// optional argument doesn't use next element from Args.
			if g.Optind >= len(g.Args) {
				// Missing argument. Handled below.
				g.Optarg = nil
			} else {
				g.Optarg = &g.Args[g.Optind]
			}
			g.Optind++
		}
	}
	if g.longOptions[match].HasArg == RequiredArgument && g.Optarg == nil {
		// Missing argument
		err := RecArgString(currentArgv)
		// XXX: GNU sets Optopt to Val regardless of Flag
		if g.longOptions[match].Flag == nil {
			g.Optopt = rune(g.longOptions[match].Val)
		} else {
			g.Optopt = 0
		}
		g.Optind--
		return BADARG, 0, err
	}
	if g.longOptions[match].Flag != nil {
		*g.longOptions[match].Flag = g.longOptions[match].Val
		return 0, match, nil
	}
	return rune(g.longOptions[match].Val), match, nil
}

func (g *Getopt) getoptInternal() (
	ch rune, longIndex int, err error) {
	nargc := len(g.Args)

	g.Optarg = nil
	if g.Optreset {
		g.nonoptStart = -1
		g.nonoptEnd = -1
	}
start:
	for {
		if g.Optreset || g.place == "" { // update scanning pointer
			g.Optreset = false
			if g.Optind >= nargc { // end of argument vector
				g.place = ""
				if g.nonoptEnd != -1 {
					// Do permutation to put skipped non-options at the end.
					permuteArgs(g.nonoptStart, g.nonoptEnd, g.Optind, g.Args)
					g.Optind -= g.nonoptEnd - g.nonoptStart
				} else if g.nonoptStart != -1 {
					// We skipped some non-options. Set Optind
					// to the first of them.
					g.Optind = g.nonoptStart
				}
				g.nonoptStart = -1
				g.nonoptEnd = -1
				return -1, 0, nil
			}
			g.place = g.Args[g.Optind]
			if !strings.HasPrefix(g.place, dash) || (len(g.place) == 1 && !g.optinfo.HasOpt('-')) {
				g.place = "" // found non-option
				if g.optinfo.scanningMode == argsInOrder {
					// GNU extension:
					// return non-option as argument to option 1
					g.Optarg = &g.Args[g.Optind]
					g.Optind++
					return INORDER, 0, nil
				}
				if g.optinfo.scanningMode == posixlyCorrect {
					// If no permutation wanted, stop parsing
					// at first non-option.
					return -1, 0, nil
				}
				// do permutation
				if g.nonoptStart == -1 {
					g.nonoptStart = g.Optind
				} else if g.nonoptEnd != -1 {
					permuteArgs(g.nonoptStart, g.nonoptEnd, g.Optind, g.Args)
					g.nonoptStart = g.Optind - (g.nonoptEnd - g.nonoptStart)
					g.nonoptEnd = -1
				}
				g.Optind++
				// process next argument
				continue start
			}
			if g.nonoptStart != -1 && g.nonoptEnd == -1 {
				g.nonoptEnd = g.Optind
			}
			// If we have "-" do nothing, if "--" we are done.
			if len(g.place) > 1 {
				g.place = g.place[1:]
				if g.place == dash {
					g.Optind++
					g.place = ""
					// We found an option (--), so if we skipped
					// non-options, we have to permute.
					if g.nonoptEnd != -1 {
						permuteArgs(g.nonoptStart, g.nonoptEnd, g.Optind, g.Args)
						g.Optind -= g.nonoptEnd - g.nonoptStart
					}
					g.nonoptStart = -1
					g.nonoptEnd = -1
					return -1, 0, nil
				}
			}
		}
		break
	}
	/*
	 * Check long options if:
	 *  1) we were passed some
	 *  2) the arg is not just "-"
	 *  3) either the arg starts with -- we are getoptLongOnly()
	 */
	if len(g.longOptions) > 0 && g.place != g.Args[g.Optind] && (strings.HasPrefix(g.place, dash) || g.optinfo.longOnly) { //revive:disable:line-length-limit
		shortToo := false
		if strings.HasPrefix(g.place, dash) {
			g.place = g.place[1:] // --foo long option
		} else if !strings.HasPrefix(g.place, ":") && g.optinfo.HasOpt([]rune(g.place)[0]) {
			shortToo = true // could be short option too
		}
		var optchar rune
		optchar, longIndex, err = g.parseLongOptions(shortToo)
		if optchar != -1 {
			g.place = ""
			return optchar, longIndex, err
		}
	}
	optchar := []rune(g.place)[0]
	g.place = g.place[1:]

	// If the user specified "-" and '-' isn't listed in
	// options, return -1 (non-option) as per POSIX.
	// Otherwise, it is an unknown option character (or ':').
	if optchar == '-' && g.place == "" {
		return -1, 0, nil
	} else if !g.optinfo.HasOpt(optchar) {
		if g.place == "" {
			g.Optind++
		}
		err := IllOptChar(optchar)
		g.Optopt = optchar
		return BADCH, 0, err
	}
	if len(g.longOptions) > 0 && optchar == 'W' && g.optinfo.w {
		// -W long-option
		if g.place != "" { // no space
			//revive:disable:empty-block NOTHING
		} else {
			g.Optind++
			if g.Optind >= nargc { // no arg
				g.place = ""
				err := RecArgChar(optchar)
				g.Optopt = optchar
				return BADARG, 0, err
			}
			// white space
			g.place = g.Args[g.Optind]
		}
		optchar, match, err := g.parseLongOptions(false)
		g.place = ""
		return optchar, match, err
	}
	if g.optinfo.opts[optchar] == NoArgument { // doesn't take argument
		if g.place == "" {
			g.Optind++
		}
	} else { // takes (optional) argument
		g.Optarg = nil
		if g.place != "" { // no white space
			g.Optarg = &g.place
			// XXX: disable test for :: if PC? (GNU doesn't)
		} else if g.optinfo.opts[optchar] == RequiredArgument { // arg not optional
			g.Optind++
			if g.Optind >= nargc { // no arg
				g.place = ""
				err := RecArgChar(optchar)
				g.Optopt = optchar
				return BADARG, 0, err
			}
			g.Optarg = &g.Args[g.Optind]
		} else if g.optinfo.scanningMode != defaultPermute {
			// If permutation is disabled, we can accept an
			// optional arg separated by whitespace so long
			// as it does not start with a dash (-).
			if g.Optind+1 < len(g.Args) && !strings.HasPrefix(g.Args[g.Optind+1], dash) {
				g.Optind++
				g.Optarg = &g.Args[g.Optind]
			}
		}
		g.place = ""
		g.Optind++
	}
	// dump back option letter
	return optchar, -1, nil
}

// Loop parses argc/argv argument vectors like GNU's getopt.
func (g *Getopt) Loop() (rune, error) {
	ch, _, err := g.getoptInternal()
	return ch, err
}

// LoopLong parses options like GNU's getopt_long.
func (g *Getopt) LoopLong() (rune, int, error) {
	return g.getoptInternal()
}
