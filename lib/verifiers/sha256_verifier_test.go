package verifiers_test

import (
	"crypto/sha256"
	"encoding/hex"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/eget/lib/download"
	"github.com/permafrost-dev/eget/lib/mockhttp"
	"github.com/permafrost-dev/eget/lib/verifiers"
)

var _ = Describe("Sha256Verifier", func() {
	var (
		mockClient download.ClientContract
		// asset      *assets.Asset
		// verifier   verifiers.Verifier
	)

	BeforeEach(func() {
		mockClient = mockhttp.NewMockHTTPClient()
		// asset = &assets.Asset{
		// 	Name:        "testAsset",
		// 	DownloadURL: "http://example.com/asset",
		// }
	})

	Context("when creating a new Sha256Verifier with valid hex string", func() {
		It("should not return an error", func() {
			expectedHex := sha256.Sum256([]byte("test"))
			hexString := hex.EncodeToString(expectedHex[:])

			v, err := verifiers.NewSha256Verifier(mockClient, hexString)
			Expect(err).NotTo(HaveOccurred())
			Expect(v).NotTo(BeNil())
		})
	})

	Context("when creating a new Sha256Verifier with invalid hex string", func() {
		It("should return an error", func() {
			_, err := verifiers.NewSha256Verifier(mockClient, "invalidhex")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when verifying an asset with matching checksum", func() {
		It("should verify successfully", func() {
			data := []byte("test")
			expectedHex := sha256.Sum256(data)
			hexString := hex.EncodeToString(expectedHex[:])
			v, _ := verifiers.NewSha256Verifier(mockClient, hexString)

			err := v.Verify(data)
			Expect(err).To(BeNil())
		})
	})

	Context("when verifying an asset with non-matching checksum", func() {
		It("should return an error", func() {
			data := []byte("test")
			expectedHex := sha256.Sum256([]byte("different"))
			hexString := hex.EncodeToString(expectedHex[:])
			v, _ := verifiers.NewSha256Verifier(mockClient, hexString)

			err := v.Verify(data)
			Expect(err).ToNot(BeNil())
		})
	})

	It("should return a string representation of the verifier", func() {
		expectedHex := sha256.Sum256([]byte("test"))
		hexString := hex.EncodeToString(expectedHex[:])
		v, _ := verifiers.NewSha256Verifier(mockClient, hexString)

		Expect(v.String()).To(Equal("sha256:" + hexString))
	})

	It("should return the correct asset", func() {
		expectedHex := sha256.Sum256([]byte("test"))
		hexString := hex.EncodeToString(expectedHex[:])
		v, _ := verifiers.NewSha256Verifier(mockClient, hexString)

		Expect(v.GetAsset()).To(BeNil())
	})
})
