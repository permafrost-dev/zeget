package utilities_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/permafrost-dev/eget/lib/utilities"
)

var _ = Describe("Filters", func() {
	Describe("FilenameToAssetFilters", func() {
		It("should return guessed filters for filenames without known patterns", func() {
			Expect(FilenameToAssetFilters("random-file")).To(Equal([]string{"file"}))
		})

		It("should return correct filters for a filename with known patterns", func() {
			filename := "test-amd64.exe"
			expectedFilters := []string{"amd64", "exe"}
			Expect(FilenameToAssetFilters(filename)).To(Equal(expectedFilters))
		})

		It("should handle multiple occurrences of the same architecture", func() {
			filename := "test-amd64-amd64"
			expectedFilters := []string{"amd64", "amd64"}
			Expect(FilenameToAssetFilters(filename)).To(ConsistOf(expectedFilters))
		})

		It("should include unknowns", func() {
			filename := "test-linux-unknown-amd64.exe"
			expectedFilters := []string{"linux", "unknown", "amd64", "exe"}
			Expect(FilenameToAssetFilters(filename)).To(Equal(expectedFilters))
		})
	})
})
