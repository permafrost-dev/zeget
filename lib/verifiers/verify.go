package verifiers

import (
	"github.com/permafrost-dev/eget/lib/download"
)

type VerifyChecksumResult byte

const (
	VerifyChecksumNone               VerifyChecksumResult = iota
	VerifyChecksumSuccess            VerifyChecksumResult = iota
	VerifyChecksumVerificationFailed VerifyChecksumResult = iota
	VerifyChecksumFailedNoVerifier   VerifyChecksumResult = iota
)

type Verifier interface {
	Verify(b []byte) error
	WithClient(client *download.Client) *Verifier
}
