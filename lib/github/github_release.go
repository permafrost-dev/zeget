package github

import "time"

type ReleaseAsset struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	DownloadURL   string `json:"browser_download_url"`
	Size          int64  `json:"size"`
	DownloadCount int64  `json:"download_count"`
	ContentType   string `json:"content_type"`
}

// A Release matches the Assets portion of Github's release API json.
type Release struct {
	Assets     []ReleaseAsset `json:"assets"`
	Prerelease bool           `json:"prerelease"`
	Tag        string         `json:"tag_name"`
	CreatedAt  time.Time      `json:"created_at"`
}
