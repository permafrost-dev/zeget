package archives_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestArchives(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Archives Suite")
}
