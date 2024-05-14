package github_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/zeget/lib/download"
	"github.com/permafrost-dev/zeget/lib/github"
	"github.com/permafrost-dev/zeget/lib/mockhttp"
)

var _ = Describe("RateLimit", func() {
	var clientBase mockhttp.HTTPClient
	var client download.ClientContract

	BeforeEach(func() {
		clientBase = mockhttp.NewMockHTTPClient()
		clientBase.AddJSONResponse("https://api.github.com/rate_limit", `{"resources":{"core":{"limit":5000,"remaining":4990,"reset":1715643356,"used":0,"resource":"core"},"graphql":{"limit":0,"remaining":0,"reset":1715643356,"used":0,"resource":"graphql"},"integration_manifest":{"limit":5000,"remaining":5000,"reset":1715643356,"used":0,"resource":"integration_manifest"},"search":{"limit":10,"remaining":10,"reset":1715639816,"used":0,"resource":"search"}},"rate":{"limit":60,"remaining":60,"reset":1715643356,"used":0,"resource":"core"}}`, 200)
		client = &clientBase
	})

	Describe("String method", func() {
		It("should format the rate limit correctly for a past reset time", func() {
			rateLimit := github.RateLimit{
				Limit:     5000,
				Remaining: 4999,
				Reset:     1715643356,
			}

			Expect(rateLimit.String()).To(ContainSubstring("Limit: 5000, Remaining: 4999, Reset:"))
		})

		It("should format the rate limit correctly for a future reset time", func() {
			rateLimit := github.RateLimit{
				Limit:     5000,
				Remaining: 4999,
				Reset:     1715643356,
			}

			Expect(rateLimit.String()).To(ContainSubstring("Limit: 5000, Remaining: 4999, Reset:"))
		})
	})

	Describe("FetchRateLimit function", func() {
		It("should fetch and parse the rate limit correctly", func() {
			rateLimit, err := github.FetchRateLimit(client)
			Expect(err).ToNot(HaveOccurred())
			Expect(rateLimit.Limit).To(Equal(5000))
			Expect(rateLimit.Remaining).To(Equal(4990))
			Expect(rateLimit.Reset).To(Equal(int64(1715643356)))
		})

		It("should return an error if the request fails", func() {
			clientBase.ResetJSONResponsesForURL("https://api.github.com/rate_limit")
			clientBase.AddJSONResponse("https://api.github.com/rate_limit", `{}`, 500)

			resp, err := github.FetchRateLimit(client)
			fmt.Printf("resp: %v\n", resp)

			Expect(err).To(HaveOccurred())
		})
	})
})
