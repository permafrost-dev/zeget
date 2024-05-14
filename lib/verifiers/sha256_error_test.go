package verifiers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/permafrost-dev/eget/lib/verifiers"
)

var _ = Describe("Sha256Error", func() {
	It("should return the correct error message", func() {
		expected := []byte{1, 2, 3, 4}
		got := []byte{4, 3, 2, 1}
		err := &verifiers.Sha256Error{
			Expected: expected,
			Got:      got,
		}

		Expect(err.Error()).To(Equal("sha256 checksum mismatch:\nexpected: 01020304\ngot:      04030201"))
	})
})
