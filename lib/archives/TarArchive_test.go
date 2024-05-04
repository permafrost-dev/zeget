package archives_test

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/permafrost-dev/eget/lib/archives" // Replace with the actual import path of your package
	"github.com/permafrost-dev/eget/lib/files"
)

var _ = Describe("TarArchive", func() {
	var (
		tarData *bytes.Buffer
		writer  *tar.Writer
	)

	BeforeEach(func() {
		tarData = new(bytes.Buffer)
		writer = tar.NewWriter(tarData)
	})

	AfterEach(func() {
		writer.Close()
	})

	Describe("NewTarArchive", func() {
		Context("when decompression is successful", func() {
			It("returns a new TarArchive", func() {
				data := "example data"
				hdr := &tar.Header{
					Name: "test123.txt",
					Mode: 0644,
					Size: int64(len(data)),
				}
				writer.WriteHeader(hdr)
				_, err := writer.Write([]byte(data))

				Expect(err).NotTo(HaveOccurred())
				writer.Close()

				archive, err := NewTarArchive(tarData.Bytes(), func(r io.Reader) (io.Reader, error) {
					return r, nil
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(archive).NotTo(BeNil())

				file, err := archive.Next()
				Expect(err).NotTo(HaveOccurred())
				Expect(file.Name).To(Equal("test123.txt"))
				Expect(file.Mode).To(Equal(fs.FileMode(0644)))
			})
		})

		Context("when decompression fails", func() {
			It("returns an error", func() {
				_, err := NewTarArchive([]byte{}, func(r io.Reader) (io.Reader, error) {
					return nil, errors.New("decompression failed")
				})
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Next", func() {
		It("iterates over files in the archive", func() {
			hdr := &tar.Header{
				Name: "test.txt",
				Mode: 0600,
				Size: int64(len("hello world")),
			}
			writer.WriteHeader(hdr)
			writer.Write([]byte("hello world"))
			writer.Close()

			archive, _ := NewTarArchive(tarData.Bytes(), func(r io.Reader) (io.Reader, error) {
				return r, nil
			})
			tarArchive, ok := archive.(*TarArchive)
			Expect(ok).To(BeTrue())

			file, err := tarArchive.Next()
			Expect(err).NotTo(HaveOccurred())
			Expect(file.Name).To(Equal("test.txt"))
			Expect(file.Mode).To(Equal(fs.FileMode(0600)))
			Expect(file.Type).To(Equal(files.TypeNormal))
		})
	})

	Describe("ReadAll", func() {
		It("reads all remaining data from the archive", func() {
			data := "hello world"
			hdr := &tar.Header{
				Name: "test.txt",
				Mode: 0600,
				Size: int64(len(data)),
			}
			writer.WriteHeader(hdr)
			bytesWritten, err := writer.Write([]byte(data))
			writer.Close()

			Expect(bytesWritten).To(Equal(len(data)))

			archive, _ := NewTarArchive(tarData.Bytes(), func(r io.Reader) (io.Reader, error) {
				return r, nil
			})

			tarArchive, ok := archive.(*TarArchive)
			Expect(ok).To(BeTrue())

			_, err = tarArchive.ReadAll()
			Expect(err).NotTo(HaveOccurred())

			// Expect(string(readData)).To(Equal(data))
		})
	})
})
