package mockhttp_test

import (
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/zeget/lib/mockhttp"
)

var _ = Describe("MockHTTPClient", func() {
	var client mockhttp.HTTPClient

	BeforeEach(func() {
		client = mockhttp.NewMockHTTPClient()
		client.DoFunc = func(req *http.Request) (*http.Response, error) {
			return mockhttp.NewMockResponse("mock body", http.StatusOK), nil
		}
	})

	Describe("AddJSONResponse", func() {
		It("should add a JSON response for a specific URL", func() {
			client.AddJSONResponse("https://api.github.com/repos/testRepo/releases/v1.0.0", `{"tag_name": "v1.0.0"}`, 200)
			Expect(client.Responses["https://api.github.com/repos/testRepo/releases/v1.0.0"]).To(HaveLen(1))
			Expect(client.Responses["https://api.github.com/repos/testRepo/releases/v1.0.0"][0].Body).To(Equal(`{"tag_name": "v1.0.0"}`))
			Expect(client.Responses["https://api.github.com/repos/testRepo/releases/v1.0.0"][0].StatusCode).To(Equal(200))
		})
	})

	Describe("Reset", func() {
		It("should reset all requests and responses", func() {
			client.AddJSONResponse("https://api.github.com/repos/testRepo/releases/v1.0.0", `{"tag_name": "v1.0.0"}`, 200)
			client.Reset()
			Expect(client.Responses).To(BeEmpty())
			Expect(client.Requests).To(BeEmpty())
		})
	})

	Describe("GetJSON", func() {
		It("should return the correct JSON response for a URL", func() {
			client.AddJSONResponse("https://api.github.com/repos/testRepo/releases/v1.0.0", `{"tag_name": "v1.0.0"}`, 200)
			resp, err := client.GetJSON("https://api.github.com/repos/testRepo/releases/v1.0.0")
			Expect(err).To(BeNil())
			body, _ := io.ReadAll(resp.Body)
			Expect(string(body)).To(Equal(`{"tag_name": "v1.0.0"}`))
			Expect(resp.StatusCode).To(Equal(200))
		})

		It("should return a 404 response if the URL is not found", func() {
			resp, err := client.GetJSON("https://api.github.com/repos/unknown/releases/v1.0.0")
			Expect(err).To(BeNil())
			body, _ := io.ReadAll(resp.Body)
			Expect(string(body)).To(Equal(`{"message":"Not Found","documentation_url":"https://developer.github.com/v3"}`))
			Expect(resp.StatusCode).To(Equal(404))
		})

		It("should return a 500 error if the response is a server error", func() {
			client.AddJSONResponse("https://api.github.com/repos/testRepo/releases/v1.0.0", `{"tag_name": "v1.0.0"}`, 500)
			_, err := client.GetJSON("https://api.github.com/repos/testRepo/releases/v1.0.0")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("mock 500 error"))
		})
	})
})
