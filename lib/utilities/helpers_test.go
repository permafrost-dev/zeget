package utilities_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/zeget/lib/assets"
	. "github.com/permafrost-dev/zeget/lib/utilities"
)

var _ = Describe("Helpers", func() {
	var (
		tempDir string
	)

	BeforeEach(func() {
		t := GinkgoT()
		tempDir = t.TempDir()
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("Cut", func() {
		It("cuts a string before and after a separator", func() {
			before, after, found := Cut("test string", " ")
			Expect(found).To(BeTrue())
			Expect(before).To(Equal("test"))
			Expect(after).To(Equal("string"))
		})

		It("returns false if the separator is not found", func() {
			_, _, found := Cut("test string", "x")
			Expect(found).To(BeFalse())
		})
	})

	Describe("Bintime", func() {
		It("returns the modification time of a file", func() {
			filePath := filepath.Join(tempDir, "testfile")
			_, err := os.Create(filePath)
			Expect(err).NotTo(HaveOccurred())
			// Ensure the file has a modification time by writing to it after creation
			time.Sleep(200 * time.Millisecond)
			err = os.WriteFile(filePath, []byte("test data"), 0644)
			Expect(err).NotTo(HaveOccurred())

			modTime := Bintime("testfile", tempDir)
			Expect(modTime).To(BeTemporally("~", time.Now(), 2*time.Second))
		})

		It("uses the EGET_BIN environment variable if set", func() {
			filePath := filepath.Join(tempDir, "testfile")
			defer os.Remove(filePath)

			err := os.WriteFile(filePath, []byte("test data"), 0644)
			Expect(err).NotTo(HaveOccurred())

			os.Setenv("EGET_BIN", tempDir)
			time.Sleep(100 * time.Millisecond)

			modTime := Bintime(filePath, "")
			Expect(modTime.String()).To(Equal("0001-01-01 00:00:00 +0000 UTC"))
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

	Describe("Github URLs", func() {
		It("identifies GitHub URLs correctly", func() {
			Expect(IsGithubURL("https://github.com/user/repo")).To(BeTrue())
			Expect(IsGithubURL("https://github.com/user/repo.git")).To(BeTrue())
			Expect(IsGithubURL("https://notgithub.com/user/repo")).To(BeFalse())
		})

		It("identifies invalid GitHub URLs correctly", func() {
			Expect(IsInvalidGithubURL("https://github.com/user")).To(BeTrue())
			Expect(IsInvalidGithubURL("https://github.com/user/repo")).To(BeFalse())
		})

		It("identifies non-GitHub URLs correctly", func() {
			Expect(IsNonGithubURL("https://example.test")).To(BeTrue())
			Expect(IsNonGithubURL("https://github.com/a/b")).To(BeFalse())
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
			ref, _ := ParseRepositoryReference("user/repo")
			Expect(ref).NotTo(BeNil())
			Expect(ref.Owner).To(Equal("user"))
			Expect(ref.Name).To(Equal("repo"))
		})

		It("returns nil for invalid references", func() {
			ref, _ := ParseRepositoryReference("user")
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

		It("returns false for non-existent paths", func() {
			Expect(IsDirectory(filepath.Join(tempDir, "nonexistent123"))).To(BeFalse())
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

	It("determines if a file is definitely not an executable", func() {
		var filePath string

		filePath = filepath.Join(tempDir, "nonexecutable")
		Expect(IsDefinitelyNotExec(filePath)).To(BeFalse())

		filePath = filepath.Join(tempDir, "nonexecutable.deb")
		Expect(IsDefinitelyNotExec(filePath)).To(BeTrue())

		filePath = filepath.Join(tempDir, "nonexecutable.1")
		Expect(IsDefinitelyNotExec(filePath)).To(BeTrue())

		filePath = filepath.Join(tempDir, "nonexecutable.txt")
		Expect(IsDefinitelyNotExec(filePath)).To(BeTrue())
	})

	It("determines if a file is executable", func() {
		filePath := filepath.Join(tempDir, "executable")
		err := os.WriteFile(filePath, []byte("#!/bin/bash\n"), 0755)
		Expect(err).NotTo(HaveOccurred())
		defer os.Remove(filePath)

		fi, err := os.Stat(filePath)
		Expect(err).NotTo(HaveOccurred())

		//check executable bit
		Expect(IsExec(filePath, fi.Mode())).To(BeTrue())

		//check file extensions
		Expect(IsExec(filePath+".exe", 0o644)).To(BeTrue())
		Expect(IsExec(filePath+".appimage", 0o644)).To(BeTrue())
		Expect(IsExec(filePath+".txt", 0o644)).To(BeFalse())
	})

	It("finds a checksum asset", func() {
		assetList := []assets.Asset{
			{Name: "file.sha256sum"},
			{Name: "file.sha256"},
			{Name: "file"},
		}
		Expect(FindChecksumAsset(assets.Asset{Name: "file"}, assetList)).To(Equal(assets.Asset{Name: "file.sha256sum"}))

		assetList = []assets.Asset{
			{Name: "file"},
			{Name: "file.sha256"},
			{Name: "file.sha256sum"},
		}
		Expect(FindChecksumAsset(assets.Asset{Name: "file"}, assetList)).To(Equal(assets.Asset{Name: "file.sha256"}))
	})

	It("returns an empty asset if no checksum asset is found", func() {
		assetList := []assets.Asset{
			{Name: "file"},
			{Name: "file.sha256sum"},
		}

		Expect(FindChecksumAsset(assets.Asset{Name: "abc"}, assetList)).To(Equal(assets.Asset{}))
	})

	It("Gets mode from file name", func() {
		Expect(ModeFrom("file.deb", 0o644)).To(Equal(fs.FileMode(0o644)))
		Expect(ModeFrom("file.1", 0o644)).To(Equal(fs.FileMode(0o644)))
		Expect(ModeFrom("file.txt", 0o644)).To(Equal(fs.FileMode(0o644)))
		Expect(ModeFrom("file", 0o644)).To(Equal(fs.FileMode(0o755)))
		Expect(ModeFrom("file.exe", 0o644)).To(Equal(fs.FileMode(0o755)))
	})

	It("returns the filename to rename to", func() {
		Expect(GetRename("file", "newname")).To(Equal("newname"))
		Expect(GetRename("file.appimage", "newname")).To(Equal("file"))
		Expect(GetRename("file.exe", "newname")).To(Equal("file.exe"))
	})

	Describe("SetIf", func() {
		It("returns the second value if condition is true", func() {
			Expect(SetIf(true, "11", "22")).To(Equal("22"))
		})

		It("returns the first value if condition is false", func() {
			Expect(SetIf(false, "11", "22")).To(Equal("11"))
		})
	})

	Describe("ExtractToolNameFromURL", func() {
		It("extracts the tool name from a URL", func() {
			Expect(ExtractToolNameFromURL("https://github.com/user1/toolname")).To(Equal("toolname"))
			Expect(ExtractToolNameFromURL("https://github.com/abc/toolnamea/")).To(Equal("toolnamea"))
			Expect(ExtractToolNameFromURL("https://github.com/abc/toolname2/")).To(Equal("toolname2"))
			Expect(ExtractToolNameFromURL("https://github.com/")).To(Equal("Unknown"))
		})
	})

	It("checks if an error is of a specific type", func() {
		var err error
		err = NewInvalidGitHubProjectURLError("test")

		Expect(IsErrorOf(InvalidGitHubProjectURL, err)).To(BeTrue())
	})

	It("gets the current working directory", func() {
		os.Chdir(tempDir)
		wd := GetCurrentDirectory()
		Expect(wd).To(Equal(tempDir))
	})

})
