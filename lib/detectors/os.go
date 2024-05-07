package detectors

import "regexp"

// An OS represents a target operating system.
type OS struct {
	name     string
	regex    *regexp.Regexp
	anti     *regexp.Regexp
	priority *regexp.Regexp // matches to priority are better than normal matches
}

// Match returns true if the given archive name is likely to store a binary for
// this OS. Also returns if this is a priority match.
func (os *OS) Match(s string) (bool, bool) {
	if os.anti != nil && os.anti.MatchString(s) {
		return false, false
	}
	if os.priority != nil {
		return os.regex.MatchString(s), os.priority.MatchString(s)
	}
	return os.regex.MatchString(s), false
}

var (
	OSDarwin = OS{
		name:  "darwin",
		regex: regexp.MustCompile(`(?i)(darwin|mac.?(os)?|osx)`),
	}
	OSWindows = OS{
		name:  "windows",
		regex: regexp.MustCompile(`(?i)([^r]win|windows)`),
	}
	OSLinux = OS{
		name:     "linux",
		regex:    regexp.MustCompile(`(?i)(linux|ubuntu)`),
		anti:     regexp.MustCompile(`(?i)(android)`),
		priority: regexp.MustCompile(`\.appimage$`),
	}
	OSNetBSD = OS{
		name:  "netbsd",
		regex: regexp.MustCompile(`(?i)(netbsd)`),
	}
	OSFreeBSD = OS{
		name:  "freebsd",
		regex: regexp.MustCompile(`(?i)(freebsd)`),
	}
	OSOpenBSD = OS{
		name:  "openbsd",
		regex: regexp.MustCompile(`(?i)(openbsd)`),
	}
	OSAndroid = OS{
		name:  "android",
		regex: regexp.MustCompile(`(?i)(android)`),
	}
	OSIllumos = OS{
		name:  "illumos",
		regex: regexp.MustCompile(`(?i)(illumos)`),
	}
	OSSolaris = OS{
		name:  "solaris",
		regex: regexp.MustCompile(`(?i)(solaris)`),
	}
	OSPlan9 = OS{
		name:  "plan9",
		regex: regexp.MustCompile(`(?i)(plan9)`),
	}
)

// a map of GOOS values to internal OS matchers
var goosmap = map[string]OS{
	"darwin":  OSDarwin,
	"windows": OSWindows,
	"linux":   OSLinux,
	"netbsd":  OSNetBSD,
	"openbsd": OSOpenBSD,
	"freebsd": OSFreeBSD,
	"android": OSAndroid,
	"illumos": OSIllumos,
	"solaris": OSSolaris,
	"plan9":   OSPlan9,
}
