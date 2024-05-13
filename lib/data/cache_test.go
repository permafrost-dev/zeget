package data_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/permafrost-dev/eget/lib/assets"
	. "github.com/permafrost-dev/eget/lib/data"
	"github.com/permafrost-dev/eget/lib/finders"
)

var _ = Describe("Cache", func() {
	var (
		cache     *Cache
		filename  string
		testEntry *RepositoryCacheEntry
	)

	BeforeEach(func() {
		filename = "test_cache.json"
		cache = NewCache(filename)
		testEntry = &RepositoryCacheEntry{
			Name:      "test",
			Target:    "target",
			Filters:   []string{"*.zip"},
			Assets:    []assets.Asset{},
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}
	})

	AfterEach(func() {
		os.Remove(filename)
	})

	Describe("NewCache", func() {
		It("should create a new cache with specified filename", func() {
			Expect(cache.Filename).To(Equal(filename))
		})
	})

	Describe("Set and Get", func() {
		It("should add and retrieve an entry from the cache", func() {
			cache.Set("test", testEntry)
			retrievedEntry, exists := cache.Get("test")
			Expect(exists).To(BeTrue())
			Expect(retrievedEntry.Name).To(Equal("test"))
		})
	})

	Describe("Has", func() {
		It("should correctly report if an entry exists based on name", func() {
			cache.Set("test", testEntry)
			Expect(cache.Has("test")).To(BeTrue())
			Expect(cache.Has("nonexistent")).To(BeFalse())
		})
	})

	Describe("SetRateLimit and SaveToFile", func() {
		It("should set the rate limit and save to file", func() {
			cache.SetRateLimit(5000, 4999, time.Now().Add(time.Hour))
			Expect(cache.Data.RateLimit.Limit).To(Equal(5000))
			Expect(cache.Data.RateLimit.Remaining).To(Equal(4999))
			_, err := os.Stat(filename)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("AddRepository", func() {
		It("should add a repository and save to file", func() {
			_, added := cache.AddRepository("owner/testrepo", "target", []string{"zip"}, &finders.FindResult{}, time.Now().Add(10*time.Minute))
			Expect(added).To(BeTrue())
			Expect(cache.Data.HasRepositoryEntryByKey("owner/testrepo")).To(BeTrue())
		})
	})

	Describe("LoadFromFile", func() {
		It("should load cache data from file", func() {
			cache.Set("test", testEntry)
			cache.SaveToFile()

			newCache := NewCache(filename)
			err := newCache.LoadFromFile()
			Expect(err).ToNot(HaveOccurred())
			_, exists := newCache.Get("test")
			Expect(exists).To(BeTrue())
		})
	})

	Describe("PurgeExpired", func() {
		It("should remove expired entries", func() {
			expiredEntry := &RepositoryCacheEntry{
				Name:      "expired",
				ExpiresAt: time.Now().Add(-1 * time.Minute),
			}
			cache.Set("expired", expiredEntry)
			cache.PurgeExpired()
			_, exists := cache.Get("expired")
			Expect(exists).To(BeFalse())
		})
	})
})
