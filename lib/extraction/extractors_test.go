package extraction_test

import (
	"fmt"
	"io/fs"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/zeget/lib/extraction"
	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"
)

type MockChooser struct {
	ChooseFn func(name string, dir bool, mode fs.FileMode) (direct bool, possible bool)
}

func (mc *MockChooser) Choose(name string, dir bool, mode fs.FileMode) (direct bool, possible bool) {
	if mc.ChooseFn != nil {
		return mc.ChooseFn(name, dir, mode)
	}
	return false, false
}

var _ = Describe("Extractor", func() {
	var testFS *vfst.TestFS
	var cleanup func()
	var err error

	BeforeEach(func() {
		testFS, cleanup, err = vfst.NewTestFS(nil)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		cleanup()
	})

	// var extractor *extraction.SingleFileExtractor

	Describe("NewExtractor", func() {
		It("should create an ArchiveExtractor for .tar.gz files", func() {
			extractor := extraction.NewExtractor(testFS, "test.tar.gz", "", nil)
			Expect(extractor).To(BeAssignableToTypeOf(&extraction.ArchiveExtractor{}))
		})

		It("should create a SingleFileExtractor for .gz files", func() {
			extractor := extraction.NewExtractor(testFS, "test.gz", "", nil)
			Expect(extractor).To(BeAssignableToTypeOf(&extraction.SingleFileExtractor{}))
		})
	})

	Describe("SingleFileExtractor", func() {
		Context("when extracting a gzip file", func() {
			It("should successfully extract the content", func() {
				wd, _ := os.Getwd()
				testArchiveFn := fmt.Sprintf("test archive: %s/../../test/test-config-toml.gz", wd)
				buf, err := os.ReadFile(testArchiveFn)

				extractor := extraction.NewExtractor(vfs.OSFS, testArchiveFn, "test-config-toml", nil)
				ef, _, err := extractor.Extract(buf, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(ef.Name).To(Equal("test-config-toml"))

				defer ef.Extract(ef.Name)
			})
		})

	})
})
