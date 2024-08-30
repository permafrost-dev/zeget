package verifiers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/zeget/lib/assets"
	"github.com/permafrost-dev/zeget/lib/mockhttp"
	"github.com/permafrost-dev/zeget/lib/verifiers"
)

var _ = Describe("Sha256SumFileAssetVerifier", func() {
	var (
		mockClient mockhttp.HTTPClient
		verifier   verifiers.Verifier
		asset      *assets.Asset
	)

	BeforeEach(func() {
		mockClient = mockhttp.NewMockHTTPClient()
		asset = &assets.Asset{
			Name:        "test-asset",
			DownloadURL: "https://example.com/test-asset",
		}
		verifier = &verifiers.Sha256SumFileAssetVerifier{
			Client:            mockClient,
			Sha256SumAssetURL: "https://example.com/sha256sums.txt",
			RealAssetURL:      "https://example.com/test-asset",
			BinaryName:        "test-asset",
			Asset:             asset,
		}

		mockClient.AddJSONResponse("https://example.com/test-asset", "fake data", 200)
	})

	Describe("Verify", func() {
		Context("when the asset's checksum matches", func() {
			It("should successfully verify the asset", func() {
				mockClient.AddJSONResponse("https://example.com/sha256sums.txt", "cac0164a3e553aafd2d84f4e83c1aa3e30289eeaa2e4627e66af9b2413fd4a06  test-asset", 200)

				data := []byte("fake data")
				err := verifier.Verify(data)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("when the asset's checksum does not match", func() {
			It("should return an error", func() {
				mockClient.AddJSONResponse("https://example.com/sha256sums.txt", "wrongchecksumhere  test-asset", 200)

				data := []byte("fake data")
				err := verifier.Verify(data)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(BeAssignableToTypeOf(&verifiers.Sha256Error{}))
			})
		})

		Context("when the sha256sums file cannot be found", func() {
			It("should return an error", func() {
				mockClient.AddJSONResponse("https://example.com/sha256sums.txt", "", 404)

				data := []byte("fake data")
				err := verifier.Verify(data)
				Expect(err).Should(HaveOccurred())
			})
		})
	})

	It("should return a string representation of the verifier", func() {
		Expect(verifier.String()).To(Equal("checksum verified with https://example.com/sha256sums.txt"))
	})

	It("should return the correct asset", func() {
		Expect(verifier.GetAsset()).To(Equal(asset))
	})
})
