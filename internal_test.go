package getopt

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
)

func optFields(ordering, w, opts types.GomegaMatcher) Fields {
	return Fields{
		"Ordering": ordering,
		"W":        w,
		"Opts":     opts,
	}
}

// This isn't supported, but we still want to have tests to _demonstrate_ that it's not used.
const PosixlyCorrect = "POSIXLY_CORRECT"

var _ = Describe("Option parsing", func() {
	Context("with nearly empty options", func() {
		DescribeTableSubtree("with no environment variabe set",
			func(opts string, expected Ordering) {
				BeforeEach(func() {
					Expect(os.Unsetenv(PosixlyCorrect)).To(Succeed())
				})

				It("parses the options", func() {
					Expect(parseShortOptionSpec(opts)).To(MatchAllFields(optFields(
						Equal(expected),
						BeFalse(),
						BeEmpty(),
					)))
				})
			},
			Entry(nil, "", Permute),
			Entry(nil, "-", ReturnInOrder),
			Entry(nil, "+", RequireOrder),
		)

		DescribeTableSubtree("ignores posixly correct mode",
			func(opts string, expected Ordering) {
				BeforeEach(func() {
					Expect(os.Setenv(PosixlyCorrect, "yes")).To(Succeed())
				})

				It("parses the options", func() {
					Expect(parseShortOptionSpec(opts)).To(MatchAllFields(optFields(
						Equal(expected),
						BeFalse(),
						BeEmpty(),
					)))
				})
			},
			Entry(nil, "", Permute),
			Entry(nil, "-", ReturnInOrder),
			Entry(nil, "+", RequireOrder),
		)

		// Leading colon should be allowed, but also have no effect.
		DescribeTable("accepts but ignores leading colon",
			func(opts string) {
				Expect(parseShortOptionSpec(opts)).To(MatchAllFields(optFields(
					Ignore(),
					BeFalse(),
					BeEmpty(),
				)))
			},
			Entry(nil, ":"),
			Entry(nil, "-:"),
			Entry(nil, "+:"),
		)
	})

	DescribeTable("handles W options",
		func(opts string, fields Fields) {
			Expect(parseShortOptionSpec(opts)).To(MatchAllFields(fields))
		},
		Entry(nil, "W;", optFields(Ignore(), BeTrue(), MatchAllKeys(Keys{
			'W': Equal(NoArgument),
		}))),
		Entry(nil, ";W", optFields(Ignore(), BeFalse(), MatchAllKeys(Keys{
			';': Equal(NoArgument),
			'W': Equal(NoArgument),
		}))),
		Entry(nil, "w;", optFields(Ignore(), BeFalse(), MatchAllKeys(Keys{
			'w': Equal(NoArgument),
			';': Equal(NoArgument),
		}))),
		Entry(nil, "W;:", optFields(Ignore(), BeTrue(), MatchAllKeys(Keys{
			'W': Equal(NoArgument),
			':': Equal(NoArgument),
		}))),
		Entry(nil, "W:", optFields(Ignore(), BeFalse(), MatchAllKeys(Keys{
			'W': Equal(RequiredArgument),
		}))),
		Entry(nil, "W::", optFields(Ignore(), BeFalse(), MatchAllKeys(Keys{
			'W': Equal(OptionalArgument),
		}))),
	)
})
