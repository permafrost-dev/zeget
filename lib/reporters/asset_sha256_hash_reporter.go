package reporters

import (
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/permafrost-dev/eget/lib/assets"
)

type AssetSha256HashReporter struct {
	Asset  *assets.Asset
	Output io.Writer
}

func (r *AssetSha256HashReporter) Report(input ...interface{}) error {
	var value string = input[0].(string)
	checksum := sha256.Sum256([]byte(value))

	fmt.Fprintf(r.Output, "› %x %s\n", checksum, r.Asset.Name)

	return nil
}

func NewAssetSha256HashReporter(asset *assets.Asset, output io.Writer) *AssetSha256HashReporter {
	return &AssetSha256HashReporter{
		Asset:  asset,
		Output: output,
	}
}
