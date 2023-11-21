package getopt

import (
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func TestEmpty(t *testing.T) {
	t.Parallel()
	cases := map[string]ordering{
		"":  Permute,
		"-": ReturnInOrder,
		"+": RequireOrder,
	}
	for k, v := range cases {
		k, v := k, v
		t.Run(k, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(parseShortOptionSpec(k)).To(MatchAllFields(Fields{
				"Ordering": Equal(v),
				"W":        BeFalse(),
				"Opts":     BeEmpty(),
			}))
		})
	}
}
