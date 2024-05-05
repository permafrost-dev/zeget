package github

import "time"

// A Release matches the Assets portion of Github's release API json.
type Release struct {
	Assets []struct {
		Name        string `json:"name"`
		URL         string `json:"url"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`

	Prerelease bool      `json:"prerelease"`
	Tag        string    `json:"tag_name"`
	CreatedAt  time.Time `json:"created_at"`
}
