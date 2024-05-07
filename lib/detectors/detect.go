package detectors

import (
	"runtime"
	"strings"

	"github.com/permafrost-dev/eget/lib/appflags"
	. "github.com/permafrost-dev/eget/lib/assets"
)

// A Detector selects an asset from a list of possibilities.
type Detector interface {
	// Detect takes a list of possible assets and returns a direct match. If a
	// single direct match is not found, it returns a list of candidates and an
	// error explaining what happened.
	Detect(assets []Asset) (Asset, []Asset, error)
}

// Determine the appropriate detector. If the --system is 'all', we use an
// AllDetector, which will just return all assets. Otherwise we use the
// --system pair provided by the user, or the runtime.GOOS/runtime.GOARCH
// pair by default (the host system OS/Arch pair).
func DetermineCorrectDetector(opts *appflags.Flags, system *SystemDetector) (detector Detector, err error) {
	if system == nil {
		system, _ = NewSystemDetector(runtime.GOOS, runtime.GOARCH)
	}

	detector = system

	if len(opts.System) > 2 && opts.System != "all" && strings.Contains(opts.System, "/") {
		split := strings.Split(opts.System, "/")
		detector, err = NewSystemDetector(split[0], split[1])
	}

	if opts.System == "all" {
		detector = &AllDetector{}
	}

	if len(opts.Asset) == 0 {
		return detector, err
	}

	detectors := make([]Detector, len(opts.Asset))

	for i, a := range opts.Asset {
		anti := strings.HasPrefix(a, "^") || strings.HasPrefix(a, "!")

		if anti {
			a = a[1:]
		}

		detectors[i] = &SingleAssetDetector{
			Asset: a,
			Anti:  anti,
		}
	}

	detector = &DetectorChain{
		detectors: detectors,
		system:    system,
	}

	return detector, err
}