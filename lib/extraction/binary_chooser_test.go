package extraction_test

import (
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/eget/lib/extraction"
)

var _ = Describe("BinaryChooser", func() {
	var chooser *extraction.BinaryChooser
	var toolName string = "testTool"

	BeforeEach(func() {
		chooser = &extraction.BinaryChooser{Tool: toolName}
	})

	Describe("Choosing files", func() {
		Context("When the file is a directory", func() {
			It("should return false, false", func() {
				direct, possible := chooser.Choose("someDir", true, 0)
				Expect(direct).To(BeFalse())
				Expect(possible).To(BeFalse())
			})
		})

		Context("When the file is an exact match", func() {
			It("should return true, true for a normal executable", func() {
				mode := fs.FileMode(0755)
				direct, possible := chooser.Choose(toolName, false, mode)
				Expect(direct).To(BeTrue())
				Expect(possible).To(BeTrue())
			})

			It("should return true, true for a .exe executable", func() {
				mode := fs.FileMode(0755)
				direct, possible := chooser.Choose(toolName+".exe", false, mode)
				Expect(direct).To(BeTrue())
				Expect(possible).To(BeTrue())
			})

			It("should return true, true for a .appimage executable", func() {
				mode := fs.FileMode(0755)
				direct, possible := chooser.Choose(toolName+".appimage", false, mode)
				Expect(direct).To(BeTrue())
				Expect(possible).To(BeTrue())
			})
		})

		Context("When the file is not a match but is executable", func() {
			It("should return false, true", func() {
				mode := fs.FileMode(0755)
				direct, possible := chooser.Choose("someOtherTool", false, mode)
				Expect(direct).To(BeFalse())
				Expect(possible).To(BeTrue())
			})
		})

		Context("When the file is not executable", func() {
			It("should return false, false", func() {
				mode := fs.FileMode(0644) // Not executable
				direct, possible := chooser.Choose("someOther.Tool", false, mode)
				Expect(direct).To(BeFalse())
				Expect(possible).To(BeFalse())
			})
		})
	})

	Describe("String representation", func() {
		It("should return the correct string representation", func() {
			Expect(chooser.String()).To(Equal("exe `testTool`"))
		})
	})
})
