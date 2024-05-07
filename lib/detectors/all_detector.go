package detectors

import (
	"fmt"

	. "github.com/permafrost-dev/eget/lib/assets"
)

// AllDetector matches every asset. If there is only one asset, it is returned
// as a direct match. If there are multiple assets they are all returned as
// candidates.
type AllDetector struct{}

func (a *AllDetector) Detect(assets []Asset) (Asset, []Asset, error) {
	if len(assets) == 1 {
		return assets[0], nil, nil
	}
	return Asset{}, assets, fmt.Errorf("%d matches found", len(assets))
}
