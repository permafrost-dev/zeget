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
)

type Cache struct {
	Filename string
	Data     ApplicationData
	mutex    sync.Mutex
}

func NewCache(filename string) *Cache {
	repoEntries := make(map[string]RepositoryCacheEntry)
	rateLimit := RateLimit{}
	return &Cache{
		Filename: filename,
		Data:     ApplicationData{RateLimit: rateLimit, Repositories: repoEntries},
	}
}

func NewCacheID() string {
	t := time.Now().UnixNano()
	entropy := ulid.Monotonic(rand.Reader, uint64(t))
	return ulid.MustNew(ulid.Now(), entropy).String()
}

// Set adds a new entry to the cache with a TTL.
func (c *Cache) Set(key string, value *RepositoryCacheEntry) *RepositoryCacheEntry {
	c.mutex.Lock()
	value.owner = c
	c.Data.Repositories[key] = *value
	c.mutex.Unlock()

	go c.expire(key, value.ExpiresAt.Sub(time.Now()))

	result, _ := c.Data.Repositories[key]

	return &result
}

// Get retrieves an entry from the cache.
func (c *Cache) Get(key string) (*RepositoryCacheEntry, bool) {
	c.mutex.Lock()
	entry, exists := c.Data.Repositories[key]
	c.mutex.Unlock()

	if !exists || time.Now().After(entry.ExpiresAt) {
		return &RepositoryCacheEntry{}, false
	}

	entry.owner = c

	return &entry, true
}

func (c *Cache) Has(name string) bool {
	// c.mutex.Lock()
	// defer c.mutex.Unlock()

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

func (c *Cache) AddRepository(name, target string, filters []string, expiresAt time.Time) (*RepositoryCacheEntry, bool) {
	// if c.Has(name) {
	// 	return c.Get(name)
	// }
	entry := RepositoryCacheEntry{
		ID:          NewCacheID(),
		Name:        name,
		Target:      target,
		Filters:     filters,
		ExpiresAt:   expiresAt,
		LastCheckAt: time.Now(),
	}

	c.Set(entry.Name, &entry)
	c.SaveToFile()

	fmt.Printf("Added repository %s, %v\n", name, entry)

	return &entry, true
}

// expire removes an entry from the cache after a duration.
func (c *Cache) expire(key string, duration time.Duration) {
	time.Sleep(duration)

	c.mutex.Lock()
	delete(c.Data.Repositories, key)
	c.mutex.Unlock()

	c.SaveToFile()
}

func (c *Cache) PurgeExpired() {
	c.mutex.Lock()
	for key, entry := range c.Data.Repositories {
		if time.Now().After(entry.ExpiresAt) {
			delete(c.Data.Repositories, key)
		}
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
	file, err := os.ReadFile(c.Filename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if os.IsNotExist(err) {
		c.SaveToFile()
		file, err = os.ReadFile(c.Filename)
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return json.Unmarshal(file, &c.Data)
}
