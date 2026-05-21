package getopt_test

import (
	"fmt"
	"slices"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	. "sweetkennedy.net/getopt"
)

var _ = Describe("Getopt", func() {
	DescribeTable("handles short options",
		func(opts string, argv []string, expectErrmsg bool) {
			gopt := New(argv, opts)
			var opt *Opt
			var err error
			for opt, err = gopt.Getopt(); err == nil && opt != nil; opt, err = gopt.Getopt() { //revive:disable-line:empty-block,line-length-limit
			}
			if expectErrmsg {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("no errors", "+ab:c", []string{"program", "-ac", "-b", "x"}, false),
		Entry("invalid option", "+ab:c", []string{"program", "-d"}, true),
		Entry("missing argument", "+ab:c", []string{"program", "-b"}, true),
	)

	DescribeTable("handles long options",
		func(opts string, longopts []Option, argv []string, expectErrmsg bool) {
			gopt := NewLong(argv, opts, longopts)
			var opt *Opt
			var err error
			for opt, err = gopt.GetoptLong(); err == nil && opt != nil; opt, err = gopt.GetoptLong() { //revive:disable-line:empty-block,line-length-limit
			}
			if expectErrmsg {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("no errors (long)", "+ab:c", []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		}, []string{"program", "-a", "--charlie", "--bravo=x"}, false,
		),
		Entry("invalid option (long)", "+ab:c", []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		}, []string{"program", "-a", "--charlie", "--dingo"}, true,
		),
		Entry("unwanted argument", "+ab:c", []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		}, []string{"program", "-a", "--charlie=dingo", "--bravo=x"}, true,
		),
		Entry("missing argument", "+ab:c", []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		}, []string{"program", "-a", "--charlie", "--bravo"}, true,
		),
		Entry("ambiguous options", "+uvw", []Option{
			{Name: "veni", HasArg: NoArgument, Val: 'u'},
			{Name: "vedi", HasArg: NoArgument, Val: 'v'},
			{Name: "veci", HasArg: NoArgument, Val: 'w'},
		}, []string{"program", "--ve"}, true,
		),
		Entry("no errors (long W)", "+ab:cW;", []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		}, []string{"program", "-a", "-W", "charlie", "-W", "bravo=x"}, false,
		),
		Entry("missing argument (W itself)", "+ab:cW;", []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		}, []string{"program", "-a", "-W", "charlie", "-W"}, true,
		),
		Entry("missing argument (W longopt)", "+ab:cW;", []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		}, []string{"program", "-a", "-W", "charlie", "-W", "bravo"}, true,
		),
		Entry("unwanted argument (W longopt)", "+ab:cW;", []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		}, []string{"program", "-a", "-W", "charlie=dingo", "-W", "bravo=x"}, true,
		),
		Entry("ambiguous options (W)", "+uvwW;", []Option{
			{Name: "veni", HasArg: NoArgument, Val: 'u'},
			{Name: "vedi", HasArg: NoArgument, Val: 'v'},
			{Name: "veci", HasArg: NoArgument, Val: 'w'},
		}, []string{"program", "-W", "ve"}, true,
		),
	)

	It("detects ambiguous arguments", func() {
		longopts := []Option{
			{Name: "one", HasArg: OptionalArgument, Val: '1'},
			{Name: "two", HasArg: NoArgument, Val: '2'},
			{Name: "one-one", HasArg: OptionalArgument, Val: '3'},
			{Name: "four", HasArg: NoArgument, Val: '4'},
			{Name: "onto", HasArg: OptionalArgument, Val: '5'},
		}
		argv := []string{"program", "--on=value"}

		gopt := NewLong(argv, "12345", longopts)
		Expect(gopt.GetoptLong()).Error().
			To(MatchError("option '--on' is ambiguous; possibilities: '--one' '--one-one' '--onto'"))
	})

	Context("detects missing arguments", func() {
		// https://sourceware.org/bugzilla/show_bug.cgi?id=11039
		It("case 1", func() {
			gopt := New([]string{"bug-getopt1", "-a"}, ":a:b")
			Expect(gopt.Getopt()).Error().To(MatchError("option '-a' requires an argument"))
		})
		It("case 2", func() {
			gopt := New([]string{"bug-getopt1", "-b", "-a"}, ":a:b")
			Expect(gopt.GetoptLong()).To(HaveValue(MatchFields(IgnoreExtras, Fields{
				"C": Equal('b'),
			})))
			Expect(gopt.Getopt()).Error().To(MatchError("option '-a' requires an argument"))
		})
	})

	It("continues after detecting error", func() {
		longOpts := []Option{
			{Name: "aaa", Val: 'a'},
			{Name: "bbb", Val: 'b'},
			{Name: "ccc", Val: 'c'},
		}
		g := NewLong([]string{"prg", "-acb"}, "", longOpts)
		Expect(g.Getopt()).Error().To(MatchError("unrecognized option '-a'"))
		Expect(g.Getopt()).Error().To(MatchError("unrecognized option '-c'"))
		Expect(g.Getopt()).Error().To(MatchError("unrecognized option '-b'"))
		Expect(g.Getopt()).To(BeNil())
	})

	It("permutes the caller's input slice", func() {
		args := []string{"program", "arg1", "-a", "arg2", "-b"}

		gopt := New(args, "ab")
		for opt, err := gopt.Getopt(); opt != nil || err != nil; opt, err = gopt.Getopt() {
			Expect(err).NotTo(HaveOccurred())
		}

		match := And(
			HaveLen(gopt.Optind()+2),
			// We only care about the positions of the non-option arguments that get permuted to the end of the list.
			MatchElementsWithIndex(IndexIdentity, IgnoreExtras, Elements{
				strconv.Itoa(gopt.Optind()):     Equal("arg1"),
				strconv.Itoa(gopt.Optind() + 1): Equal("arg2"),
			}),
		)
		// In practice, both slices refer to the same data. That is, &args[0] == &gopt.Args[0]. That details isn't
		// strictly required, so long as both slices _look_ the same.
		Expect(args).To(match)
		Expect(gopt.Args).To(match)
	})

	It("returns options and non-options interspersed in ReturnInOrder mode", func() {
		args := []string{"program", "x", "-a", "y", "-b", "z"}
		original := slices.Clone(args)

		gopt := New(args, "-ab")
		got := []string{}
		for opt, err := gopt.Getopt(); opt != nil || err != nil; opt, err = gopt.Getopt() {
			Expect(err).NotTo(HaveOccurred())
			if opt.C == 1 {
				Expect(opt.Arg).NotTo(BeNil())
				got = append(got, "1:"+*opt.Arg)
				continue
			}
			got = append(got, string(opt.C))
		}

		Expect(got).To(HaveExactElements("1:x", "a", "1:y", "b", "1:z"))
		Expect(args).To(Equal(original))
		Expect(gopt.Args).To(Equal(original))
	})

	Context("handles W; options", func() {
		longopts := []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "beta", HasArg: RequiredArgument, Val: 'b'},
		}

		It("case 1", func() {
			gopt := NewLong([]string{"bug-getopt3", "-a;"}, "ab:W;", longopts)
			Expect(gopt.GetoptLong()).To(HaveValue(MatchFields(IgnoreExtras, Fields{
				"C": Equal('a'),
			})))
			Expect(gopt.GetoptLong()).Error().To(MatchError("unrecognized option '-;'"))
		})

		It("case 2", func() {
			gopt := NewLong([]string{"bug-getopt3", "-a:"}, "ab:W;", longopts)
			Expect(gopt.GetoptLong()).To(HaveValue(MatchFields(IgnoreExtras, Fields{
				"C": Equal('a'),
			})))
			Expect(gopt.GetoptLong()).Error().To(MatchError("unrecognized option '-:'"))
		})

		It("case 3", func() {
			gopt := New([]string{"program", "-W", "-;"}, "W;")
			Expect(gopt.Getopt()).To(HaveValue(MatchFields(IgnoreExtras, Fields{
				"C": Equal('W'),
			})))
			Expect(gopt.Getopt()).Error().To(MatchError("unrecognized option '-;'"))
		})

		It("reads third argument", func() {
			gopt := NewLong([]string{"program", "-W", "opt", "arg"}, "W;", []Option{
				{Name: "opt", HasArg: RequiredArgument},
			})
			Expect(gopt.Getopt()).To(HaveValue(MatchAllFields(Fields{
				"C":       Equal(rune(0)),
				"Arg":     HaveValue(Equal("arg")),
				"LongInd": Equal(0),
			})))
		})
	})
})

func ExampleGetopt_Getopt() {
	args := []string{"program", "-n", "-t", "3", "4", "5"}
	gopt := New(args, "nt:")

	nsecs := 0
	tfnd := 0
	flags := 0

	var opt *Opt
	var err error
	for opt, err = gopt.Getopt(); err == nil && opt != nil; opt, err = gopt.Getopt() {
		switch opt.C {
		case 'n':
			flags = 1
		case 't':
			nsecs, _ = strconv.Atoi(*opt.Arg)
			tfnd = 1
		default:
		}
	}
	_, _ = fmt.Printf("flags=%d; tfnd=%d; nsecs=%d; optind=%d\n", flags, tfnd, nsecs, gopt.Optind())

	// Output:
	// flags=1; tfnd=1; nsecs=3; optind=4
}

func ExampleGetopt_GetoptLong() {
	longOptions := []Option{
		{Name: "add", HasArg: RequiredArgument},
		{Name: "append", HasArg: NoArgument},
		{Name: "delete", HasArg: RequiredArgument},
		{Name: "verbose", HasArg: NoArgument},
		{Name: "create", HasArg: RequiredArgument, Val: 'c'},
		{Name: "file", HasArg: RequiredArgument},
	}

	args := []string{"program", "--add", "a", "--append", "-0", "--delete", "d", "-2", "--verbose", "-cc", "--", "xyz"}
	gopt := NewLong(args, "abc:d:012", longOptions)

	digitOptind := 0

	var opt *Opt
	var err error
	for opt, err = gopt.GetoptLong(); err == nil && opt != nil; opt, err = gopt.GetoptLong() {
		thisOptionOptind := gopt.Optind()

		switch opt.C {
		case 0:
			_, _ = fmt.Printf("option %s", longOptions[opt.LongInd].Name)
			if opt.Arg != nil {
				_, _ = fmt.Printf(" with arg %s", *opt.Arg)
			}
			_, _ = fmt.Println()

		case '0', '1', '2':
			if digitOptind != 0 && digitOptind != thisOptionOptind {
				_, _ = fmt.Println("digits occur in two different argv-elements.")
			}
			digitOptind = thisOptionOptind
			_, _ = fmt.Printf("option %c\n", opt.C)

		case 'a':
			_, _ = fmt.Println("option a")

		case 'b':
			_, _ = fmt.Println("option b")

		case 'c':
			_, _ = fmt.Printf("option c with value '%s'\n", *opt.Arg)

		case 'd':
			_, _ = fmt.Printf("option d with value '%s'\n", *opt.Arg)

		case '?':

		default:
			_, _ = fmt.Printf("?? getopt returned character code 0%o ??\n", opt.C)
		}
	}

	if err != nil {
		_, _ = fmt.Printf("error: %v\n", err)
	}

	if gopt.Optind() < len(args) {
		_, _ = fmt.Print("non-option ARGV-elements: ")
		for optind := gopt.Optind(); optind < len(args); optind++ {
			_, _ = fmt.Printf("%s ", args[optind])
		}
		_, _ = fmt.Println()
	}

	// Output:
	// option add with arg a
	// option append
	// option 0
	// option delete with arg d
	// digits occur in two different argv-elements.
	// option 2
	// option verbose
	// option c with value 'c'
	// non-option ARGV-elements: xyz
}
