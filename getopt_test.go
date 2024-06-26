package getopt_test

import (
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	. "github.com/rkennedy/go-getopt"
)

type testShort struct {
	label        string
	opts         string
	argv         []string
	expectErrmsg bool
}

type testLong struct {
	label        string
	opts         string
	longopts     []Option
	argv         []string
	expectErrmsg bool
}

var shortTests = []testShort{
	{
		label:        "no errors",
		opts:         "+ab:c",
		argv:         []string{"program", "-ac", "-b", "x"},
		expectErrmsg: false,
	},
	{
		label:        "invalid option",
		opts:         "+ab:c",
		argv:         []string{"program", "-d"},
		expectErrmsg: true,
	},
	{
		label:        "missing argument",
		opts:         "+ab:c",
		argv:         []string{"program", "-b"},
		expectErrmsg: true,
	},
}

var longTests = []testLong{
	{
		label: "no errors (long)",
		opts:  "+ab:c",
		longopts: []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		},
		argv:         []string{"program", "-a", "--charlie", "--bravo=x"},
		expectErrmsg: false,
	},

	{
		label: "invalid option (long)",
		opts:  "+ab:c",
		longopts: []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		},
		argv:         []string{"program", "-a", "--charlie", "--dingo"},
		expectErrmsg: true,
	},

	{
		label: "unwanted argument",
		opts:  "+ab:c",
		longopts: []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		},
		argv:         []string{"program", "-a", "--charlie=dingo", "--bravo=x"},
		expectErrmsg: true,
	},

	{
		label: "missing argument",
		opts:  "+ab:c",
		longopts: []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		},
		argv:         []string{"program", "-a", "--charlie", "--bravo"},
		expectErrmsg: true,
	},

	{
		label: "ambiguous options",
		opts:  "+uvw",
		longopts: []Option{
			{Name: "veni", HasArg: NoArgument, Val: 'u'},
			{Name: "vedi", HasArg: NoArgument, Val: 'v'},
			{Name: "veci", HasArg: NoArgument, Val: 'w'}},
		argv:         []string{"program", "--ve"},
		expectErrmsg: true,
	},

	{
		label: "no errors (long W)",
		opts:  "+ab:cW;",
		longopts: []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		},
		argv:         []string{"program", "-a", "-W", "charlie", "-W", "bravo=x"},
		expectErrmsg: false,
	},

	{
		label: "missing argument (W itself)",
		opts:  "+ab:cW;",
		longopts: []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		},
		argv:         []string{"program", "-a", "-W", "charlie", "-W"},
		expectErrmsg: true,
	},

	{
		label: "missing argument (W longopt)",
		opts:  "+ab:cW;",
		longopts: []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		},
		argv:         []string{"program", "-a", "-W", "charlie", "-W", "bravo"},
		expectErrmsg: true,
	},

	{
		label: "unwanted argument (W longopt)",
		opts:  "+ab:cW;",
		longopts: []Option{
			{Name: "alpha", HasArg: NoArgument, Val: 'a'},
			{Name: "bravo", HasArg: RequiredArgument, Val: 'b'},
			{Name: "charlie", HasArg: NoArgument, Val: 'c'},
		},
		argv:         []string{"program", "-a", "-W", "charlie=dingo", "-W", "bravo=x"},
		expectErrmsg: true,
	},

	{
		label: "ambiguous options (W)",
		opts:  "+uvwW;",
		longopts: []Option{
			{Name: "veni", HasArg: NoArgument, Val: 'u'},
			{Name: "vedi", HasArg: NoArgument, Val: 'v'},
			{Name: "veci", HasArg: NoArgument, Val: 'w'}},
		argv:         []string{"program", "-W", "ve"},
		expectErrmsg: true,
	},
}

func TestShort(t *testing.T) {
	t.Parallel()
	for _, tc := range shortTests {
		tc := tc
		t.Run(tc.label, func(t *testing.T) {
			g := NewWithT(t)
			gopt := New(tc.argv, tc.opts)
			var opt *Opt
			var err error
			for opt, err = gopt.Getopt(); err == nil && opt != nil; opt, err = gopt.Getopt() { //revive:disable-line:empty-block,line-length-limit
			}
			if tc.expectErrmsg {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
}

func TestLong(t *testing.T) {
	t.Parallel()
	for _, tc := range longTests {
		tc := tc
		t.Run(tc.label, func(t *testing.T) {
			g := NewWithT(t)
			gopt := NewLong(tc.argv, tc.opts, tc.longopts)
			var opt *Opt
			var err error
			for opt, err = gopt.GetoptLong(); err == nil && opt != nil; opt, err = gopt.GetoptLong() { //revive:disable-line:empty-block,line-length-limit
			}
			if tc.expectErrmsg {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
}

func TestAmbiguous(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	longopts := []Option{
		{Name: "one", HasArg: NoArgument, Val: '1'},
		{Name: "two", HasArg: NoArgument, Val: '2'},
		{Name: "one-one", HasArg: NoArgument, Val: '3'},
		{Name: "four", HasArg: NoArgument, Val: '4'},
		{Name: "onto", HasArg: NoArgument, Val: '5'},
	}
	argv := []string{"program", "--on"}

	gopt := NewLong(argv, "12345", longopts)
	g.Expect(gopt.GetoptLong()).Error().
		To(MatchError("option '--on' is ambiguous; possibilities: '--one' '--one-one' '--onto'"))
}

func TestMissingArg(t *testing.T) {
	// https://sourceware.org/bugzilla/show_bug.cgi?id=11039
	t.Parallel()

	t.Run("case 1", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		gopt := New([]string{"bug-getopt1", "-a"}, ":a:b")
		g.Expect(gopt.Getopt()).Error().To(MatchError("option '-a' requires an argument"))
	})
	t.Run("case 2", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		gopt := New([]string{"bug-getopt1", "-b", "-a"}, ":a:b")
		g.Expect(gopt.GetoptLong()).To(HaveValue(MatchFields(IgnoreExtras, Fields{
			"C": Equal('b'),
		})))
		g.Expect(gopt.Getopt()).Error().To(MatchError("option '-a' requires an argument"))
	})
}

func TestWSemicolon(t *testing.T) {
	t.Parallel()

	longopts := []Option{
		{Name: "alpha", HasArg: NoArgument, Val: 'a'},
		{Name: "beta", HasArg: RequiredArgument, Val: 'b'},
	}
	t.Run("1", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		gopt := NewLong([]string{"bug-getopt3", "-a;"}, "ab:W;", longopts)
		g.Expect(gopt.GetoptLong()).To(HaveValue(MatchFields(IgnoreExtras, Fields{
			"C": Equal('a'),
		})))
		g.Expect(gopt.GetoptLong()).Error().To(MatchError("unrecognized option '-;'"))
	})
	t.Run("2", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		gopt := NewLong([]string{"bug-getopt3", "-a:"}, "ab:W;", longopts)
		g.Expect(gopt.GetoptLong()).To(HaveValue(MatchFields(IgnoreExtras, Fields{
			"C": Equal('a'),
		})))
		g.Expect(gopt.GetoptLong()).Error().To(MatchError("unrecognized option '-:'"))
	})
	t.Run("3", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		gopt := New([]string{"program", "-W", "-;"}, "W;")
		g.Expect(gopt.Getopt()).To(HaveValue(MatchFields(IgnoreExtras, Fields{
			"C": Equal('W'),
		})))
		g.Expect(gopt.Getopt()).Error().To(MatchError("unrecognized option '-;'"))
	})
	t.Run("reads third argument", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		gopt := NewLong([]string{"program", "-W", "opt", "arg"}, "W;", []Option{
			{Name: "opt", HasArg: RequiredArgument},
		})
		g.Expect(gopt.Getopt()).To(HaveValue(MatchAllFields(Fields{
			"C":       Equal(rune(0)),
			"Arg":     HaveValue(Equal("arg")),
			"LongInd": Equal(0),
		})))
	})
}
