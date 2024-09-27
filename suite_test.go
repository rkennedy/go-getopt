package getopt_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGetoptPackage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Getopt Suite")
}
