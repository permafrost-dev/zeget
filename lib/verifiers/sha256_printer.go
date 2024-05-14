package verifiers

import (
	"crypto/sha256"
	"fmt"
	"go/types"

	"github.com/permafrost-dev/eget/lib/assets"
	"github.com/permafrost-dev/eget/lib/download"
)

type Sha256Printer struct {
	Asset *assets.Asset
	types.Type
}

func (s256 Sha256Printer) GetAsset() *assets.Asset {
	return s256.Asset
}

func (s256 Sha256Printer) WithClient(_ download.ClientContract) Verifier {
	return s256
}

func (s256 Sha256Printer) Verify(b []byte) error {
	sum := sha256.Sum256(b)
	fmt.Printf("%x\n", sum)
	return nil
}

func (s256 Sha256Printer) String() string {
	return "sha256:print"
}
