package finders

import (
	"github.com/permafrost-dev/eget/lib/assets"
	"github.com/permafrost-dev/eget/lib/download"
)

// A DirectAssetFinder returns the embedded URL directly as the only asset.
type DirectAssetFinder struct {
	URL string
}

func (f DirectAssetFinder) Find(_ download.ClientContract) *FindResult {
	asset := assets.Asset{
		Name:        f.URL,
		DownloadURL: f.URL,
	}

	return NewFindResult([]assets.Asset{asset}, nil)
}
