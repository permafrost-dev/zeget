package app_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/permafrost-dev/eget/app"
)

var _ = Describe("Config", func() {
	var (
		tempDir      string
		err          error
		configPath   string
		configSample = `
[global]
all = true
download_only = true
file = "test.txt"
github_token = "testtoken"
quiet = true
show_hash = true
download_source = true
system = "linux"
target = "/tmp"
upgrade_only = true

[repositories]
  [repositories.repo1]
  all = false
  asset_filters = ["*.zip", "*.tar.gz"]
  download_only = false
  file = "repo1.txt"
  name = "First Repo"
  quiet = false
  show_hash = false
  download_source = false
  system = "darwin"
  tag = "v1.0.0"
  target = "/var"
  upgrade_only = false
  verify_sha256 = "abc123"
  disable_ssl = true
`
	)

	BeforeEach(func() {
		tempDir, err = ioutil.TempDir("", "configTest")
		Expect(err).NotTo(HaveOccurred())

		configPath = filepath.Join(tempDir, "config.toml")
		err = ioutil.WriteFile(configPath, []byte(configSample), 0644)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("Loading configuration file", func() {
		Context("When file is correctly formatted", func() {
			It("Should load global and repository configurations successfully", func() {
				config, err := LoadConfigurationFile(configPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(config).NotTo(BeNil())
				Expect(config.Global.All).To(Equal(true))
				Expect(config.Global.DownloadOnly).To(Equal(true))
				Expect(config.Repositories).To(HaveKey("repo1"))
				repo1 := config.Repositories["repo1"]
				Expect(repo1.All).To(Equal(false))
				Expect(repo1.AssetFilters).To(ConsistOf([]string{"*.zip", "*.tar.gz"}))
				Expect(repo1.DisableSSL).To(Equal(true))
			})
		})

		Context("When file does not exist", func() {
			It("Should return an error", func() {
				_, err := LoadConfigurationFile("nonexistent.toml")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Building configuration filename", func() {
		Context("With specified paths", func() {
			It("Should correctly build the configuration filename", func() {
				filename := BuildConfigurationFilename(tempDir)
				Expect(filename).To(Equal(filepath.Join(tempDir, "."+ApplicationName+".toml")))
			})
		})

		Context("Without specified paths", func() {
			It("Should return only a filename", func() {
				filename := BuildConfigurationFilename()
				Expect(filename).To(Equal("." + ApplicationName + ".toml"))
			})
		})
	})

	Describe("Getting OS specific configuration path", func() {
		It("Should return the correct path for the OS", func() {
			expectedPath := GetOSConfigPath(tempDir)
			Expect(expectedPath).To(ContainSubstring(tempDir))
		})
	})
})
