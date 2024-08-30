package github

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/permafrost-dev/zeget/lib/download"
)

type RateLimitJSON struct {
	Resources map[string]RateLimit
}

type RateLimit struct {
	Limit     int
	Remaining int
	Reset     int64
	ResetsAt  time.Time
}

func (r RateLimit) ResetTime() time.Time {
	return time.Unix(r.Reset, 0)
}

func (r RateLimit) String() string {
	now := time.Now()
	rTime := r.ResetTime()

	if rTime.Before(now) {
		return fmt.Sprintf("Limit: %d, Remaining: %d, Reset: %v", r.Limit, r.Remaining, rTime)
	}

	return fmt.Sprintf(
		"Limit: %d, Remaining: %d, Reset: %v (%v)",
		r.Limit, r.Remaining, rTime, rTime.Sub(now).Round(time.Second),
	)
}

func FetchRateLimit(client download.ClientContract) (*RateLimit, error) {
	resp, err := client.GetJSON("https://api.github.com/rate_limit")

	if err != nil {
		return &RateLimit{}, err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return &RateLimit{}, err
	}

	var parsed RateLimitJSON
	err = json.Unmarshal(b, &parsed)

	result := parsed.Resources["core"]
	result.ResetsAt = result.ResetTime()

	return &result, err
}
