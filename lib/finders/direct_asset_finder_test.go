package finders_test

import (
	"bytes"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/eget/lib/assets"
	. "github.com/permafrost-dev/eget/lib/finders"
)

type MockHTTPRequestData struct {
	Method string
	URL    string
}

type MockHTTPClient struct {
	Requests []MockHTTPRequestData
	DoFunc   func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	return newMockResponse("mock body", http.StatusOK), nil
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.Requests = append(m.Requests, MockHTTPRequestData{Method: req.Method, URL: req.URL.String()})
	return m.DoFunc(req)
}

// Utility function to create a mock HTTP response
func newMockResponse(body string, statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

var _ = Describe("DirectAssetFinder", func() {
	var (
		directAssetFinder *DirectAssetFinder
		//downloadClient    interface{}
	)

	BeforeEach(func() {
		directAssetFinder = &DirectAssetFinder{URL: "https://example.com/asset.zip"}
		// downloadClient = &MockHTTPClient{
		// 	DoFunc: func(req *http.Request) (*http.Response, error) {
		// 		return newMockResponse("mock body", http.StatusOK), nil
		// 	},
		// }
	})

	Describe("Find", func() {
		It("should return an asset with the same URL as the DirectAssetFinder URL", func() {
			expectedAsset := assets.Asset{
				Name:        "https://example.com/asset.zip",
				DownloadURL: "https://example.com/asset.zip",
			}

			assets, err := directAssetFinder.Find(nil)

			Expect(err).To(BeNil())
			Expect(assets).To(HaveLen(1))
			Expect(assets[0]).To(Equal(expectedAsset))
		})
	})
})
