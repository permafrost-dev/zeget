package reporters_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/zeget/lib/assets"
	"github.com/permafrost-dev/zeget/lib/reporters"
)

var _ = Describe("AssetSha256HashReporter", func() {
	var (
		asset    *assets.Asset
		buffer   *bytes.Buffer
		reporter *reporters.AssetSha256HashReporter
	)

	BeforeEach(func() {
		asset = &assets.Asset{Name: "TestAsset"}
		buffer = new(bytes.Buffer)
		reporter = reporters.NewAssetSha256HashReporter(asset, buffer)
	})

	Describe("Report", func() {
		It("writes the SHA256 hash of the input and the asset name to the output", func() {
			err := reporter.Report("hello world")
			Expect(err).NotTo(HaveOccurred())
			Expect(buffer.String()).To(ContainSubstring("TestAsset"))
			Expect(buffer.String()).To(MatchRegexp(`› [a-f0-9]{64} TestAsset\n`))
		})
	})
})
