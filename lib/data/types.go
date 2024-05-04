package data

import (
	"time"
)

type RateLimit struct {
	Service   string     `json:"service"`
	Limit     int        `json:"limit"`
	Remaining int        `json:"remaining"`
	Reset     *time.Time `json:"reset"`
}

type RepositoryCacheEntry struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	LastCheckAt     time.Time `json:"last_check_at"`
	LastDownloadAt  time.Time `json:"last_download_at"`
	LastDownloadTag string    `json:"last_download_tag"`
	ExpiresAt       time.Time `json:"expires_at"`
	Target          string    `json:"target"`
	Filters         []string  `json:"filters"`
	owner           *Cache
}

type ApplicationData struct {
	RateLimit    RateLimit                       `json:"rate_limit"`
	Repositories map[string]RepositoryCacheEntry `json:"repositories"`
}

func (rce *RepositoryCacheEntry) UpdateCheckedAt() {
	rce.LastCheckAt = time.Now()

	// if rce.owner != nil {
	// 	rce.owner.SaveToFile()
	// }
}

func (rce *RepositoryCacheEntry) UpdateDownloadedAt(tag string) {
	rce.LastDownloadAt = time.Now()
	rce.LastDownloadTag = tag

	// if rce.owner != nil {
	// 	rce.owner.SaveToFile()
	// }
}
