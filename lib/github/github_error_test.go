package github_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/eget/lib/github"
)

var _ = Describe("Github Errors", func() {
	var (
		err *github.Error
	)

	BeforeEach(func() {
		err = &github.Error{}
	})

	Describe("Error method", func() {
		Context("when the status code is 403", func() {
			BeforeEach(func() {
				err.Code = http.StatusForbidden
				err.Status = "Forbidden"
				err.Body = []byte(`{"message": "API rate limit exceeded", "documentation_url": "https://developer.github.com/v3/#rate-limiting"}`)
			})

			It("returns the correct error message", func() {
				Expect(err.Error()).To(Equal("Forbidden: API rate limit exceeded: https://developer.github.com/v3/#rate-limiting"))
			})
		})

		Context("when the status code is not 403", func() {
			BeforeEach(func() {
				err.Code = http.StatusNotFound
				err.Status = "Not Found"
				err.URL = "https://api.github.com/repos/nonexistent"
			})

			It("returns a generic error message", func() {
				Expect(err.Error()).To(Equal("Not Found (URL: https://api.github.com/repos/nonexistent)"))
			})
		})
	})
})
