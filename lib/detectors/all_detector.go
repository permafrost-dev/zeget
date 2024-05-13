package detectors

import (
	"fmt"

	. "github.com/permafrost-dev/eget/lib/assets"
)

// AllDetector matches every asset. If there is only one asset, it is returned
// as a direct match. If there are multiple assets they are all returned as
// candidates.
type AllDetector struct{}

func (a *AllDetector) Detect(assets []Asset) (DetectionResult, error) {
	if len(assets) == 1 {
		return NewDetectionResult(&assets[0], nil), nil
	}
	return NewDetectionResult(&Asset{}, assets), fmt.Errorf("%d matches found", len(assets))
}
