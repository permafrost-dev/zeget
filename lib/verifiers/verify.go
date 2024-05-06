package verifiers

import (
	"go/types"

	"github.com/permafrost-dev/eget/lib/assets"
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
	GetAsset() *assets.Asset
	String() string
	types.Type
}
