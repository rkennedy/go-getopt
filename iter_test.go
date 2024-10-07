package getopt_test

import (
	"fmt"
	"iter"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/rkennedy/go-getopt"
)

type Pair[K comparable, V any] struct {
	K K
	V V
}

func collect[K comparable, V any](items iter.Seq2[K, V]) (result []Pair[K, V]) {
	for k, v := range items {
		result = append(result, Pair[K, V]{k, v})
	}
	return result
}

var _ = Describe("Getopt iterator interface", func() {
	It("returns in order", func() {
		var remaining []string
		opts := collect(getopt.Iterate([]string{"prg", "-ba", "-c"}, "abc", &remaining))
		Expect(opts).To(HaveExactElements(
			MatchAllFields(Fields{
				"K": PointTo(MatchFields(IgnoreExtras, Fields{"C": Equal('b')})),
				"V": BeNil(),
			}),
			MatchAllFields(Fields{
				"K": PointTo(MatchFields(IgnoreExtras, Fields{"C": Equal('a')})),
				"V": BeNil(),
			}),
			MatchAllFields(Fields{
				"K": PointTo(MatchFields(IgnoreExtras, Fields{"C": Equal('c')})),
				"V": BeNil(),
			}),
		))
	})

	It("continues on error", func() {
		var remaining []string
		opts := collect(getopt.Iterate([]string{"prg", "-acb"}, "", &remaining))
		Expect(opts).To(HaveExactElements(
			MatchAllFields(Fields{
				"K": BeNil(),
				"V": MatchError("unrecognized option '-a'"),
			}),
			MatchAllFields(Fields{
				"K": BeNil(),
				"V": MatchError("unrecognized option '-c'"),
			}),
			MatchAllFields(Fields{
				"K": BeNil(),
				"V": MatchError("unrecognized option '-b'"),
			}),
		))
	})

	It("returns remaining arguments", func() {
		var remaining []string
		opts := collect(getopt.Iterate([]string{"prg", "-a", "arg1", "-b", "arg2"}, "ab", &remaining))
		Expect(opts).To(HaveLen(2))
		Expect(remaining).To(HaveExactElements("arg1", "arg2"))
	})
})

func ExampleIterate() {
	args := []string{"prg", "-a", "arg1", "-b", "arg2"}
	optionDefinition := "ab"

	var remaining []string
	for opt, err := range getopt.Iterate(args, optionDefinition, &remaining) {
		if err != nil {
			_, _ = fmt.Println(err.Error())
			break
		}
		switch opt.C {
		case 'a':
			_, _ = fmt.Println("got option a")
		case 'b':
			_, _ = fmt.Println("got option b")
		}
	}
	_, _ = fmt.Printf("Remaining arguments: %v", remaining)
	// Output:
	// got option a
	// got option b
	// Remaining arguments: [arg1 arg2]
}

func ExampleIterateLong() {
	args := []string{"prg", "-d", "--bbb", "-a"}
	optionDefinition := "ab"
	longOpts := []getopt.Option{
		{Name: "aaa", Val: 'a'},
		{Name: "bbb", Val: 'b'},
	}
	var remaining []string
	for opt, err := range getopt.IterateLong(args, optionDefinition, longOpts, &remaining) {
		if err != nil {
			_, _ = fmt.Println(err.Error())
			continue
		}
		if opt.LongInd == -1 {
			_, _ = fmt.Printf("Got short option '%c'\n", opt.C)
		} else {
			_, _ = fmt.Printf("Got long option '%s'\n", longOpts[opt.LongInd].Name)
		}
	}
	// Output:
	// unrecognized option '-d'
	// Got long option 'bbb'
	// Got short option 'a'
}
