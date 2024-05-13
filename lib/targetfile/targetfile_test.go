package targetfile_test

import (
	"fmt"
	"io"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/permafrost-dev/eget/lib/targetfile"
	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"
)

var fs vfs.FS
var testFs *vfst.TestFS
var tempFile *os.File
var fsCleanup func()

var _ = Describe("TargetFile", func() {
	BeforeEach(func() {
		tempFileT := &vfst.File{
			Contents: []byte("test"),
			Perm:     0o644,
		}

		fs, fsCleanup, _ = vfst.NewTestFS(map[string]interface{}{
			"/newfile.txt": tempFileT,
		})

		testFs = fs.(*vfst.TestFS)
		tempFile, _ = testFs.Create("/newfile2.txt")
		tempFile.WriteString("test")
		tempFile.Seek(0, 0)
	})

	AfterEach(func() {
		tempFile.Close()
		fsCleanup()
	})

	It("should cleanup a targetfile", func() {
		filename := tempFile.Name()
		tf := &targetfile.TargetFile{
			File:        tempFile,
			Filename:    &filename,
			ShouldClose: true,
			Fs:          fs,
		}

		err := tf.Cleanup()
		Expect(err).To(BeNil())
		Expect(tf.File).To(BeNil())
		Expect(tf.ShouldClose).To(BeFalse())
	})

	It("should write to a targetfile", func() {
		filename := tempFile.Name()
		tf := &targetfile.TargetFile{
			File:        tempFile,
			Filename:    &filename,
			ShouldClose: false,
			Fs:          fs,
		}

		data := []byte("hello world")
		err := tf.Write(data, false)
		Expect(err).To(BeNil())

		// Read back the data
		tempFile.Seek(0, 0) // rewind to read the file
		readData, err := io.ReadAll(tempFile)
		Expect(err).To(BeNil())
		Expect(data).To(Equal(readData))
	})

	It("should write with an error to a targetfile", func() {
		filename := tempFile.Name()
		tf := &targetfile.TargetFile{
			File:        tempFile,
			Filename:    &filename,
			ShouldClose: false,
			Fs:          fs,
		}

		tf.WithError(fmt.Errorf("test error"))

		data := []byte("hello world")
		err := tf.Write(data, false)
		Expect(err).ToNot(BeNil())
		Expect(tf.Err).ToNot(BeNil())
	})

	It("should write and cleanup a targetfile", func() {
		filename := tempFile.Name()
		tf := &targetfile.TargetFile{
			File:        tempFile,
			Filename:    &filename,
			ShouldClose: true,
			Fs:          fs,
		}

		data := []byte("hello cleanup")
		err := tf.Write(data, true)
		Expect(err).To(BeNil())

		Expect(tf.File).To(BeNil())
		Expect(tf.ShouldClose).To(BeFalse())
	})

	It("should not write to a nil file", func() {
		filename := "nonexistent.file"
		tf := &targetfile.TargetFile{
			File:        nil,
			Filename:    &filename,
			ShouldClose: false,
			Fs:          fs,
		}

		err := tf.Write([]byte("data"), false)
		Expect(err).ToNot(BeNil())
	})

	It("should check if a targetfile has a filename", func() {
		filename := "nonexistent.file"
		var tf *targetfile.TargetFile

		tf = &targetfile.TargetFile{
			File:        nil,
			Filename:    &filename,
			ShouldClose: false,
			Fs:          fs,
		}

		Expect(tf.HasFilename()).To(BeTrue())

		tf = &targetfile.TargetFile{
			File:        nil,
			Filename:    nil,
			ShouldClose: false,
			Fs:          fs,
		}

		Expect(tf.HasFilename()).To(BeFalse())
	})

	It("should get the name of a targetfile", func() {
		var tf *targetfile.TargetFile
		filename := "nonexistent.file"
		tf = &targetfile.TargetFile{
			File:        nil,
			Filename:    &filename,
			ShouldClose: false,
			Fs:          fs,
		}
		Expect(tf.Name()).To(Equal(filename))

		tf = &targetfile.TargetFile{
			File:        nil,
			Filename:    nil,
			ShouldClose: false,
			Fs:          fs,
		}
		Expect(tf.Name()).To(Equal(""))
	})

	It("should set an error on a targetfile", func() {
		var tf *targetfile.TargetFile
		err := io.EOF
		tf = &targetfile.TargetFile{}
		tf.WithError(err)
		Expect(tf.Err).To(Equal(err))
	})

	It("should not close a targetfile on cleanup", func() {
		tempFile, err := os.CreateTemp("", "targetfile_test")
		if err != nil {
			Fail(fmt.Sprintf("Failed to create temp file: %v", err))
		}
		defer os.Remove(tempFile.Name()) // clean up

		filename := tempFile.Name()
		tf := &targetfile.TargetFile{
			File:        tempFile,
			Filename:    &filename,
			ShouldClose: false,
			Fs:          fs,
		}

		err = tf.Cleanup()
		Expect(err).To(BeNil())
	})

})
