package getopt

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
)

const PosixlyCorrect = "POSIXLY_CORRECT"

func optFields(ordering, w, opts types.GomegaMatcher) Fields {
	return Fields{
		"Ordering": ordering,
		"W":        w,
		"Opts":     opts,
	}
}

func TestEmpty(t *testing.T) {
	t.Setenv(PosixlyCorrect, "foo")
	g := NewWithT(t)
	g.Expect(os.Unsetenv(PosixlyCorrect)).To(Succeed())

	cases := map[string]ordering{
		"":  Permute,
		"-": ReturnInOrder,
		"+": RequireOrder,
	}
	for k, v := range cases {
		k, v := k, v
		t.Run(k, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(parseShortOptionSpec(k)).To(MatchAllFields(optFields(
				Equal(v),
				BeFalse(),
				BeEmpty(),
			)))
		})
	}
}

func TestPosixEnvironment(t *testing.T) {
	t.Setenv(PosixlyCorrect, "yes")

	cases := map[string]ordering{
		"":  RequireOrder,
		"-": ReturnInOrder, // This override environment variable.
		"+": RequireOrder,
	}
	for k, v := range cases {
		k, v := k, v
		t.Run(k, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(parseShortOptionSpec(k)).To(MatchAllFields(optFields(
				Equal(v),
				BeFalse(),
				BeEmpty(),
			)))
		})
	}
}

func TestLeadingColon(t *testing.T) {
	// Leading colon should be allowed, but also have no effect.
	t.Parallel()
	cases := []string{
		":",
		"-:",
		"+:",
	}
	for _, s := range cases {
		s := s
		t.Run(s, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			g.Expect(parseShortOptionSpec(s)).To(MatchAllFields(optFields(
				Ignore(),
				BeFalse(),
				BeEmpty(),
			)))
		})
	}
}

func TestWExtension(t *testing.T) {
	t.Parallel()

	cases := map[string]Fields{
		"W;": optFields(Ignore(), BeTrue(), MatchAllKeys(Keys{
			'W': Equal(NoArgument),
		})),
		";W": optFields(Ignore(), BeFalse(), MatchAllKeys(Keys{
			';': Equal(NoArgument),
			'W': Equal(NoArgument),
		})),
		"w;": optFields(Ignore(), BeFalse(), MatchAllKeys(Keys{
			'w': Equal(NoArgument),
			';': Equal(NoArgument),
		})),
		"W;:": optFields(Ignore(), BeTrue(), MatchAllKeys(Keys{
			'W': Equal(NoArgument),
			':': Equal(NoArgument),
		})),
		"W:": optFields(Ignore(), BeFalse(), MatchAllKeys(Keys{
			'W': Equal(RequiredArgument),
		})),
		"W::": optFields(Ignore(), BeFalse(), MatchAllKeys(Keys{
			'W': Equal(OptionalArgument),
		})),
	}

	for k, v := range cases {
		k, v := k, v
		t.Run(k, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			g.Expect(parseShortOptionSpec(k)).To(MatchAllFields(v))
		})
	}
}
