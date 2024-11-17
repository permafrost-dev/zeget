package download_test

import (
	"bytes"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	g "github.com/onsi/ginkgo/v2"

	// . "github.com/onsi/gomega"
	gm "github.com/onsi/gomega"

	// "testing"

	. "github.com/permafrost-dev/zeget/lib/download"
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

var _ = Describe("DownloadClient", func() {
	It("should create a new DownloadClient", func() {
		token := "test-token"
		dc := NewClient(token)

		gm.Expect(dc.Token).To(gm.Equal(token))
		gm.Expect(dc.GetTokenType()).To(gm.Equal("Bearer"))
	})

	It("should set headers", func() {
		dc := NewClient("")
		headers := []string{"header1:value1", "header2:value2"}

		dc.SetHeaders(headers)

		gm.Expect(dc.Headers).To(gm.Equal(headers))
	})

	It("should set a token", func() {
		dc := NewClient("")
		token := "test-token"
		dc.SetToken(token)
		gm.Expect(dc.Token).To(gm.Equal(token))
	})

	g.It("should set a token type", func() {
		dc := NewClient("")
		tokenType := "TestType" // the first char of type is auto-capitalized
		dc.SetTokenType(tokenType)
		gm.Expect(dc.GetTokenType()).To(gm.Equal(tokenType))
	})

	g.It("should set accept", func() {
		dc := NewClient("")
		dc.SetAccept(AcceptGitHubJSON)

		gm.Expect(dc.Accept).To(gm.Equal(string(AcceptGitHubJSON)))
	})

	It("should add a header", func() {
		dc := NewClient("")
		dc.AddHeader("Test-Header", "value")

		gm.Expect(dc.Headers).To(gm.Equal([]string{"Test-Header:value"}))
	})

	g.It("should get a URL", func() {
		client := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return newMockResponse("mock body", http.StatusOK), nil
			},
		}

		dc := &Client{Token: "test-token"}
		dc.SetDisableSSL(true) // To avoid dealing with TLS in tests
		dc.AddHeader("Test-Header", "value")
		dc.AddHeader("X-Test", "123")

		// Override the getClient method to use the mock client
		originalGetClient := dc.GetClient
		dc.CreateClient = func() *http.Client {
			return &http.Client{Transport: client}
		}
		defer func() { dc.CreateClient = originalGetClient }()

		resp, err := dc.Get("https://github.com")
		gm.Expect(err).To(gm.BeNil())
		gm.Expect(resp.StatusCode).To(gm.Equal(http.StatusOK))

		body, _ := io.ReadAll(resp.Body)
		gm.Expect(string(body)).To(gm.Equal("mock body"))
	})

	g.It("should get a JSON URL", func() {
		client := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return newMockResponse("mock body", http.StatusOK), nil
			},
		}

		dc := &Client{}
		dc.SetDisableSSL(true) // To avoid dealing with TLS in tests

		// Override the getClient method to use the mock client
		originalGetClient := dc.GetClient
		dc.CreateClient = func() *http.Client {
			return &http.Client{Transport: client}
		}
		defer func() { dc.CreateClient = originalGetClient }()

		resp, err := dc.GetJSON("https://github.com")
		gm.Expect(err).To(gm.BeNil())
		gm.Expect(resp.StatusCode).To(gm.Equal(http.StatusOK))

		body, _ := io.ReadAll(resp.Body)
		gm.Expect(string(body)).To(gm.Equal("mock body"))
	})

	It("should get a binary file URL", func() {
		client := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return newMockResponse("mock body", http.StatusOK), nil
			},
		}

		dc := &Client{}
		dc.SetDisableSSL(true) // To avoid dealing with TLS in tests

		// Override the getClient method to use the mock client
		originalGetClient := dc.GetClient
		dc.CreateClient = func() *http.Client {
			return &http.Client{Transport: client}
		}
		defer func() { dc.CreateClient = originalGetClient }()
		resp, err := dc.GetBinaryFile("https://github.com")
		gm.Expect(err).To(gm.BeNil())
		gm.Expect(resp.StatusCode).To(gm.Equal(http.StatusOK))

		body, _ := io.ReadAll(resp.Body)
		gm.Expect(string(body)).To(gm.Equal("mock body"))
	})

	It("should get a text file URL", func() {
		client := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return newMockResponse("mock body", http.StatusOK), nil
			},
		}

		dc := &Client{}
		dc.SetDisableSSL(true) // To avoid dealing with TLS in tests

		// Override the getClient method to use the mock client
		originalGetClient := dc.GetClient
		dc.CreateClient = func() *http.Client {
			return &http.Client{Transport: client}
		}
		defer func() { dc.CreateClient = originalGetClient }()
		resp, err := dc.GetText("https://github.com")
		gm.Expect(err).To(gm.BeNil())
		gm.Expect(resp.StatusCode).To(gm.Equal(http.StatusOK))

		body, _ := io.ReadAll(resp.Body)
		gm.Expect(string(body)).To(gm.Equal("mock body"))
	})

	// It("should Download a file", func() {
	// 	client := &MockHTTPClient{
	// 		DoFunc: func(req *http.Request) (*http.Response, error) {
	// 			return newMockResponse("mock body", http.StatusOK), nil
	// 		},
	// 		Requests: []MockHTTPRequestData{},
	// 	}

	// 	dc := &Client{CreateClient: func() *http.Client { return &http.Client{Transport: client} }}
	// 	dc.SetDisableSSL(true) // To avoid dealing with TLS in tests

	// 	// Override the getClient method to use the mock client
	// 	originalGetClient := dc.GetClient
	// 	dc.CreateClient = func() *http.Client {
	// 		return &http.Client{Transport: client}
	// 	}
	// 	defer func() { dc.CreateClient = originalGetClient }()

	// 	var buf bytes.Buffer
	// 	err := dc.Download("https://github.com", &buf, func(size int64) *pb.ProgressBar {
	// 		return pb.NewOptions64(size, pb.OptionSetWriter(io.Discard))
	// 	})
	// 	gm.Expect(err).To(gm.BeNil())
	// 	gm.Expect(buf.String()).To(gm.Equal("mock body"))
	// })
})
