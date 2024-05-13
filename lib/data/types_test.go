package data_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/eget/lib/data"
)

var _ = Describe("ApplicationData", func() {
	var (
		appData data.ApplicationData
		owner   *data.Cache // Assuming Cache is a struct in your package with a mutex
	)

	BeforeEach(func() {
		appData = data.ApplicationData{
			RateLimit:    data.RateLimit{},
			Repositories: make(map[string]*data.RepositoryCacheEntry),
		}
		owner = &data.Cache{} // Setup owner assuming a valid constructor or initial state
	})

	Describe("HasRepositoryEntryByKey", func() {
		It("should return false if the key does not exist", func() {
			Expect(appData.HasRepositoryEntryByKey("nonexistent")).To(BeFalse())
		})

		It("should return true if the key exists", func() {
			appData.Repositories["exists"] = &data.RepositoryCacheEntry{}
			Expect(appData.HasRepositoryEntryByKey("exists")).To(BeTrue())
		})
	})

	Describe("GetRepositoryEntryByKey", func() {
		It("should return a new entry if the key does not exist", func() {
			entry := appData.GetRepositoryEntryByKey("nonexistent", owner)
			Expect(entry).NotTo(BeNil())
			Expect(entry.GetOwner().Filename).To(Equal(owner.Filename))
		})

		It("should return the existing entry if the key exists", func() {
			expectedEntry := &data.RepositoryCacheEntry{}
			appData.Repositories["exists"] = expectedEntry
			entry := appData.GetRepositoryEntryByKey("exists", owner)
			Expect(entry).To(Equal(expectedEntry))
			Expect(entry.GetOwner().Filename).To(Equal(owner.Filename))
		})
	})

	Describe("SetRepositoryEntryByKey", func() {
		It("should set the repository entry for a given key", func() {
			entry := &data.RepositoryCacheEntry{}
			appData.SetRepositoryEntryByKey("newEntry", entry, owner)
			Expect(appData.Repositories).To(HaveKeyWithValue("newEntry", entry))
			Expect(entry.GetOwner().Filename).To(Equal(owner.Filename))
		})
	})
})

var _ = Describe("RepositoryCacheEntry", func() {
	var (
		entry *data.RepositoryCacheEntry
	)

	BeforeEach(func() {
		entry = &data.RepositoryCacheEntry{}
	})

	Describe("UpdateCheckedAt", func() {
		It("should update LastCheckAt to the current time", func() {
			beforeUpdate := time.Now().Add(-time.Minute)
			entry.UpdateCheckedAt()
			Expect(entry.LastCheckAt).To(BeTemporally(">", beforeUpdate))
		})
	})

	Describe("UpdateDownloadedAt", func() {
		It("should update LastDownloadAt and LastDownloadTag correctly", func() {
			tag := "v1.0.0"
			beforeUpdate := time.Now().Local().Add(-time.Minute)
			entry.UpdateDownloadedAt(tag)
			Expect(entry.LastDownloadAt).To(BeTemporally(">", beforeUpdate))
			Expect(entry.LastDownloadTag).To(Equal(tag))
		})
	})

	Describe("Exists", func() {
		It("should return false by default", func() {
			Expect(entry.Exists()).To(BeFalse())
		})
	})
})
