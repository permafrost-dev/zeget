package verifiers

import (
	"go/types"

	"github.com/permafrost-dev/zeget/lib/assets"
	"github.com/permafrost-dev/zeget/lib/download"
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
	WithClient(client download.ClientContract) Verifier
	GetAsset() *assets.Asset
	String() string
	types.Type
}
