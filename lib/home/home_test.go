package home_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/permafrost-dev/eget/lib/home"
)

func TestHome(t *testing.T) {
	expectedHomeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("getting user home dir for test setup failed: %v", err)
	}

	homeDir, err := home.Home()
	if err != nil {
		t.Errorf("Home() error = %v, wantErr = false", err)
		return
	}
	if homeDir != expectedHomeDir {
		t.Errorf("Home() = %v, want %v", homeDir, expectedHomeDir)
	}
}

func TestExpand(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("getting user home dir for test setup failed: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "With tilde",
			path:    "~/test",
			want:    filepath.Join(homeDir, "test"),
			wantErr: false,
		},
		{
			name:    "Without tilde",
			path:    "/test",
			want:    "/test",
			wantErr: false,
		},
		// This test assumes running as a non-root user; might need adjustments based on test environment
		{
			name:    "Tilde with username",
			path:    "~nonexistentuser/test",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := home.Expand(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Expand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expand() = %v, want %v", got, tt.want)
			}
		})
	}
}
