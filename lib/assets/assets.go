package assets

import "time"

type Asset struct {
	Name        string
	DownloadURL string
	ReleaseDate time.Time
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
