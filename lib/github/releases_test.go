package github_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/zeget/lib/assets"
	. "github.com/permafrost-dev/zeget/lib/github"
)

var _ = Describe("Release", func() {
	var (
		release Release
	)

	BeforeEach(func() {
		release = Release{
			Assets: []ReleaseAsset{
				{
					Name:          "asset1",
					URL:           "http://example.com/asset1",
					DownloadURL:   "http://example.com/download/asset1",
					Size:          1024,
					DownloadCount: 100,
					ContentType:   "application/octet-stream",
				},
				{
					Name:          "asset2",
					URL:           "http://example.com/asset2",
					DownloadURL:   "http://example.com/download/asset2",
					Size:          2048,
					DownloadCount: 200,
					ContentType:   "application/octet-stream",
				},
			},
			Prerelease:  false,
			Tag:         "v1.0.0",
			CreatedAt:   time.Now(),
			PublishedAt: time.Now(),
		}
	})

	Describe("ProcessReleaseAssets", func() {
		It("should correctly associate each asset with its release", func() {
			release.ProcessReleaseAssets()
			for _, asset := range release.Assets {
				Expect(asset.Release).To(Equal(&release))
			}
		})
	})
})

var _ = Describe("ReleaseAsset", func() {
	var (
		releaseAsset ReleaseAsset
		release      Release
	)

	BeforeEach(func() {
		release = Release{
			Assets:      nil,
			Prerelease:  false,
			Tag:         "v1.0.0",
			CreatedAt:   time.Now(),
			PublishedAt: time.Now(),
		}

		releaseAsset = ReleaseAsset{
			Release:       &release,
			Name:          "asset",
			URL:           "http://example.com/asset",
			DownloadURL:   "http://example.com/download/asset",
			Size:          1024,
			DownloadCount: 100,
			ContentType:   "application/octet-stream",
		}
	})

	Describe("CopyToNewAsset", func() {
		It("should correctly copy ReleaseAsset to Asset", func() {
			copiedAsset := releaseAsset.CopyToNewAsset()
			Expect(copiedAsset).To(Equal(assets.Asset{
				Name:        releaseAsset.Name,
				DownloadURL: releaseAsset.DownloadURL,
				ReleaseDate: releaseAsset.Release.PublishedAt,
			}))
		})
	})
})
