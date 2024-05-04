package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/permafrost-dev/eget/lib/home"
	pb "github.com/schollz/progressbar/v3"
)

func tokenFrom(s string) (string, error) {
	if strings.HasPrefix(s, "@") {
		f, err := home.Expand(s[1:])
		if err != nil {
			return "", err
		}

		b, _ := os.ReadFile(f)
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

type RateLimitJSON struct {
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
	rTime := r.ResetTime()

	if rTime.Before(now) {
		return fmt.Sprintf("Limit: %d, Remaining: %d, Reset: %v", r.Limit, r.Remaining, rTime)
	}

	return fmt.Sprintf(
		"Limit: %d, Remaining: %d, Reset: %v (%v)",
		r.Limit, r.Remaining, rTime, rTime.Sub(now).Round(time.Second),
	)
}

func (app *Application) GetRateLimit() (RateLimit, error) {
	resp, err := app.DownloadClient().GetJSON("https://api.github.com/rate_limit")

	if err != nil {
		return RateLimit{}, err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return RateLimit{}, err
	}

	var parsed RateLimitJSON
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
	return app.DownloadClient().Download(url, out, app.getDownloadProgressBar)
}
