package github

import (
	"time"

	"github.com/permafrost-dev/zeget/lib/assets"
)

type ReleaseAsset struct {
	Release       *Release
	Name          string `json:"name"`
	URL           string `json:"url"`
	DownloadURL   string `json:"browser_download_url"`
	Size          int64  `json:"size"`
	DownloadCount int64  `json:"download_count"`
	ContentType   string `json:"content_type"`
}

// A Release matches the Assets portion of Github's release API json.
type Release struct {
	Assets      []ReleaseAsset `json:"assets"`
	Prerelease  bool           `json:"prerelease"`
	Tag         string         `json:"tag_name"`
	CreatedAt   time.Time      `json:"created_at"`
	PublishedAt time.Time      `json:"published_at"`
}

func (r *Release) ProcessReleaseAssets() {
	for i, asset := range r.Assets {
		asset.Release = r
		r.Assets[i] = asset
	}
}

func (ra *ReleaseAsset) CopyToNewAsset() assets.Asset {
	return assets.Asset{
		Name:        ra.Name,
		DownloadURL: ra.DownloadURL,
		ReleaseDate: ra.Release.PublishedAt,
	}
}
