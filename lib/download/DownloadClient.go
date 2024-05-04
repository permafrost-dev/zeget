package download

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	pb "github.com/schollz/progressbar/v3"
)

type Client struct {
	Headers      []string
	Token        string
	Accept       string
	DisableSSL   bool
	tokenType    string
	CreateClient func() *http.Client
}

func NewClient(token string) *Client {
	result := &Client{
		Token:      token,
		tokenType:  "Bearer",
		DisableSSL: false,
	}

	result.CreateClient = result.GetClient
	return result
}

func setIf[T interface{}](condition bool, original T, newValue T) T {
	if condition {
		return newValue
	}
	return original
}

func (dc *Client) GetTokenType() string {
	return dc.tokenType
}

func (dc *Client) SetHeaders(headers []string) *Client {
	dc.Headers = headers
	return dc
}

func (dc *Client) SetAccept(accept AcceptContentType) *Client {
	dc.Accept = strings.TrimSpace(string(accept))
	return dc
}

func (dc *Client) SetToken(token string) *Client {
	dc.Token = strings.TrimSpace(token)
	return dc
}

func (dc *Client) SetDisableSSL(disableSSL bool) *Client {
	dc.DisableSSL = disableSSL
	return dc
}

func (dc *Client) SetTokenType(tokenType string) *Client {
	dc.tokenType = strings.TrimSpace(tokenType)

	// capitalize the first letter, like "Bearer" or "Token"
	if len(dc.tokenType) > 1 {
		dc.tokenType = strings.ToUpper(string(dc.tokenType[0])) + dc.tokenType[1:]
	}

	return dc
}

func (dc *Client) AddHeader(header string, value string) *Client {
	dc.Headers = append(dc.Headers, header+":"+value)
	return dc
}

func (dc *Client) initRequest(req *http.Request) *http.Request {
	if strings.Contains(dc.Accept, "/") {
		req.Header.Set("Accept", dc.Accept)
	}

	if dc.Token != "" {
		tokenTypeStr := setIf(len(dc.tokenType) > 0, dc.tokenType, "Bearer")
		req.Header.Set("Authorization", fmt.Sprintf("%s %s", tokenTypeStr, dc.Token))
	}

	for _, header := range dc.Headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(parts[0], parts[1])
		}
	}

	return req
}

func (dc *Client) createRequest(method string, url string) (*http.Request, error) {
	result, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	return dc.initRequest(result), nil
}

func (dc *Client) GetClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: dc.DisableSSL},
		},
	}
}

func (dc *Client) Get(url string) (*http.Response, error) {
	req, err := dc.createRequest("GET", url)
	if err != nil {
		return nil, err
	}

	return dc.CreateClient().Do(req)
}

func (dc *Client) GetJSON(url string) (*http.Response, error) {
	return dc.
		SetAccept(AcceptGitHubJSON).
		Get(url)
}

func (dc *Client) GetBinaryFile(url string) (*http.Response, error) {
	return dc.
		SetAccept(AcceptBinary).
		Get(url)
}

func (dc *Client) GetText(url string) (*http.Response, error) {
	return dc.
		SetAccept(AcceptText).
		Get(url)
}

func (dc *Client) Download(url string, out io.Writer, progressBarCallback func(size int64) *pb.ProgressBar) error {
	if isLocalFile(url) {
		f, err := os.Open(url)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(out, f)
		return err
	}

	resp, err := dc.GetBinaryFile(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("download error: %d: %s", resp.StatusCode, body)
	}

	bar := progressBarCallback(resp.ContentLength)
	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)

	return err
}
