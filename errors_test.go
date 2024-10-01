package getopt_test

import (
	"fmt"

	. "github.com/rkennedy/go-getopt"
)

func ExampleAmbiguousOptionError() {
	longopts := []Option{
		{Name: "one", HasArg: NoArgument, Val: '1'},
		{Name: "two", HasArg: NoArgument, Val: '2'},
		{Name: "one-one", HasArg: NoArgument, Val: '3'},
		{Name: "four", HasArg: NoArgument, Val: '4'},
		{Name: "onto", HasArg: NoArgument, Val: '5'},
	}
	argv := []string{"program", "--on"}

	gopt := NewLong(argv, "12345", longopts)
	_, err := gopt.GetoptLong()
	_, _ = fmt.Println(err.Error())
	// Output: option '--on' is ambiguous; possibilities: '--one' '--one-one' '--onto'
}

func ExampleUnrecognizedOptionError() {
	argv := []string{"program", "-c"}

	gopt := New(argv, "ab")
	_, err := gopt.Getopt()
	_, _ = fmt.Println(err.Error())
	// Output: unrecognized option '-c'
}

func ExampleArgumentNotAllowedError() {
	longopts := []Option{
		{Name: "sample", HasArg: NoArgument},
	}
	argv := []string{"program", "--sample=arg"}

	gopt := NewLong(argv, "", longopts)
	_, err := gopt.Getopt()
	_, _ = fmt.Println(err.Error())
	// Output: option '--sample' doesn't allow an argument
}

func ExampleArgumentRequiredError() {
	argv := []string{"program", "-a"}

	gopt := New(argv, "a:")
	_, err := gopt.Getopt()
	_, _ = fmt.Println(err.Error())
	// Output: option '-a' requires an argument
}

func ExampleGetopt_Optind() {
	argv := []string{"program", "-a", "f1", "f2", "f3"}

	gopt := New(argv, "a:")
	for opt, err := gopt.Getopt(); opt != nil && err == nil; opt, err = gopt.Getopt() {
		_, _ = fmt.Printf("Got argument: %s\n", *opt.Arg)
	}
	_, _ = fmt.Printf("Remaining arguments: %v", gopt.Args[gopt.Optind():])
	// Output:
	// Got argument: f1
	// Remaining arguments: [f2 f3]
}
