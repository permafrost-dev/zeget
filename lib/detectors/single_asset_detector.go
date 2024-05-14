package detectors

import (
	"fmt"
	"path"
	"strings"

	. "github.com/permafrost-dev/zeget/lib/assets"
)

// SingleAssetDetector finds a single named asset. If Anti is true it finds all
// assets that don't contain Asset.
type SingleAssetDetector struct {
	Asset string
	Anti  bool
}

func (s *SingleAssetDetector) Detect(assets []Asset) (DetectionResult, error) {
	var candidates []Asset
	for _, a := range assets {
		if !s.Anti && path.Base(a.Name) == s.Asset {
			return NewDetectionResult(&a, nil), nil
		}
		if !s.Anti && strings.Contains(path.Base(a.Name), s.Asset) {
			candidates = append(candidates, a)
		}
		if s.Anti && path.Base(a.Name) != s.Asset && len(assets) == 2 {
			return NewDetectionResult(&a, nil), nil
		}
		if s.Anti && !strings.Contains(path.Base(a.Name), s.Asset) {
			candidates = append(candidates, a)
		}
	}

	if len(candidates) == 1 {
		return NewDetectionResult(&candidates[0], nil), nil
	}

	if len(candidates) > 1 {
		return NewDetectionResult(&Asset{}, candidates), fmt.Errorf("%d candidates found for asset `%s`", len(candidates), s.Asset)
	}

	return NewDetectionResult(&Asset{}, nil), fmt.Errorf("asset `%s` not found", s.Asset)
}
