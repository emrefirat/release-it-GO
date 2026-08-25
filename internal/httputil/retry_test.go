package httputil

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func noSleep(delays *[]time.Duration) func(time.Duration) {
	return func(d time.Duration) { *delays = append(*delays, d) }
}

func TestDo_RetriesOn503_ThenSucceeds(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"x":1}` {
			t.Errorf("attempt %d lost the request body: %q", atomic.LoadInt32(&attempts), body)
		}
		if atomic.AddInt32(&attempts, 1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, _ := http.NewRequest("POST", server.URL, bytes.NewReader([]byte(`{"x":1}`)))
	var delays []time.Duration
	resp, err := Do(server.Client(), req, Options{Sleep: noSleep(&delays)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
	if len(delays) != 2 || delays[0] >= delays[1] {
		t.Errorf("expected 2 growing backoff delays, got %v", delays)
	}
}

func TestDo_HonorsRetryAfterHeader(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&attempts, 1) == 1 {
			w.Header().Set("Retry-After", "7")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	var delays []time.Duration
	resp, err := Do(server.Client(), req, Options{Sleep: noSleep(&delays)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if len(delays) != 1 || delays[0] != 7*time.Second {
		t.Errorf("expected the server's Retry-After (7s) to be honored, got %v", delays)
	}
}

func TestDo_NonRetryableStatus_ReturnsResponse(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := Do(server.Client(), req, Options{Sleep: func(time.Duration) {}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("4xx (non-429) must not be retried; attempts = %d", attempts)
	}
}

func TestDo_ConnError_POST_NotRetried(t *testing.T) {
	// Port 1: connection refused. A POST that never got a response must not
	// be replayed — the server may have received the first one.
	req, _ := http.NewRequest("POST", "http://127.0.0.1:1/", strings.NewReader("x"))
	calls := 0
	_, err := Do(&http.Client{}, req, Options{Sleep: func(time.Duration) { calls++ }})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 0 {
		t.Errorf("POST conn error must fail fast, slept %d times", calls)
	}
}

func TestDo_ConnError_GET_Retried(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://127.0.0.1:1/", nil)
	var delays []time.Duration
	_, err := Do(&http.Client{}, req, Options{Sleep: noSleep(&delays)})
	if err == nil {
		t.Fatal("expected error after retries")
	}
	if len(delays) != 2 {
		t.Errorf("GET conn error should be retried (2 sleeps for 3 attempts), got %d", len(delays))
	}
}
