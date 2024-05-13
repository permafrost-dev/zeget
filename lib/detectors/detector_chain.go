package detectors

import (
	"fmt"

	. "github.com/permafrost-dev/eget/lib/assets"
)

type DetectorChain struct {
	detectors []Detector
	system    Detector
}

func (dc *DetectorChain) Detect(assets []Asset) (DetectionResult, error) {
	for _, d := range dc.detectors {
		detected, err := d.Detect(assets)
		if len(detected.Candidates) == 0 && err != nil {
			return DetectionResult{}, err
		}
		if len(detected.Candidates) == 0 {
			return detected, nil
		}
		assets = detected.Candidates
	}

	detected, err := dc.system.Detect(assets)
	if len(detected.Candidates) == 0 && err != nil {
		return DetectionResult{}, err
	}

	return detected, fmt.Errorf("%d candidates found for asset chain", len(assets))
}
