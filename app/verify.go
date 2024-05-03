package app

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
)

type Verifier interface {
	Verify(b []byte) error
}

type NoVerifier struct{}

func (n *NoVerifier) Verify(b []byte) error {
	return nil
}

type Sha256Error struct {
	Expected []byte
	Got      []byte
}

func (e *Sha256Error) Error() string {
	return fmt.Sprintf("sha256 checksum mismatch:\nexpected: %x\ngot:      %x", e.Expected, e.Got)
}

type Sha256Verifier struct {
	Expected []byte
}

func NewSha256Verifier(expectedHex string) (*Sha256Verifier, error) {
	expected, _ := hex.DecodeString(expectedHex)
	if len(expected) != sha256.Size {
		return nil, fmt.Errorf("sha256sum (%s) too small: %d bytes decoded", expectedHex, len(expectedHex))
	}
	return &Sha256Verifier{
		Expected: expected,
	}, nil
}

func (s256 *Sha256Verifier) Verify(b []byte) error {
	sum := sha256.Sum256(b)
	if bytes.Equal(sum[:], s256.Expected) {
		return nil
	}
	return &Sha256Error{
		Expected: s256.Expected,
		Got:      sum[:],
	}
}

type Sha256Printer struct{}

func (s256 *Sha256Printer) Verify(b []byte) error {
	sum := sha256.Sum256(b)
	fmt.Printf("%x\n", sum)
	return nil
}

type Sha256AssetVerifier struct {
	AssetURL string
}

func (s256 *Sha256AssetVerifier) Verify(b []byte) error {
	resp, err := GetJson(s256.AssetURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	expected := make([]byte, sha256.Size)

	n, _ := hex.Decode(expected, data)
	if n < sha256.Size {
		return &Sha256Error{
			Expected: expected[:n],
			Got:      []byte{0},
		}
		// return fmt.Errorf("sha256sum (%s) too small: %d bytes decoded", string(data), n)
	}
	sum := sha256.Sum256(b)
	if bytes.Equal(sum[:], expected[:n]) {
		return nil
	}
	return &Sha256Error{
		Expected: expected[:n],
		Got:      sum[:],
	}
}

func (n *Sha256AssetVerifier) String() string {
	return fmt.Sprintf("checksum verified with %s", n.AssetURL)
}

type Sha256SumFileAssetVerifier struct {
	Sha256SumAssetURL string
	RealAssetURL      string
	BinaryName        string
}

func (s256 *Sha256SumFileAssetVerifier) Verify(b []byte) error {
	got := sha256.Sum256(b)
	resp1, err := GetJson(s256.Sha256SumAssetURL)
	if err != nil {
		return err
	}
	defer resp1.Body.Close()

	// follow the "redirect" in the JSON provided by "browser_download_url":
	body, _ := io.ReadAll(resp1.Body)
	urlpattern := regexp.MustCompile(`"(https://github.com/[\w-_]+/[\w-_]+/releases/download/.+)"`)
	downloadMatch := urlpattern.FindStringSubmatch(string(body))
	if downloadMatch == nil {
		return fmt.Errorf("no download url found in %s", s256.Sha256SumAssetURL)
	}

	s256.RealAssetURL = downloadMatch[1]

	resp, err := GetText(s256.RealAssetURL)
	if err != nil {
		return err
	}

	expectedFound := false
	scanner := bufio.NewScanner(resp.Body)
	sha256sumLinePattern := regexp.MustCompile(fmt.Sprintf("(%x)\\s+([\\w_\\-\\.]+)", got)) //, s256.BinaryName
	for scanner.Scan() {
		line := scanner.Bytes()
		matches := sha256sumLinePattern.FindStringSubmatch(string(line))
		if matches == nil {
			continue
		}
		expectedFound = true
		break
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read sha256sum %s: %w", s256.Sha256SumAssetURL, err)
	}
	if !expectedFound {
		return &Sha256Error{
			Got: got[:],
		}
	}
	return nil
}

func (n *Sha256SumFileAssetVerifier) String() string {
	return fmt.Sprintf("checksum verified with %s", n.Sha256SumAssetURL)
}
