package finders_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/eget/lib/download"
	. "github.com/permafrost-dev/eget/lib/finders"
	"github.com/permafrost-dev/eget/lib/github"
	pb "github.com/schollz/progressbar/v3"
)

var (
	server *httptest.Server
)

type MockHTTPRequestData struct {
	Method string
	URL    string
}

type MockHTTPClient struct {
	Requests []MockHTTPRequestData
	BaseURL  string
	DoFunc   func(req *http.Request) (*http.Response, error)
	download.ClientContract
}

func (m MockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	return newMockResponse("mock body", http.StatusOK), nil
}

func (m MockHTTPClient) GetClient() *http.Client {
	return nil
}

func (m MockHTTPClient) GetJSON(url string) (*http.Response, error) {
	if strings.HasSuffix(url, "nonexistent") {
		return nil, &github.Error{
			Status: "404 Not Found",
			Code:   http.StatusNotFound,
			Body:   []byte(`{"message":"Not Found","documentation_url":"https://developer.github.com/v3"}`),
			URL:    url,
		}
	}

	var js string

	switch strings.Replace(url, "https://github.com", "", 1) {
	case "https://api.github.com/repos/testRepo/releases?page=1":
		js = `[{"tag_name": "v1.0.0", "assets": [{"name": "asset1", "browser_download_url": "http://example.com/asset1"}], "created_at": "2020-01-01T00:00:00Z"}]`
		break
	case "https://api.github.com/repos/testRepo/releases/latest":
		js = `{"tag_name": "v1.0.0", "assets": [{"name": "asset1", "browser_download_url": "http://example.com/asset1"}], "created_at": "2020-01-01T00:00:00Z"}`
		break
	case "https://api.github.com/repos/testRepo/releases/v1.0.0":
		js = `{"tag_name": "v1.0.0", "assets": [{"name": "asset1", "browser_download_url": "http://example.com/asset1"}], "created_at": "2020-01-01T00:00:00Z"}`
		break
	default:
		return nil, &github.Error{
			Status: "404 Not Found",
			Code:   http.StatusNotFound,
			Body:   []byte(`{"message":"Not Found","documentation_url":"https://developer.github.com/v3"}`),
			URL:    url,
		}
	}

	// fmt.Printf("request: %s\n%s\n", url, js)

	return newMockResponse(js, http.StatusOK), nil
}

func (m MockHTTPClient) GetBinaryFile(url string) (*http.Response, error) {
	return newMockResponse("mock body", http.StatusOK), nil
}

func (m MockHTTPClient) GetText(url string) (*http.Response, error) {
	return newMockResponse("mock body", http.StatusOK), nil
}

func (m MockHTTPClient) Get(url string) (*http.Response, error) {
	return newMockResponse("mock body", http.StatusOK), nil
}

func (m MockHTTPClient) Download(url string, out io.Writer, progressBarCallback func(size int64) *pb.ProgressBar) error {
	return nil
}

func (m MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
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

var _ = Describe("GithubAssetFinder", func() {
	var (
		client      *MockHTTPClient
		assetFinder *GithubAssetFinder
	)

	BeforeEach(func() {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/repos/testRepo/releases":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[{"tag_name": "v1.0.0", "assets": [{"name": "asset1", "browser_download_url": "http://example.com/asset1"}], "created_at": "2020-01-01T00:00:00Z"}]`))
			case "/repos/testRepo/releases/latest":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"tag_name": "v1.0.0", "assets": [{"name": "asset1", "browser_download_url": "http://example.com/asset1"}], "created_at": "2020-01-01T00:00:00Z"}`))
			case "/repos/testRepo/releases/tags/v1.0.0":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"tag_name": "v1.0.0", "assets": [{"name": "asset1", "browser_download_url": "http://example.com/asset1"}], "created_at": "2020-01-01T00:00:00Z"}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}

			// fmt.Printf("request: %s\n", r.URL.Path)
		}))
		client = &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return newMockResponse("mock body", http.StatusOK), nil
			},
			BaseURL: server.URL + "/",
		}

		assetFinder = &GithubAssetFinder{
			Repo:    "testRepo",
			Tag:     "latest",
			MinTime: time.Date(2019, 12, 31, 23, 59, 59, 0, time.UTC),
		}
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Find", func() {
		Context("with a valid tag", func() {
			It("should return assets", func() {
				// assetFinder.Tag = "latest"
				// baseURL := server.URL + "/"
				assets, _ := assetFinder.Find(client)
				// Expect(err).ToNot(HaveOccurred())
				//>0
				// Expect(assets).To(BeNumerically(">", 0))
				Expect(assets[0].Name).To(Equal("asset1"))
				Expect(assets[0].DownloadURL).To(Equal("http://example.com/asset1"))
			})
		})

		Context("with a tag that does not exist", func() {
			It("should return an error", func() {
				// dlclient := client.(download.ClientContract)
				assetFinder.Tag = "nonexistent"
				//client.BaseURL = server.URL + "/"
				_, err := assetFinder.Find(client)
				Expect(err).To(HaveOccurred())
			})
		})

		// FindMatch:

		Context("find a match", func() {
			It("should return assets", func() {
				// baseURL := server.URL + "/"
				assetFinder.Tag = "v1.0.0"
				assets, err := assetFinder.FindMatch(client)
				Expect(err).ToNot(HaveOccurred())
				Expect(assets).To(HaveLen(1))
				Expect(assets[0].Name).To(Equal("asset1"))
				Expect(assets[0].DownloadURL).To(Equal("http://example.com/asset1"))
			})
		})
	})
})
