package files_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/permafrost-dev/eget/lib/files"
)

var _ = Describe("File", func() {
	It("should return a File with the correct values", func() {
		file := files.File{
			Name:     "test",
			LinkName: "testlink",
			Mode:     0o644,
			Type:     files.TypeNormal,
		}
		Expect(file.Name).To(Equal("test"))
		Expect(file.LinkName).To(Equal("testlink"))
		Expect(file.Type).To(Equal(files.TypeNormal))
	})

	It("Should return true for a directory", func() {
		file := files.File{
			Name:     "test",
			LinkName: "testlink",
			Mode:     0o644,
			Type:     files.TypeDir,
		}
		Expect(file.Dir()).To(BeTrue())
	})

	It("Should return false for a non-directory", func() {
		file := files.File{
			Name:     "test",
			LinkName: "testlink",
			Mode:     0o644,
			Type:     files.TypeNormal,
		}
		Expect(file.Dir()).To(BeFalse())
	})

})
