package extraction_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/permafrost-dev/eget/lib/extraction"
)

var _ = Describe("ExtractedFile", func() {
	var (
		extractedFile ExtractedFile
	)

	BeforeEach(func() {
		extractedFile = ExtractedFile{
			Name:        "testfile.txt",
			ArchiveName: "archive/testfile.txt",
			Extract:     func(to string) error { return nil },
			Dir:         false,
		}

		extractedFile.SetMode(0o644)
	})

	Describe("Mode", func() {
		It("returns the correct file mode", func() {
			var expectedMode os.FileMode = 0
			Expect(extractedFile.Mode()).To(Equal(expectedMode))
		})
	})

	Describe("String", func() {
		It("returns the correct archive name", func() {
			Expect(extractedFile.String()).To(Equal("archive/testfile.txt"))
		})
	})

	Describe("Extract functionality", func() {
		Context("When extracting a file", func() {
			It("should not return an error", func() {
				Expect(extractedFile.Extract("destination/path")).To(Succeed())
			})
		})
	})

	Describe("Dir flag", func() {
		Context("When the Dir flag is set", func() {
			It("indicates the ExtractedFile represents a directory", func() {
				extractedFile.Dir = true
				Expect(extractedFile.Dir).To(BeTrue())
			})
		})

		Context("When the Dir flag is not set", func() {
			It("indicates the ExtractedFile does not represent a directory", func() {
				extractedFile.Dir = false
				Expect(extractedFile.Dir).To(BeFalse())
			})
		})
	})
})
