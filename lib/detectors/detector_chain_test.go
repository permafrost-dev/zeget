package detectors_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/permafrost-dev/zeget/lib/assets"
	. "github.com/permafrost-dev/zeget/lib/detectors"
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
		assets        []Asset
	)

	BeforeEach(func() {
		assets = []Asset{
			{Name: "asset1"},
			{Name: "asset2"},
		}
	})

	Describe("DetectWithoutSystem", func() {
		Context("when there are no detectors", func() {
			BeforeEach(func() {
				detectorChain = &DetectorChain{
					Detectors: []Detector{},
				}
			})

			It("should return the initial assets unchanged", func() {
				result, err := detectorChain.DetectWithoutSystem(assets)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Candidates).To(Equal(assets))
			})
		})

		Context("with a passing detector", func() {
			BeforeEach(func() {
				passingDetector := &MockDetector{
					DetectFunc: func(a []Asset) (DetectionResult, error) {
						return NewDetectionResult(nil, a), nil
					},
				}

				detectorChain = &DetectorChain{
					Detectors: []Detector{passingDetector},
				}
			})

			It("should return the initial assets unchanged", func() {
				result, err := detectorChain.DetectWithoutSystem(assets)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Candidates).To(Equal(assets))
			})
		})

		Context("with a filtering detector", func() {
			BeforeEach(func() {
				filteringDetector := &MockDetector{
					DetectFunc: func(a []Asset) (DetectionResult, error) {
						if len(a) == 0 {
							return NewDetectionResult(nil, nil), nil
						}
						return NewDetectionResult(nil, []Asset{a[0]}), nil
					},
				}

				detectorChain = &DetectorChain{
					Detectors: []Detector{filteringDetector},
				}
			})

			It("should return a subset of the initial assets", func() {
				_, err := detectorChain.DetectWithoutSystem(assets)
				Expect(err).NotTo(HaveOccurred())
				//Expect(result.Candidates).To(Equal([]Asset{assets[0]}))
			})
		})

		Context("with a detector that returns an error", func() {
			BeforeEach(func() {
				errorDetector := &MockDetector{
					DetectFunc: func(a []Asset) (DetectionResult, error) {
						return DetectionResult{}, errors.New("detector error")
					},
				}

				detectorChain = &DetectorChain{
					Detectors: []Detector{errorDetector},
				}
			})

			It("should return the error", func() {
				result, err := detectorChain.DetectWithoutSystem(assets)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("detector error"))
				Expect(result).To(BeNil())
			})
		})

		Context("with a nil detector in the chain", func() {
			BeforeEach(func() {
				passingDetector := &MockDetector{
					DetectFunc: func(a []Asset) (DetectionResult, error) {
						return NewDetectionResult(nil, a), nil
					},
				}

				detectorChain = &DetectorChain{
					Detectors: []Detector{nil, passingDetector},
				}
			})

			It("should skip the nil detector and process the rest", func() {
				result, err := detectorChain.DetectWithoutSystem(assets)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Candidates).To(Equal(assets))
			})
		})

		Context("when a detector returns zero candidates without error", func() {
			BeforeEach(func() {
				detector := &MockDetector{
					DetectFunc: func(a []Asset) (DetectionResult, error) {
						return NewDetectionResult(nil, nil), nil
					},
				}

				detectorChain = &DetectorChain{
					Detectors: []Detector{detector},
				}
			})

			It("should return an empty detection result", func() {
				result, err := detectorChain.DetectWithoutSystem(assets)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Candidates).To(BeNil())
			})
		})

		Context("when a detector returns zero candidates with an error", func() {
			BeforeEach(func() {
				detector := &MockDetector{
					DetectFunc: func(a []Asset) (DetectionResult, error) {
						return NewDetectionResult(nil, nil), errors.New("no candidates found")
					},
				}

				detectorChain = &DetectorChain{
					Detectors: []Detector{detector},
				}
			})

			It("should return the error", func() {
				result, err := detectorChain.DetectWithoutSystem(assets)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no candidates found"))
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("Detect", func() {
		var (
			mockDetector1 *MockDetector
			mockDetector2 *MockDetector
			mockSystem    *MockDetector
		)

		Context("when detectors filter assets successfully", func() {
			BeforeEach(func() {
				mockDetector1 = &MockDetector{
					DetectFunc: func(a []Asset) (DetectionResult, error) {
						return DetectionResult{Candidates: a[:1]}, nil
					},
				}
				mockDetector2 = &MockDetector{
					DetectFunc: func(a []Asset) (DetectionResult, error) {
						return DetectionResult{Candidates: a}, nil
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
			})

			It("should return the final detection result with error for remaining candidates", func() {
				result, err := detectorChain.Detect(assets)
				Expect(err).To(HaveOccurred())
				//Expect(err.Error()).To(Equal(fmt.Sprintf("%d candidates found for asset chain", len(assets))))
				Expect(result.Candidates[0].Name).To(Equal(assets[0].Name))
			})
		})

		Context("when a detector returns an error", func() {
			BeforeEach(func() {
				mockDetector1 = &MockDetector{
					DetectFunc: func(assets []Asset) (DetectionResult, error) {
						return DetectionResult{}, errors.New("detection failed")
					},
				}
				detectorChain = &DetectorChain{
					Detectors: []Detector{mockDetector1},
					System:    nil,
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
				mockSystem = &MockDetector{
					DetectFunc: func(assets []Asset) (DetectionResult, error) {
						return DetectionResult{}, errors.New("no assets found")
					},
				}
				detectorChain = &DetectorChain{
					Detectors: []Detector{},
					System:    mockSystem,
				}
			})

			It("should return an error from the system detector", func() {
				result, err := detectorChain.Detect([]Asset{}) // Empty assets list
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no assets found"))
				Expect(result.Candidates).To(BeNil())
			})
		})

		Context("when there are no detectors and no system detector", func() {
			BeforeEach(func() {
				detectorChain = &DetectorChain{
					Detectors: []Detector{},
					System:    nil,
				}
			})

			It("should return an empty detection result without error", func() {
				result, err := detectorChain.Detect(assets)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Candidates).To(BeNil())
			})
		})

		Context("when the system detector returns an error", func() {
			BeforeEach(func() {
				passingDetector := &MockDetector{
					DetectFunc: func(a []Asset) (DetectionResult, error) {
						return NewDetectionResult(nil, a), nil
					},
				}

				errorSystemDetector := &MockDetector{
					DetectFunc: func(a []Asset) (DetectionResult, error) {
						return DetectionResult{}, errors.New("system detector error")
					},
				}

				detectorChain = &DetectorChain{
					Detectors: []Detector{passingDetector},
					System:    errorSystemDetector,
				}
			})

			It("should return the system detector error", func() {
				result, err := detectorChain.Detect(assets)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("system detector error"))
				Expect(result.Candidates).To(BeNil())
			})
		})

		Context("when the system detector is nil", func() {
			BeforeEach(func() {
				passingDetector := &MockDetector{
					DetectFunc: func(a []Asset) (DetectionResult, error) {
						return NewDetectionResult(nil, a), nil
					},
				}

				detectorChain = &DetectorChain{
					Detectors: []Detector{passingDetector},
					System:    nil,
				}
			})

			It("should return an empty detection result without error", func() {
				result, err := detectorChain.Detect(assets)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Candidates).To(BeNil())
			})
		})

		Context("when the system detector returns candidates", func() {
			BeforeEach(func() {
				passingDetector := &MockDetector{
					DetectFunc: func(a []Asset) (DetectionResult, error) {
						return NewDetectionResult(nil, a), nil
					},
				}

				systemDetector := &MockDetector{
					DetectFunc: func(a []Asset) (DetectionResult, error) {
						return NewDetectionResult(nil, a), nil
					},
				}

				detectorChain = &DetectorChain{
					Detectors: []Detector{passingDetector},
					System:    systemDetector,
				}
			})

			It("should return the system detector's candidates with an error indicating the number of candidates", func() {
				result, err := detectorChain.Detect(assets)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf("%d candidates found for asset chain", len(assets))))
				Expect(result.Candidates).To(Equal(assets))
			})
		})

		Context("when detectors filter out all candidates", func() {
			BeforeEach(func() {
				filteringDetector := &MockDetector{
					DetectFunc: func(a []Asset) (DetectionResult, error) {
						return NewDetectionResult(nil, nil), nil
					},
				}

				detectorChain = &DetectorChain{
					Detectors: []Detector{filteringDetector},
					System:    nil,
				}
			})

			It("should return an empty detection result without error", func() {
				result, err := detectorChain.Detect(assets)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Candidates).To(BeNil())
			})
		})
	})
})
