package detectors_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/permafrost-dev/eget/lib/detectors"

	"github.com/permafrost-dev/eget/lib/assets"
)

var _ = Describe("lib/detectors > DetectionResult", func() {
	It("should return a DetectionResult with the correct asset and candidates 1", func() {
		asset := NewDetectionResult(&assets.Asset{}, []assets.Asset{})
		Expect(asset.Asset).To(Equal(assets.Asset{}))
		Expect(asset.Candidates).To(BeEmpty())
	})

	It("should return a DetectionResult with the correct asset and candidates 2", func() {
		asset := NewDetectionResult(nil, []assets.Asset{assets.Asset{Name: "test"}})
		Expect(asset.Asset.Name).To(Equal("test"))
		Expect(len(asset.Candidates)).To(Equal(0))
	})
})
