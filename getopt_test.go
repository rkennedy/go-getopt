package getopt_test

import (
	"bytes"
	"testing"

	. "github.com/onsi/gomega"
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
			gopt := New()
			var buf bytes.Buffer
			gopt.Out = &buf
			for ch, err := gopt.Loop(tc.argv, tc.opts); err == nil && ch != -1; ch, err = gopt.Loop(tc.argv, tc.opts) { //revive:disable-line:empty-block,line-length-limit
			}
			if tc.expectErrmsg {
				g.Expect(buf.String()).NotTo(Equal(""))
			} else {
				g.Expect(buf.String()).To(Equal(""))
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
			gopt := New()
			var buf bytes.Buffer
			gopt.Out = &buf
			for ch, _, err := gopt.LoopLong(tc.argv, tc.opts, tc.longopts); err == nil && ch != -1; ch, _, err = gopt.LoopLong(tc.argv, tc.opts, tc.longopts) { //revive:disable-line:empty-block,line-length-limit
			}
			if tc.expectErrmsg {
				g.Expect(buf.String()).NotTo(Equal(""))
			} else {
				g.Expect(buf.String()).To(Equal(""))
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

	gopt := New()
	var buf bytes.Buffer
	gopt.Out = &buf
	ch, _, _ := gopt.LoopLong(argv, "12345", longopts)

	g.Expect(ch).To(Equal('?'))
	g.Expect(buf.String()).To(Equal("ambiguous option -- on\n"))
}