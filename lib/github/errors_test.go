package github_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/eget/lib/github"
)

var _ = Describe("GitHub Errors", func() {

	Describe("NewInvalidGitHubProjectURLError", func() {
		It("returns the InvalidGitHubProjectURLError", func() {
			err := github.NewInvalidGitHubProjectURLError("https://github.com/not/a/real/project")
			Expect(err).To(MatchError(github.InvalidGitHubProjectURLError))
		})
	})

	Describe("NewInvalidGitHubProjectReferenceError", func() {
		It("returns the InvalidGitHubProjectReferenceError", func() {
			err := github.NewInvalidGitHubProjectReferenceError("master")
			Expect(err).To(MatchError(github.InvalidGitHubProjectReferenceError))
		})
	})
})
