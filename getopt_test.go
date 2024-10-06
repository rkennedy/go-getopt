package getopt_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	. "github.com/rkennedy/go-getopt"
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
			{Name: "one", HasArg: NoArgument, Val: '1'},
			{Name: "two", HasArg: NoArgument, Val: '2'},
			{Name: "one-one", HasArg: NoArgument, Val: '3'},
			{Name: "four", HasArg: NoArgument, Val: '4'},
			{Name: "onto", HasArg: NoArgument, Val: '5'},
		}
		argv := []string{"program", "--on"}

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
			{Name: "bbb", Val: 'd'},
			{Name: "ccc", Val: 'c'},
		}
		g := NewLong([]string{"prg", "-acb"}, "", longOpts)
		Expect(g.Getopt()).Error().To(MatchError("unrecognized option '-a'"))
		Expect(g.Getopt()).Error().To(MatchError("unrecognized option '-c'"))
		Expect(g.Getopt()).Error().To(MatchError("unrecognized option '-b'"))
		Expect(g.Getopt()).To(BeNil())
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
