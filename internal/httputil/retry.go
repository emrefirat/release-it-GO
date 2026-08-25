// Package httputil provides a small retrying HTTP transport helper shared by
// the GitHub/GitLab clients and webhook notifications. A transient 429/5xx or
// a dropped connection no longer aborts a release mid-pipeline (after the tag
// was already pushed) when a second attempt would have succeeded.
package httputil

import (
	"io"
	"net/http"
	"strconv"
	"time"
)

// defaultMaxAttempts bounds the total number of tries (first + retries).
const defaultMaxAttempts = 3

// defaultBaseDelay is the first backoff delay; it doubles per attempt.
const defaultBaseDelay = time.Second

// retryableStatuses are transient server-side conditions worth retrying.
var retryableStatuses = map[int]bool{
	http.StatusTooManyRequests:    true, // 429
	http.StatusBadGateway:         true, // 502
	http.StatusServiceUnavailable: true, // 503
	http.StatusGatewayTimeout:     true, // 504
}

// Options tunes Do. The zero value is a sensible default.
type Options struct {
	// MaxAttempts is the total number of tries (default 3).
	MaxAttempts int
	// BaseDelay is the first backoff delay, doubled per attempt (default 1s).
	// A server-provided Retry-After header takes precedence.
	BaseDelay time.Duration
	// Sleep overrides time.Sleep (tests). Nil uses time.Sleep.
	Sleep func(time.Duration)
}

// Do executes req, retrying on 429/502/503/504 responses (honoring
// Retry-After) with exponential backoff. Transport-level errors (no response
// received) are retried only for idempotent methods — replaying a POST whose
// fate is unknown could create a duplicate release. Request bodies are
// replayed via req.GetBody, which http.NewRequest sets automatically for
// bytes/strings readers.
func Do(client *http.Client, req *http.Request, opts Options) (*http.Response, error) {
	maxAttempts := opts.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = defaultMaxAttempts
	}
	baseDelay := opts.BaseDelay
	if baseDelay <= 0 {
		baseDelay = defaultBaseDelay
	}
	sleep := opts.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}

	var resp *http.Response
	var err error
	delay := baseDelay

	for attempt := 1; ; attempt++ {
		if attempt > 1 && req.GetBody != nil {
			body, bodyErr := req.GetBody()
			if bodyErr != nil {
				return resp, err // cannot replay the body — return the last outcome
			}
			req.Body = body
		}

		resp, err = client.Do(req)

		if err != nil {
			// No response received. Only idempotent methods are safe to replay.
			if !isIdempotent(req.Method) || attempt >= maxAttempts {
				return nil, err
			}
			sleep(delay)
			delay *= 2
			continue
		}

		if !retryableStatuses[resp.StatusCode] || attempt >= maxAttempts {
			return resp, nil
		}

		// Transient server condition — drain, back off, retry.
		wait := delay
		if ra := retryAfter(resp); ra > 0 {
			wait = ra
		}
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		sleep(wait)
		delay *= 2
	}
}

// isIdempotent reports whether the method is safe to replay when no response
// was received.
func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}

// retryAfter parses a delay-seconds Retry-After header (0 when absent or not
// a plain number — HTTP-date values fall back to the local backoff).
func retryAfter(resp *http.Response) time.Duration {
	value := resp.Header.Get("Retry-After")
	if value == "" {
		return 0
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds < 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}
