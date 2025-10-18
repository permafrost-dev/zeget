package download

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/schollz/progressbar/v3"
)

const (
	minParallelSize = int64(1 << 20) // 1 MiB threshold to enable parallel
	defaultPartSize = int64(4 << 20) // 4 MiB chunks
	maxRetries      = 2
)

// DownloadSmart chooses a parallel ranged download when beneficial and supported.
// maxConcurrency caps parallel range workers (e.g., 4–16). If <=0, it defaults to 8.
func (dc *Client) DownloadSmart(
	url string,
	out io.Writer,
	progressBarCallback func(size int64) *progressbar.ProgressBar,
	maxConcurrency int,
) error {
	if isLocalFile(url) {
		// Reuse your simple downloader for local files.
		return dc.Download(url, out, func(total int64) *progressbar.ProgressBar {
			return defaultBar(total)
		})
	}

	dc.SetTokenFromEnv()

	// Probe server for size & range support.
	size, acceptsRanges, err := dc.headProbe(url)
	if err != nil || size <= 0 || size < minParallelSize || !acceptsRanges {
		// Fallback to single stream.
		return dc.Download(url, out, func(total int64) *progressbar.ProgressBar {
			if progressBarCallback != nil {
				return progressBarCallback(total)
			}
			return defaultBar(total)
		})
	}

	// Ensure we have a WriterAt; otherwise stream to temp file then copy.
	var wAt io.WriterAt
	var tmp *os.File
	switch t := out.(type) {
	case io.WriterAt:
		wAt = t
	default:
		tf, err := os.CreateTemp("", "rangeddl-*")
		if err != nil {
			return err
		}
		defer func() {
			tf.Close()
			os.Remove(tf.Name())
		}()
		wAt = tf
		tmp = tf
	}

	if maxConcurrency <= 0 {
		maxConcurrency = 8
	}

	partSize := defaultPartSize
	parts := int64(math.Ceil(float64(size) / float64(partSize)))
	if parts < 1 {
		parts = 1
	}
	if parts > int64(maxConcurrency) {
		parts = int64(maxConcurrency)
		partSize = int64(math.Ceil(float64(size) / float64(parts)))
	}

	// Progress bar
	var bar *progressbar.ProgressBar
	if progressBarCallback != nil {
		bar = progressBarCallback(size)
	}
	if bar == nil {
		bar = defaultBar(size)
	}

	type job struct {
		start int64
		end   int64 // inclusive
	}
	jobs := make(chan job, parts)
	for i := int64(0); i < parts; i++ {
		start := i * partSize
		end := start + partSize - 1
		if end >= size {
			end = size - 1
		}
		jobs <- job{start: start, end: end}
	}
	close(jobs)

	client := dc.GetClient()

	// Shared written count for any external use; bar.Add64 is already synchronized.
	var written int64
	errCh := make(chan error, maxConcurrency)

	var wg sync.WaitGroup
	worker := func() {
		fmt.Printf("[debug] worker started\n")
		defer wg.Done()
		for j := range jobs {
			var lastErr error
			for attempt := 0; attempt < maxRetries; attempt++ {
				req, rerr := dc.createRequest("GET", url)
				if rerr != nil {
					lastErr = rerr
					continue
				}
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", j.start, j.end))

				resp, derr := client.Do(req)
				if derr != nil {
					lastErr = derr
					backoff(attempt)
					continue
				}
				func() {
					defer resp.Body.Close()
					if resp.StatusCode != http.StatusPartialContent {
						lastErr = fmt.Errorf("expected 206 Partial Content, got %d", resp.StatusCode)
						return
					}
					// Writer that targets the correct offset
					section := &writerAtSection{W: wAt, Off: j.start}
					// Counting writer increments bar + total
					cw := countingWriter(func(n int) {
						atomic.AddInt64(&written, int64(n))
						_ = bar.Add64(int64(n)) // ignore minor draw errors
					})

					cw(0)

					if _, err := io.Copy(io.MultiWriter(section, cw), resp.Body); err != nil {
						lastErr = err
						return
					}
					lastErr = nil
				}()
				if lastErr == nil {
					break
				}
				backoff(attempt)
			}
			if lastErr != nil {
				errCh <- lastErr
				return
			}
		}
	}

	workerCount := int(parts)
	if workerCount > maxConcurrency {
		workerCount = maxConcurrency
	}
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go worker()
	}
	wg.Wait()

	select {
	case e := <-errCh:
		// drain channel
		for range errCh {
		}
		_ = bar.Finish()
		return e
	default:
	}

	// Ensure bar completes exactly at size.
	cur := bar.State().CurrentBytes
	if int64(cur) != int64(size) {
		bar.Set64(int64(size))
		// _ = bar.Add64(int64(size) - int64(cur))
	}
	_ = bar.Finish()

	// If we used a temp file, copy to final writer.
	if tmp != nil {
		if _, err := tmp.Seek(0, io.SeekStart); err != nil {
			return err
		}
		_, err := io.Copy(out, tmp)
		return err
	}
	return nil
}

// --- helpers ---

// headProbe does HEAD; falls back to byteProbe on error/unhelpful status.
func (dc *Client) headProbe(url string) (size int64, acceptsRanges bool, err error) {
	req, err := dc.createRequest("HEAD", url)
	if err != nil {
		return 0, false, err
	}
	resp, err := dc.GetClient().Do(req)
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()

	fmt.Printf("[debug] headprobe url: %s\n", req.URL.String())
	fmt.Printf("[debug] headprobe request status: %v\n", resp.StatusCode)

	if resp.StatusCode >= 400 &&
		resp.StatusCode != http.StatusMethodNotAllowed &&
		resp.StatusCode != http.StatusNotImplemented {
		return dc.byteProbe(url)
	}

	size = resp.ContentLength
	acceptsRanges = strings.EqualFold(resp.Header.Get("Accept-Ranges"), "bytes")
	if size <= 0 || !acceptsRanges {
		// Try a tiny ranged GET to confirm.
		if s, ok, e := dc.byteProbe(url); e == nil {
			if size <= 0 {
				size = s
			}
			if !acceptsRanges {
				acceptsRanges = ok
			}
		}
	}
	return size, acceptsRanges, nil
}

// byteProbe: GET Range: bytes=0-0 to test ranges & parse total from Content-Range.
func (dc *Client) byteProbe(url string) (size int64, acceptsRanges bool, err error) {
	req, err := dc.createRequest("GET", url)
	if err != nil {
		return 0, false, err
	}
	req.Header.Set("Range", "bytes=0-0")

	resp, err := dc.GetClient().Do(req)
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()

	fmt.Printf("[debug] byteprobe url: %s\n", req.URL.String())
	fmt.Printf("[debug] byteprobe request status: %v\n", resp.StatusCode)

	if resp.StatusCode != http.StatusPartialContent {
		return resp.ContentLength, false, nil
	}
	var total int64
	if _, scanErr := fmt.Sscanf(resp.Header.Get("Content-Range"), "bytes %*d-%*d/%d", &total); scanErr == nil {
		return total, true, nil
	}
	return -1, true, nil
}

// writerAtSection adapts an io.WriterAt into an io.Writer that writes sequentially from Off.
type writerAtSection struct {
	W   io.WriterAt
	Off int64
}

func (w *writerAtSection) Write(p []byte) (int, error) {
	n, err := w.W.WriteAt(p, w.Off)
	w.Off += int64(n)
	return n, err
}

// countingWriter calls f with bytes written.
type countingWriter func(int)

func (cw countingWriter) Write(p []byte) (int, error) {
	n := len(p)
	cw(n)
	return n, nil
}

func backoff(attempt int) {
	if attempt <= 0 {
		return
	}
	time.Sleep(time.Duration(1200*min(attempt, 15)) * time.Millisecond)
}

// Optional: sensible default bar if caller doesn't provide one.
func defaultBar(total int64) *progressbar.ProgressBar {
	return progressbar.NewOptions64(
		total,
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
	)
}
