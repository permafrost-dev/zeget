package finders_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/eget/lib/assets"
	"github.com/permafrost-dev/eget/lib/finders"
	"github.com/permafrost-dev/eget/lib/mockhttp"
)

var _ = Describe("GithubSourceFinder", func() {
	var (
		githubFinder *finders.GithubSourceFinder
		client       mockhttp.HTTPClient
	)

	BeforeEach(func() {
		githubFinder = &finders.GithubSourceFinder{
			Tool: "exampleTool",
			Repo: "example/repo",
			Tag:  "v1.0.0",
		}

		client = mockhttp.NewMockHTTPClient()
		client.DoFunc = func(req *http.Request) (*http.Response, error) {
			return mockhttp.NewMockResponse("mock body", http.StatusOK), nil
		}

		client.AddJSONResponse("https://api.github.com/repos/testRepo/releases/v1.0.0", `{"tag_name": "v1.0.0", "prerelease": false, "assets": [{"name": "exampleTool.tar.gz", "browser_download_url": "https://github.com/example/repo/tarball/v1.0.0/exampleTool.tar.gz"}], "created_at": "2020-01-01T00:00:00Z"}`, 200)
	})

	Describe("Finding assets", func() {
		It("should correctly construct the asset's download URL", func() {
			expectedAsset := assets.Asset{
				Name:        "exampleTool.tar.gz",
				DownloadURL: "https://github.com/example/repo/tarball/v1.0.0/exampleTool.tar.gz",
			}

			client := mockhttp.NewMockHTTPClient()
			findResult := githubFinder.Find(client)

			Expect(findResult.Error).NotTo(HaveOccurred())
			Expect(findResult.Assets).To(HaveLen(1))
			Expect(findResult.Assets[0]).To(Equal(expectedAsset))
		})
	})
})
