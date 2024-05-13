package targetfile_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/eget/lib/targetfile"
	"github.com/twpayne/go-vfs/v5/vfst"
)

var _ = Describe("TargetFile Helpers", func() {
	It("should create a new targetfile", func() {
		fs, cleanup, _ := vfst.NewTestFS(map[string]interface{}{
			"/newfile.txt": &vfst.File{
				Contents: []byte("test"),
				Perm:     0o644,
			},
		})
		defer cleanup()

		file, _ := fs.OpenFile("/newfile.txt", os.O_RDWR, 0o644)
		defer file.Close()

		tf := targetfile.NewTargetFile(fs, file, file.Name(), true)
		Expect(tf.File).To(Equal(file))
		Expect(*tf.Filename).To(Equal(file.Name()))
		Expect(tf.ShouldClose).To(BeTrue())
	})

	It("should get a targetfile", func() {
		fs, cleanup, _ := vfst.NewTestFS(map[string]interface{}{
			"/newfile.txt": &vfst.File{
				Contents: []byte("test"),
				Perm:     0o644,
			},
		})
		defer cleanup()

		tf := targetfile.GetTargetFile(fs, "/newfile.txt", 0644, false)
		Expect(tf.Err).To(BeNil())
		Expect(*tf.Filename).To(Equal("/newfile.txt"))

		_, err := fs.Stat("/newfile.txt")
		Expect(err).To(BeNil())

		tf = targetfile.GetTargetFile(fs, "-", 0644, false)
		Expect(tf.Err).To(BeNil())
		Expect(tf.File).To(Equal(os.Stdout))
	})

	It("should remove an existing targetfile", func() {
		fs, cleanup, _ := vfst.NewTestFS(map[string]interface{}{
			"/newfile.txt": &vfst.File{
				Contents: []byte("test"),
				Perm:     0o644,
			},
		})
		defer cleanup()

		tf := targetfile.GetTargetFile(fs, "/newfile.txt", 0644, true)
		Expect(tf.Err).To(BeNil())

		_, err := fs.Stat("/newfile.txt")
		Expect(err).To(BeNil())
	})
})

// func TestNewTargetFile(t *testing.T) {
// 	// Test with non-empty filename
// 	// file, err := os.CreateTemp("", "testfile")
// 	// if err != nil {
// 	// 	t.Fatalf("Failed to create temp file: %v", err)
// 	// }
// 	// defer os.Remove(file.Name()) // clean up

// 	fn := "/newfile.txt"
// 	fs, cleanup, _ := vfst.NewTestFS(map[string]interface{}{
// 		fn: &vfst.File{
// 			Contents: []byte("test"),
// 			Perm:     0o644,
// 		},
// 	})
// 	defer cleanup()

// 	file, _ := fs.OpenFile(fn, os.O_RDWR, 0o644)
// 	defer file.Close()

// 	tf := targetfile.NewTargetFile(fs, file, file.Name(), true)
// 	if tf.File != file || *tf.Filename != file.Name() || !tf.ShouldClose {
// 		t.Errorf("NewTargetFile did not properly initialize with non-empty filename")
// 	}

// 	// Test with "-" as filename
// 	tf = targetfile.NewTargetFile(fs, os.Stdout, "-", false)
// 	if tf.File != os.Stdout || tf.Filename != nil || tf.ShouldClose {
// 		t.Errorf("NewTargetFile did not properly handle '-' as filename")
// 	}
// }

// func TestGetTargetFile(t *testing.T) {
// 	filename := "/newfile.txt"
// 	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
// 		filename: &vfst.File{
// 			Contents: []byte("test"),
// 			Perm:     0o644,
// 		},
// 	})
// 	defer cleanup()

// 	tf := targetfile.GetTargetFile(fs, filename, 0644, false)
// 	if tf.Err != nil || tf == nil || *tf.Filename != filename {
// 		t.Errorf("GetTargetFile failed to create new file: %v", tf)
// 	}

// 	// Verify the file exists
// 	if _, err := fs.Stat(filename); os.IsNotExist(err) {
// 		t.Errorf("GetTargetFile did not create the file on disk")
// 	}

// 	// Test with "-" as filename (should use os.Stdout)
// 	tf = targetfile.GetTargetFile(fs, "-", 0644, false)
// 	if tf.Err != nil || tf.File != os.Stdout {
// 		t.Errorf("GetTargetFile did not properly handle '-' as filename")
// 	}

// 	fs.Remove(filename)

// 	tf = targetfile.GetTargetFile(fs, filename, 0644, true)
// 	if tf.Err != nil || tf == nil {
// 		t.Errorf("GetTargetFile failed to handle existing file removal: %v", err)
// 	}
// }
