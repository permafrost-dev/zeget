package utilities_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/permafrost-dev/eget/lib/utilities"
)

var _ = Describe("Helpers", func() {
	var (
		tempDir string
		err     error
	)

	BeforeEach(func() {
		tempDir, err = os.MkdirTemp("", "helpers")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("Bintime", func() {
		It("returns the modification time of a file", func() {
			filePath := filepath.Join(tempDir, "testfile")
			_, err := os.Create(filePath)
			Expect(err).NotTo(HaveOccurred())
			// Ensure the file has a modification time by writing to it after creation
			time.Sleep(1 * time.Second)
			err = os.WriteFile(filePath, []byte("test data"), 0644)
			Expect(err).NotTo(HaveOccurred())

			modTime := Bintime("testfile", tempDir)
			Expect(modTime).To(BeTemporally("~", time.Now(), 2*time.Second))
		})
	})

	Describe("IsURL", func() {
		It("returns true for valid URLs", func() {
			Expect(IsURL("http://example.com")).To(BeTrue())
			Expect(IsURL("https://example.com")).To(BeTrue())
		})

		It("returns false for invalid URLs", func() {
			Expect(IsURL("not a url")).To(BeFalse())
			Expect(IsURL("ftp://example.com")).To(BeTrue()) // FTP is a valid URL scheme
		})
	})

	Describe("IsGithubURL", func() {
		It("identifies GitHub URLs correctly", func() {
			Expect(IsGithubURL("https://github.com/user/repo")).To(BeTrue())
			Expect(IsGithubURL("https://github.com/user/repo.git")).To(BeTrue())
			Expect(IsGithubURL("https://notgithub.com/user/repo")).To(BeFalse())
		})
	})

	Describe("IsInvalidGithubURL", func() {
		It("identifies invalid GitHub URLs correctly", func() {
			Expect(IsInvalidGithubURL("https://github.com/user")).To(BeTrue())
			Expect(IsInvalidGithubURL("https://github.com/user/repo")).To(BeFalse())
		})
	})

	Describe("RepositoryNameFromGithubURL", func() {
		It("extracts repository names from GitHub URLs", func() {
			name, found := RepositoryNameFromGithubURL("https://github.com/user/repo")
			Expect(found).To(BeTrue())
			Expect(name).To(Equal("user/repo"))
		})

		It("returns false if the URL is not a GitHub URL", func() {
			_, found := RepositoryNameFromGithubURL("https://notgithub.com/user/repo")
			Expect(found).To(BeFalse())
		})
	})

	Describe("IsValidRepositoryReference", func() {
		It("validates repository references correctly", func() {
			Expect(IsValidRepositoryReference("user/repo")).To(BeTrue())
			Expect(IsValidRepositoryReference("user")).To(BeFalse())
		})
	})

	Describe("ParseRepositoryReference", func() {
		It("parses valid repository references", func() {
			ref := ParseRepositoryReference("user/repo")
			Expect(ref).NotTo(BeNil())
			Expect(ref.Owner).To(Equal("user"))
			Expect(ref.Name).To(Equal("repo"))
		})

		It("returns nil for invalid references", func() {
			ref := ParseRepositoryReference("user")
			Expect(ref).To(BeNil())
		})
	})

	Describe("IsLocalFile", func() {
		It("checks if a file exists locally", func() {
			filePath := filepath.Join(tempDir, "existent")
			_, err := os.Create(filePath)
			Expect(err).NotTo(HaveOccurred())

			Expect(IsLocalFile(filePath)).To(BeTrue())
			Expect(IsLocalFile(filepath.Join(tempDir, "nonexistent"))).To(BeFalse())
		})
	})

	Describe("IsDirectory", func() {
		It("verifies if a path is a directory", func() {
			Expect(IsDirectory(tempDir)).To(BeTrue())
			filePath := filepath.Join(tempDir, "file")
			_, err := os.Create(filePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(IsDirectory(filePath)).To(BeFalse())
		})
	})

	Describe("IsExec", func() {
		It("determines if a file is an executable", func() {
			filePath := filepath.Join(tempDir, "executable")
			err := os.WriteFile(filePath, []byte("#!/bin/bash\n"), 0755)
			Expect(err).NotTo(HaveOccurred())

			fi, err := os.Stat(filePath)
			Expect(err).NotTo(HaveOccurred())

			Expect(IsExec(filePath, fi.Mode())).To(BeTrue())
		})
	})
})
