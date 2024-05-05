package verifiers

import "github.com/permafrost-dev/eget/lib/download"

type NoVerifier struct{}

func (n *NoVerifier) Verify(_ []byte) error {
	return nil
}

func (n *NoVerifier) WithClient(_ *download.Client) *Verifier {
	var intf interface{} = n
	var result Verifier = intf.(Verifier)
	return &result
}
