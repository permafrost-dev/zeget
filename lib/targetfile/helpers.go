package targetfile

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/twpayne/go-vfs/v5"
)

// NewTargetFile returns a new TargetFile instance for the given file.
// If the filename is "-", then:
//   - result.File will be os.Stdout
//   - result.shouldClose will be false.
//   - result.Filename will be nil.
//
// If the filename is empty, then:
//   - result.File will be the given file
//   - result.shouldClose will be false.
//   - result.Filename will be nil.
func NewTargetFile(fs vfs.FS, file *os.File, filename string, shouldClose bool) *TargetFile {
	fn := &filename

	if filename == "" || filename == "-" {
		fn = nil
		shouldClose = false
	}

	return &TargetFile{
		File:        file,
		Filename:    fn,
		ShouldClose: shouldClose,
		Err:         nil,
		Fs:          &fs,
	}
}

// GetTargetFile returns a new TargetFile instance for the given filename. If the filename is "-", the file will be os.Stdout.
// The file will be opened with the given mode; result.Cleanup() should be called to close the file.
// If removeExisting is true, the existing file will be removed before creating a new one.
// If the target directory does not exist, it will be created recursively with mode 0755.
func GetTargetFile(fs vfs.FS, filename string, mode fs.FileMode, removeExisting bool) (tf *TargetFile) {
	if filename == "-" {
		return NewTargetFile(fs, os.Stdout, filename, false)
	}

	if removeExisting {
		fs.Remove(filename)
	}

	fs.Mkdir(filepath.Dir(filename), 0755)

	file, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	// if err != nil {
	// 	return NewTargetFile(fs, nil, "", false).WithError(err)
	// }

	return NewTargetFile(fs, file, filename, err != nil).WithError(err)

	// return NewTargetFile(fs, file, filename, true)
}
