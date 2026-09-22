package help

import (
	"strings"

	"sweetkennedy.net/getopt"
)

// ShortOptions returns a string suitable for use with [getopt.New] corresponding to the options described in up.
// Entries with no Short value set are omitted. Entries with a Var value will include a colon to indicate a required
// argument, or two colons for an optional argument if the Var entry ends with a question mark.
func (up *UsageParams) ShortOptions() string {
	result := ""
	for _, entry := range up.Entries {
		if entry.Short != 0 {
			result += string(entry.Short)
			if entry.Var != "" {
				result += ":"
				if strings.HasSuffix(entry.Var, "?") {
					result += ":"
				}
			}
		}
	}
	return result
}

// LongOptions returns a list of values suitable for use with [getopt.NewLong] corresponding to options described in up.
// Entries with no Long value set are omitted. Entries with a Var value will have HasArg set to RequiredArgument or
// OptionalArgument. Entries will have Val set to the Short value (which might be 0).
func (up *UsageParams) LongOptions() []getopt.Option {
	var result []getopt.Option
	for _, entry := range up.Entries {
		if entry.Long != "" {
			opt := getopt.Option{
				Name: entry.Long,
				Val:  entry.Short,
			}
			if strings.HasSuffix(entry.Var, "?") {
				opt.HasArg = getopt.OptionalArgument
			} else if entry.Var != "" {
				opt.HasArg = getopt.RequiredArgument
			}
			result = append(result, opt)
		}
	}
	return result
}
