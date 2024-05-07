package detectors

import "regexp"

// An Arch represents a system architecture, such as amd64, i386, arm or others.
type Arch struct {
	name  string
	regex *regexp.Regexp
}

// Match returns true if this architecture is likely supported by the given
// archive name.
func (a *Arch) Match(s string) bool {
	return a.regex.MatchString(s)
}

var (
	ArchAMD64 = Arch{
		name:  "amd64",
		regex: regexp.MustCompile(`(?i)(x64|amd64|x86(-|_)?64)`),
	}
	ArchI386 = Arch{
		name:  "386",
		regex: regexp.MustCompile(`(?i)(x32|amd32|x86(-|_)?32|i?386)`),
	}
	ArchArm = Arch{
		name:  "arm",
		regex: regexp.MustCompile(`(?i)(arm32|armv6|arm\b)`),
	}
	ArchArm64 = Arch{
		name:  "arm64",
		regex: regexp.MustCompile(`(?i)(arm64|armv8|aarch64)`),
	}
	ArchRiscv64 = Arch{
		name:  "riscv64",
		regex: regexp.MustCompile(`(?i)(riscv64)`),
	}
)

// a map from GOARCH values to internal architecture matchers
var goarchmap = map[string]Arch{
	"amd64":   ArchAMD64,
	"386":     ArchI386,
	"arm":     ArchArm,
	"arm64":   ArchArm64,
	"riscv64": ArchRiscv64,
}
