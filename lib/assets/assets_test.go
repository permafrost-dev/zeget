package assets_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/permafrost-dev/eget/lib/assets"
)

var _ = Describe("lib/assets > Asset", func() {
	It("should return an AssetWrapper with the correct values", func() {
		aw := NewAssetWrapper([]Asset{Asset{Name: "one"}, Asset{Name: "two"}})
		Expect(len(aw.Assets)).To(Equal(2))
		Expect(aw.Assets[0].Name).To(Equal("one"))
		Expect(aw.Assets[1].Name).To(Equal("two"))
	})
})
