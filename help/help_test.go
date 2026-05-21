package help_test

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"

	"sweetkennedy.net/getopt/help"
)

var _ = Describe("TokenizeWords", func() {
	DescribeTable("tokenizes whitespace separately",
		func(input string, expected []string) {
			Expect(help.TokenizeWords(input)).To(HaveExactElements(expected))
		},
		Entry("simple", "a b c", []string{"a", " ", "b", " ", "c"}),
		Entry("multiple spaces", "a  b   c", []string{"a", "  ", "b", "   ", "c"}),
		Entry("leading and trailing spaces", "  a b c  ", []string{"  ", "a", " ", "b", " ", "c", "  "}),
		Entry("tabs and newlines", "a\tb\nc", []string{"a", "\t", "b", "\n", "c"}),
		Entry("newlines separate from space", "a \n\n b", []string{"a", " ", "\n", "\n", " ", "b"}),
	)
})

var _ = Describe("Usage", func() {
	DescribeTable("aligns the first column",
		func(short rune, long string, meta string) {
			params := help.UsageParams{
				Entries: []help.Entry{
					{Short: short, Long: long, Var: meta, Help: "a short message"},
				},
			}
			var buf bytes.Buffer
			help.Usage(&buf, &params, 80)
			Expect(buf.String()).To(WithTransform(func(s string) int {
				return strings.Index(s, "a short message")
			}, Equal(29)), fmt.Sprintf("%#v", buf.String()))
		},
		Entry("short only, no var", 'a', "", ""),
		Entry("long only, no var", rune(0), "long", ""),
		Entry("both, no var", 'a', "long", ""),
		Entry("short only, required", 'a', "", "V"),
		Entry("long only, required", rune(0), "long", "V"),
		Entry("both, required", 'a', "long", "V"),
		Entry("short only, optional", 'a', "", "V?"),
		Entry("long only, optional", rune(0), "long", "V?"),
		Entry("both, optional", 'a', "long", "V?"),
	)

	It("pushes first line by 1", func() {
		params := help.UsageParams{
			Entries: []help.Entry{
				{Short: rune(0), Long: "indicator-style", Var: "WORD", Help: "a short message"},
			},
		}
		var buf bytes.Buffer
		help.Usage(&buf, &params, 80)
		Expect(buf.String()).To(WithTransform(func(s string) int {
			return strings.Index(s, "a short message")
		}, Equal(30)), fmt.Sprintf("%#v", buf.String()))
		Expect(strings.Index(buf.String(), "a short message")).To(Equal(30))
	})

	DescribeTable("starts help on next line",
		func(short rune, long string, meta string) {
			params := help.UsageParams{
				Entries: []help.Entry{
					{Short: short, Long: long, Var: meta, Help: "a short message"},
				},
			}
			var buf bytes.Buffer
			help.Usage(&buf, &params, 80)
			lines := strings.Split(buf.String(), "\n")
			Expect(lines).To(HaveLen(3))
			Expect(lines[0]).NotTo(ContainSubstring("a short message"))
			Expect(lines[1]).To(WithTransform(func(s string) int {
				return strings.Index(s, "a short message")
			}, Equal(29)), fmt.Sprintf("%#v", lines[1]))
		},
		Entry("exact width", rune(0), "indicator-style", "WORDS"),
		Entry("also exact width", 'p', "indicator-style=slash", ""),
		Entry("obviously too wide", 'H', "dereference-command-line", ""),
	)

	It("wraps long messages to the next line", func() {
		param := help.UsageParams{
			Entries: []help.Entry{
				{Short: 'a', Help: "a very long message that should wrap to the next line"},
			},
		}
		var buf bytes.Buffer
		help.Usage(&buf, &param, 80)
		lines := strings.Split(buf.String(), "\n")
		Expect(lines).To(HaveLen(3))
		Expect(lines[0]).To(MatchRegexp("^.{29}a very long message that should wrap"))
		Expect(lines[1]).To(MatchRegexp("^ {29}line"))
	})

	It("wraps combining-mark text by display width", func() {
		param := help.UsageParams{
			Entries: []help.Entry{
				{Short: 'a', Help: "the Ka\u0308se a"}, // "Käse"
			},
		}
		var buf bytes.Buffer
		// 29 cols before help + 8 cols for first two words. If the second word's width is incorrectly calculated as 5
		// instead of 4, then it will wrap to the second line.
		help.Usage(&buf, &param, 37)
		lines := strings.Split(buf.String(), "\n")
		Expect(lines).To(HaveLen(3))
		Expect(lines[0]).To(MatchRegexp("^.{29}the Käse$"))
		Expect(lines[1]).To(MatchRegexp("^ {29}a$"))
	})

	It("wraps emoji and flags by display width", func() {
		param := help.UsageParams{
			Entries: []help.Entry{
				{Short: 'a', Help: "\U0001f3f3\ufe0f\u200d\U0001f308 \U0001f1e9\U0001f1ea e"}, // 🏳️‍🌈 🇩🇪
			},
		}
		var buf bytes.Buffer
		// 29 cols before help + 5 cols for "🏳️‍🌈 🇩🇪" + 1 more, in case widths are calculated incorrectly and the
		// trailing "e" gets put on the wrong line.
		help.Usage(&buf, &param, 35)
		lines := strings.Split(buf.String(), "\n")
		Expect(lines).To(HaveLen(3))
		Expect(lines[0]).To(MatchRegexp("^.{29}🏳️‍🌈 🇩🇪$"))
		Expect(lines[1]).To(MatchRegexp("^ {29}e$"))
	})

	It("honors newlines in the help text", func() {
		param := help.UsageParams{
			Entries: []help.Entry{
				{Short: 'a', Help: "a message\nthat contains\nnewlines"},
			},
		}
		var buf bytes.Buffer
		help.Usage(&buf, &param, 80)
		lines := strings.Split(buf.String(), "\n")
		Expect(lines).To(HaveLen(4))
		Expect(lines[0]).To(MatchRegexp("^.{29}a message$"))
		Expect(lines[1]).To(MatchRegexp("^ {29}that contains$"))
		Expect(lines[2]).To(MatchRegexp("^ {29}newlines$"))
	})

	It("omits padding with empty help text", func() {
		param := help.UsageParams{
			Entries: []help.Entry{
				{Short: 'a', Help: ""},
			},
		}
		var buf bytes.Buffer
		help.Usage(&buf, &param, 80)
		Expect(buf.String()).To(Equal("  -a\n"))
	})

	It("converts nbsp to regular space", func() {
		param := help.UsageParams{
			Entries: []help.Entry{
				{Short: rune(0), Long: "indicator-style", Var: "WORD", Help: "append indicator with style WORD to " +
					"entry names: none\u00a0(default), slash\u00a0(-p), file-type\u00a0(--file-type), " +
					"classify\u00a0(-F)"},
			},
		}
		var buf bytes.Buffer
		help.Usage(&buf, &param, 80)
		// At 80 columns, the natural word wrap will put a break between "file-type" and "(--file-type)". Ensure the
		// nbsp kept the two tokens together, but that we didn't actually print the nbsp.
		Expect(buf.String()).To(ContainSubstring("file-type (--file-type)"))
	})

	It("wraps long usage messages", func() {
		param := help.UsageParams{
			Usage: "a long usage message that should wrap to the next line",
		}
		var buf bytes.Buffer
		help.Usage(&buf, &param, 40)
		lines := strings.Split(buf.String(), "\n")
		Expect(lines).To(HaveExactElements(
			"Usage: a long usage message that should",
			"       wrap to the next line",
			"",
		))
	})

	It("wraps long descriptions and trailers", func() {
		param := help.UsageParams{
			Description: "a long description that should wrap to the next line",
			Trailer:     "a long trailer that should also wrap to the next line",
		}
		var buf bytes.Buffer
		help.Usage(&buf, &param, 30)
		lines := strings.Split(buf.String(), "\n")
		Expect(lines).To(HaveExactElements(
			"a long description that should",
			"wrap to the next line",
			"a long trailer that should",
			"also wrap to the next line",
			"",
		))
	})

	When("width limit is non-positive", func() {
		It("does not wrap", func() {
			param := help.UsageParams{
				Usage:       "a long usage message that would normally wrap at reasonable widths like 80 columns",
				Description: "a long description with enough text to trigger wrapping if the width limit were 80",
				Entries: []help.Entry{
					{Short: 'a', Help: "a help message that contains enough text that it would normally wrap"},
				},
			}
			var buf bytes.Buffer
			help.Usage(&buf, &param, 0)
			lines := strings.Split(buf.String(), "\n")
			Expect(lines).To(HaveLen(4))
			Expect(lines[0]).To(Equal(
				"Usage: a long usage message that would normally wrap at reasonable widths like 80 columns"))
			Expect(lines[1]).To(Equal(
				"a long description with enough text to trigger wrapping if the width limit were 80"))
			Expect(lines[2]).To(HavePrefix("  -a"))
		})

		It("honors explicit newlines", func() {
			param := help.UsageParams{
				Entries: []help.Entry{
					{Short: 'a', Help: "first line\nsecond line with more text\nthird line"},
				},
			}
			var buf bytes.Buffer
			help.Usage(&buf, &param, 0)
			lines := strings.Split(buf.String(), "\n")
			Expect(lines).To(HaveLen(4))
			Expect(lines[0]).To(MatchRegexp("^  -a +first line$"))
			Expect(lines[1]).To(MatchRegexp("^ {29}second line with more text$"))
			Expect(lines[2]).To(MatchRegexp("^ {29}third line$"))
		})
	})

	It("formats ls output", func() {
		format.CharactersAroundMismatchToInclude = 30
		params := help.UsageParams{
			Usage:       "ls [options] [file...]",
			Description: "List information about the FILEs (the current directory by default).",
			Trailer:     "Report bugs to <https://bugs.example.com/ls>.",
			Entries: []help.Entry{
				{Short: 'a', Long: "all", Help: "do not ignore entries starting with ."},
				{Short: 'A', Long: "almost-all", Help: "do not list implied . and .."},
				{Long: "author", Help: "with -l, print the author of each file"},
				{Short: 'b', Long: "escape", Help: "print C-style escapes for nongraphic characters"},
				{Long: "block-size", Var: "SIZE", Help: "with -l, scale sizes by SIZE when printing them; e.g., " +
					"'--block-size=M'; see SIZE format below"},
				{Short: 'B', Long: "ignore-backups", Help: "do not list implied entries ending with ~"},
				{Short: 'c', Help: "with -lt: sort by, and show, ctime (time of last modification of file status " +
					"information);\nwith -l: show ctime and sort by name;\notherwise: sort by ctime, newest first"},
				{Short: 'C', Help: "list entries by columns"},
				{Long: "color", Var: "WHEN?", Help: "colorize the output; WHEN can be 'always' (default if omitted), " +
					"'auto', or 'never'; more info below"},
				{Long: "group-directories-first", Help: "group directories before files;\ncan be augmented with a " +
					"--sort option, but any use of --sort=none (-U) disables grouping"},
				{Short: 'G', Long: "no-group", Help: "in a long listing, don't print group names"},
				{Short: 'h', Long: "human-readable", Help: "with -l and -s, print sizes like 1K 234M 2G etc."},
				{Long: "si", Help: "likewise, but use powers of 1000 not 1024"},
				{Short: 'H', Long: "dereference-command-line",
					Help: "follow symbolic links listed on the command line"},
				{Long: "dereference-command-line-symlink-to-dir", Help: "follow each command line symbolic link that " +
					"points to a directory"},
				{Long: "hide", Var: "PATTERN", Help: "do not list implied entries matching shell PATTERN (overridden " +
					"by -a or -A)"},
				{Long: "hyperlink", Var: "WHEN?", Help: "hyperlink file names; WHEN can be 'always' (default if " +
					"omitted), 'auto', or 'never'"},
				{Long: "indicator-style", Var: "WORD", Help: "append indicator with style WORD to entry names: " +
					"none\u00a0(default), slash\u00a0(-p), file-type\u00a0(--file-type), classify\u00a0(-F)"},
				{Short: 'i', Long: "inode", Help: "print the index number of each file"},
				{Short: 'I', Long: "ignore", Var: "PATTERN", Help: "do not list implied entries matching shell " +
					"PATTERN"},
				{Short: 'k', Long: "kibibytes", Help: "default to 1024-byte blocks for disk usage; used only with -s " +
					"and per directory totals"},
				{Short: 'l', Help: "use a long listing format"},
				{Short: 'L', Long: "dereference", Help: "when showing file information for a symbolic link, show " +
					"information for the file the link references rather than for the link itself"},
				{Short: 'm', Help: "fill width with a comma separated list of entries"},
				{Short: 'n', Long: "numeric-uid-gid", Help: "like -l, but list numeric user and group IDs"},
				{Short: 'N', Long: "literal", Help: "print entry names without quoting"},
				{Short: 'o', Help: "like -l, but do not list group information"},
				{Short: 'p', Long: "indicator-style=slash", Help: "append / indicator to directories"},
			},
		}
		var buf bytes.Buffer
		help.Usage(&buf, &params, 80)
		Expect(buf.String()).To(Equal(heredoc.Doc(`
			Usage: ls [options] [file...]
			List information about the FILEs (the current directory by default).
			  -a, --all                  do not ignore entries starting with .
			  -A, --almost-all           do not list implied . and ..
			      --author               with -l, print the author of each file
			  -b, --escape               print C-style escapes for nongraphic characters
			      --block-size SIZE      with -l, scale sizes by SIZE when printing them;
			                             e.g., '--block-size=M'; see SIZE format below
			  -B, --ignore-backups       do not list implied entries ending with ~
			  -c                         with -lt: sort by, and show, ctime (time of last
			                             modification of file status information);
			                             with -l: show ctime and sort by name;
			                             otherwise: sort by ctime, newest first
			  -C                         list entries by columns
			      --color[=WHEN]         colorize the output; WHEN can be 'always' (default
			                             if omitted), 'auto', or 'never'; more info below
			      --group-directories-first
			                             group directories before files;
			                             can be augmented with a --sort option, but any use
			                             of --sort=none (-U) disables grouping
			  -G, --no-group             in a long listing, don't print group names
			  -h, --human-readable       with -l and -s, print sizes like 1K 234M 2G etc.
			      --si                   likewise, but use powers of 1000 not 1024
			  -H, --dereference-command-line
			                             follow symbolic links listed on the command line
			      --dereference-command-line-symlink-to-dir
			                             follow each command line symbolic link that points
			                             to a directory
			      --hide PATTERN         do not list implied entries matching shell PATTERN
			                             (overridden by -a or -A)
			      --hyperlink[=WHEN]     hyperlink file names; WHEN can be 'always' (default
			                             if omitted), 'auto', or 'never'
			      --indicator-style WORD  append indicator with style WORD to entry names:
			                             none (default), slash (-p),
			                             file-type (--file-type), classify (-F)
			  -i, --inode                print the index number of each file
			  -I, --ignore PATTERN       do not list implied entries matching shell PATTERN
			  -k, --kibibytes            default to 1024-byte blocks for disk usage; used
			                             only with -s and per directory totals
			  -l                         use a long listing format
			  -L, --dereference          when showing file information for a symbolic link,
			                             show information for the file the link references
			                             rather than for the link itself
			  -m                         fill width with a comma separated list of entries
			  -n, --numeric-uid-gid      like -l, but list numeric user and group IDs
			  -N, --literal              print entry names without quoting
			  -o                         like -l, but do not list group information
			  -p, --indicator-style=slash
			                             append / indicator to directories
			Report bugs to <https://bugs.example.com/ls>.
		`)))
	})
})
