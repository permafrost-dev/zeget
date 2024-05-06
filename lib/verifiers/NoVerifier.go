package verifiers

import (
	"go/types"

	"github.com/permafrost-dev/eget/lib/assets"
	"github.com/permafrost-dev/eget/lib/download"
)

type NoVerifier struct {
	Asset *assets.Asset
	types.Type
}

func (n *NoVerifier) GetAsset() *assets.Asset {
	return n.Asset
}

func (n *NoVerifier) Verify(_ []byte) error {
	return nil
}

func (n *NoVerifier) WithClient(_ *download.Client) *Verifier {
	var intf interface{} = n
	var result Verifier = intf.(Verifier)
	return &result
}
