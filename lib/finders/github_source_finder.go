package finders

import (
	"fmt"

	"github.com/permafrost-dev/eget/lib/assets"
	"github.com/permafrost-dev/eget/lib/download"
)

type GithubSourceFinder struct {
	Finder

	Tool string
	Repo string
	Tag  string
}

func (f GithubSourceFinder) Find(_ download.ClientContract) *FindResult {
	name := fmt.Sprintf("%s.tar.gz", f.Tool)

	asset := assets.Asset{
		Name:        name,
		DownloadURL: fmt.Sprintf("https://github.com/%s/tarball/%s/%s", f.Repo, f.Tag, name),
	}

	return NewFindResult([]assets.Asset{asset}, nil)
}
