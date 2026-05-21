package help_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGetoptHelpPackage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Getopt Help Suite")
}
