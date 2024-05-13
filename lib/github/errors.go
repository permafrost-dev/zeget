package github

import (
	"errors"
)

type InvalidGitHubProjectURL = error
type InvalidGitHubProjectReference = error

var (
	InvalidGitHubProjectURLError       InvalidGitHubProjectURL       = errors.New("Invalid GitHub project URL")
	InvalidGitHubProjectReferenceError InvalidGitHubProjectReference = errors.New("Invalid GitHub project reference")
)

func NewInvalidGitHubProjectURLError(URL string) InvalidGitHubProjectURL {
	return InvalidGitHubProjectURLError
}

func NewInvalidGitHubProjectReferenceError(reference string) InvalidGitHubProjectReference {
	return InvalidGitHubProjectReferenceError
}
