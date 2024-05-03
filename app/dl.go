package app

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/permafrost-dev/eget/lib/home"
	pb "github.com/schollz/progressbar/v3"
)

const (
	AcceptBinary     = "application/octet-stream"
	AcceptGitHubJSON = "application/vnd.github+json"
	AcceptText       = "text/plain"
)

func tokenFrom(s string) (string, error) {
	if strings.HasPrefix(s, "@") {
		f, err := home.Expand(s[1:])
		if err != nil {
			return "", err
		}
		b, err := os.ReadFile(f)
		return strings.TrimRight(string(b), "\r\n"), nil
	}
	return s, nil
}

var ErrNoToken = errors.New("no github token")

func getGithubToken() (string, error) {
	if os.Getenv("EGET_GITHUB_TOKEN") != "" {
		return tokenFrom(os.Getenv("EGET_GITHUB_TOKEN"))
	}
	if os.Getenv("GITHUB_TOKEN") != "" {
		return tokenFrom(os.Getenv("GITHUB_TOKEN"))
	}
	return "", ErrNoToken
}

func SetAuthHeader(req *http.Request, disableSSL bool) *http.Request {
	token, err := getGithubToken()
	if err != nil && !errors.Is(err, ErrNoToken) {
		fmt.Fprintln(os.Stderr, "warning: not using github token:", err)
	}

	if req.URL.Scheme == "https" && req.Host == "api.github.com" && err == nil {
		if disableSSL {
			fmt.Fprintln(os.Stderr, "error: cannot use GitHub token if SSL verification is disabled")
			os.Exit(1)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	return req
}

func Get(url, accept string, disableSSL bool) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", accept)

	req = SetAuthHeader(req, disableSSL)

	proxyClient := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: disableSSL},
		},
	}

	return proxyClient.Do(req)
}

func GetJson(url string) (*http.Response, error) {
	return Get(url, AcceptGitHubJSON, false)
}

func GetBinaryFile(url string) (*http.Response, error) {
	return Get(url, AcceptBinary, false)
}

func GetText(url string) (*http.Response, error) {
	return Get(url, AcceptText, false)
}

type RateLimitJson struct {
	Resources map[string]RateLimit
}

type RateLimit struct {
	Limit     int
	Remaining int
	Reset     int64
}

func (r RateLimit) ResetTime() time.Time {
	return time.Unix(r.Reset, 0)
}

func (r RateLimit) String() string {
	now := time.Now()
	rtime := r.ResetTime()
	if rtime.Before(now) {
		return fmt.Sprintf("Limit: %d, Remaining: %d, Reset: %v", r.Limit, r.Remaining, rtime)
	} else {
		return fmt.Sprintf(
			"Limit: %d, Remaining: %d, Reset: %v (%v)",
			r.Limit, r.Remaining, rtime, rtime.Sub(now).Round(time.Second),
		)
	}
}

func GetRateLimit() (RateLimit, error) {
	resp, err := GetJson("https://api.github.com/rate_limit")
	if err != nil {
		return RateLimit{}, err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return RateLimit{}, err
	}

	var parsed RateLimitJson
	err = json.Unmarshal(b, &parsed)

	return parsed.Resources["core"], err
}

func (app *Application) getDownloadProgressBar(size int64) *pb.ProgressBar {
	var pbout io.Writer = os.Stderr
	if app.Opts.Quiet {
		pbout = io.Discard
	}

	return pb.NewOptions64(
		size,
		pb.OptionSetWriter(pbout),
		pb.OptionShowBytes(true),
		pb.OptionSetWidth(10),
		pb.OptionThrottle(65*time.Millisecond),
		pb.OptionShowCount(),
		pb.OptionSpinnerType(14),
		pb.OptionFullWidth(),
		pb.OptionSetDescription("Downloading"),
		pb.OptionOnCompletion(func() {
			fmt.Fprint(pbout, "\n")
		}),
		pb.OptionSetTheme(pb.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
}

// Download the file at 'url' and write the http response body to 'out'. The
// 'getbar' function allows the caller to construct a progress bar given the
// size of the file being downloaded, and the download will write to the
// returned progress bar.
func (app *Application) Download(url string, out io.Writer) error {
	if IsLocalFile(url) {
		f, err := os.Open(url)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(out, f)
		return err
	}

	resp, err := GetBinaryFile(url)
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

	bar := app.getDownloadProgressBar(resp.ContentLength)
	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)

	return err
}
