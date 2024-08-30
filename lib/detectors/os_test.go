package detectors_test

import (
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/permafrost-dev/zeget/lib/detectors"
)

var _ = Describe("OS", func() {
	var (
		osDarwin  OS
		osWindows OS
		osLinux   OS
	)

	BeforeEach(func() {
		osDarwin = OS{
			Name:  "darwin",
			Regex: regexp.MustCompile(`(?i)(darwin|mac.?(os)?|osx)`),
		}
		osWindows = OS{
			Name:  "windows",
			Regex: regexp.MustCompile(`(?i)([^r]win|windows)`),
		}
		osLinux = OS{
			Name:     "linux",
			Regex:    regexp.MustCompile(`(?i)(linux|ubuntu)`),
			Anti:     regexp.MustCompile(`(?i)(android)`),
			Priority: regexp.MustCompile(`\.appimage$`),
		}
	})

	Describe("Match", func() {
		Context("with an OS that matches the given string", func() {
			It("should return true for a matching OS string", func() {
				Expect(osDarwin.Match("macOS")).To(Equal(true))
				Expect(osWindows.Match("windows10")).To(Equal(true))
			})

			It("should return false for a non-matching OS string", func() {
				Expect(osDarwin.Match("linux")).To(BeFalse())
				Expect(osWindows.Match("darwin")).To(BeFalse())
			})

			It("should respect anti-patterns", func() {
				Expect(osLinux.Match("android")).To(BeFalse())
			})

			It("should detect priority matches", func() {
				_, priority := osLinux.Match("app.appimage")
				Expect(priority).To(BeTrue())
			})
		})
	})
})
