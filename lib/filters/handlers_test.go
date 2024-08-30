package filters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/permafrost-dev/zeget/lib/assets"
	. "github.com/permafrost-dev/zeget/lib/filters"
)

var anyHandler = FilterMap["any"].Handler
var allHandler = FilterMap["all"].Handler
var hasHandler = FilterMap["has"].Handler
var noneHandler = FilterMap["none"].Handler
var extensionHandler = FilterMap["ext"].Handler

var _ = Describe("Filters", func() {
	Context("Handlers", func() {
		asset1 := assets.Asset{Name: "file1.txt"}
		asset2 := assets.Asset{Name: "file2.exe"}

		It("anyHandler should return true if any argument matches the asset name", func() {
			Expect(anyHandler(asset1, []string{"file1.txt", "file2.exe"})).To(BeTrue())
			Expect(anyHandler(asset2, []string{"file1.txt", "file2.exe"})).To(BeTrue())
			Expect(anyHandler(asset1, []string{"file2.exe"})).To(BeFalse())
		})

		It("allHandler should return true if all arguments match the asset name", func() {
			Expect(allHandler(asset1, []string{"file1.txt", "file1.txt"})).To(BeTrue())
			Expect(allHandler(asset2, []string{"file2.exe", "file2.exe"})).To(BeTrue())
			Expect(allHandler(asset1, []string{"file1.txt", "file2.exe"})).To(BeFalse())
		})

		It("hasHandler should return true if any argument matches the asset name", func() {
			Expect(hasHandler(asset1, []string{"file1.txt"})).To(BeTrue())
			Expect(hasHandler(asset2, []string{"file2.exe"})).To(BeTrue())
			Expect(hasHandler(asset1, []string{"file2.exe"})).To(BeFalse())
		})

		It("noneHandler should return true if no arguments match the asset name", func() {
			Expect(noneHandler(asset1, []string{"file2.exe"})).To(BeTrue())
			Expect(noneHandler(asset2, []string{"file1.txt"})).To(BeTrue())
			Expect(noneHandler(asset1, []string{"file1.txt"})).To(BeFalse())
		})

		It("extensionHandler should return true if the extension matches the asset name", func() {
			Expect(extensionHandler(asset1, []string{".txt"})).To(BeTrue())
			Expect(extensionHandler(asset2, []string{".exe"})).To(BeTrue())
			Expect(extensionHandler(asset1, []string{".exe"})).To(BeFalse())
		})
	})

})
