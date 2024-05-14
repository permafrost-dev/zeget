package filters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/permafrost-dev/eget/lib/assets"
	. "github.com/permafrost-dev/eget/lib/filters"
)

var _ = Describe("Filters", func() {
	Context("NewFilter", func() {
		It("should create a new filter with the given parameters", func() {
			handler := func(a assets.Asset, args []string) bool { return true }
			filter := NewFilter("test", handler, FilterActionInclude, "arg1", "arg2")

			Expect(filter.Name).To(Equal("test"))
			Expect(filter.Action).To(Equal(FilterActionInclude))
			Expect(filter.Args).To(Equal([]string{"arg1", "arg2"}))
			Expect(filter.Definition).To(Equal("test(arg1,arg2)"))
		})
	})

	Context("Filter Apply", func() {
		It("should apply the handler to the asset", func() {
			handler := func(a assets.Asset, args []string) bool { return a.Name == "test" }
			filter := NewFilter("test", handler, FilterActionInclude)
			asset := assets.Asset{Name: "test"}

			Expect(filter.Apply(asset)).To(BeTrue())
		})
	})

	Context("Filter WithArgs", func() {
		It("should set the arguments and return the filter", func() {
			handler := func(a assets.Asset, args []string) bool { return true }
			filter := NewFilter("test", handler, FilterActionInclude)
			filter.WithArgs("newArg1", "newArg2")

			Expect(filter.Args).To(Equal([]string{"newArg1", "newArg2"}))
		})
	})

	Context("Parser", func() {
		parser := NewParser()

		It("should parse multiple definitions", func() {
			definitions := "all(file1.txt);none(file2.exe)"
			filters := parser.ParseDefinitions(definitions)

			Expect(filters).To(HaveLen(2))
			Expect(filters[0].Name).To(Equal("all"))
			Expect(filters[1].Name).To(Equal("none"))
		})

		It("should parse a single definition", func() {
			definition := "all(file1.txt)"
			filter := parser.ParseDefinition(definition)

			Expect(filter).NotTo(BeNil())
			Expect(filter.Name).To(Equal("all"))
			Expect(filter.Args).To(Equal([]string{"file1.txt"}))
		})

		It("should return nil for invalid definitions", func() {
			definition := "invalid()"
			filter := parser.ParseDefinition(definition)

			Expect(filter).To(BeNil())
		})
	})
})
