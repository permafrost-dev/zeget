package utilities

import (
	"errors"
	"fmt"
)

type InvalidGitHubProjectURLError = error
type InvalidGitHubProjectReference = error

var InvalidGitHubProjectURL InvalidGitHubProjectURLError = errors.New("Invalid GitHub URL").(InvalidGitHubProjectURLError)

func NewInvalidGitHubProjectURLError(_ string) InvalidGitHubProjectURLError {
	return InvalidGitHubProjectURL
}

func NewInvalidGitHubProjectReferenceError(reference string) InvalidGitHubProjectReference {
	return fmt.Errorf("Invalid GitHub project reference: %s", reference)
}
