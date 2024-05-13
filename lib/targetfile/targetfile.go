package targetfile

import (
	"errors"
	"os"

	"github.com/twpayne/go-vfs/v5"
)

var ErrInvalidated = errors.New("target file has been invalidated")

type TargetFile struct {
	File        *os.File
	Filename    *string
	ShouldClose bool
	Err         error
	Fs          vfs.FS
}

// Cleanup closes the file if it should be closed and sets the file to nil.
// The TargetFile instance should be considered invalid after calling this method.
func (tf *TargetFile) Cleanup() error {
	if !tf.ShouldClose || tf.IsInvalid() {
		return nil
	}

	if err := tf.File.Close(); err != nil {
		return err
	}

	tf.Invalidate()

	return nil
}

// Write writes the given data to the target file. If cleanup is true, the file will be closed after
// writing the data and the TargetFile instance should be considered invalid.
func (tf *TargetFile) Write(data []byte, cleanup bool) error {
	if tf.HasError() {
		return tf.Err
	}

	if tf.IsInvalid() {
		return ErrInvalidated
	}

	_, err := tf.File.Write(data)

	if cleanup {
		defer tf.Cleanup()
	}

	return err
}

func (tf *TargetFile) WithError(err error) *TargetFile {
	tf.Err = err

	return tf
}

func (tf *TargetFile) Invalidate() {
	tf.File = nil
	tf.ShouldClose = false
}

func (tf *TargetFile) IsInvalid() bool {
	return tf.File == nil
}

func (tf *TargetFile) HasError() bool {
	return tf.Err != nil
}

func (tf *TargetFile) HasFilename() bool {
	return tf.Filename != nil
}

func (tf *TargetFile) Name() string {
	if tf.Filename == nil {
		return ""
	}

	return *tf.Filename
}
