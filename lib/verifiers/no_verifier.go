package verifiers

import (
	"github.com/permafrost-dev/eget/lib/assets"
	"github.com/permafrost-dev/eget/lib/download"
)

type NoVerifier struct {
	Asset *assets.Asset
	Verifier
	// types.Type
}

func (n NoVerifier) GetAsset() *assets.Asset {
	return n.Asset
}

func (n NoVerifier) Verify(_ []byte) error {
	return nil
}

func (n NoVerifier) WithClient(_ download.ClientContract) Verifier {
	return n
}

func (n NoVerifier) String() string {
	return "NoVerifier"
}
