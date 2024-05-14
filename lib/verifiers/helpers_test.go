package verifiers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/permafrost-dev/eget/lib/verifiers"
)

var _ = Describe("Helpers", func() {
	Describe("DetermineHashTypeByLength function", func() {
		It("returns MD5 for 16 byte hashes", func() {
			Expect(verifiers.DetermineHashTypeByLength("1234567890123456")).To(Equal(verifiers.MD5))
		})

		It("returns SHA1 for 20 byte hashes", func() {
			Expect(verifiers.DetermineHashTypeByLength("12345678901234567890")).To(Equal(verifiers.SHA1))
		})

		It("returns SHA256 for 32 byte hashes", func() {
			Expect(verifiers.DetermineHashTypeByLength("12345678901234567890123456789012")).To(Equal(verifiers.SHA256))
		})

		It("returns SHA512 for 64 byte hashes", func() {
			Expect(verifiers.DetermineHashTypeByLength("1234567890123456789012345678901212345678901234567890123456789012")).To(Equal(verifiers.SHA512))
		})

		It("returns Unknown for other lengths", func() {
			Expect(verifiers.DetermineHashTypeByLength("123")).To(Equal(verifiers.Unknown))
		})
	})
})
