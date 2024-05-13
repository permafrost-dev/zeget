package extraction_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/eget/lib/extraction"
)

var _ = Describe("GlobChooser", func() {
	var (
		testDir string
		err     error
	)

	BeforeEach(func() {
		testDir, err = os.MkdirTemp("", "globchooser")
		Expect(err).NotTo(HaveOccurred())

		files := []string{
			"file.txt",
			"test.jpg",
			"another_test.JPG",
			"subdir/file_in_subdir.txt",
			"subdir/another_file.png",
		}

		for _, file := range files {
			path := filepath.Join(testDir, file)
			err := os.MkdirAll(filepath.Dir(path), 0755)
			Expect(err).NotTo(HaveOccurred())
			_, err = os.Create(path)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	AfterEach(func() {
		os.RemoveAll(testDir)
	})

	Describe("NewGlobChooser", func() {
		It("creates a new GlobChooser instance", func() {
			gc, err := extraction.NewGlobChooser("*.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(gc).NotTo(BeNil())
		})
	})

	Describe("Choose", func() {
		Context("with a glob matching everything", func() {
			It("should match all files", func() {
				gc, _ := extraction.NewGlobChooser("*")
				for _, file := range []string{"file.txt", "test.jpg", "subdir/another_file.png"} {
					match, _ := gc.Choose(filepath.Join(testDir, file), false, 0)
					Expect(match).To(BeTrue())
				}
			})
		})

		Context("with a specific extension", func() {
			It("matches files with that extension", func() {
				gc, _ := extraction.NewGlobChooser("*.txt")
				match, possible := gc.Choose(filepath.Join(testDir, "file.txt"), false, 0)
				Expect(possible).To(BeTrue())

				match, _ = gc.Choose(filepath.Join(testDir, "test.svg"), false, 0)
				Expect(match).To(BeFalse())
			})
		})

		Context("with a complex pattern", func() {
			It("matches files according to the pattern", func() {
				gc, _ := extraction.NewGlobChooser("subdir/another_file.png")
				match, _ := gc.Choose("subdir/another_file.png", false, 0)
				Expect(match).To(BeTrue())

				match, _ = gc.Choose(filepath.Join(testDir, "a/b/c.1"), false, 0)
				Expect(match).To(BeFalse())
			})
		})
	})
})
