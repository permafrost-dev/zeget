package finders

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	. "github.com/permafrost-dev/eget/lib/assets"
	"github.com/permafrost-dev/eget/lib/download"
	"github.com/permafrost-dev/eget/lib/github"
	"github.com/permafrost-dev/eget/lib/utilities"
)

// A GithubAssetFinder finds assets for the given Repo at the given tag. Tags
// must be given as 'tag/<tag>'. Use 'latest' to get the latest release.

type GithubAssetFinder struct {
	Finder

	Repo       string
	Tag        string
	Prerelease bool
	MinTime    time.Time // release must be after MinTime to be found
}

func NewGithubAssetFinder(repo *utilities.RepositoryReference, tag string, prerelease bool, minTime time.Time) *GithubAssetFinder {
	return &GithubAssetFinder{
		Repo:       repo.String(),
		Tag:        tag,
		Prerelease: prerelease,
		MinTime:    minTime,
	}
}

var ErrNoUpgrade = errors.New("requested release is not more recent than current version")

func (f GithubAssetFinder) Find(client download.ClientContract) *FindResult {
	if f.Prerelease && f.Tag == "latest" {
		tag, err := f.GetLatestTag(client)
		if err != nil {
			return NewInvalidFindResult(err)
		}
		f.Tag = "tags/" + tag
	}

	// query github's API for this repo/tag pair.
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/%s", f.Repo, f.Tag)
	resp, err := client.GetJSON(url)

	if err != nil {
		return NewInvalidFindResult(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return NewInvalidFindResult(err)
		}
		if strings.HasPrefix(f.Tag, "tags/") && resp.StatusCode == http.StatusNotFound {
			return f.FindMatch(client)
		}
		return NewInvalidFindResult(&github.Error{
			Status: resp.Status,
			Code:   resp.StatusCode,
			Body:   body,
			URL:    url,
		})
	}

	// read and unmarshal the resulting json
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewInvalidFindResult(err)
	}

	var release github.Release
	err = json.Unmarshal(body, &release)
	if err != nil {
		return NewInvalidFindResult(err)
	}

	release.ProcessReleaseAssets()

	if release.CreatedAt.Before(f.MinTime) {
		return NewInvalidFindResult(ErrNoUpgrade)
	}

	// accumulate all assets from the json into a slice
	assets := make([]Asset, len(release.Assets))
	for idx, a := range release.Assets {
		assets[idx] = a.CopyToNewAsset()
	}

	return NewFindResult(assets, nil)
}

func (f *GithubAssetFinder) FindMatch(client download.ClientContract) *FindResult {
	var tag = f.Tag

	if strings.HasPrefix(f.Tag, "tags/") {
		tag = f.Tag[len("tags/"):]
	}

	for page := 1; ; page++ {
		url := fmt.Sprintf("https://api.github.com/repos/%s/releases?page=%d", f.Repo, page)
		resp, err := client.GetJSON(url)
		if err != nil {
			return NewInvalidFindResult(err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return NewInvalidFindResult(err)
			}
			return NewInvalidFindResult(&github.Error{
				Status: resp.Status,
				Code:   resp.StatusCode,
				Body:   body,
				URL:    url,
			})
		}

		// read and unmarshal the resulting json
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return NewInvalidFindResult(err)
		}

		var releases []github.Release
		if err = json.Unmarshal(body, &releases); err != nil {
			return NewInvalidFindResult(err)
		}

		for _, r := range releases {
			r.ProcessReleaseAssets()
			if !f.Prerelease && r.Prerelease {
				continue
			}
			if strings.Contains(r.Tag, tag) && !r.CreatedAt.Before(f.MinTime) {
				// we have a winner
				assets := make([]Asset, 0, len(r.Assets))

				for _, a := range r.Assets {
					assets = append(assets, a.CopyToNewAsset())
				}

				return NewFindResult(assets, nil)
			}
		}

		if len(releases) < 30 || page > 20 {
			break
		}
	}

	return NewInvalidFindResult(fmt.Errorf("no matching tag for '%s'", tag))
}

// finds the latest pre-release and returns the tag
func (f *GithubAssetFinder) GetLatestTag(client download.ClientContract) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", f.Repo)
	resp, err := client.GetJSON(url)
	if err != nil {
		return "", fmt.Errorf("pre-release finder: %w", err)
	}

	var rel github.Release

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("pre-release finder: %w", err)
	}
	err = json.Unmarshal(body, &rel)
	if err != nil {
		return "", fmt.Errorf("pre-release finder: %w", err)
	}

	rel.ProcessReleaseAssets()

	// if len(rel) <= 0 {
	// 	return "", fmt.Errorf("no releases found")
	// }

	return rel.Tag, nil
}
