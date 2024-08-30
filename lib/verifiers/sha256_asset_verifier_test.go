package verifiers_test

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/zeget/lib/assets"
	"github.com/permafrost-dev/zeget/lib/mockhttp"
	"github.com/permafrost-dev/zeget/lib/verifiers"
)

var _ = Describe("Sha256AssetVerifier", func() {
	var (
		mockClient mockhttp.HTTPClient
		verifier   verifiers.Sha256AssetVerifier
		assetURL   string
		asset      assets.Asset
	)

	BeforeEach(func() {
		mockClient = mockhttp.NewMockHTTPClient()

		assetURL = "https://example.com/asset"
		asset = assets.Asset{Name: "test", DownloadURL: assetURL, Filters: nil, ReleaseDate: time.Now().Round(time.Second)}
		verifier = verifiers.Sha256AssetVerifier{
			AssetURL: assetURL,
			Asset:    &asset,
		}
		verifier.WithClient(&mockClient)
	})

	Context("Verify", func() {
		It("should verify asset with correct sha256 checksum", func() {
			// Simulate downloading asset with known checksum
			data := []byte("test data")
			sum := sha256.Sum256(data)
			hexSum := hex.EncodeToString(sum[0:])

			mockClient.AddJSONResponse(assetURL, ``+hexSum+``, 200)
			verifier.WithClient(&mockClient)

			verifier.AssetURL = assetURL
			err := verifier.Verify(data)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should fail to verify asset with incorrect sha256 checksum", func() {
			// Simulate downloading asset with incorrect checksum
			data := []byte("test data")
			mockClient.AddJSONResponse(assetURL, `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`, 200)
			verifier.WithClient(&mockClient)
			verifier.AssetURL = assetURL

			err := verifier.Verify(data)
			Expect(err).Should(HaveOccurred())
			Expect(err).Should(BeAssignableToTypeOf(&verifiers.Sha256Error{}))
		})

		It("should fail if asset URL is not reachable", func() {
			// Simulate HTTP error
			mockClient.AddJSONResponse(assetURL, ``, 500)
			verifier.WithClient(&mockClient)

			verifier.AssetURL = assetURL
			err := verifier.Verify([]byte("test data"))
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("String", func() {
		It("should return a descriptive string", func() {
			str := verifier.String()
			Expect(str).Should(ContainSubstring("checksum verified with " + assetURL))
		})
	})

	Context("WithClient", func() {
		It("should set the client", func() {
			newMockClient := mockhttp.NewMockHTTPClient()
			newVerifier := verifier.WithClient(&newMockClient)
			Expect(newVerifier).ShouldNot(BeNil())
		})
	})

	It("should return the correct asset", func() {
		Expect(verifier.GetAsset()).To(Equal(&asset))
	})
})
