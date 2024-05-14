package data

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	ulid "github.com/oklog/ulid/v2"
	"github.com/permafrost-dev/zeget/lib/finders"
	"github.com/permafrost-dev/zeget/lib/utilities"
)

type Cache struct {
	Filename string
	Data     ApplicationData
	Debug    bool
	mutex    sync.Mutex
}

func NewCache(filename string) *Cache {
	repoEntries := make(map[string]*RepositoryCacheEntry)
	rateLimit := RateLimit{}
	return &Cache{
		Filename: filename,
		Data:     ApplicationData{RateLimit: rateLimit, Repositories: repoEntries},
		Debug:    false,
	}
}

func NewCacheID() string {
	t := time.Now().UnixNano()
	entropy := ulid.Monotonic(rand.Reader, uint64(t))
	return ulid.MustNew(ulid.Now(), entropy).String()
}

// Set adds a new entry to the cache with a TTL.
func (c *Cache) Set(key string, value *RepositoryCacheEntry) *RepositoryCacheEntry {
	c.Data.SetRepositoryEntryByKey(key, value, c)

	go c.expire(key, value.ExpiresAt.Sub(time.Now()), 0)

	return value
}

// Get retrieves an entry from the cache.
func (c *Cache) Get(key string) (*RepositoryCacheEntry, bool) {
	entry, exists := c.Data.Repositories[key]

	if !exists || time.Now().After(entry.ExpiresAt) {
		return &RepositoryCacheEntry{}, false
	}

	entry.owner = c

	return entry, true
}

func (c *Cache) Has(name string) bool {
	for _, repo := range c.Data.Repositories {
		if strings.EqualFold(repo.Name, name) {
			return true
		}
	}

	return false
}

func (c *Cache) SetRateLimit(limit int, remaining int, reset time.Time) {
	c.mutex.Lock()
	c.Data.RateLimit = RateLimit{
		Service:   "github",
		Limit:     limit,
		Remaining: remaining,
		Reset:     &reset,
	}
	c.mutex.Unlock()

	c.SaveToFile()
}

func (c *Cache) AddRepository(name, target string, filters []string, findResult *finders.FindResult, expiresAt time.Time) (*RepositoryCacheEntry, bool) {
	if !utilities.IsValidRepositoryReference(name) {
		return &RepositoryCacheEntry{}, false
	}

	entry := RepositoryCacheEntry{
		ID:          NewCacheID(),
		Name:        name,
		Target:      target,
		Filters:     filters,
		Assets:      findResult.Assets,
		FindError:   findResult.Error,
		ExpiresAt:   expiresAt,
		LastCheckAt: time.Now(),
	}

	if len(findResult.Assets) > 0 {
		entry.LastReleaseDate = findResult.Assets[0].ReleaseDate
	}

	c.Set(entry.Name, &entry)
	c.SaveToFile()

	if c.Debug {
		fmt.Printf("Added repository %s, %v\n", name, entry)
	}

	return &entry, true
}

// expire removes an entry from the cache after a duration.
func (c *Cache) expire(key string, duration time.Duration, counter int) {
	time.Sleep(duration)

	locked := c.mutex.TryLock()

	if !locked && counter > 10 {
		return // give up after 10 tries
	}
	if !locked {
		go c.expire(key, duration, counter+1)
		return
	}
	delete(c.Data.Repositories, key)
	c.mutex.Unlock()

	c.SaveToFile()
}

func (c *Cache) PurgeExpired() {
	c.mutex.Lock()
	for key := range c.Data.Repositories {
		entry := c.Data.Repositories[key]
		if time.Now().After(entry.ExpiresAt) {
			delete(c.Data.Repositories, key)
		}
	}

	// check for expired rate limit
	if c.Data.RateLimit.Reset != nil && time.Now().After(*c.Data.RateLimit.Reset) {
		c.Data.RateLimit.Remaining = c.Data.RateLimit.Limit
		c.Data.RateLimit.Reset = nil
	}

	c.mutex.Unlock()

	c.SaveToFile()
}

// SaveToFile saves the entire cache to a JSON file.
func (c *Cache) SaveToFile() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	file, err := json.MarshalIndent(c.Data, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.Filename, file, 0644)
}

func (c *Cache) LoadFromFile() error {
	if !utilities.IsLocalFile(c.Filename) {
		c.SaveToFile()
	}

	file, err := os.ReadFile(c.Filename)
	if err != nil {
		return err
	}

	result := json.Unmarshal(file, &c.Data)

	c.PurgeExpired() // remove any expired entries after the file

	return result
}
