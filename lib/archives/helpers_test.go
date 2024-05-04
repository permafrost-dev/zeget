package archives_test

import (
	"archive/tar"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/eget/lib/archives"
	"github.com/permafrost-dev/eget/lib/files"
)

var _ = Describe("Helpers", func() {
	Describe("TarFileType function", func() {
		It("should return TypeNormal for tar.TypeReg", func() {
			Expect(archives.TarFileType(tar.TypeReg)).To(Equal(files.TypeNormal))
		})

		It("should return TypeDir for tar.TypeDir", func() {
			Expect(archives.TarFileType(tar.TypeDir)).To(Equal(files.TypeDir))
		})

		It("should return TypeLink for tar.TypeLink", func() {
			Expect(archives.TarFileType(tar.TypeLink)).To(Equal(files.TypeLink))
		})

		It("should return TypeSymlink for tar.TypeSymlink", func() {
			Expect(archives.TarFileType(tar.TypeSymlink)).To(Equal(files.TypeSymlink))
		})

		It("should return TypeOther for undefined tar types", func() {
			Expect(archives.TarFileType(0)).To(Equal(files.TypeOther))
		})
	})
})
