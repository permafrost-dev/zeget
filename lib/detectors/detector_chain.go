package detectors

import (
	"fmt"

	. "github.com/permafrost-dev/eget/lib/assets"
)

type DetectorChain struct {
	detectors []Detector
	system    Detector
}

func (dc *DetectorChain) Detect(assets []Asset) (Asset, []Asset, error) {
	for _, d := range dc.detectors {
		choice, candidates, err := d.Detect(assets)
		if len(candidates) == 0 && err != nil {
			return Asset{}, nil, err
		}
		if len(candidates) == 0 {
			return choice, nil, nil
		}
		assets = candidates
	}

	choice, candidates, err := dc.system.Detect(assets)
	if len(candidates) == 0 && err != nil {
		return Asset{}, nil, err
	}
	if len(candidates) == 0 {
		return choice, nil, nil
	}
	if len(candidates) >= 1 {
		assets = candidates
	}

	return Asset{}, assets, fmt.Errorf("%d candidates found for asset chain", len(assets))
}
