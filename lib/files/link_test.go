package files_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/permafrost-dev/eget/lib/files"
)

var _ = Describe("Link", func() {
	var (
		link *files.Link
	)

	BeforeEach(func() {
		t := GinkgoT()
		t.Helper()

		fileSystem, cleanup, err := vfst.NewTestFS(map[string]interface{}{
			"/one": &vfst.File{
				Contents: []byte("test"),
				Perm:     0o644,
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		link = files.NewLink("/one", "/two")
		link.Fs = fileSystem
		link.Cleanup = cleanup
	})

	AfterEach(func() {
		defer link.Cleanup()
	})

	Describe("Creating a new link", func() {
		Context("with valid parameters", func() {
			It("should create a Link object", func() {
				Expect(link).NotTo(BeNil())
				Expect(link.Oldname).To(Equal("/one"))
				Expect(link.Newname).To(Equal("/two"))
				Expect(link.Sym).To(BeTrue())
			})
		})
	})

	Describe("Writing a link", func() {
		Context("with an initialized filesystem", func() {
			It("should not return an error", func() {
				err := link.Write()
				fmt.Printf("error: %v\n", err)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should create a new symlink", func() {
				_ = link.Write()
				_, err := link.Fs.Stat("/two")
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
