package help_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"sweetkennedy.net/getopt"
	"sweetkennedy.net/getopt/help"
)

var _ = Describe("Options", func() {
	Context("ShortOptions", func() {
		It("returns simple options", func() {
			params := &help.UsageParams{
				Entries: []help.Entry{
					{Short: 'h', Long: "help"},
					{Short: 'v', Long: "version"},
					{Short: 'o', Long: "output"},
				},
			}
			Expect(params.ShortOptions()).To(Equal("hvo"))
		})

		It("skips entries lacking short options", func() {
			params := &help.UsageParams{
				Entries: []help.Entry{
					{Short: 'h', Long: "help"},
					{Long: "version"},
					{Short: 'o', Long: "output"},
				},
			}
			Expect(params.ShortOptions()).To(Equal("ho"))
		})

		It("includes colons for required arguments", func() {
			params := &help.UsageParams{
				Entries: []help.Entry{
					{Short: 'h', Long: "help"},
					{Short: 'o', Long: "output", Var: "FILE"},
					{Short: 'n', Long: "number", Var: "NUM"},
				},
			}
			Expect(params.ShortOptions()).To(Equal("ho:n:"))
		})

		It("includes double colons for optional arguments", func() {
			params := &help.UsageParams{
				Entries: []help.Entry{
					{Short: 'h', Long: "help"},
					{Short: 'o', Long: "output", Var: "FILE?"},
					{Short: 'n', Long: "number", Var: "NUM"},
				},
			}
			Expect(params.ShortOptions()).To(Equal("ho::n:"))
		})
	})

	Context("LongOptions", func() {
		It("returns simple options", func() {
			params := &help.UsageParams{
				Entries: []help.Entry{
					{Short: 'h', Long: "help"},
					{Short: 'v', Long: "version"},
					{Short: 'o', Long: "output"},
				},
			}
			Expect(params.LongOptions()).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Name": Equal("help"),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Name": Equal("version"),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Name": Equal("output"),
				}),
			))
		})

		It("skips entries lacking long options", func() {
			params := &help.UsageParams{
				Entries: []help.Entry{
					{Short: 'h', Long: "help"},
					{Short: 'v'},
					{Short: 'o', Long: "output"},
				},
			}
			Expect(params.LongOptions()).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Name": Equal("help"),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Name": Equal("output"),
				}),
			))
		})

		It("includes required arguments", func() {
			params := &help.UsageParams{
				Entries: []help.Entry{
					{Short: 'h', Long: "help"},
					{Short: 'o', Long: "output", Var: "FILE"},
					{Short: 'n', Long: "number", Var: "NUM"},
				},
			}
			Expect(params.LongOptions()).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Name":   Equal("help"),
					"HasArg": Equal(getopt.NoArgument),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Name":   Equal("output"),
					"HasArg": Equal(getopt.RequiredArgument),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Name":   Equal("number"),
					"HasArg": Equal(getopt.RequiredArgument),
				}),
			))
		})

		It("includes optional arguments", func() {
			params := &help.UsageParams{
				Entries: []help.Entry{
					{Short: 'h', Long: "help"},
					{Short: 'o', Long: "output", Var: "FILE?"},
					{Short: 'n', Long: "number", Var: "NUM"},
				},
			}
			Expect(params.LongOptions()).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Name":   Equal("help"),
					"HasArg": Equal(getopt.NoArgument),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Name":   Equal("output"),
					"HasArg": Equal(getopt.OptionalArgument),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Name":   Equal("number"),
					"HasArg": Equal(getopt.RequiredArgument),
				}),
			))
		})

		It("stores short options in Val", func() {
			params := &help.UsageParams{
				Entries: []help.Entry{
					{Short: 'h', Long: "help"},
					{Long: "output"},
					{Short: 'n', Long: "number"},
				},
			}
			Expect(params.LongOptions()).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Name": Equal("help"),
					"Val":  Equal('h'),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Name": Equal("output"),
					"Val":  Equal(rune(0)),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Name": Equal("number"),
					"Val":  Equal('n'),
				}),
			))
		})
	})
})
