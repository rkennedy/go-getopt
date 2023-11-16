package main

import (
	"fmt"
	"os"
	"strings"
)

type ArgumentDisposition int

const (
	NoArgument ArgumentDisposition = iota
	RequiredArgument
	OptionalArgument
)

type Option struct {
	// name of long option
	Name string
	// one of no_argument, required_argument, and optional_argument:
	// whether option takes an argument
	HasArg ArgumentDisposition
	// if not nil, set *Flag to Val when option found
	Flag *int
	// if Flag not nil, value to set *Flag to; else return value
	Val int
}

var (
	Optarg   *string        /* argument associated with option */
	Opterr   bool    = true /* if error message should be printed */
	Optind   int     = 1    /* index into parent argv vector */
	Optopt   rune    = '?'  /* character checked for validity */
	Optreset bool           /* reset getopt */
)

type ScanningMode int

const (
	DefaultPermute ScanningMode = iota /* permute non-options to the end of argv */
	PosixlyCorrect
	ArgsInOrder /* treat non-options as args to option "-1" */
)

/* return values */
const (
	BADCH   rune = '?'
	INORDER rune = 1
)

func BADARG(options *optinfo) rune {
	if options.suppressPrintingErrors {
		return ':'
	}
	return '?'
}

var place string // option letter processing

/* XXX: set Optreset to true rather than these two */
var (
	nonopt_start int = -1 // first non option argument (for permute)
	nonopt_end   int = -1 // first option after non options (for permute)
)

/* Error messages */
type RecArgChar rune

func (e RecArgChar) Error() string {
	return fmt.Sprintf("option requires an argument -- %c", e)
}

type RecArgString string

func (e RecArgString) Error() string {
	return fmt.Sprintf("option requires an argument -- %s", string(e))
}

type Ambig struct {
	Length  int
	Message string
}

func (e Ambig) Error() string {
	return fmt.Sprintf("ambiguous option -- %.*s", e.Length, e.Message)
}

type Noarg struct {
	Length  int
	Message string
}

func (e Noarg) Error() string {
	return fmt.Sprintf("option doesn't take an argument -- %.*s", e.Length, e.Message)
}

type IllOptChar rune

func (e IllOptChar) Error() string {
	return fmt.Sprintf("unknown option -- %c", e)
}

type IllOptString string

func (e IllOptString) Error() string {
	return fmt.Sprintf("unknown option -- %s", string(e))
}

// Exchange the block from nonopt_start to nonopt_end with the block
// from nonopt_end to opt_end (keeping the same order of arguments
// in each block).
func permuteArgs(panonopt_start int, panonopt_end int, opt_end int, nargv []string) {
	newArgs := nargv[0:panonopt_start]
	newArgs = append(newArgs, nargv[panonopt_end:opt_end]...)
	newArgs = append(newArgs, nargv[panonopt_start:panonopt_end]...)
	newArgs = append(newArgs, nargv[opt_end:]...)
	copy(nargv, newArgs)
}

type optinfo struct {
	scanningMode           ScanningMode
	suppressPrintingErrors bool
	w                      bool
	longOnly               bool
	opts                   map[rune]ArgumentDisposition
}

func parseShortOptionSpec(options string) optinfo {
	var result optinfo
	_, ok := os.LookupEnv("POSIXLY_CORRECT")
	if ok || strings.HasPrefix(options, "+") {
		result.scanningMode = PosixlyCorrect
	} else if strings.HasPrefix(options, "-") {
		result.scanningMode = ArgsInOrder
	}
	if strings.HasPrefix(options, "+") || strings.HasPrefix(options, "-") {
		options = options[1:]
	}
	if strings.HasPrefix(options, ":") {
		result.suppressPrintingErrors = true
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
 * parse_long_options --
 *	Parse long options in argc/argv argument vector.
 * Returns -1 if short_too is set and the option does not match long_options.
 */
func parseLongOptions(nargv []string, options *optinfo, long_options []Option, idx *int, short_too bool) (rune, error) {
	var current_argv_len int
	currentArgv := place
	match := -1
	Optind++
	hasEqual := strings.IndexRune(currentArgv, '=')
	if hasEqual >= 0 {
		/* argument found (--option=arg) */
		current_argv_len = hasEqual
		hasEqual++
	} else {
		current_argv_len = len(currentArgv)
	}
	for i, _ := range long_options {
		/* find matching long option */
		if currentArgv[0:current_argv_len] != long_options[i].Name[0:current_argv_len] {
			continue
		}
		if len(long_options[i].Name) == current_argv_len {
			/* exact match */
			match = i
			break
		}
		// If this is a known short option, don't allow
		// a partial match of a single character.
		if short_too && current_argv_len == 1 {
			continue
		}
		if match == -1 { // partial match
			match = i
		} else {
			// ambiguous abbreviation
			err := Ambig{Length: current_argv_len, Message: currentArgv}
			if Opterr && !options.suppressPrintingErrors {
				fmt.Fprintln(os.Stderr, err.Error())
			}
			Optopt = 0
			return BADCH, err
		}
	}
	if match != -1 { /* option found */
		if long_options[match].HasArg == NoArgument && hasEqual >= 0 {
			err := Noarg{Length: current_argv_len, Message: currentArgv}
			if Opterr && !options.suppressPrintingErrors {
				fmt.Fprintln(os.Stderr, err.Error())
			}
			// XXX: GNU sets Optopt to Val regardless of Flag
			if long_options[match].Flag == nil {
				Optopt = rune(long_options[match].Val)
			} else {
				Optopt = 0
			}
			return BADARG(options), err
		}
		if long_options[match].HasArg == RequiredArgument || long_options[match].HasArg == OptionalArgument {
			if hasEqual >= 0 {
				argpart := currentArgv[hasEqual:]
				Optarg = &argpart
			} else if long_options[match].HasArg == RequiredArgument {
				/*
				 * optional argument doesn't use next nargv
				 */
				Optarg = &nargv[Optind]
				Optind++
			}
		}
		if long_options[match].HasArg == RequiredArgument && Optarg == nil {
			/*
			 * Missing argument; leading ':' indicates no error
			 * should be generated.
			 */
			err := RecArgString(currentArgv)
			if Opterr && !options.suppressPrintingErrors {
				fmt.Fprintln(os.Stderr, err.Error())
			}
			/*
			 * XXX: GNU sets Optopt to Val regardless of Flag
			 */
			if long_options[match].Flag == nil {
				Optopt = rune(long_options[match].Val)
			} else {
				Optopt = 0
			}
			Optind--
			return BADARG(options), err
		}
	} else { /* unknown option */
		if short_too {
			Optind--
			return -1, nil
		}
		err := IllOptString(currentArgv)
		if Opterr && !options.suppressPrintingErrors {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		Optopt = 0
		return BADCH, err
	}
	if idx != nil {
		*idx = match
	}
	if long_options[match].Flag != nil {
		*long_options[match].Flag = long_options[match].Val
		return 0, nil
	}
	return rune(long_options[match].Val), nil
}

type permuteMode int

const (
	PosixPermute permuteMode = iota
	Permute
	AllArgs
)

/*
 * Getopt parses argc/argv argument vector.  Called by user level routines.
 */
func getopt_internal(nargv []string, info *optinfo, long_options []Option, idx *int) (rune, error) {
	nargc := len(nargv)

	Optarg = nil
	if Optreset {
		nonopt_start = -1
		nonopt_end = -1
	}
start:
	for {
		if Optreset || place != "" { /* update scanning pointer */
			Optreset = false
			if Optind >= nargc { /* end of argument vector */
				place = ""
				if nonopt_end != -1 {
					/* do permutation, if we have to */
					permuteArgs(nonopt_start, nonopt_end, Optind, nargv)
					Optind -= nonopt_end - nonopt_start
				} else if nonopt_start != -1 {
					/*
					 * If we skipped non-options, set Optind
					 * to the first of them.
					 */
					Optind = nonopt_start
				}
				nonopt_start = -1
				nonopt_end = -1
				return -1, nil
			}
			place = nargv[Optind]
			if !strings.HasPrefix(place, "-") || (len(place) == 1 && !info.HasOpt('-')) {
				place = "" /* found non-option */
				if info.scanningMode == ArgsInOrder {
					// GNU extension:
					// return non-option as argument to option 1
					Optarg = &nargv[Optind]
					Optind++
					return INORDER, nil
				}
				if info.scanningMode == PosixlyCorrect {
					// If no permutation wanted, stop parsing
					// at first non-option.
					return -1, nil
				}
				/* do permutation */
				if nonopt_start == -1 {
					nonopt_start = Optind
				} else if nonopt_end != -1 {
					permuteArgs(nonopt_start, nonopt_end, Optind, nargv)
					nonopt_start = Optind - (nonopt_end - nonopt_start)
					nonopt_end = -1
				}
				Optind++
				/* process next argument */
				continue start
			}
			if nonopt_start != -1 && nonopt_end == -1 {
				nonopt_end = Optind
			}
			// If we have "-" do nothing, if "--" we are done.
			if len(place) > 1 {
				place = place[1:]
				if place == "-" {
					Optind++
					place = ""
					// We found an option (--), so if we skipped
					// non-options, we have to permute.
					if nonopt_end != -1 {
						permuteArgs(nonopt_start, nonopt_end, Optind, nargv)
						Optind -= nonopt_end - nonopt_start
					}
					nonopt_start = -1
					nonopt_end = -1
					return -1, nil
				}
			}
		}
		break
	}
	/*
	 * Check long options if:
	 *  1) we were passed some
	 *  2) the arg is not just "-"
	 *  3) either the arg starts with -- we are getopt_long_only()
	 */
	var optchar rune
	if len(long_options) > 0 && place != nargv[Optind] && (strings.HasPrefix(place, "-") || info.longOnly) {
		short_too := false
		if strings.HasPrefix(place, "-") {
			place = place[1:] /* --foo long option */
		} else if !strings.HasPrefix(place, ":") && info.HasOpt([]rune(place)[0]) {
			short_too = true /* could be short option too */
		}
		var err error
		optchar, err = parseLongOptions(nargv, info, long_options, idx, short_too)
		if optchar != -1 {
			place = ""
			return optchar, err
		}
	}
	optchar = []rune(place)[0]
	place = place[1:]

	if optchar == ':' || (optchar == '-' && place == "") || !info.HasOpt(optchar) {
		/*
		 * If the user specified "-" and  '-' isn't listed in
		 * options, return -1 (non-option) as per POSIX.
		 * Otherwise, it is an unknown option character (or ':').
		 */
		if optchar == '-' && place == "" {
			return -1, nil
		}
		if place == "" {
			Optind++
		}
		err := IllOptChar(optchar)
		if Opterr && !info.suppressPrintingErrors {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		Optopt = optchar
		return BADCH, err
	}
	if len(long_options) > 0 && optchar == 'W' && info.w {
		/* -W long-option */
		if place != "" { /* no space */
			/* NOTHING */
		} else {
			Optind++
			if Optind >= nargc { /* no arg */
				place = ""
				err := RecArgChar(optchar)
				if Opterr && !info.suppressPrintingErrors {
					fmt.Fprintln(os.Stderr, err.Error())
				}
				Optopt = optchar
				return BADARG(info), err
			} else { /* white space */
				place = nargv[Optind]
			}
		}
		optchar, err := parseLongOptions(nargv, info, long_options, idx, false)
		place = ""
		return optchar, err
	}
	if info.opts[optchar] == NoArgument { /* doesn't take argument */
		if place == "" {
			Optind++
		}
	} else { /* takes (optional) argument */
		Optarg = nil
		if place != "" { /* no white space */
			Optarg = &place
			/* XXX: disable test for :: if PC? (GNU doesn't) */
		} else if info.opts[optchar] == RequiredArgument { /* arg not optional */
			Optind++
			if Optind >= nargc { /* no arg */
				place = ""
				err := RecArgChar(optchar)
				if Opterr && !info.suppressPrintingErrors {
					fmt.Fprintln(os.Stderr, err.Error())
				}
				Optopt = optchar
				return BADARG(info), err
			} else {
				Optarg = &nargv[Optind]
			}
		} else if info.scanningMode != DefaultPermute {
			/*
			 * If permutation is disabled, we can accept an
			 * optional arg separated by whitespace so long
			 * as it does not start with a dash (-).
			 */
			if Optind+1 < len(nargv) && !strings.HasPrefix(nargv[Optind+1], "-") {
				Optind++
				Optarg = &nargv[Optind]
			}
		}
		place = ""
		Optind++
	}
	/* dump back option letter */
	return optchar, nil
}

/*
 * getopt --
 *	Parse argc/argv argument vector.
 *
 * [eventually this will replace the BSD getopt]
 */
func getopt(nargv []string, options string) (rune, error) {
	info := parseShortOptionSpec(options)
	if !strings.HasPrefix(options, "+") && !strings.HasPrefix(options, "-") {
		/*
		 * We don't pass FLAG_PERMUTE to getopt_internal() since
		 * the BSD getopt(3) (unlike GNU) has never done this.
		 *
		 * Furthermore, since many privileged programs call getopt()
		 * before dropping privileges it makes sense to keep things
		 * as simple (and bug-free) as possible.
		 */
		info.scanningMode = PosixlyCorrect
	}
	return getopt_internal(nargv, &info, nil, nil)
}

/*
 * getopt_long --
 *	Parse argc/argv argument vector.
 */
func getopt_long(nargv []string, options string, long_options []Option, idx *int) (rune, error) {
	info := parseShortOptionSpec(options)
	return getopt_internal(nargv, &info, long_options, idx)
}

/*
 * getopt_long_only --
 *	Parse argc/argv argument vector.
 */
func getopt_long_only(nargv []string, options string, long_options []Option, idx *int) (rune, error) {
	info := parseShortOptionSpec(options)
	info.longOnly = true
	return getopt_internal(nargv, &info, long_options, idx)
}
