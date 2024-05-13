package github

import (
	"fmt"
)

type InvalidGitHubProjectURL = error
type InvalidGitHubProjectReference = error

func NewInvalidGitHubProjectURLError(URL string) InvalidGitHubProjectURL {
	return fmt.Errorf("Invalid GitHub URL: %s", URL)
}

func NewInvalidGitHubProjectReferenceError(reference string) InvalidGitHubProjectReference {
	return fmt.Errorf("Invalid GitHub project reference: %s", reference)
}
