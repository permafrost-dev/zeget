package detectors

import "regexp"

// An OS represents a target operating system.
type OS struct {
	Name     string
	Regex    *regexp.Regexp
	Anti     *regexp.Regexp
	Priority *regexp.Regexp // matches to priority are better than normal matches
}

// Match returns true if the given archive name is likely to store a binary for
// this OS. Also returns if this is a priority match.
func (os *OS) Match(s string) (bool, bool) {
	if os.Anti != nil && os.Anti.MatchString(s) {
		return false, false
	}
	if os.Priority != nil {
		return os.Regex.MatchString(s), os.Priority.MatchString(s)
	}
	return os.Regex.MatchString(s), false
}

var (
	OSDarwin = OS{
		Name:  "darwin",
		Regex: regexp.MustCompile(`(?i)(darwin|mac.?(os)?|osx)`),
	}
	OSWindows = OS{
		Name:  "windows",
		Regex: regexp.MustCompile(`(?i)([^r]win|windows)`),
	}
	OSLinux = OS{
		Name:     "linux",
		Regex:    regexp.MustCompile(`(?i)(linux|ubuntu)`),
		Anti:     regexp.MustCompile(`(?i)(android)`),
		Priority: regexp.MustCompile(`\.appimage$`),
	}
	OSNetBSD = OS{
		Name:  "netbsd",
		Regex: regexp.MustCompile(`(?i)(netbsd)`),
	}
	OSFreeBSD = OS{
		Name:  "freebsd",
		Regex: regexp.MustCompile(`(?i)(freebsd)`),
	}
	OSOpenBSD = OS{
		Name:  "openbsd",
		Regex: regexp.MustCompile(`(?i)(openbsd)`),
	}
	OSAndroid = OS{
		Name:  "android",
		Regex: regexp.MustCompile(`(?i)(android)`),
	}
	OSIllumos = OS{
		Name:  "illumos",
		Regex: regexp.MustCompile(`(?i)(illumos)`),
	}
	OSSolaris = OS{
		Name:  "solaris",
		Regex: regexp.MustCompile(`(?i)(solaris)`),
	}
	OSPlan9 = OS{
		Name:  "plan9",
		Regex: regexp.MustCompile(`(?i)(plan9)`),
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
