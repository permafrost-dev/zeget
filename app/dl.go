package app

import (
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

func (app *Application) getDownloadProgressBar(size int64) *pb.ProgressBar {
	var pbout io.Writer = app.Outputs.Stderr
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
