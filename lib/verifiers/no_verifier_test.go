package verifiers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/zeget/lib/assets"
	"github.com/permafrost-dev/zeget/lib/download"
	. "github.com/permafrost-dev/zeget/lib/verifiers"
)

var _ = Describe("NoVerifier", func() {
	var (
		noVerifier *NoVerifier
		asset      *assets.Asset
	)

	BeforeEach(func() {
		asset = &assets.Asset{Name: "test-asset", DownloadURL: "http://example.com/test-asset"}
		noVerifier = &NoVerifier{Asset: asset}
	})

	Describe("GetAsset", func() {
		It("should return the correct asset", func() {
			Expect(noVerifier.GetAsset()).To(Equal(asset))
		})
	})

	Describe("Verify", func() {
		It("should always return nil", func() {
			Expect(noVerifier.Verify(nil)).To(BeNil())
		})
	})

	Describe("WithClient", func() {
		It("should return a Verifier type", func() {
			client := &download.Client{}
			Expect(noVerifier.WithClient(client)).To(BeAssignableToTypeOf(NoVerifier{}))
		})

		It("should not panic", func() {
			client := &download.Client{}
			Expect(func() { noVerifier.WithClient(client) }).ToNot(Panic())
		})
	})
})
