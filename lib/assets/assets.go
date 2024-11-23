package assets

import "time"

type Asset struct {
	Name        string    `json:"name"`
	DownloadURL string    `json:"download_url"`
	ReleaseDate time.Time `json:"release_date"`
	Filters     []string  `json:"filters"`
}

type AssetWrapper struct {
	Assets []Asset
	Asset  *Asset
}

func NewAssetWrapper(assets []Asset) *AssetWrapper {
	return &AssetWrapper{
		Assets: assets,
		Asset:  nil,
	}
}
