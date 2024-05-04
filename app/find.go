package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/permafrost-dev/eget/lib/download"
)

type Asset struct {
	Name        string
	DownloadURL string
}

// A Finder returns a list of URLs making up a project's assets.
type Finder interface {
	Find(client *download.Client) ([]Asset, error)
}

// A GithubRelease matches the Assets portion of Github's release API json.
type GithubRelease struct {
	Assets []struct {
		Name        string `json:"name"`
		URL         string `json:"url"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`

	Prerelease bool      `json:"prerelease"`
	Tag        string    `json:"tag_name"`
	CreatedAt  time.Time `json:"created_at"`
}

// A GithubAssetFinder finds assets for the given Repo at the given tag. Tags
// must be given as 'tag/<tag>'. Use 'latest' to get the latest release.
type GithubAssetFinder struct {
	Repo       string
	Tag        string
	Prerelease bool
	MinTime    time.Time // release must be after MinTime to be found
}

var ErrNoUpgrade = errors.New("requested release is not more recent than current version")

func (f *GithubAssetFinder) Find(client *download.Client) ([]Asset, error) {
	if f.Prerelease && f.Tag == "latest" {
		tag, err := f.getLatestTag(client)
		if err != nil {
			return nil, err
		}
		f.Tag = fmt.Sprintf("tags/%s", tag)
	}

	// query github's API for this repo/tag pair.
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/%s", f.Repo, f.Tag)
	resp, err := client.GetJSON(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(f.Tag, "tags/") && resp.StatusCode == http.StatusNotFound {
			return f.FindMatch(client)
		}
		return nil, &GithubError{
			Status: resp.Status,
			Code:   resp.StatusCode,
			Body:   body,
			URL:    url,
		}
	}

	// read and unmarshal the resulting json
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release GithubRelease
	err = json.Unmarshal(body, &release)
	if err != nil {
		return nil, err
	}

	if release.CreatedAt.Before(f.MinTime) {
		return nil, ErrNoUpgrade
	}

	// accumulate all assets from the json into a slice
	assets := make([]Asset, 0, len(release.Assets))
	for _, a := range release.Assets {
		assets = append(assets, Asset{Name: a.Name, DownloadURL: a.URL})
	}

	return assets, nil
}

func (f *GithubAssetFinder) FindMatch(client *download.Client) ([]Asset, error) {
	tag := f.Tag[len("tags/"):]

	for page := 1; ; page++ {
		url := fmt.Sprintf("https://api.github.com/repos/%s/releases?page=%d", f.Repo, page)
		resp, err := client.GetJSON(url)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			return nil, &GithubError{
				Status: resp.Status,
				Code:   resp.StatusCode,
				Body:   body,
				URL:    url,
			}
		}

		// read and unmarshal the resulting json
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var releases []GithubRelease
		err = json.Unmarshal(body, &releases)
		if err != nil {
			return nil, err
		}

		for _, r := range releases {
			if !f.Prerelease && r.Prerelease {
				continue
			}
			if strings.Contains(r.Tag, tag) && !r.CreatedAt.Before(f.MinTime) {
				// we have a winner
				assets := make([]Asset, 0, len(r.Assets))
				for _, a := range r.Assets {
					assets = append(assets, Asset{Name: a.Name, DownloadURL: a.URL})
				}
				return assets, nil
			}
		}

		if len(releases) < 30 {
			break
		}

		if page > 20 {
			break
		}
	}

	return nil, fmt.Errorf("no matching tag for '%s'", tag)
}

// finds the latest pre-release and returns the tag
func (f *GithubAssetFinder) getLatestTag(client *download.Client) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", f.Repo)
	resp, err := client.GetJSON(url)
	if err != nil {
		return "", fmt.Errorf("pre-release finder: %w", err)
	}

	var releases []GithubRelease

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("pre-release finder: %w", err)
	}
	err = json.Unmarshal(body, &releases)
	if err != nil {
		return "", fmt.Errorf("pre-release finder: %w", err)
	}

	if len(releases) <= 0 {
		return "", fmt.Errorf("no releases found")
	}

	return releases[0].Tag, nil
}

// A DirectAssetFinder returns the embedded URL directly as the only asset.
type DirectAssetFinder struct {
	URL string
}

func (f *DirectAssetFinder) Find(_ *download.Client) ([]Asset, error) {
	asset := Asset{
		Name:        f.URL,
		DownloadURL: f.URL,
	}

	return []Asset{asset}, nil
}

type GithubSourceFinder struct {
	Tool string
	Repo string
	Tag  string
}

func (f *GithubSourceFinder) Find(_ *download.Client) ([]Asset, error) {
	name := fmt.Sprintf("%s.tar.gz", f.Tool)
	asset := Asset{
		Name:        name,
		DownloadURL: fmt.Sprintf("https://github.com/%s/tarball/%s/%s", f.Repo, f.Tag, name),
	}

	return []Asset{asset}, nil
}
