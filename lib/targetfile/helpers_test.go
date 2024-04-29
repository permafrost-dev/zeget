package targetfile_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/permafrost-dev/eget/lib/targetfile"
)

func TestNewTargetFile(t *testing.T) {
	// Test with non-empty filename
	file, err := ioutil.TempFile("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name()) // clean up

	tf := targetfile.NewTargetFile(file, file.Name(), true)
	if tf.File != file || *tf.Filename != file.Name() || !tf.ShouldClose {
		t.Errorf("NewTargetFile did not properly initialize with non-empty filename")
	}

	// Test with "-" as filename
	tf = targetfile.NewTargetFile(os.Stdout, "-", false)
	if tf.File != os.Stdout || tf.Filename != nil || tf.ShouldClose {
		t.Errorf("NewTargetFile did not properly handle '-' as filename")
	}
}

func TestGetTargetFile(t *testing.T) {
	// Test creating a new file
	tempDir, err := ioutil.TempDir("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // clean up

	filename := filepath.Join(tempDir, "newfile.txt")
	tf := targetfile.GetTargetFile(filename, 0644, false)
	if tf.Err != nil || tf == nil || *tf.Filename != filename {
		t.Errorf("GetTargetFile failed to create new file: %v", err)
	}

	// Verify the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("GetTargetFile did not create the file on disk")
	}

	// Test with "-" as filename (should use os.Stdout)
	tf = targetfile.GetTargetFile("-", 0644, false)
	if tf.Err != nil || tf.File != os.Stdout {
		t.Errorf("GetTargetFile did not properly handle '-' as filename")
	}

	// Test removing existing file
	file, err := os.CreateTemp(tempDir, "existingfile")
	if err != nil {
		t.Fatalf("Failed to create temp file for removal test: %v", err)
	}
	file.Close()

	tf = targetfile.GetTargetFile(file.Name(), 0644, true)
	if tf.Err != nil || tf == nil {
		t.Errorf("GetTargetFile failed to handle existing file removal: %v", err)
	}
}
