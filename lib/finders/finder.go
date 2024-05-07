package finders

import (
	"github.com/permafrost-dev/eget/lib/assets"
	"github.com/permafrost-dev/eget/lib/download"
)

// A Finder returns a list of URLs making up a project's assets.
type Finder interface {
	Find(client download.ClientContract) *FindResult
}

type FindResult struct {
	Assets []assets.Asset
	Error  error
}

func NewFindResult(assets []assets.Asset, err error) *FindResult {
	return &FindResult{
		Assets: assets,
		Error:  err,
	}
}

func NewInvalidFindResult(err error) *FindResult {
	return NewFindResult([]assets.Asset{}, err)
}

type ValidFinder struct {
	Finder Finder
	Tool   string
}

func NewValidFinder(finder Finder, tool string) ValidFinder {
	return ValidFinder{
		Finder: finder,
		Tool:   tool,
	}
}
