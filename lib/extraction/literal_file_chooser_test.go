package extraction_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/permafrost-dev/zeget/lib/extraction"
)

var _ = Describe("LiteralFileChooser", func() {
	var chooser *extraction.LiteralFileChooser

	BeforeEach(func() {
		chooser = &extraction.LiteralFileChooser{File: "target.txt"}
	})

	Describe("Choosing files", func() {
		Context("When the file matches exactly", func() {
			It("should choose the file", func() {
				selected, continueWalk := chooser.Choose("path/to/target.txt", false, 0)
				Expect(selected).To(BeFalse())
				Expect(continueWalk).To(BeTrue())
			})
		})

		Context("When the file does not match", func() {
			It("should not choose the file", func() {
				selected, continueWalk := chooser.Choose("path/to/not_target.txt", false, 0)
				Expect(selected).To(BeFalse())
				Expect(continueWalk).To(BeFalse())
			})
		})

		Context("When the file matches but in a different directory", func() {
			It("should choose the file", func() {
				selected, continueWalk := chooser.Choose("another/path/to/target.txt", false, 0)
				Expect(selected).To(BeFalse())
				Expect(continueWalk).To(BeTrue())
			})
		})

		Context("When the base name matches but the suffix does not", func() {
			It("should not choose the file", func() {
				selected, continueWalk := chooser.Choose("path/to/target.jpg", false, 0)
				Expect(selected).To(BeFalse())
				Expect(continueWalk).To(BeFalse())
			})
		})
	})

	Describe("String representation", func() {
		It("should return a formatted string", func() {
			str := chooser.String()
			Expect(str).To(Equal("`target.txt`"))
		})
	})
})
