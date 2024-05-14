package utilities_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/permafrost-dev/zeget/lib/utilities"
)

var _ = Describe("Utilities/Errors", func() {
	Describe("InvalidGitHubProjectURL", func() {
		It("should be an error type", func() {
			var err error
			e := errors.New("test").(InvalidGitHubProjectURLError)
			err = NewInvalidGitHubProjectURLError("test")
			Expect(err).To(BeAssignableToTypeOf(e))
		})

		It("should have the correct error message", func() {
			err := NewInvalidGitHubProjectURLError("test")
			Expect(err.Error()).To(Equal("Invalid GitHub URL"))
		})
	})

	Describe("InvalidGitHubProjectReference", func() {
		It("should be an error type", func() {
			var err error
			e := errors.New("test").(InvalidGitHubProjectReference)
			err = NewInvalidGitHubProjectReferenceError("test")
			Expect(err).To(BeAssignableToTypeOf(e))
		})

		It("should have the correct error message", func() {
			err := NewInvalidGitHubProjectReferenceError("test")
			Expect(err.Error()).To(Equal("Invalid GitHub project reference: test"))
		})
	})
})
