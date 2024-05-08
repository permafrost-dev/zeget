package data

import (
	"time"

	"github.com/permafrost-dev/eget/lib/assets"
)

type RateLimit struct {
	Service   string     `json:"service"`
	Limit     int        `json:"limit"`
	Remaining int        `json:"remaining"`
	Reset     *time.Time `json:"reset"`
}

type RepositoryCacheEntry struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	LastCheckAt     time.Time      `json:"last_check_at"`
	LastDownloadAt  time.Time      `json:"last_download_at"`
	LastDownloadTag string         `json:"last_download_tag"`
	LastReleaseDate time.Time      `json:"last_release_date"`
	ExpiresAt       time.Time      `json:"expires_at"`
	Target          string         `json:"target"`
	Filters         []string       `json:"filters"`
	Assets          []assets.Asset `json:"assets"`
	FindError       error          `json:"find_error"`
	owner           *Cache
}

type ApplicationData struct {
	RateLimit    RateLimit                        `json:"rate_limit"`
	Repositories map[string]*RepositoryCacheEntry `json:"repositories"`
}

func (ad *ApplicationData) HasRepositoryEntryByKey(key string) bool {
	_, found := ad.Repositories[key]
	return found
}

func (ad *ApplicationData) GetRepositoryEntryByKey(key string, owner *Cache) *RepositoryCacheEntry {
	result, found := ad.Repositories[key]

	if found {
		result.owner = owner
	}

	return result
}

func (ad *ApplicationData) SetRepositoryEntryByKey(key string, entry *RepositoryCacheEntry, owner *Cache) {
	owner.mutex.Lock()
	defer owner.mutex.Unlock()

	entry.owner = owner
	ad.Repositories[key] = entry
}

func (rce *RepositoryCacheEntry) UpdateCheckedAt() {
	rce.LastCheckAt = time.Now()

	// if rce.owner != nil {
	// 	rce.owner.SaveToFile()
	// }
}

func (rce *RepositoryCacheEntry) UpdateDownloadedAt(tag string) {
	rce.LastDownloadAt = time.Now().Local()
	rce.LastDownloadTag = tag

	if rce.owner != nil {
		rce.owner.SaveToFile()
	}
}
