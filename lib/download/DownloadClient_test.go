package download_test

import (
	"bytes"
	"io"
	"net/http"

	_ "github.com/onsi/ginkgo/v2"
	g "github.com/onsi/ginkgo/v2"
	_ "github.com/onsi/gomega"
	gm "github.com/onsi/gomega"

	// "testing"

	. "github.com/permafrost-dev/eget/lib/download"
)

type MockHTTPRequestData struct {
	Method string
	URL    string
}

// Mock HTTP Client
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

var _ = g.Describe("DownloadClient", func() {
	g.It("should create a new DownloadClient", func() {
		token := "test-token"
		dc := NewClient(token)

		gm.Expect(dc.Token).To(gm.Equal(token))
		gm.Expect(dc.GetTokenType()).To(gm.Equal("Bearer"))
	})

	g.It("should set headers", func() {
		dc := NewClient("")
		headers := []string{"header1:value1", "header2:value2"}

		dc.SetHeaders(headers)

		gm.Expect(dc.Headers).To(gm.Equal(headers))
	})

	g.It("should set accept", func() {
		dc := NewClient("")
		dc.SetAccept(AcceptGitHubJSON)

		gm.Expect(dc.Accept).To(gm.Equal(string(AcceptGitHubJSON)))
	})

	g.It("should add a header", func() {
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

		dc := &Client{}
		dc.SetDisableSSL(true) // To avoid dealing with TLS in tests

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
})

// func TestSetAccept(t *testing.T) {
// 	dc := NewDownloadClient("")

// 	// Assuming AcceptContentType is a string for simplicity
// 	dc.SetAccept(AcceptGitHubJSON)

// 	if dc.Accept != string(AcceptGitHubJSON) {
// 		t.Errorf("Expected Accept to be %s, got %s", string(AcceptGitHubJSON), dc.Accept)
// 	}
// }

// func TestAddHeader(t *testing.T) {
// 	dc := NewDownloadClient("")
// 	dc.AddHeader("Test-Header", "value")

// 	if len(dc.Headers) != 1 || dc.Headers[0] != "Test-Header:value" {
// 		t.Errorf("Expected header 'Test-Header:value', got %+v", dc.Headers)
// 	}
// }

// func TestGet(t *testing.T) {
// 	// Mock HTTP client
// 	client := &MockHTTPClient{
// 		DoFunc: func(req *http.Request) (*http.Response, error) {
// 			return newMockResponse("mock body", http.StatusOK), nil
// 		},
// 	})

// 	dc := &DownloadClient{}
// 	dc.SetDisableSSL(true) // To avoid dealing with TLS in tests

// 	// Override the getClient method to use the mock client
// 	originalGetClient := dc.GetClient
// 	dc.CreateClient = func() *http.Client {
// 		return &http.Client{Transport: client}
// 	}
// 	defer func() { dc.CreateClient = originalGetClient }()

// 	resp, err := dc.Get("https://github.com")
// 	if err != nil {
// 		t.Errorf("Expected no error, got %v", err)
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
// 	}

// 	body, _ := io.ReadAll(resp.Body)
// 	if string(body) != "mock body" {
// 		t.Errorf("Expected body to be 'mock body', got %s", body)
// 	}
// }

// func TestGetJSON(t *testing.T) {
// 	//use snapshot testing:
// 	//https://pkg.go.dev/github.com/google/go-cmp/cmp#hdr-CanonicalJSON
// 	// Mock HTTP client
// 	client := &MockHTTPClient{
// 		DoFunc: func(req *http.Request) (*http.Response, error) {
// 			return newMockResponse("mock body", http.StatusOK), nil
// 		},
// 	})

// 	dc := &DownloadClient{}
// 	dc.SetDisableSSL(true) // To avoid dealing with TLS in tests

// 	// Override the getClient method to use the mock client
// 	originalGetClient := dc.GetClient
// 	dc.CreateClient = func() *http.Client {
// 		return &http.Client{Transport: client}
// 	}
// 	defer func() { dc.CreateClient = originalGetClient }()

// 	resp, err := dc.GetJSON("https://github.com")
// 	if err != nil {
// 		t.Errorf("Expected no error, got %v", err)
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
// 	}

// 	body, _ := io.ReadAll(resp.Body)
// 	if string(body) != "mock body" {
// 		t.Errorf("Expected body to be 'mock body', got %s", body)
// 	}
// }
