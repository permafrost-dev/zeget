package targetfile_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/permafrost-dev/eget/lib/targetfile"
)

func TestCleanup(t *testing.T) {
	tempFile, err := os.CreateTemp("", "targetfile_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // clean up

	filename := tempFile.Name()
	tf := &targetfile.TargetFile{
		File:        tempFile,
		Filename:    &filename,
		ShouldClose: true,
	}

	if err := tf.Cleanup(); err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}

	if tf.File != nil || tf.ShouldClose {
		t.Errorf("Cleanup did not set File to nil or ShouldClose to false")
	}
}

func TestWrite(t *testing.T) {
	tempFile, err := os.CreateTemp("", "targetfile_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // clean up

	filename := tempFile.Name()
	tf := &targetfile.TargetFile{
		File:        tempFile,
		Filename:    &filename,
		ShouldClose: false,
	}

	data := []byte("hello world")
	if err := tf.Write(data, false); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read back the data
	tempFile.Seek(0, 0) // rewind to read the file
	readData, err := io.ReadAll(tempFile)
	if err != nil {
		t.Fatalf("Failed to read back data: %v", err)
	}

	if !bytes.Equal(data, readData) {
		t.Errorf("Written data and read data do not match. Got %s, want %s", readData, data)
	}
}

func TestWriteWithError(t *testing.T) {
	tempFile, err := os.CreateTemp("", "targetfile_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // clean up

	filename := tempFile.Name()
	tf := &targetfile.TargetFile{
		File:        tempFile,
		Filename:    &filename,
		ShouldClose: false,
	}

	tf.WithError(fmt.Errorf("test error"))

	data := []byte("hello world")
	if err := tf.Write(data, false); err == nil {
		t.Fatalf("Expected error when writing with error, got nil")
	}

	if tf.Err == nil {
		t.Errorf("Expected error to be set after writing with error, got nil")
	}
}

func TestWriteAndCleanup(t *testing.T) {
	tempFile, err := os.CreateTemp("", "targetfile_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	filename := tempFile.Name()
	defer os.Remove(filename) // clean up

	tf := &targetfile.TargetFile{
		File:        tempFile,
		Filename:    &filename,
		ShouldClose: true,
	}

	data := []byte("hello cleanup")
	if err := tf.Write(data, true); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if tf.File != nil || tf.ShouldClose {
		t.Errorf("Cleanup after write did not set File to nil or ShouldClose to false")
	}
}

func TestNoFileWrite(t *testing.T) {
	filename := "nonexistent.file"
	tf := &targetfile.TargetFile{
		File:        nil,
		Filename:    &filename,
		ShouldClose: false,
	}

	err := tf.Write([]byte("data"), false)
	if err == nil {
		t.Errorf("Expected error when writing to a nil file, got nil")
	}
}

func TestHasFilename(t *testing.T) {
	filename := "nonexistent.file"
	var tf *targetfile.TargetFile

	tf = &targetfile.TargetFile{
		File:        nil,
		Filename:    &filename,
		ShouldClose: false,
	}

	if !tf.HasFilename() {
		t.Errorf("Expected HasFilename to return true, got false")
	}

	tf = &targetfile.TargetFile{
		File:        nil,
		Filename:    nil,
		ShouldClose: false,
	}

	if tf.HasFilename() {
		t.Errorf("Expected HasFilename to return false, got true")
	}
}

func TestName(t *testing.T) {
	var tf *targetfile.TargetFile
	filename := "nonexistent.file"
	tf = &targetfile.TargetFile{
		File:        nil,
		Filename:    &filename,
		ShouldClose: false,
	}
	if tf.Name() != filename {
		t.Errorf("Expected Name() to return %s, got %s", filename, tf.Name())
	}

	tf = &targetfile.TargetFile{
		File:        nil,
		Filename:    nil,
		ShouldClose: false,
	}
	if tf.Name() != "" {
		t.Errorf("Expected Name() to return '', got %s", tf.Name())
	}
}

func TestWithError(t *testing.T) {
	var tf *targetfile.TargetFile
	err := io.EOF
	tf = &targetfile.TargetFile{}
	tf.WithError(err)
	if tf.Err != err {
		t.Errorf("Expected error to be %v, got %v", err, tf.Err)
	}
}

func TestShouldNotCloseCleanup(t *testing.T) {
	tempFile, err := os.CreateTemp("", "targetfile_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // clean up

	filename := tempFile.Name()
	tf := &targetfile.TargetFile{
		File:        tempFile,
		Filename:    &filename,
		ShouldClose: false,
	}

	if err := tf.Cleanup(); err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}
}
