package help_test

//revive:disable:add-constant

import (
	"os"

	"golang.org/x/term"

	"sweetkennedy.net/getopt/help"
)

func ExampleUsage() {
	params := &help.UsageParams{
		Usage:       "myprog [options] <input>",
		Description: "This is a description of myprog.",
		Entries: []help.Entry{
			{Short: 'h', Long: "help", Help: "Show this help message and exit."},
			{Short: 'v', Long: "version", Help: "Show version information and exit."},
			{Short: 'o', Long: "output", Var: "FILE", Help: "Write output to FILE instead of standard output."},
			{Long: "verbose", Help: "Enable verbose mode."},
			{Short: 'n', Var: "NUM?", Help: "Process at most NUM items. If NUM is not provided, process all items."},
		},
		Trailer: "For more information, see the manual page.",
	}
	help.Usage(os.Stdout, params, 80)
	// Output:
	// Usage: myprog [options] <input>
	// This is a description of myprog.
	//   -h, --help                 Show this help message and exit.
	//   -v, --version              Show version information and exit.
	//   -o, --output FILE          Write output to FILE instead of standard output.
	//       --verbose              Enable verbose mode.
	//   -n[=NUM]                   Process at most NUM items. If NUM is not provided,
	//                              process all items.
	// For more information, see the manual page.
}

func ExampleUsage_terminal() {
	params := &help.UsageParams{
		Usage:       "myprog [options] <input>",
		Description: "This is a description of myprog.",
		Entries: []help.Entry{
			{Short: 'h', Long: "help", Help: "Show this help message and exit."},
			{Short: 'v', Long: "version", Help: "Show version information and exit."},
			{Short: 'o', Long: "output", Var: "FILE", Help: "Write output to FILE instead of standard output."},
			{Long: "verbose", Help: "Enable verbose mode."},
			{Short: 'n', Var: "NUM?", Help: "Process at most NUM items. If NUM is not provided, process all items."},
		},
		Trailer: "For more information, see the manual page.",
	}

	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80
	}

	help.Usage(os.Stdout, params, width)
	// Output:
	// Usage: myprog [options] <input>
	// This is a description of myprog.
	//   -h, --help                 Show this help message and exit.
	//   -v, --version              Show version information and exit.
	//   -o, --output FILE          Write output to FILE instead of standard output.
	//       --verbose              Enable verbose mode.
	//   -n[=NUM]                   Process at most NUM items. If NUM is not provided,
	//                              process all items.
	// For more information, see the manual page.
}
