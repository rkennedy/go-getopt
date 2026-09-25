// Package help formats help and usage information.
package help

import (
	"fmt"
	"io"
	"strings"

	"github.com/rivo/uniseg"
)

// Entry represents a help entry. It has an optional short name, and optional long name, an optional metavariable, and a
// help description. If the metavariable name ends with '?', then it is formatted as an optional argument instead of
// required.
type Entry struct {
	Short rune
	Long  string
	Var   string
	Help  string
}

// UsageParams control how [Usage] formats output.
type UsageParams struct {
	Usage       string
	Description string
	Entries     []Entry
	Trailer     string
}

// Usage prints usage information to the given output. Output is wrapped at width columns, unless the width is
// non-positive, in which case the natural wrapping of the terminal (if any) will apply implicitly. The Usage string is
// printed first, prefixed by "Usage:", and then the Description field is printed. The Entries are printed in order,
// with the help aligned with a hanging indent. Finally, the Trailer field is printed. Any empty fields are omitted.
func Usage(out io.Writer, params *UsageParams, width int) {
	if params.Usage != "" {
		prefix := "Usage: "
		printWrappedWithPrefixes(out, prefix, strings.Repeat(" ", len(prefix)), params.Usage, width)
	}
	if params.Description != "" {
		printWrappedWithPrefixes(out, "", "", params.Description, width)
	}

	for _, entry := range params.Entries {
		names := formatNames(entry.Short, entry.Long, entry.Var)
		printEntry(out, names, entry.Help, width)
	}

	if params.Trailer != "" {
		printWrappedWithPrefixes(out, "", "", params.Trailer, width)
	}
}

// formatNames arranges the short and long names of the help entry, followed by a metavariable to indicate the argument,
// if any. The option names are immune to the terminal width limit. Format is inspired by that of `ls --help`.
func formatNames(short rune, long string, meta string) string {
	var result string
	if short != 0 {
		result = fmt.Sprintf("  -%c", short)
	} else {
		result = "    "
	}
	if long != "" {
		if short != 0 {
			result += ", "
		} else {
			result += "  "
		}
		result += fmt.Sprintf("--%s", long)
	}
	if meta != "" {
		name, cut := strings.CutSuffix(meta, "?")
		if cut {
			result += fmt.Sprintf("[=%s]", name)
		} else {
			result += fmt.Sprintf(" %s", name)
		}
	}
	return result
}

// HelpStart is the column where help text will be aligned following an option name.
const HelpStart = 29

// printEntry prints the given option names followed by the help text, wrapped to `width` columns. The help text is
// aligned to column 30 (i.e., there are 29 columns of between the left edge and the start of the text, because we start
// counting columns at 1). There is a minimum of two spaces between the end of the option text and the start of the help
// text. If the options encroach into the help area, then the start of the help text will be pushed right by up to one
// space. Any more than that, and the help text will start on the next line instead, still aligned to column 30.
// Newlines in the help text are printed verbatim.
func printEntry(out io.Writer, names, help string, width int) {
	// Minimum space between option names and help text to make sure it's clear they're separate things.
	const minTextPadding = 2

	if help == "" {
		_, _ = fmt.Fprintln(out, names)
		return
	}

	indent := strings.Repeat(" ", HelpStart)

	// If the option names are "too wide," then we'll start the help text on the next line. But if they're just a
	// "little" too wide -- within one space of the limit -- then we'll just shove the help text over one space to keep
	// it starting on the same line. (Still need to honor the overall maximum width, though.)
	padding := HelpStart - uniseg.StringWidth(names)
	if padding <= 0 {
		// The option names have used up all of or more than the column allocated for them, abutting or overlapping with
		// the help-text column. Print just the options, no padding, and start the help text on the next line.
		_, _ = fmt.Fprintln(out, names)
		printWrappedWithPrefixes(out, indent, indent, help, width)
		return
	}
	padding = max(padding, minTextPadding)

	printWrappedWithPrefixes(
		out,
		names+strings.Repeat(" ", padding),
		indent,
		help,
		width,
	)
}

// printWrappedWithPrefixes prints wrapped text with different prefixes for the first and continuation lines.
//
//revive:disable-next-line:argument-limit
func printWrappedWithPrefixes(out io.Writer, firstPrefix, continuationPrefix, text string, width int) {
	firstWidth, continuationWidth := -1, -1
	if width > 0 {
		continuationWidth = max(1, width-uniseg.StringWidth(continuationPrefix))
		firstWidth = max(width-uniseg.StringWidth(firstPrefix), continuationWidth)
	}

	lines := wrapWords(text, firstWidth, continuationWidth)
	if len(lines) == 0 {
		_, _ = fmt.Fprintln(out, firstPrefix)
		return
	}
	_, _ = fmt.Fprintf(out, "%s%s\n", firstPrefix, lines[0])
	for _, line := range lines[1:] {
		_, _ = fmt.Fprintf(out, "%s%s\n", continuationPrefix, line)
	}
}

func wrapWords(text string, firstWidth, continuationWidth int) []string {
	var lines []string
	for rawLine := range strings.SplitSeq(text, "\n") {
		lines = append(lines, wrapOneLine(rawLine, firstWidth, continuationWidth)...)
		firstWidth = continuationWidth
	}
	return lines
}

func wrapOneLine(line string, firstWidth, continuationWidth int) []string {
	if line == "" {
		return nil
	}

	if firstWidth < 0 || continuationWidth < 0 {
		// No width limit.
		return []string{normalizeLine(line)}
	}

	tokens := TokenizeWords(line)
	var (
		lines        []string
		current      strings.Builder
		currentWidth int
	)
	activeWidth := firstWidth

	for i := 0; i < len(tokens); {
		// Skip runs of spaces. Then gather tokens into a chunk until we reach the end, or an all-spaces token. Consider
		// whether there's room to add this new chunk onto the current line.
		if isAllSpaces(tokens[i]) {
			i++
			continue
		}

		var chunk strings.Builder
		for i < len(tokens) && !isAllSpaces(tokens[i]) {
			_, _ = chunk.WriteString(tokens[i])
			i++
		}

		chunkText := chunk.String()
		chunkWidth := uniseg.StringWidth(chunkText)

		// Check for line overflow, but only if there's already something else on the line. The first token on a line is
		// allowed to exceed the width limit, or else we'd never print anything.
		if current.Len() > 0 {
			if currentWidth+1+chunkWidth > activeWidth {
				// The current chunk would put the current line over the limit. Ship the line we've accumulated so far,
				// and then prepare to start a new line.
				lines = append(lines, normalizeLine(current.String()))
				current.Reset()
				currentWidth = 0
				activeWidth = continuationWidth
			} else {
				// Tokens with multiple spaces are collapsed down to a single space.
				_ = current.WriteByte(' ')
				currentWidth++
			}
		}

		_, _ = current.WriteString(chunkText)
		currentWidth += chunkWidth
	}

	// Ship the remainder of the last line.
	if current.Len() > 0 {
		lines = append(lines, normalizeLine(current.String()))
	}

	return lines
}

// TokenizeWords splits the input into tokens according to Unicode word boundaries. Consecutive spaces are returned as a
// single token, but line breaks count as separate tokens.
func TokenizeWords(text string) []string {
	var tokens []string
	state := -1
	for len(text) > 0 {
		token, rest, nextState := uniseg.FirstWordInString(text, state)
		text = rest
		state = nextState
		if token == "" && len(text) == 0 {
			break
		}
		tokens = append(tokens, token)
	}
	return tokens
}

// isAllSpaces returns true if the input is non-empty and consists entirely of spaces.
func isAllSpaces(token string) bool {
	return !(token == "" || strings.ContainsFunc(token, func(r rune) bool { return r != ' ' }))
}

// normalizeLine replaces non-breaking spaces with regular spaces. This transformation occurs late in the process so the
// actual output contains spaces, but nbsp characters can be used in the input to influence word splitting.
func normalizeLine(line string) string {
	return strings.ReplaceAll(line, "\u00a0", " ")
}
