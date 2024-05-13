package detectors_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/permafrost-dev/eget/lib/assets"
	. "github.com/permafrost-dev/eget/lib/detectors"
)

type MockDetector struct {
	DetectFunc func([]Asset) (DetectionResult, error)
}

func (m *MockDetector) Detect(assets []Asset) (DetectionResult, error) {
	return m.DetectFunc(assets)
}

var _ = Describe("DetectorChain", func() {
	var (
		detectorChain *DetectorChain
		mockDetector1 *MockDetector
		mockDetector2 *MockDetector
		mockSystem    *MockDetector
		assets        []Asset
	)

	BeforeEach(func() {
		mockDetector1 = &MockDetector{
			DetectFunc: func(a []Asset) (DetectionResult, error) {
				return DetectionResult{Candidates: a[:1]}, nil
			},
		}
		mockDetector2 = &MockDetector{
			DetectFunc: func(a []Asset) (DetectionResult, error) {
				return DetectionResult{Candidates: assets}, nil
			},
		}
		mockSystem = &MockDetector{
			DetectFunc: func(assets []Asset) (DetectionResult, error) {
				if len(assets) == 0 {
					return DetectionResult{}, errors.New("no assets found")
				}
				return DetectionResult{Candidates: assets}, nil
			},
		}

		detectorChain = &DetectorChain{
			Detectors: []Detector{mockDetector1, mockDetector2},
			System:    mockSystem,
		}

		assets = []Asset{{}, {}, {}}
	})

	Describe("Detect", func() {
		Context("when detectors filter assets successfully", func() {
			It("should return the final detection result with error for remaining candidates", func() {
				result, err := detectorChain.Detect(assets)
				Expect(err).To(HaveOccurred())
				Expect(result.Asset).To(Equal(assets[0]))
			})
		})

		Context("when a detector returns an error", func() {
			BeforeEach(func() {
				mockDetector1.DetectFunc = func(assets []Asset) (DetectionResult, error) {
					return DetectionResult{}, errors.New("detection failed")
				}
			})

			It("should return an error immediately", func() {
				_, err := detectorChain.Detect(assets)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("detection failed"))
			})
		})

		Context("when no candidates are found", func() {
			BeforeEach(func() {
				mockSystem.DetectFunc = func(assets []Asset) (DetectionResult, error) {
					return DetectionResult{}, errors.New("no assets found")
				}
			})

			It("should return an error from the system detector", func() {
				result, err := detectorChain.Detect(assets[:0]) // Empty assets list
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no assets found"))
				Expect(len(result.Candidates)).To(Equal(0))
			})
		})
	})
})
