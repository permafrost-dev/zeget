package home_test

import (
	"os/user"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/permafrost-dev/eget/lib/home"
)

var _ = Describe("lib/home > PathExpander", func() {
	var (
		pathExpander *PathExpander
	)

	BeforeEach(func() {
		pathExpander = NewPathExpander()
	})

	Describe("NewPathExpander", func() {
		Context("when no arguments are provided", func() {
			It("should return a PathExpander with an empty homePath", func() {
				Expect(pathExpander.GetHomePath()).To(BeEmpty())
			})

			It("should create a PathExpander with the provided homePath", func() {
				pathExpander = NewPathExpander("/test/path")
				Expect(pathExpander.GetHomePath()).To(Equal("/test/path"))
			})
		})
	})

	Describe("SetHomePath", func() {
		It("should set the home path correctly", func() {
			testPath := "/test/path"
			pathExpander.SetHomePath(testPath)
			Expect(pathExpander.GetHomePath()).To(Equal(testPath))
		})
	})

	Describe("HomeDirectory", func() {
		Context("when homePath is set", func() {
			It("returns the set homePath without error", func() {
				testPath := "/custom/home"
				pathExpander.SetHomePath(testPath)
				homeDir, err := pathExpander.HomeDirectory()
				Expect(err).NotTo(HaveOccurred())
				Expect(homeDir).To(Equal(testPath))
			})
		})

		Context("when homePath is not set", func() {
			It("returns the current user's home directory", func() {
				currentUser, err := user.Current()
				Expect(err).NotTo(HaveOccurred())
				Expect(currentUser.HomeDir).NotTo(BeEmpty())

				homeDir, err := pathExpander.HomeDirectory()
				Expect(err).NotTo(HaveOccurred())
				Expect(homeDir).To(Equal(currentUser.HomeDir))
			})
		})
	})

	Describe("Expand", func() {
		Context("when path does not start with ~", func() {
			It("returns the same path without error", func() {
				testPath := "/some/path"
				expandedPath, err := pathExpander.Expand(testPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(expandedPath).To(Equal(testPath))
			})
		})

		Context("when path starts with ~ and homePath is set", func() {
			It("replaces ~ with the home directory", func() {
				t := GinkgoT()
				homeDir := t.TempDir()
				// Expect(err).NotTo(HaveOccurred())

				testPath := "~/some/path"
				expectedPath := homeDir + "/some/path"

				pathExpander.SetHomePath(homeDir)
				expandedPath, err := pathExpander.Expand(testPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(expandedPath).To(Equal(expectedPath))
			})
		})

		Context("when path starts with ~ and homePath is unset", func() {
			It("replaces ~ with the home directory", func() {
				currentUser, err := user.Current()
				homeDir := currentUser.HomeDir
				// Expect(err).NotTo(HaveOccurred())

				testPath := "~/some/path"
				expectedPath := homeDir + "/some/path"

				pathExpander.SetHomePath("")
				expandedPath, err := pathExpander.Expand(testPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(expandedPath).To(Equal(expectedPath))
			})
		})
	})
})
