package finders_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/permafrost-dev/eget/lib/finders"
	. "github.com/permafrost-dev/eget/lib/mockhttp"
)

var _ = Describe("GithubAssetFinder", func() {
	var (
		client      HTTPClient
		assetFinder *GithubAssetFinder
	)

	BeforeEach(func() {
		client = NewMockHTTPClient()
		client.DoFunc = func(req *http.Request) (*http.Response, error) {
			return NewMockResponse("mock body", http.StatusOK), nil
		}

		client.AddJSONResponse("https://api.github.com/repos/testRepo/releases", `[{"tag_name": "v1.0.0", "prerelease": false, "assets": [{"name": "asset1", "browser_download_url": "http://example.com/asset1"}], "created_at": "2020-01-01T00:00:00Z"}]`, 200)
		client.AddJSONResponse("https://api.github.com/repos/testRepo/releases/latest", `{"tag_name": "v1.0.0", "prerelease": false, "assets": [{"name": "asset1", "browser_download_url": "http://example.com/asset1"}], "created_at": "2020-01-01T00:00:00Z"}`, 200)
		client.AddJSONResponse("https://api.github.com/repos/testRepo/releases/v1.0.0", `{"tag_name": "v1.0.0", "prerelease": false, "assets": [{"name": "asset1", "browser_download_url": "http://example.com/asset1"}], "created_at": "2020-01-01T00:00:00Z"}`, 200)
		client.AddJSONResponse("https://api.github.com/repos/testRepo/releases/v1.1.0", `{"tag_name": "v1.0.0", "prerelease": true, "assets": [{"name": "asset1", "browser_download_url": "http://example.com/asset1"}], "created_at": "2020-01-01T00:00:00Z"}`, 200)

		assetFinder = &GithubAssetFinder{
			Repo:       "testRepo",
			Tag:        "latest",
			Prerelease: false,
			MinTime:    time.Date(2019, 12, 31, 23, 59, 59, 0, time.UTC),
		}
	})

	AfterEach(func() {
		client.Reset()
	})

	Describe("Find", func() {
		Context("with a valid tag", func() {
			It("should return assets", func() {
				assetFinder.Tag = "v1.0.0"
				assetFinder.MinTime = time.Date(2018, 12, 31, 23, 59, 59, 0, time.UTC)
				assets, _ := assetFinder.Find(client)

				Expect(assets[0].Name).To(Equal("asset1"))
				Expect(assets[0].DownloadURL).To(Equal("http://example.com/asset1"))
			})
		})

		Context("with a tag that does not exist", func() {
			It("should return an error", func() {
				// dlclient := client.(download.ClientContract)
				assetFinder.Tag = "nonexistent"
				_, err := assetFinder.Find(client)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("find a match", func() {
			It("should return assets", func() {
				assetFinder.Tag = "tags/v1.0.0"
				assets, err := assetFinder.FindMatch(client)
				Expect(err).ToNot(HaveOccurred())
				Expect(assets).To(HaveLen(1))
				Expect(assets[0].Name).To(Equal("asset1"))
				Expect(assets[0].DownloadURL).To(Equal("http://example.com/asset1"))
			})
		})

		Context("find a match with a tag that does not exist", func() {
			It("should return an error", func() {
				assetFinder.Tag = "nonexistent"
				_, err := assetFinder.FindMatch(client)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("find latest tag", func() {
			It("should return a tag string", func() {
				tag, err := assetFinder.GetLatestTag(client)
				Expect(err).ToNot(HaveOccurred())
				Expect(tag).To(Equal("v1.0.0"))
			})
		})

		Context("request a tag that does not exist", func() {
			It("should return a an error", func() {
				assetFinder.Prerelease = false
				assetFinder.Tag = "v3.1.2"
				_, err := assetFinder.Find(client)
				Expect(err).To(HaveOccurred())
			})

			It("FindMatch return a an error", func() {
				assetFinder.Repo = "otherRepo"
				assetFinder.Prerelease = false
				assetFinder.Tag = "v3.1.2"
				_, err := assetFinder.FindMatch(client)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
