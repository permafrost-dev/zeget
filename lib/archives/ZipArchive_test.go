package archives_test

import (
	"archive/zip"
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/permafrost-dev/zeget/lib/archives"
	"github.com/permafrost-dev/zeget/lib/files"
)

var _ = Describe("ZipArchive", func() {
	var (
		zipBytes []byte
		err      error
		archive  Archive
	)

	BeforeEach(func() {
		// Create a new zip archive in memory
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)

		var files = []struct {
			Name, Body string
		}{
			{"readme.txt", "This archive contains some text files."},
			{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
			{"todo.txt", "Get animal handling licence.\nWrite more examples."},
		}

		for _, file := range files {
			f, err := w.Create(file.Name)
			Expect(err).NotTo(HaveOccurred())
			_, err = f.Write([]byte(file.Body))
			Expect(err).NotTo(HaveOccurred())
		}

		err := w.Close()
		Expect(err).NotTo(HaveOccurred())

		zipBytes = buf.Bytes()
	})

	Describe("NewZipArchive", func() {
		It("should create a new ZipArchive successfully", func() {
			archive, err = NewZipArchive(zipBytes, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(archive).NotTo(BeNil())
		})
	})

	Describe("Next", func() {
		It("should iterate over files correctly", func() {
			archive, _ = NewZipArchive(zipBytes, nil)
			file, err := archive.Next()
			Expect(err).NotTo(HaveOccurred())
			Expect(file.Name).To(Equal("readme.txt"))
			Expect(file.Type).To(Equal(files.TypeNormal))

			file, err = archive.Next()
			Expect(err).NotTo(HaveOccurred())
			Expect(file.Name).To(Equal("gopher.txt"))
		})

		It("should return EOF after the last file", func() {
			archive, _ = NewZipArchive(zipBytes, nil)
			for {
				_, err := archive.Next()
				if err != nil {
					Expect(err).To(Equal(fmt.Errorf("EOF")))
					break
				}
			}
		})
	})

	Describe("ReadAll", func() {
		It("should read file contents correctly", func() {
			archive, _ = NewZipArchive(zipBytes, nil)
			_, _ = archive.Next() // Move to the first file
			data, err := archive.ReadAll()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("This archive contains some text files."))
		})
	})
})
