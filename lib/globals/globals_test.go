package globals_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/zeget/lib/globals"
)

var _ = Describe("Globals", func() {
	Context("ApplicationName", func() {
		It("should have the correct application name", func() {
			Expect(len(globals.ApplicationName)).To(BeNumerically(">", 1))
		})
	})

	Context("ApplicationRepository", func() {
		It("should have the correct application repository", func() {
			Expect(globals.ApplicationRepository).To(ContainSubstring("permafrost-dev/"))
		})
	})

	Context("GetApplicationName", func() {
		It("should return the correct application name", func() {
			Expect(globals.GetApplicationName()).To(ContainSubstring(globals.ApplicationName))
		})
	})

	Context("Version", func() {
		It("should have the correct version", func() {
			Expect(globals.Version).ToNot(BeEmpty())
		})
	})
})
