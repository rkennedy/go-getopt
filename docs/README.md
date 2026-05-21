# getopt

[![Go Reference](https://pkg.go.dev/badge/sweetkennedy.net/getopt.svg)](https://pkg.go.dev/sweetkennedy.net/getopt)

This Go package provides argument parsing in the style of GNU getopt,
characterized by a loop that returns successive command-line options and their
arguments, possibly consolidating non-option arguments for processing
afterward. It handles short and long options.

## Usage

```go
import "sweetkennedy.net/getopt"
```

To begin, create a new `getopt.Getopt`, specifying the command-line arguments
and the option definitions. Use `New` for processing only short options, or
`NewLong` for processing both short and long options. Not every logn option
needs an associated short option, and vice versa.

```go
var help rune
shortOptions = "hf:c::"
longOptions = []getopt.Option{
    {Name: "help", Flag: &help, Val: 'h'},
    {Name: "file", HasArg: getopt.RequiredArgument, Val: 'f'},
    {Name: "color", HasArg: getopt.OptionalArgument, Val: 'c'},
}
gopt := getopt.NewLong(os.Args, shortOptions, longOptions)
```

Next, iterate using `Getopt`, `GetoptLong`, or `GetoptLongOnly` to receive each
option. If the next option is short, or if it's long and it has an associated
`Option.Val` value, then that rune will be available in the `C` field of the
returned option information. If it's long and there's no `Option.Val` for it,
then the `C` field will be zero; use the `LongInd` field to determine which
long option was provided. If an argument was provided for the option, it will
be in the `Arg` field.

```go
var opt *Opt
var err error
for opt, err = gopt.GetoptLong(); err == nil && opt != nil; opt, err = gopt.GetoptLong() {
    switch opt.C {
    case 'h':
        fmt.Println("help requested")
    case 'f':
        fmt.Printf("file argument: %s\n", opt.Arg)
    case 'c':
        if opt.Arg != "" {
            fmt.Printf("color argument: %s\n", opt.Arg)
        } else {
            fmt.Println("color argument: (none)")
        }
    case 0:
        // long option --help was given. `help` variable has been assigned the
        // value 'h' from the `Val` field.
        fmt.Println("help requested")
	}
}
if err != nil {
	fmt.Printf("error: %s\n", err.Error())
}
```

When iteration is complete, the index of the first non-option argument is
available via `Optind()`.

```go
if gopt.Optind() < len(os.Args) {
	fmt.Print("non-option ARGV-elements: ")
	for optind := gopt.Optind(); optind < len(os.Args); optind++ {
		fmt.Printf("- %s\n", os.Args[optind])
	}
}
```

## Iterator style

This library also supports a Go iterator style of processing arguments. Instead
of creating a `Getopt`, just go directly to the loop by calling `Iterate`,
`IterateLong`, or `IterateLongOnly` using the same options that would have been
passed to the corresponding `New` functions, plus a pointer to a string slice
to receiving any remaining non-option arguments.

```go
var remaining []string
for opt, err := range getopt.IterateLong(os.Args, shortOptions, longOptions, &remaining) {
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		continue
	}
    switch opt.C {
        // ...
	}
}
fmt.Printf("Remaining arguments: %v", remaining)
```

Notice that in this style, error handling happens _inside_ the loop because
`err` has been declared in the `for` header.

## Differences from GNU

There are a number of key differences between this module and GNU's C library:

1. There are no global variables. All operations are performed on a `Getopt`
   struct that maintains state between successive calls. To read an option's
   argument value, read `Opt.Arg` instead of `optarg`. To reset option-parsing,
   create a new `Getopt` struct instead of assigning `optreset`.
2. The `opterr` setting is permanently false. Errors are never printed anywhere
   by this library. Instead, errors are returned and the caller can choose what
   to do with them. The text of the errors corresponds to messages that would
   be printed by GNU getopt. The leading ':' in the option spec that controls
   error-reporting is accepted for compatibility, but it's ignored.
3. A struct is returned instead of just the matched option character. The
   struct includes the option character, any value that would have been in
   `optarg`, as well as any value that would have been returned in the
   `longindex` argument to `getopt_long`.
4. The `optopt` value is not used, and getopt does not return '?' for
   unrecognized options. Instead, it returns an `UnrecognizedOptionError`,
   which will hold the relevant unrecognized character or option name in its
   `Option` field.
5. The `Flag` and `Val` fields of `Option` have type `rune`, not `int`.
6. The list of options and arguments cannot be changed in the middle of
   parsing. The argument list and option definition are set once at the start,
   and then you just call `Getopt` or `GetoptLong` with no parameters.
7. The `POSIXLY_CORRECT` environment variable is ignored. The library runs as
   though the environment variable is never set. Use leading '+' or '-'
   characters in the option specification instead; see the `Ordering` type for
   more.
