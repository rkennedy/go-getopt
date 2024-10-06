package getopt

import (
	"strings"
)

// optinfo holds the result of parsing the short option specs. Fields are exported so they can be tested with
// gomega/gstruct, but the type itself is not exported.
type optinfo struct {
	Ordering Ordering
	W        bool
	Opts     map[rune]ArgumentDisposition
}

func (inf *optinfo) HasOpt(c rune) bool {
	_, ok := inf.Opts[c]
	return ok
}

func parseShortOptionSpec(options string) optinfo {
	const (
		inorderPrefix = "-"
		posixPrefix   = "+"
	)
	var result optinfo
	switch {
	case strings.HasPrefix(options, inorderPrefix):
		result.Ordering = ReturnInOrder
	case strings.HasPrefix(options, posixPrefix):
		result.Ordering = RequireOrder
	default:
		result.Ordering = Permute
	}
	if strings.HasPrefix(options, posixPrefix) || strings.HasPrefix(options, inorderPrefix) {
		options = options[1:]
	}
	if strings.HasPrefix(options, ":") {
		// Here we would mark to suppress printing errors, but we just always do that.
		options = options[1:]
	}
	result.Opts = map[rune]ArgumentDisposition{}
	optrunes := []rune(options)
	for i := 0; i < len(optrunes); {
		c := optrunes[i]
		result.Opts[c] = NoArgument
		i++
		if c == 'W' && i < len(optrunes) && optrunes[i] == ';' {
			result.W = true
			i++
		} else {
			if i < len(optrunes) && optrunes[i] == ':' {
				result.Opts[c] = RequiredArgument
				i++
			}
			if i < len(optrunes) && optrunes[i] == ':' {
				result.Opts[c] = OptionalArgument
				i++
			}
		}
	}
	return result
}
