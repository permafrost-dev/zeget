package detectors

import (
	"fmt"
	"strings"

	. "github.com/permafrost-dev/eget/lib/assets"
)

// A SystemDetector matches a particular OS/Arch system pair.
type SystemDetector struct {
	Os   OS
	Arch Arch
}

// NewSystemDetector returns a new detector for the given OS/Arch as given by
// Go OS/Arch names.
func NewSystemDetector(sos, sarch string) (*SystemDetector, error) {
	os, ok := goosmap[sos]
	if !ok {
		return nil, fmt.Errorf("unsupported target OS: %s", sos)
	}
	arch, ok := goarchmap[sarch]
	if !ok {
		return nil, fmt.Errorf("unsupported target arch: %s", sarch)
	}
	return &SystemDetector{
		Os:   os,
		Arch: arch,
	}, nil
}

// Detect extracts the assets that match this detector's OS/Arch pair. If one
// direct OS/Arch match is found, it is returned.  If multiple OS/Arch matches
// are found they are returned as candidates. If multiple assets that only
// match the OS are found, and no full OS/Arch matches are found, the OS
// matches are returned as candidates. Otherwise all assets are returned as
// candidates.
func (d *SystemDetector) Detect(assets []Asset) (Asset, []Asset, error) {
	var priority []Asset
	var matches []Asset
	var candidates []Asset
	all := make([]Asset, 0, len(assets))
	for _, a := range assets {
		if strings.HasSuffix(a.Name, ".sha256") || strings.HasSuffix(a.Name, ".sha256sum") {
			// skip checksums (they will be checked later by the verifier)
			continue
		}

		os, extra := d.Os.Match(a.Name)
		if extra {
			priority = append(priority, a)
		}
		arch := d.Arch.Match(a.Name)
		if os && arch {
			matches = append(matches, a)
		}
		if os {
			candidates = append(candidates, a)
		}
		all = append(all, a)
	}

	if len(priority) == 1 {
		return priority[0], nil, nil
	}

	if len(priority) > 1 {
		return Asset{}, priority, nil //fmt.Errorf("%d priority matches found", len(matches))
	}

	if len(matches) == 1 {
		return matches[0], nil, nil
	}

	if len(matches) > 1 {
		return Asset{}, matches, nil
	}

	if len(candidates) == 1 {
		return candidates[0], nil, nil
	}

	if len(candidates) > 1 {
		return Asset{}, candidates, nil //fmt.Errorf("%d candidates found (unsure architecture)", len(candidates))
	}

	if len(all) == 1 {
		return all[0], nil, nil
	}

	return Asset{}, all, fmt.Errorf("no candidates found")
}
