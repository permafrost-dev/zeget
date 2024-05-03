package app_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/permafrost-dev/eget/app"
)

func TestBintime(t *testing.T) {
	// Create a temporary directory and file for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "testfile")
	os.WriteFile(tempFile, []byte("test content"), 0644)

	// Change modification time of the file
	newTime := time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC)
	os.Chtimes(tempFile, newTime, newTime)

	t.Run("file modification time", func(t *testing.T) {
		if got := app.Bintime("testfile", tempDir); !got.Equal(newTime) {
			t.Errorf("Bintime() = %v, want %v", got, newTime)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		if got := app.Bintime("nonexistent", tempDir); !got.IsZero() {
			t.Errorf("Expected zero time for non-existent file, got %v", got)
		}
	})
}

func TestIsUrl(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"valid http url", "http://example.com", true},
		{"valid https url", "https://example.com", true},
		{"invalid url", "://badurl", false},
		{"empty string", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := app.IsURL(tt.s); got != tt.want {
				t.Errorf("IsUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCut(t *testing.T) {
	tests := []struct {
		s     string
		sep   string
		want1 string
		want2 string
		found bool
	}{
		{"a,b,c", ",", "a", "b,c", true},
		{"abcdef", "z", "abcdef", "", false},
	}
	for _, tt := range tests {
		got1, got2, found := app.Cut(tt.s, tt.sep)
		if got1 != tt.want1 || got2 != tt.want2 || found != tt.found {
			t.Errorf("Cut() got1 = %v, want1 %v, got2 = %v, want2 %v, found = %v, wantFound %v", got1, tt.want1, got2, tt.want2, found, tt.found)
		}
	}
}

func TestIsGithubUrl(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"https://github.com/user/repo", true},
		{"https://github.com/user/repo.git", true},
		{"http://notgithub.com/user/repo", false},
		{"https://github.com/", false},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := app.IsGithubURL(tt.s); got != tt.want {
				t.Errorf("IsGithubUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsLocalFile(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	t.Run("existing file", func(t *testing.T) {
		if got := app.IsLocalFile(tempFile.Name()); !got {
			t.Errorf("IsLocalFile() = %v, want %v", got, true)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		if got := app.IsLocalFile("nonexistentfile"); got {
			t.Errorf("IsLocalFile() = %v, want %v", got, false)
		}
	})
}

func TestIsDirectory(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("existing directory", func(t *testing.T) {
		if got := app.IsDirectory(tempDir); !got {
			t.Errorf("IsDirectory() = %v, want %v", got, true)
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		if got := app.IsDirectory(filepath.Join(tempDir, "nonexistent")); got {
			t.Errorf("IsDirectory() = %v, want %v", got, false)
		}
	})
}

func TestFindChecksumAsset(t *testing.T) {
	tests := []struct {
		asset  app.Asset
		assets []app.Asset
		want   string
	}{
		{app.Asset{Name: "file", DownloadURL: ""}, []app.Asset{app.Asset{Name: "file.sha256", DownloadURL: ""}}, "file.sha256"},
		{app.Asset{Name: "file", DownloadURL: ""}, []app.Asset{app.Asset{Name: "otherfile.sha256", DownloadURL: ""}}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.asset.Name, func(t *testing.T) {
			if got := app.FindChecksumAsset(tt.asset, tt.assets); got.Name != tt.want {
				t.Errorf("FindChecksumAsset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsExec(t *testing.T) {
	tests := []struct {
		name string
		mode os.FileMode
		want bool
	}{
		{"executable file", 0755, true},
		{"executable.exe", 0644, true},
		{"executable.appimage", 0644, true},
		{"no-extension", 0655, true},
		{"nonexecutable.txt", 0600, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := app.IsExec(tt.name, tt.mode); got != tt.want {
				t.Errorf("IsExec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDefinitelyNotExec(t *testing.T) {
	tests := []struct {
		name string
		file string
		want bool
	}{
		{".deb file", "file.deb", true},
		{".1 file", "file.1", true},
		{".txt file", "file.txt", true},
		{"other file", "file", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := app.IsDefinitelyNotExec(tt.file); got != tt.want {
				t.Errorf("IsDefinitelyNotExec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModeFrom(t *testing.T) {
	tests := []struct {
		name  string
		fname string
		mode  os.FileMode
		want  os.FileMode
	}{
		{"executable-file", "file.exe", 0644, 0755},
		{"non-executable-file", "file.txt", 0644, 0644},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := app.ModeFrom(tt.fname, tt.mode); got != tt.want {
				t.Errorf("ModeFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRename(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		nameguess string
		want      string
	}{
		{"appimage file", "file.appimage", "file", "file"},
		{"exe file", "file.exe", "file", "file.exe"},
		{"other file", "file.txt", "file", "file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := app.GetRename(tt.file, tt.nameguess); got != tt.want {
				t.Errorf("GetRename() = %v, want %v", got, tt.want)
			}
		})
	}
}
