package mockhttp

import (
	"bytes"
	"io"
	"net/http"

	"github.com/permafrost-dev/eget/lib/download"
	"github.com/permafrost-dev/eget/lib/utilities"
	pb "github.com/schollz/progressbar/v3"
)

/*
  - * *
    // basic usage example (in some_file_test.go):

    client = mockhttp.NewMockHTTPClient()

    client.DoFunc = func(req *http.Request) (*http.Response, error) {
    //
    }

    client.AddJSONResponse("https://api.github.com/repos/testRepo/releases/v1.0.0", `{"tag_name": "v1.0.0", "prerelease": false, "assets": [{"name": "exampleTool.tar.gz", "browser_download_url": "https://github.com/example/repo/tarball/v1.0.0/exampleTool.tar.gz"}], "created_at": "2020-01-01T00:00:00Z"}`, 200)
*/
type RequestData struct {
	Method string
	URL    string
}

type JSONResponse struct {
	Body       string
	StatusCode int
}

type HTTPClient struct {
	Requests  []RequestData
	BaseURL   string
	DoFunc    func(req *http.Request) (*http.Response, error)
	Responses map[string][]JSONResponse

	download.ClientContract
}

func NewMockHTTPClient() HTTPClient {
	return HTTPClient{
		Responses: make(map[string][]JSONResponse),
	}
}

func (m HTTPClient) Reset() {
	copy(m.Requests, make([]RequestData, 0))

	for k := range m.Responses {
		delete(m.Responses, k)
	}
}

func (m HTTPClient) AddJSONResponse(url string, json string, statusCode int) {
	m.Responses[url] = append(m.Responses[url], JSONResponse{Body: json, StatusCode: statusCode})
}

func (m HTTPClient) RoundTrip(_ *http.Request) (*http.Response, error) {
	return NewMockResponse("mock body", http.StatusOK), nil
}

func (m HTTPClient) GetClient() *http.Client {
	return nil
}

func (m HTTPClient) GetJSON(url string) (*http.Response, error) {
	before, _, _ := utilities.Cut(url, "?")
	url = before

	for k, v := range m.Responses {
		if k == url {
			return NewMockResponse(v[0].Body, v[0].StatusCode), nil
		}
	}

	js := `{"message":"Not Found","documentation_url":"https://developer.github.com/v3"}`

	return NewMockResponse(js, http.StatusNotFound), nil
}

func (m HTTPClient) GetBinaryFile(_ string) (*http.Response, error) {
	return NewMockResponse("mock body", http.StatusOK), nil
}

func (m HTTPClient) GetText(_ string) (*http.Response, error) {
	return NewMockResponse("mock body", http.StatusOK), nil
}

func (m HTTPClient) Get(_ string) (*http.Response, error) {
	return NewMockResponse("mock body", http.StatusOK), nil
}

func (m HTTPClient) Download(_ string, _ io.Writer, _ func(size int64) *pb.ProgressBar) error {
	return nil
}

func (m HTTPClient) Do(req *http.Request) (*http.Response, error) {
	_ = append(m.Requests, RequestData{Method: req.Method, URL: req.URL.String()})
	return m.DoFunc(req)
}

// Utility function to create a mock HTTP response
func NewMockResponse(body string, statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}
