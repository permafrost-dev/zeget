package targetfile_test

import (
	"bytes"
	"io/ioutil"
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
	tempFile, err := ioutil.TempFile("", "targetfile_test")
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
	readData, err := ioutil.ReadAll(tempFile)
	if err != nil {
		t.Fatalf("Failed to read back data: %v", err)
	}

	if !bytes.Equal(data, readData) {
		t.Errorf("Written data and read data do not match. Got %s, want %s", readData, data)
	}
}

func TestWriteAndCleanup(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "targetfile_test")
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
