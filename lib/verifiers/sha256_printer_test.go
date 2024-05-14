package verifiers_test

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/permafrost-dev/eget/lib/assets"
	. "github.com/permafrost-dev/eget/lib/verifiers"
)

var _ = ginkgo.Describe("Sha256Printer", func() {
	var (
		asset         *assets.Asset
		sha256Printer Sha256Printer
	)

	ginkgo.BeforeEach(func() {
		// Setup your asset and Sha256Printer here
		asset = &assets.Asset{
			Name: "TestAsset",
			// Populate other necessary fields if needed
		}
		sha256Printer = Sha256Printer{
			Asset: asset,
		}
	})

	ginkgo.Describe("GetAsset", func() {
		ginkgo.It("should return the correct asset", func() {
			gomega.Expect(sha256Printer.GetAsset()).To(gomega.Equal(asset))
		})
	})

	ginkgo.Describe("WithClient", func() {
		ginkgo.It("should return a new Sha256Printer instance with the same asset", func() {
			newSha256Printer := sha256Printer.WithClient(nil)
			gomega.Expect(newSha256Printer).ToNot(gomega.BeNil())
			gomega.Expect(newSha256Printer.GetAsset()).To(gomega.Equal(asset))
		})
	})

	ginkgo.Describe("Verify", func() {
		ginkgo.It("should not return an error", func() {
			err := sha256Printer.Verify([]byte("test"))
			gomega.Expect(err).To(gomega.BeNil())
		})

		// Note: Since Verify prints the SHA256 sum, we won't check the output here.
		// In a real test, you might use a mock to capture stdout or modify the method to return the sum for testing.
	})

	ginkgo.Describe("String", func() {
		ginkgo.It("should return the correct string representation", func() {
			gomega.Expect(sha256Printer.String()).To(gomega.Equal("sha256:print"))
		})
	})
})
