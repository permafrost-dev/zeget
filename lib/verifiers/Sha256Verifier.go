package verifiers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/permafrost-dev/eget/lib/download"
)

type Sha256Verifier struct {
	Expected []byte
	client   *download.Client
}

func NewSha256Verifier(client *download.Client, expectedHex string) (*Sha256Verifier, error) {
	expected, _ := hex.DecodeString(expectedHex)
	if len(expected) != sha256.Size {
		return nil, fmt.Errorf("sha256sum (%s) too small: %d bytes decoded", expectedHex, len(expectedHex))
	}
	return &Sha256Verifier{
		client:   client,
		Expected: expected,
	}, nil
}

func (s256 *Sha256Verifier) WithClient(client *download.Client) *Verifier {
	s256.client = client
	var intf interface{} = s256
	var result Verifier = intf.(Verifier)
	return &result
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
