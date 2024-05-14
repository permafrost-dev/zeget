package detectors

import "github.com/permafrost-dev/zeget/lib/assets"

type DetectionResult struct {
	Asset      assets.Asset
	Candidates []assets.Asset
}

func NewDetectionResult(asset *assets.Asset, candidates []assets.Asset) DetectionResult {
	if asset == nil {
		asset = &assets.Asset{}
	}

	if len(candidates) == 1 {
		return DetectionResult{Asset: candidates[0], Candidates: nil}
	}

	return DetectionResult{
		Asset:      *asset,
		Candidates: candidates,
	}
}
