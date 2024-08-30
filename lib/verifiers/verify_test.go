package verifiers_test

import (
	"testing"

	"github.com/permafrost-dev/zeget/lib/download"
	. "github.com/permafrost-dev/zeget/lib/verifiers"
)

func TestNoVerifier_Verify(t *testing.T) {
	nv := &NoVerifier{}
	err := nv.Verify([]byte("test"))
	if err != nil {
		t.Errorf("NoVerifier.Verify() error = %v, wantErr %v", err, nil)
	}
}

func TestNewSha256Verifier(t *testing.T) {
	client := download.NewClient("")
	// Valid SHA256 hex string (64 characters long)
	validHex := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	_, err := NewSha256Verifier(client, validHex)
	if err != nil {
		t.Errorf("NewSha256Verifier() error = %v, wantErr %v", err, nil)
	}

	// Invalid SHA256 hex string (not 64 characters)
	invalidHex := "12345"
	_, err = NewSha256Verifier(client, invalidHex)
	if err == nil {
		t.Errorf("NewSha256Verifier() error = %v, wantErr %v", nil, "error expected")
	}
}

func TestSha256Verifier_Verify(t *testing.T) {
	client := download.NewClient("")

	// Assuming the expected hex corresponds to the hash of "test" input
	expectedHex := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	s256, _ := NewSha256Verifier(client, expectedHex)
	err := s256.Verify([]byte("test"))
	if err != nil {
		t.Errorf("Sha256Verifier.Verify() error = %v, wantErr %v", err, nil)
	}

	// Test with incorrect input
	err = s256.Verify([]byte("wrong"))
	if err == nil {
		t.Errorf("Sha256Verifier.Verify() got no error, want error")
	}
}
