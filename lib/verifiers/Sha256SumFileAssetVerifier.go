package verifiers

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"regexp"

	"github.com/permafrost-dev/eget/lib/download"
)

type Sha256SumFileAssetVerifier struct {
	Client            *download.Client
	Sha256SumAssetURL string
	RealAssetURL      string
	BinaryName        string
}

func (s256 *Sha256SumFileAssetVerifier) WithClient(client *download.Client) *Verifier {
	s256.Client = client
	var intf interface{} = s256
	var result Verifier = intf.(Verifier)
	return &result
}

func (s256 *Sha256SumFileAssetVerifier) Verify(b []byte) error {
	got := sha256.Sum256(b)
	resp1, err := s256.Client.GetJSON(s256.Sha256SumAssetURL)
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

	resp, err := s256.Client.GetText(s256.RealAssetURL)
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

func (s256 *Sha256SumFileAssetVerifier) String() string {
	return fmt.Sprintf("checksum verified with %s", s256.Sha256SumAssetURL)
}
