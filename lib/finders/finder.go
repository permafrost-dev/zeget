package finders

import (
	"github.com/permafrost-dev/eget/lib/assets"
	"github.com/permafrost-dev/eget/lib/download"
)

// A Finder returns a list of URLs making up a project's assets.
type Finder interface {
	Find(client *download.Client) ([]assets.Asset, error)
}
