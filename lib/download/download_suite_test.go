package download_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDownloadPackage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "test suite: lib/download package")
}
