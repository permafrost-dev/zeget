package verifiers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/permafrost-dev/eget/lib/assets"
	"github.com/permafrost-dev/eget/lib/download"
)

type Sha256AssetVerifier struct {
	client   download.ClientContract
	AssetURL string
	Asset    *assets.Asset
	Verifier
}

func (s256 *Sha256AssetVerifier) GetAsset() *assets.Asset {
	return s256.Asset
}

func (s256 *Sha256AssetVerifier) WithClient(client download.ClientContract) Verifier {
	s256.client = client
	return s256
}

func (s256 *Sha256AssetVerifier) Verify(b []byte) error {
	resp, err := s256.client.GetJSON(s256.AssetURL)

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

func (s256 *Sha256AssetVerifier) String() string {
	return fmt.Sprintf("checksum verified with %s", s256.AssetURL)
}
