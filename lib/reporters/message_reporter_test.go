package reporters_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/eget/lib/reporters"
)

var _ = Describe("MessageReporter", func() {
	var (
		outputBuffer *bytes.Buffer
		reporter     *reporters.MessageReporter
	)

	BeforeEach(func() {
		outputBuffer = new(bytes.Buffer)
	})

	Describe("Report method", func() {
		Context("with no format string and no arguments", func() {
			It("should not write anything to the output", func() {
				reporter = reporters.NewMessageReporter(outputBuffer, "", nil)
				Expect(reporter.Report()).To(Succeed())
				Expect(outputBuffer.String()).To(BeEmpty())
			})
		})

		Context("with format string but no additional arguments", func() {
			It("should write formatted string to the output", func() {
				formatString := "Test message"
				reporter = reporters.NewMessageReporter(outputBuffer, formatString)
				Expect(reporter.Report()).To(Succeed())
				Expect(outputBuffer.String()).To(ContainSubstring(formatString))
			})
		})

		Context("with format string and additional arguments", func() {
			It("should write formatted string with arguments to the output", func() {
				formatString := "Test %s"
				arg := "message"
				reporter = reporters.NewMessageReporter(outputBuffer, formatString, arg)
				Expect(reporter.Report()).To(Succeed())
				Expect(outputBuffer.String()).To(ContainSubstring("Test message"))
			})
		})

		Context("with additional arguments but no format string", func() {
			It("should write additional arguments to the output", func() {
				reporter = reporters.NewMessageReporter(outputBuffer, "")
				arg := "message"
				Expect(reporter.Report(arg)).To(Succeed())
				Expect(outputBuffer.String()).To(ContainSubstring(arg))
			})
		})
	})
})
