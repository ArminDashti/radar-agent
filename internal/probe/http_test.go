package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPCountsAnyResponseAsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	latency, ok := HTTP(context.Background(), server.URL, time.Second)
	if !ok || latency == nil || *latency < 0 {
		t.Fatalf("latency = %v, ok = %v", latency, ok)
	}
}

func TestHTTPPrependsHTTPSWhenSchemeMissing(t *testing.T) {
	_, ok := HTTP(context.Background(), "127.0.0.1:1", 20*time.Millisecond)
	if ok {
		t.Fatal("expected connection failure")
	}
	if normalizeURL("example.com") != "https://example.com" || !strings.HasPrefix(normalizeURL("http://example.com"), "http://") {
		t.Fatal("unexpected URL normalization")
	}
}
