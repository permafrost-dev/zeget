package detectors

import (
	"fmt"

	. "github.com/permafrost-dev/zeget/lib/assets"
)

type DetectorChain struct {
	Detectors []Detector
	System    Detector
}

func (dc *DetectorChain) DetectWithoutSystem(assets []Asset) (*DetectionResult, error) {
	for _, d := range dc.Detectors {
		if d == nil {
			continue
		}

		detected, err := d.Detect(assets)
		if len(detected.Candidates) == 0 && err != nil {
			return nil, err
		}
		if len(detected.Candidates) == 0 {
			return &detected, nil
		}
		assets = detected.Candidates
	}

	ptr := NewDetectionResult(nil, assets)
	return &ptr, nil
}

func (dc *DetectorChain) Detect(assets []Asset) (DetectionResult, error) {
	for _, d := range dc.Detectors {
		if d == nil {
			continue
		}

		detected, err := d.Detect(assets)
		if len(detected.Candidates) == 0 && err != nil {
			return DetectionResult{}, err
		}
		if len(detected.Candidates) == 0 {
			return detected, nil
		}
		assets = detected.Candidates
	}

	if dc.System == nil {
		return DetectionResult{}, nil
	}

	detected, err := dc.System.Detect(assets)
	if len(detected.Candidates) == 0 && err != nil {
		return DetectionResult{}, err
	}

	return detected, fmt.Errorf("%d candidates found for asset chain", len(assets))
}
