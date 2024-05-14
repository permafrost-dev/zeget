package detectors_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/permafrost-dev/zeget/lib/assets"
	. "github.com/permafrost-dev/zeget/lib/detectors"
)

var _ = Describe("SystemDetector", func() {
	var (
		systemDetector *SystemDetector
		err            error
		assets         []Asset
		result         DetectionResult
	)

	BeforeEach(func() {
		assets = []Asset{
			{Name: "asset-linux-amd64.tar.gz"},
			{Name: "asset-linux-arm64.tar.gz"},
			{Name: "asset-windows-amd64.zip"},
			{Name: "asset-windows-arm64.zip"},
			{Name: "asset-darwin-amd64.tar.gz"},
			{Name: "asset-darwin-arm64.tar.gz"},
			{Name: "other-file.sha256"},
		}
	})

	Context("when OS and Arch are supported", func() {
		BeforeEach(func() {
			systemDetector, err = NewSystemDetector("linux", "amd64")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should detect the correct asset for linux/amd64", func() {
			result, err = systemDetector.Detect(assets)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Asset).To(Equal(assets[0]))
		})

		It("should return candidates if multiple matches found", func() {
			systemDetector, _ = NewSystemDetector("linux", "amd64")
			result, err = systemDetector.Detect(assets)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Asset).To(Equal(assets[0]))
			Expect(result.Candidates).To(BeEmpty())
		})
	})

	Context("when OS or Arch is unsupported", func() {
		It("should return an error for unsupported OS", func() {
			_, err = NewSystemDetector("unsupported-os", "amd64")
			Expect(err).To(HaveOccurred())
		})

		It("should return an error for unsupported Arch", func() {
			_, err = NewSystemDetector("linux", "unsupported-arch")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when no exact match is found", func() {
		BeforeEach(func() {
			systemDetector, err = NewSystemDetector("linux", "riscv64")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return all assets as candidates if no matches are found", func() {
			result, err = systemDetector.Detect(assets)
			Expect(result.Candidates).To(HaveLen(2))
		})
	})
})
