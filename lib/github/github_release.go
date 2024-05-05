package github

import "time"

type ReleaseAsset struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	DownloadURL string `json:"browser_download_url"`
}

// A Release matches the Assets portion of Github's release API json.
type Release struct {
	Assets     []ReleaseAsset `json:"assets"`
	Prerelease bool           `json:"prerelease"`
	Tag        string         `json:"tag_name"`
	CreatedAt  time.Time      `json:"created_at"`
}
